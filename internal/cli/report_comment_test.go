package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/report"
	"github.com/glassops-platform/glassops/internal/core/run"
)

// fakeIssueAPI is a synthetic issue comments API written from GitHub's REST documentation, not
// captured from a repository [unverified].
type fakeIssueAPI struct {
	mu       sync.Mutex
	comments []map[string]any
	status   int // when non zero, writes answer this status
	requests []string
	srv      *httptest.Server
}

func newFakeIssueAPI(t *testing.T) *fakeIssueAPI {
	t.Helper()
	f := &fakeIssueAPI{}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.requests = append(f.requests, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet && f.status != 0 {
			w.WriteHeader(f.status)
			_, _ = w.Write([]byte(`{"message":"Resource not accessible by integration"}`))
			return
		}
		var in struct{ Body string }
		if r.Body != nil {
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &in)
		}
		switch {
		case r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(f.comments)
		case r.Method == http.MethodPost:
			c := map[string]any{"id": float64(500 + len(f.comments)), "body": in.Body}
			f.comments = append(f.comments, c)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(c)
		case r.Method == http.MethodPatch:
			f.comments[0]["body"] = in.Body
			_ = json.NewEncoder(w).Encode(f.comments[0])
		}
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeIssueAPI) writes() int {
	n := 0
	for _, r := range f.requests {
		if !strings.HasPrefix(r, "GET ") {
			n++
		}
	}
	return n
}

// githubPR sets a GitHub pull_request environment on the harness.
func githubPR(h *harness, api *fakeIssueAPI, event string) {
	h.env["GITHUB_ACTIONS"] = "true"
	h.env["GITHUB_EVENT_NAME"] = event
	h.env["GITHUB_REPOSITORY"] = "o/r"
	h.env["GITHUB_REF"] = "refs/pull/7/merge"
	h.env["GITHUB_TOKEN"] = "tok-secret"
	h.env["GITHUB_API_URL"] = api.srv.URL
}

func TestReportPRCommentCreatesThenUpdates(t *testing.T) {
	api := newFakeIssueAPI(t)
	h := newHarness(t)
	githubPR(h, api, "pull_request")
	seedResult(t, h, "doctor", run.StatusPass)

	for i := 0; i < 2; i++ {
		if code := Execute(h.app, []string{"report", "--pr-comment", "--dir", h.dir, "--ci", "github"}); code != exitcode.OK {
			t.Fatalf("run %d: exit %d\n%s", i, code, h.out)
		}
	}
	if len(api.comments) != 1 {
		t.Fatalf("want one comment after two runs, got %d", len(api.comments))
	}
	body, _ := api.comments[0]["body"].(string)
	if !strings.HasPrefix(body, report.CommentMarker) || !strings.Contains(body, "## glassops report: pass") {
		t.Fatalf("comment body wrong:\n%s", body)
	}
	if api.writes() != 2 || !strings.Contains(strings.Join(api.requests, ","), "PATCH /repos/o/r/issues/comments/500") {
		t.Fatalf("want a create then an update: %v", api.requests)
	}
	if strings.Contains(h.out.String(), "tok-secret") {
		t.Fatal("token reached the output")
	}
}

func TestReportPRCommentFromConfig(t *testing.T) {
	api := newFakeIssueAPI(t)
	h := newHarness(t)
	githubPR(h, api, "pull_request")
	h.write(t, "sfdx-project.json", "{}")
	h.write(t, "glassops.yaml", "version: 1\nreport:\n  pr_comment: true\n")
	seedResult(t, h, "doctor", run.StatusPass)
	if code := Execute(h.app, []string{"report", "--dir", h.dir, "--ci", "github"}); code != exitcode.OK {
		t.Fatalf("exit %d\n%s", code, h.out)
	}
	if len(api.comments) != 1 {
		t.Fatalf("config did not turn the comment on: %v", api.requests)
	}
}

func TestReportPRCommentSkippedOnPush(t *testing.T) {
	api := newFakeIssueAPI(t)
	h := newHarness(t)
	githubPR(h, api, "push")
	h.env["GITHUB_REF"] = "refs/heads/main"
	seedResult(t, h, "doctor", run.StatusPass)
	if code := Execute(h.app, []string{"report", "--pr-comment", "--dir", h.dir, "--ci", "github"}); code != exitcode.OK {
		t.Fatalf("exit %d\n%s", code, h.out)
	}
	if len(api.requests) != 0 {
		t.Fatalf("a push run called the API: %v", api.requests)
	}
}

func TestReportPRCommentOffByDefault(t *testing.T) {
	api := newFakeIssueAPI(t)
	h := newHarness(t)
	githubPR(h, api, "pull_request")
	seedResult(t, h, "doctor", run.StatusPass)
	if code := Execute(h.app, []string{"report", "--dir", h.dir, "--ci", "github"}); code != exitcode.OK {
		t.Fatalf("exit %d\n%s", code, h.out)
	}
	if len(api.requests) != 0 {
		t.Fatalf("no flag and no config, but the API was called: %v", api.requests)
	}
}

func TestReportPRComment403WarnsAndKeepsExit(t *testing.T) {
	api := newFakeIssueAPI(t)
	api.status = http.StatusForbidden
	h := newHarness(t)
	githubPR(h, api, "pull_request")
	seedResult(t, h, "doctor", run.StatusPass)
	if code := Execute(h.app, []string{"report", "--pr-comment", "--dir", h.dir, "--ci", "github"}); code != exitcode.OK {
		t.Fatalf("a 403 changed the exit to %d\n%s", code, h.out)
	}
	out := h.out.String()
	if !strings.Contains(out, "::warning::") || !strings.Contains(out, "pull-requests: write") {
		t.Fatalf("no warning about the permission:\n%s", out)
	}
	if strings.Contains(out, "tok-secret") {
		t.Fatal("token reached the output")
	}
}

func TestReportPRCommentFailureKeepsWorstExit(t *testing.T) {
	api := newFakeIssueAPI(t)
	api.status = http.StatusForbidden
	h := newHarness(t)
	githubPR(h, api, "pull_request")
	seedResult(t, h, "validate", run.StatusFail, run.Check{Name: "p", Status: run.StatusFail, Detail: "bad"})
	if code := Execute(h.app, []string{"report", "--pr-comment", "--dir", h.dir, "--ci", "github"}); code != exitcode.Fail {
		t.Fatalf("exit %d, want the step failure %d\n%s", code, exitcode.Fail, h.out)
	}
}
