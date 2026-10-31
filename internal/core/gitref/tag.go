package gitref

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// DefaultRemote is the remote the deployed tag is pushed to.
const DefaultRemote = "origin"

// tagTarget is what a target name must look like to be used in a tag name. It keeps the tag
// a valid git ref and stops a name from reaching into another part of refs/tags.
var tagTarget = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*(\.[A-Za-z0-9_-]+)*$`)

// DeployedTag returns the tag name that marks what target last received, such as
// glassops/deployed/qa.
func DeployedTag(target string) (string, error) {
	if !tagTarget.MatchString(target) || strings.HasSuffix(target, ".lock") {
		return "", fmt.Errorf("target name %q cannot be used in a git tag", target)
	}
	return DeployedTagPrefix + target, nil
}

// LastDeployed returns the commit the deployed tag of target points at. ok is false when
// the tag does not exist. A shallow clone is ErrShallow, because a tag outside the fetched
// history would look missing, and a repository git cannot read is an error.
func LastDeployed(run Runner, dir, target string) (sha string, ok bool, err error) {
	tag, err := DeployedTag(target)
	if err != nil {
		return "", false, err
	}
	shallow, err := run(dir, "rev-parse", "--is-shallow-repository")
	if err != nil {
		return "", false, fmt.Errorf("not a git repository: %w", err)
	}
	if strings.TrimSpace(shallow) == "true" {
		return "", false, ErrShallow
	}
	sha, err = commit(run, dir, "refs/tags/"+tag)
	if err != nil {
		return "", false, nil
	}
	return sha, true, nil
}

// MoveDeployedTag points the deployed tag of target at commit in the local repository. The
// tag is lightweight, so it needs no git identity on the runner.
func MoveDeployedTag(run Runner, dir, target, commit string) error {
	tag, err := DeployedTag(target)
	if err != nil {
		return err
	}
	if commit == "" {
		return errors.New("move deployed tag: no commit given")
	}
	if _, err := run(dir, "tag", "-f", tag, commit); err != nil {
		return fmt.Errorf("move tag %s to %s: %w", tag, commit, err)
	}
	return nil
}

// PushDeployedTag force pushes the deployed tag of target to remote, replacing the one there.
func PushDeployedTag(run Runner, dir, remote, target string) error {
	tag, err := DeployedTag(target)
	if err != nil {
		return err
	}
	if remote == "" {
		remote = DefaultRemote
	}
	ref := "refs/tags/" + tag
	if _, err := run(dir, "push", "-f", remote, ref); err != nil {
		return fmt.Errorf("push %s to %s: %w", ref, remote, err)
	}
	return nil
}
