package ci

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// maxCommentPages caps how many pages of issue comments FindComment reads.
const maxCommentPages = 50

// apiComment is the part of an issue comment the client reads.
type apiComment struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
}

// FindComment returns the Id of the first comment on issue or pull request number whose body
// contains marker, from GET /repos/{repo}/issues/{number}/comments. It returns 0 when there is
// none. Reading comments on a public repository needs no extra permission.
func (a *GitHubAPI) FindComment(ctx context.Context, repo string, number int, marker string) (int64, error) {
	if !validRepo(repo) {
		return 0, fmt.Errorf("github api: repository %q is not owner/name", repo)
	}
	if number <= 0 {
		return 0, fmt.Errorf("github api: issue number %d is not valid", number)
	}
	if marker == "" {
		return 0, fmt.Errorf("github api: the comment marker is empty")
	}
	path := "/repos/" + repo + "/issues/" + strconv.Itoa(number) + "/comments"
	url := a.baseURL + path + "?per_page=100"
	for page := 0; url != ""; page++ {
		if page == maxCommentPages {
			return 0, fmt.Errorf("github api: GET %s: more than %d pages of comments", path, maxCommentPages)
		}
		// The token goes only to the API root it was configured for.
		if !strings.HasPrefix(url, a.baseURL+"/") {
			return 0, fmt.Errorf("github api: GET %s: the next page is outside the API root", path)
		}
		var comments []apiComment
		_, next, err := a.fetch(ctx, url, path, &comments)
		if err != nil {
			return 0, err
		}
		for _, c := range comments {
			if strings.Contains(c.Body, marker) {
				return c.ID, nil
			}
		}
		url = next
	}
	return 0, nil
}

// CreateComment adds a comment to issue or pull request number, with
// POST /repos/{repo}/issues/{number}/comments, and returns its Id. It needs the
// pull-requests: write (or issues: write) permission.
func (a *GitHubAPI) CreateComment(ctx context.Context, repo string, number int, body string) (int64, error) {
	if !validRepo(repo) {
		return 0, fmt.Errorf("github api: repository %q is not owner/name", repo)
	}
	if number <= 0 {
		return 0, fmt.Errorf("github api: issue number %d is not valid", number)
	}
	path := "/repos/" + repo + "/issues/" + strconv.Itoa(number) + "/comments"
	var c apiComment
	if err := a.send(ctx, http.MethodPost, path, http.StatusCreated, body, &c); err != nil {
		return 0, err
	}
	return c.ID, nil
}

// UpdateComment replaces the body of comment id, with
// PATCH /repos/{repo}/issues/comments/{id}.
func (a *GitHubAPI) UpdateComment(ctx context.Context, repo string, id int64, body string) error {
	if !validRepo(repo) {
		return fmt.Errorf("github api: repository %q is not owner/name", repo)
	}
	if id <= 0 {
		return fmt.Errorf("github api: comment Id %d is not valid", id)
	}
	path := "/repos/" + repo + "/issues/comments/" + strconv.FormatInt(id, 10)
	var c apiComment
	return a.send(ctx, http.MethodPatch, path, http.StatusOK, body, &c)
}

// UpsertComment updates the comment that holds marker, or creates one when none does. body must
// contain marker, or the next run would not find the comment and would add another. It reports
// whether the comment was created.
func (a *GitHubAPI) UpsertComment(ctx context.Context, repo string, number int, marker, body string) (created bool, err error) {
	if !strings.Contains(body, marker) {
		return false, fmt.Errorf("github api: the comment body does not contain its marker")
	}
	id, err := a.FindComment(ctx, repo, number, marker)
	if err != nil {
		return false, err
	}
	if id == 0 {
		_, err := a.CreateComment(ctx, repo, number, body)
		return true, err
	}
	return false, a.UpdateComment(ctx, repo, id, body)
}

// send writes body as {"body": ...} to path and decodes the answer into v. wantStatus is the
// success status. The token goes only in the Authorization header.
func (a *GitHubAPI) send(ctx context.Context, method, path string, wantStatus int, body string, v any) error {
	payload, err := json.Marshal(map[string]string{"body": body})
	if err != nil {
		return fmt.Errorf("github api: %s %s: %w", method, path, err)
	}
	req, err := http.NewRequestWithContext(ctx, method, a.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("github api: %s %s: %w", method, path, err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "glassops")
	if a.token != "" {
		req.Header.Set("Authorization", "Bearer "+a.token)
	}
	resp, err := a.http.Do(req)
	if err != nil {
		return fmt.Errorf("github api: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxAPIBody))
	if err != nil {
		return fmt.Errorf("github api: %s %s: read body: %w", method, path, err)
	}
	if resp.StatusCode != wantStatus {
		var e apiError
		_ = json.Unmarshal(data, &e)
		msg := e.Message
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return &HTTPError{Method: method, Path: path, Status: resp.StatusCode, Message: msg}
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("github api: %s %s: parse body: %w", method, path, err)
	}
	return nil
}

// HTTPError is a write that GitHub answered with an unexpected status.
type HTTPError struct {
	Method, Path string
	Status       int
	Message      string
}

// Error implements error.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("github api: %s %s: HTTP %d: %s", e.Method, e.Path, e.Status, e.Message)
}
