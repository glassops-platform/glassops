package ci

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// fixtureServer serves files from testdata/github by request path and records the
// Authorization header of every request.
func fixtureServer(t *testing.T, routes map[string]struct {
	status int
	file   string
}) (*httptest.Server, *[]string) {
	t.Helper()
	var auth []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = append(auth, r.Header.Get("Authorization"))
		route, ok := routes[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"Not Found"}`))
			return
		}
		data, err := os.ReadFile(filepath.Join("testdata", "github", route.file))
		if err != nil {
			t.Errorf("fixture %s: %v", route.file, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(route.status)
		_, _ = w.Write(data)
	}))
	t.Cleanup(srv.Close)
	return srv, &auth
}

type route = struct {
	status int
	file   string
}

const (
	testRepo   = "glassops-platform/glassops"
	mergeSHA   = "20459158b4f16da4cddc816ac7c1d695177af196"
	labelSHA   = "24cb560cd6f9d354f4d2a592e36e4a66db8fe55b"
	unknownSHA = "0000000000000000000000000000000000000000"
)

func TestPullForCommit(t *testing.T) {
	srv, auth := fixtureServer(t, map[string]route{
		"/repos/" + testRepo + "/commits/" + mergeSHA + "/pulls":   {200, "commit-pulls.json"},
		"/repos/" + testRepo + "/commits/" + labelSHA + "/pulls":   {200, "commit-pulls-labelled.json"},
		"/repos/" + testRepo + "/commits/" + unknownSHA + "/pulls": {422, "commit-pulls-unknown.json"},
	})
	api := NewGitHubAPI(srv.URL, "synthetic-token", srv.Client())

	cases := []struct {
		name       string
		sha        string
		wantNumber int
		wantAuthor string
		wantLabels []string
	}{
		{"merged pull request", mergeSHA, 46, "rdbumstead", []string{}},
		{"pull request with labels", labelSHA, 45, "dependabot[bot]", []string{"dependencies", "github_actions"}},
		{"unknown commit is HTTP 422, no pull request", unknownSHA, 0, "", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			n, author, labels, err := api.PullForCommit(context.Background(), testRepo, tc.sha)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if n != tc.wantNumber || author != tc.wantAuthor || !reflect.DeepEqual(labels, tc.wantLabels) {
				t.Fatalf("got %d %q %#v, want %d %q %#v", n, author, labels, tc.wantNumber, tc.wantAuthor, tc.wantLabels)
			}
		})
	}
	for _, a := range *auth {
		if a != "Bearer synthetic-token" {
			t.Fatalf("Authorization header %q", a)
		}
	}
}

func TestPullForCommitPicksTheMergedPull(t *testing.T) {
	// Synthetic body: two pull requests list the commit, and only the second was merged as it.
	body := `[{"number":7,"merge_commit_sha":"aaaaaaa","merged_at":null,"user":{"login":"first"},"labels":[]},
	{"number":8,"merge_commit_sha":"` + mergeSHA + `","merged_at":"2026-10-29T02:41:37Z","user":{"login":"second"},"labels":[{"name":"approved"}]}]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()
	n, author, labels, err := NewGitHubAPI(srv.URL, "t", srv.Client()).PullForCommit(context.Background(), testRepo, mergeSHA)
	if err != nil || n != 8 || author != "second" || !reflect.DeepEqual(labels, []string{"approved"}) {
		t.Fatalf("got %d %q %v %v", n, author, labels, err)
	}
}

func TestPullForCommitErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Bad credentials"}`))
	}))
	defer srv.Close()
	api := NewGitHubAPI(srv.URL, "secret-token-value", srv.Client())

	cases := []struct {
		name, repo, sha, want string
	}{
		{"401 is an error", testRepo, mergeSHA, "HTTP 401: Bad credentials"},
		{"bad repo", "not a repo", mergeSHA, "not owner/name"},
		{"path in repo", "a/b/../c", mergeSHA, "not owner/name"},
		{"dot segments", "../..", mergeSHA, "not owner/name"},
		{"bad sha", testRepo, "main", "not a SHA"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, err := api.PullForCommit(context.Background(), tc.repo, tc.sha)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
			if strings.Contains(err.Error(), "secret-token-value") {
				t.Fatal("the token leaked into an error")
			}
		})
	}
}

func TestRunApprovals(t *testing.T) {
	srv, _ := fixtureServer(t, map[string]route{
		"/repos/" + testRepo + "/actions/runs/41165934934/approvals": {200, "run-approvals.json"},
	})
	api := NewGitHubAPI(srv.URL, "t", srv.Client())
	got, err := api.RunApprovals(context.Background(), testRepo, "41165934934")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("approvals: %+v", got)
	}
	a := got[0]
	if a.Mechanism != "environment" || a.Approver != "rdbumstead" || a.State != "approved" || a.Environment != "glassops-deploy" || a.Comment == "" {
		t.Fatalf("approval: %+v", a)
	}

	if _, err := api.RunApprovals(context.Background(), testRepo, "999"); err == nil || !strings.Contains(err.Error(), "HTTP 404") {
		t.Fatalf("missing run: err = %v", err)
	}
	if _, err := api.RunApprovals(context.Background(), testRepo, "1/../2"); err == nil {
		t.Fatal("a run Id with a path was accepted")
	}
}

func TestRunApprovalsOnePerEnvironment(t *testing.T) {
	// Synthetic body: one review covering two environments, and one covering none.
	body := `[{"state":"approved","comment":"","user":{"login":"a"},"environments":[{"name":"one"},{"name":"two"}]},
	{"state":"rejected","comment":"no","user":{"login":"b"},"environments":[]}]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()
	got, err := NewGitHubAPI(srv.URL, "t", srv.Client()).RunApprovals(context.Background(), testRepo, "1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0].Environment != "one" || got[1].Environment != "two" || got[2].State != "rejected" || got[2].Environment != "" {
		t.Fatalf("approvals: %+v", got)
	}
}

func TestRunApprovalsBadBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"not":"a list"}`))
	}))
	defer srv.Close()
	if _, err := NewGitHubAPI(srv.URL, "t", srv.Client()).RunApprovals(context.Background(), testRepo, "1"); err == nil || !strings.Contains(err.Error(), "parse body") {
		t.Fatalf("err = %v", err)
	}
}

func TestPull(t *testing.T) {
	srv, _ := fixtureServer(t, map[string]route{
		"/repos/" + testRepo + "/pulls/46": {200, "pull.json"},
	})
	api := NewGitHubAPI(srv.URL, "t", srv.Client())
	author, labels, err := api.Pull(context.Background(), testRepo, 46)
	if err != nil || author != "rdbumstead" || !reflect.DeepEqual(labels, []string{}) {
		t.Fatalf("got %q %#v %v", author, labels, err)
	}
	if _, _, err := api.Pull(context.Background(), testRepo, 47); err == nil || !strings.Contains(err.Error(), "HTTP 404") {
		t.Fatalf("missing pull: err = %v", err)
	}
	if _, _, err := api.Pull(context.Background(), testRepo, 0); err == nil {
		t.Fatal("pull request 0 was accepted")
	}
	if _, _, err := api.Pull(context.Background(), "../..", 46); err == nil {
		t.Fatal("a bad repo was accepted")
	}
}

func TestLabelAdderFixture(t *testing.T) {
	srv, _ := fixtureServer(t, map[string]route{
		"/repos/" + testRepo + "/issues/45/events": {200, "issue-events-labelled.json"},
	})
	api := NewGitHubAPI(srv.URL, "t", srv.Client())
	cases := []struct{ label, want string }{
		{"dependencies", "dependabot[bot]"},
		{"github_actions", "dependabot[bot]"},
		{"approved", ""},
	}
	for _, tc := range cases {
		got, err := api.LabelAdder(context.Background(), testRepo, 45, tc.label)
		if err != nil || got != tc.want {
			t.Fatalf("label %s: got %q %v, want %q", tc.label, got, err, tc.want)
		}
	}
}

func TestLabelAdderOrderAndRemoval(t *testing.T) {
	// Synthetic events, served out of order across two pages.
	page1 := `[{"id":3,"event":"labeled","created_at":"2026-10-03T00:00:00Z","actor":{"login":"carol"},"label":{"name":"ok"}},
	{"id":1,"event":"labeled","created_at":"2026-10-01T00:00:00Z","actor":{"login":"alice"},"label":{"name":"ok"}}]`
	page2 := `[{"id":2,"event":"unlabeled","created_at":"2026-10-02T00:00:00Z","actor":{"login":"bob"},"label":{"name":"ok"}},
	{"id":4,"event":"labeled","created_at":"2026-10-04T00:00:00Z","actor":{"login":"dave"},"label":{"name":"gone"}},
	{"id":5,"event":"unlabeled","created_at":"2026-10-05T00:00:00Z","actor":{"login":"erin"},"label":{"name":"gone"}}]`
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "2" {
			_, _ = w.Write([]byte(page2))
			return
		}
		w.Header().Set("Link", `<`+srv.URL+r.URL.Path+`?page=2>; rel="next", <`+srv.URL+r.URL.Path+`?page=2>; rel="last"`)
		_, _ = w.Write([]byte(page1))
	}))
	defer srv.Close()
	api := NewGitHubAPI(srv.URL, "t", srv.Client())
	for label, want := range map[string]string{"ok": "carol", "gone": ""} {
		got, err := api.LabelAdder(context.Background(), testRepo, 9, label)
		if err != nil || got != want {
			t.Fatalf("label %s: got %q %v, want %q", label, got, err, want)
		}
	}
}

func TestLabelAdderRefusesForeignNextPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Link", `<https://elsewhere.example/events?page=2>; rel="next"`)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()
	_, err := NewGitHubAPI(srv.URL, "t", srv.Client()).LabelAdder(context.Background(), testRepo, 9, "ok")
	if err == nil || !strings.Contains(err.Error(), "outside the API root") {
		t.Fatalf("err = %v", err)
	}
}

func TestNextLink(t *testing.T) {
	cases := map[string]string{
		"": "",
		`<https://api.github.com/x?page=2>; rel="next", <https://api.github.com/x?page=5>; rel="last"`: "https://api.github.com/x?page=2",
		`<https://api.github.com/x?page=1>; rel="prev"`:                                                "",
	}
	for h, want := range cases {
		if got := nextLink(h); got != want {
			t.Fatalf("nextLink(%q) = %q, want %q", h, got, want)
		}
	}
}
