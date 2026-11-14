package ci

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The bodies below are synthetic, written from GitHub's REST documentation for issue comments
// and not captured from a real repository [unverified]. Capture command for the orchestrator:
// gh api repos/glassops-platform/glassops/issues/47/comments

const testMarker = "<!-- glassops-report -->"

// commentServer is a fake issue comments API. It keeps comments in memory.
type commentServer struct {
	comments []apiComment
	nextID   int64
	status   int // when non zero, every write answers this status
	requests []string
	auth     []string
	srv      *httptest.Server
}

func newCommentServer(t *testing.T, pageSize int) *commentServer {
	t.Helper()
	cs := &commentServer{nextID: 100}
	cs.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cs.requests = append(cs.requests, r.Method+" "+r.URL.Path)
		cs.auth = append(cs.auth, r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/o/r/issues/7/comments":
			page := 1
			fmt.Sscanf(r.URL.Query().Get("page"), "%d", &page)
			lo := (page - 1) * pageSize
			hi := lo + pageSize
			if lo > len(cs.comments) {
				lo = len(cs.comments)
			}
			if hi > len(cs.comments) {
				hi = len(cs.comments)
			}
			if hi < len(cs.comments) {
				w.Header().Set("Link", fmt.Sprintf(`<%s%s?per_page=100&page=%d>; rel="next"`, cs.srv.URL, r.URL.Path, page+1))
			}
			_ = json.NewEncoder(w).Encode(cs.comments[lo:hi])
		case r.Method == http.MethodPost && r.URL.Path == "/repos/o/r/issues/7/comments":
			if cs.status != 0 {
				w.WriteHeader(cs.status)
				_, _ = w.Write([]byte(`{"message":"Resource not accessible by integration"}`))
				return
			}
			var in struct{ Body string }
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &in)
			cs.nextID++
			c := apiComment{ID: cs.nextID, Body: in.Body}
			cs.comments = append(cs.comments, c)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(c)
		case r.Method == http.MethodPatch && strings.HasPrefix(r.URL.Path, "/repos/o/r/issues/comments/"):
			if cs.status != 0 {
				w.WriteHeader(cs.status)
				_, _ = w.Write([]byte(`{"message":"Resource not accessible by integration"}`))
				return
			}
			var in struct{ Body string }
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &in)
			for i := range cs.comments {
				if r.URL.Path == fmt.Sprintf("/repos/o/r/issues/comments/%d", cs.comments[i].ID) {
					cs.comments[i].Body = in.Body
					_ = json.NewEncoder(w).Encode(cs.comments[i])
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"Not Found"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"Not Found"}`))
		}
	}))
	t.Cleanup(cs.srv.Close)
	return cs
}

func TestUpsertCreatesThenUpdates(t *testing.T) {
	cs := newCommentServer(t, 100)
	cs.comments = []apiComment{{ID: 1, Body: "looks good"}}
	api := NewGitHubAPI(cs.srv.URL, "tok", nil)
	ctx := context.Background()

	created, err := api.UpsertComment(ctx, "o/r", 7, testMarker, testMarker+"\nfirst")
	if err != nil || !created {
		t.Fatalf("first run: created=%v err=%v", created, err)
	}
	created, err = api.UpsertComment(ctx, "o/r", 7, testMarker, testMarker+"\nsecond")
	if err != nil || created {
		t.Fatalf("second run: created=%v err=%v", created, err)
	}
	if len(cs.comments) != 2 {
		t.Fatalf("want 2 comments (one foreign, one ours), got %d", len(cs.comments))
	}
	if cs.comments[1].Body != testMarker+"\nsecond" || cs.comments[0].Body != "looks good" {
		t.Fatalf("bodies wrong: %+v", cs.comments)
	}
	for _, a := range cs.auth {
		if a != "Bearer tok" {
			t.Fatalf("authorization header %q", a)
		}
	}
}

func TestFindCommentFollowsPages(t *testing.T) {
	cs := newCommentServer(t, 2)
	for i := 1; i <= 5; i++ {
		cs.comments = append(cs.comments, apiComment{ID: int64(i), Body: "c"})
	}
	cs.comments[4].Body = "x " + testMarker
	api := NewGitHubAPI(cs.srv.URL, "tok", nil)
	id, err := api.FindComment(context.Background(), "o/r", 7, testMarker)
	if err != nil || id != 5 {
		t.Fatalf("id=%d err=%v", id, err)
	}
}

func TestUpsertRejectsBodyWithoutMarker(t *testing.T) {
	cs := newCommentServer(t, 100)
	api := NewGitHubAPI(cs.srv.URL, "tok", nil)
	if _, err := api.UpsertComment(context.Background(), "o/r", 7, testMarker, "no marker"); err == nil {
		t.Fatal("want an error")
	}
	if len(cs.requests) != 0 {
		t.Fatalf("no request should be sent: %v", cs.requests)
	}
}

func TestCommentWriteForbiddenIsTypedAndHidesToken(t *testing.T) {
	cs := newCommentServer(t, 100)
	cs.status = http.StatusForbidden
	api := NewGitHubAPI(cs.srv.URL, "secret-token", nil)
	_, err := api.UpsertComment(context.Background(), "o/r", 7, testMarker, testMarker)
	he, ok := err.(*HTTPError)
	if !ok || he.Status != 403 {
		t.Fatalf("want a 403 HTTPError, got %v", err)
	}
	if strings.Contains(err.Error(), "secret-token") {
		t.Fatalf("token in error: %v", err)
	}
}

func TestCommentInputValidation(t *testing.T) {
	api := NewGitHubAPI("http://127.0.0.1:1", "tok", nil)
	ctx := context.Background()
	if _, err := api.FindComment(ctx, "../x", 7, testMarker); err == nil {
		t.Error("bad repo accepted")
	}
	if _, err := api.FindComment(ctx, "o/r", 0, testMarker); err == nil {
		t.Error("bad number accepted")
	}
	if _, err := api.FindComment(ctx, "o/r", 7, ""); err == nil {
		t.Error("empty marker accepted")
	}
	if _, err := api.CreateComment(ctx, "o/r", -1, "x"); err == nil {
		t.Error("bad number accepted")
	}
	if err := api.UpdateComment(ctx, "o/r", 0, "x"); err == nil {
		t.Error("bad id accepted")
	}
}
