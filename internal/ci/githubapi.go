package ci

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/glassops-platform/glassops/internal/core/facts"
)

// DefaultGitHubAPIURL is the REST API root used when GITHUB_API_URL is not set.
const DefaultGitHubAPIURL = "https://api.github.com"

// maxAPIBody caps how much of a response is read, so a broken proxy cannot exhaust memory.
const maxAPIBody = 10 << 20

var (
	repoPattern  = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
	shaPattern   = regexp.MustCompile(`^[0-9a-fA-F]{7,64}$`)
	runIDPattern = regexp.MustCompile(`^[0-9]+$`)
)

// GitHubAPI reads approval evidence from the GitHub REST API. It sends the token only in the
// Authorization header, and no error it returns contains the token.
type GitHubAPI struct {
	baseURL string
	http    *http.Client
	token   string
}

// NewGitHubAPI returns a client for the REST API at baseURL, authenticated by token. An empty
// baseURL means DefaultGitHubAPIURL. A nil client means one with a 30 second timeout.
func NewGitHubAPI(baseURL, token string, hc *http.Client) *GitHubAPI {
	if baseURL == "" {
		baseURL = DefaultGitHubAPIURL
	}
	if hc == nil {
		hc = &http.Client{Timeout: 30 * time.Second}
	}
	return &GitHubAPI{baseURL: strings.TrimRight(baseURL, "/"), http: hc, token: token}
}

// apiPull is the part of a pull request the client reads.
type apiPull struct {
	Number         int     `json:"number"`
	MergeCommitSHA string  `json:"merge_commit_sha"`
	MergedAt       *string `json:"merged_at"`
	User           struct {
		Login string `json:"login"`
	} `json:"user"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

// PullForCommit returns the pull request a commit came from, from
// GET /repos/{repo}/commits/{sha}/pulls. It picks the merged pull request whose merge commit is
// sha, else the first one listed. A commit with no pull request, or one GitHub does not know
// (HTTP 422), gives number 0 and no error.
func (a *GitHubAPI) PullForCommit(ctx context.Context, repo, sha string) (number int, author string, labels []string, err error) {
	if !validRepo(repo) {
		return 0, "", nil, fmt.Errorf("github api: repository %q is not owner/name", repo)
	}
	if !shaPattern.MatchString(sha) {
		return 0, "", nil, fmt.Errorf("github api: commit %q is not a SHA", sha)
	}
	var pulls []apiPull
	status, err := a.get(ctx, "/repos/"+repo+"/commits/"+sha+"/pulls", &pulls)
	if status == http.StatusUnprocessableEntity {
		// GitHub answers 422 "No commit found for SHA" for a commit it does not have.
		return 0, "", nil, nil
	}
	if err != nil {
		return 0, "", nil, err
	}
	if len(pulls) == 0 {
		return 0, "", nil, nil
	}
	pick := pulls[0]
	for _, p := range pulls {
		if p.MergedAt != nil && p.MergeCommitSHA != "" && strings.EqualFold(p.MergeCommitSHA, sha) {
			pick = p
			break
		}
	}
	labels = make([]string, 0, len(pick.Labels))
	for _, l := range pick.Labels {
		labels = append(labels, l.Name)
	}
	return pick.Number, pick.User.Login, labels, nil
}

// apiApproval is one entry from the run approvals endpoint.
type apiApproval struct {
	State   string `json:"state"`
	Comment string `json:"comment"`
	User    struct {
		Login string `json:"login"`
	} `json:"user"`
	Environments []struct {
		Name string `json:"name"`
	} `json:"environments"`
}

// RunApprovals returns the environment reviews of a workflow run, from
// GET /repos/{repo}/actions/runs/{run_id}/approvals. A review that covered several
// environments gives one approval per environment.
func (a *GitHubAPI) RunApprovals(ctx context.Context, repo, runID string) ([]facts.Approval, error) {
	if !validRepo(repo) {
		return nil, fmt.Errorf("github api: repository %q is not owner/name", repo)
	}
	if !runIDPattern.MatchString(runID) {
		return nil, fmt.Errorf("github api: run Id %q is not a number", runID)
	}
	var raw []apiApproval
	if _, err := a.get(ctx, "/repos/"+repo+"/actions/runs/"+runID+"/approvals", &raw); err != nil {
		return nil, err
	}
	out := []facts.Approval{}
	for _, r := range raw {
		base := facts.Approval{
			Mechanism: "environment",
			Approver:  r.User.Login,
			State:     r.State,
			Comment:   r.Comment,
		}
		if len(r.Environments) == 0 {
			out = append(out, base)
			continue
		}
		for _, e := range r.Environments {
			ap := base
			ap.Environment = e.Name
			out = append(out, ap)
		}
	}
	return out, nil
}

// validRepo reports whether repo is owner/name with no path segments such as "..".
func validRepo(repo string) bool {
	if !repoPattern.MatchString(repo) {
		return false
	}
	for _, part := range strings.Split(repo, "/") {
		if strings.Trim(part, ".") == "" {
			return false
		}
	}
	return true
}

// apiError is the body GitHub sends with an error status.
type apiError struct {
	Message string `json:"message"`
}

// get fetches path and decodes a 200 body into v. It returns the HTTP status whenever a
// response arrived, so callers can tell a 422 apart from other failures.
func (a *GitHubAPI) get(ctx context.Context, path string, v any) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+path, nil)
	if err != nil {
		return 0, fmt.Errorf("github api: GET %s: %w", path, err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "glassops")
	if a.token != "" {
		req.Header.Set("Authorization", "Bearer "+a.token)
	}
	resp, err := a.http.Do(req)
	if err != nil {
		// url.Error carries the URL, which holds no token, and the transport error.
		return 0, fmt.Errorf("github api: GET %s: %w", path, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxAPIBody))
	if err != nil {
		return resp.StatusCode, fmt.Errorf("github api: GET %s: read body: %w", path, err)
	}
	if resp.StatusCode != http.StatusOK {
		var e apiError
		_ = json.Unmarshal(body, &e)
		msg := e.Message
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return resp.StatusCode, fmt.Errorf("github api: GET %s: HTTP %d: %s", path, resp.StatusCode, msg)
	}
	if err := json.Unmarshal(body, v); err != nil {
		return resp.StatusCode, fmt.Errorf("github api: GET %s: parse body: %w", path, err)
	}
	return resp.StatusCode, nil
}
