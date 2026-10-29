// Package gitref resolves the git refs a plan compares. A ref can be given explicitly or as
// "auto", which picks the base from the deploy tag, the target's branch or the pull request
// base branch. Git is reached through a Runner, so tests need no repository.
package gitref

import (
	"errors"
	"fmt"
	"strings"
)

// Auto asks Resolve to pick a ref.
const Auto = "auto"

// DeployedTagPrefix starts the tag that marks what a target last received. Phase 2 only
// reads it.
const DeployedTagPrefix = "glassops/deployed/"

// Runner runs git with args in dir and returns stdout. A non-zero exit is an error.
type Runner func(dir string, args ...string) (string, error)

// Options says what to resolve.
type Options struct {
	// Dir is the repository.
	Dir string
	// From and To are the requested refs, "auto" or empty meaning pick one.
	From string
	To   string
	// Target is the target the plan is for, if any.
	Target string
	// Branch is the target's configured branch, if any.
	Branch string
	// BaseBranch is the pull request base branch, if the run is a pull request.
	BaseBranch string
}

// Refs are the resolved commits.
type Refs struct {
	From string
	To   string
	// FromSource says how From was found: "explicit", "tag" or "merge-base".
	FromSource string
}

// ErrShallow means the clone has no full history.
var ErrShallow = errors.New("shallow clone: the history needed to compare refs is missing. " +
	"Fetch full history (GitHub: actions/checkout with fetch-depth: 0)")

// Resolve returns the commits to compare. A shallow clone is always an error, because a
// delta over missing history would be empty or wrong.
func Resolve(run Runner, o Options) (Refs, error) {
	shallow, err := run(o.Dir, "rev-parse", "--is-shallow-repository")
	if err != nil {
		return Refs{}, fmt.Errorf("not a git repository: %w", err)
	}
	if strings.TrimSpace(shallow) == "true" {
		return Refs{}, ErrShallow
	}

	to := o.To
	if to == "" || to == Auto {
		to = "HEAD"
	}
	toSHA, err := commit(run, o.Dir, to)
	if err != nil {
		return Refs{}, fmt.Errorf("--to %s: %w", to, err)
	}

	if o.From != "" && o.From != Auto {
		fromSHA, err := commit(run, o.Dir, o.From)
		if err != nil {
			return Refs{}, fmt.Errorf("--from %s: %w", o.From, err)
		}
		return Refs{From: fromSHA, To: toSHA, FromSource: "explicit"}, nil
	}

	from, source, err := autoFrom(run, o)
	if err != nil {
		return Refs{}, err
	}
	return Refs{From: from, To: toSHA, FromSource: source}, nil
}

func autoFrom(run Runner, o Options) (string, string, error) {
	const hint = "Pass --from."
	if o.Target != "" {
		tag := DeployedTagPrefix + o.Target
		if sha, err := commit(run, o.Dir, "refs/tags/"+tag); err == nil {
			return sha, "tag", nil
		}
		if o.Branch != "" {
			if sha, ok := mergeBase(run, o.Dir, o.Branch); ok {
				return sha, "merge-base", nil
			}
		}
		return "", "", fmt.Errorf("--from auto: target %s has no tag %s and %s. %s", o.Target, tag, branchNote(o.Branch), hint)
	}
	if o.BaseBranch != "" {
		if sha, ok := mergeBase(run, o.Dir, o.BaseBranch); ok {
			return sha, "merge-base", nil
		}
		return "", "", fmt.Errorf("--from auto: no merge base with the pull request base branch %s (tried origin/%s and %s). %s", o.BaseBranch, o.BaseBranch, o.BaseBranch, hint)
	}
	return "", "", fmt.Errorf("--from auto needs a target or a pull request base branch. %s", hint)
}

func branchNote(branch string) string {
	if branch == "" {
		return "the target sets no branch"
	}
	return fmt.Sprintf("there is no merge base with origin/%s or %s", branch, branch)
}

// mergeBase tries origin/<branch>, then <branch>.
func mergeBase(run Runner, dir, branch string) (string, bool) {
	for _, ref := range []string{"origin/" + branch, branch} {
		out, err := run(dir, "merge-base", "HEAD", ref)
		if sha := strings.TrimSpace(out); err == nil && sha != "" {
			return sha, true
		}
	}
	return "", false
}

// commit resolves ref to a commit SHA.
func commit(run Runner, dir, ref string) (string, error) {
	out, err := run(dir, "rev-parse", "--verify", "--quiet", ref+"^{commit}")
	sha := strings.TrimSpace(out)
	if err != nil || sha == "" {
		return "", fmt.Errorf("%q is not a commit in this repository", ref)
	}
	return sha, nil
}
