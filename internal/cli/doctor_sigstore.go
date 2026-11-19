package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/core/run"
)

func init() {
	doctorExtras = append(doctorExtras, doctorSigstore)
}

// doctorSigstore checks what sigstore PERMIT signing needs when policy.permit.signing is
// sigstore: cosign on PATH, a CI where sigstore works, the job's OIDC token on GitHub Actions,
// an identity and issuer shaped like a GitHub workflow's (the reusable deploy workflow's when
// the project calls it), and the source repository verifying pins. It is silent for hmac. It
// reads only whether the OIDC variables are set, never their values.
func doctorSigstore(app *App, cfg *config.Config) []run.Check {
	if permit.Method(cfg) != config.SigningSigstore {
		return nil
	}
	return []run.Check{app.cosignCheck(), app.sigstoreCICheck(),
		sigstoreIdentityCheck(cfg.Policy.Permit, reusableWorkflowRefs(projectDirOf(cfg))),
		app.sigstoreRepositoryCheck(cfg.Policy.Permit)}
}

// ReusableDeployWorkflow is the GlassOps reusable deploy workflow a consumer repository calls
// with uses: <ReusableDeployWorkflow>@<ref>. A PERMIT signed inside it carries its identity,
// https://github.com/<ReusableDeployWorkflow>@<full ref>, never the caller's own workflow.
const ReusableDeployWorkflow = "glassops-platform/glassops/.github/workflows/glassops-deploy.yml"

// reusableUses matches a uses: line that calls ReusableDeployWorkflow and captures its ref.
var reusableUses = regexp.MustCompile(`(?m)^\s*(?:-\s*)?uses:\s*["']?` + regexp.QuoteMeta(ReusableDeployWorkflow) + `@([^\s"'#]+)`)

// reusableWorkflowRefs returns the refs at which the workflows in dir/.github/workflows call
// ReusableDeployWorkflow, sorted and without repeats. A missing or unreadable directory
// means none.
func reusableWorkflowRefs(dir string) []string {
	if dir == "" {
		dir = "."
	}
	seen := map[string]bool{}
	for _, pattern := range []string{"*.yml", "*.yaml"} {
		files, _ := filepath.Glob(filepath.Join(dir, ".github", "workflows", pattern))
		for _, f := range files {
			data, err := os.ReadFile(f)
			if err != nil {
				continue
			}
			for _, m := range reusableUses.FindAllStringSubmatch(string(data), -1) {
				seen[m[1]] = true
			}
		}
	}
	refs := make([]string, 0, len(seen))
	for r := range seen {
		refs = append(refs, r)
	}
	sort.Strings(refs)
	return refs
}

// sigstoreRepositoryCheck reports the source repository permit verify and deploy pin a
// signature to. A configured repository that differs from the running one is a warning,
// because PERMITs gate signs here would then never verify. No repository to pin is a warning
// too, because verifying off GitHub Actions then stops with a config error.
func (app *App) sigstoreRepositoryCheck(p config.PermitSigning) run.Check {
	name := "permit.repository"
	repo, source, err := permit.PinnedRepository(p, app.Getenv)
	if errors.Is(err, permit.ErrNoRepository) {
		return run.Check{Name: name, Status: run.StatusPass,
			Detail: "warning: no repository to pin, so permit verify and deploy stop here. Set policy.permit.repository to owner/name to verify off GitHub Actions"}
	}
	if err != nil {
		return run.Check{Name: name, Status: run.StatusPass, Detail: "warning: " + strings.TrimPrefix(err.Error(), "permit: ")}
	}
	if running := app.Getenv(permit.RepositoryEnv); app.Getenv("GITHUB_ACTIONS") == "true" && running != "" && running != repo {
		return run.Check{Name: name, Status: run.StatusPass,
			Detail: fmt.Sprintf("warning: policy.permit.repository is %s but this run is in %s, so PERMITs gate signs here will not verify", repo, running)}
	}
	return run.Check{Name: name, Status: run.StatusPass, Detail: fmt.Sprintf("%s (from %s)", repo, source)}
}

// cosignCheck runs "cosign version" and reports its GitVersion line.
func (app *App) cosignCheck() run.Check {
	name := "tool.cosign"
	out, err := app.Run(".", permit.Cosign, "version")
	if err != nil {
		return run.Check{Name: name, Status: run.StatusFail,
			Detail: "cosign not found, sigstore signing needs it (https://docs.sigstore.dev/cosign/system_config/installation/)"}
	}
	for _, line := range strings.Split(string(out), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "GitVersion:"); ok {
			return run.Check{Name: name, Status: run.StatusPass, Detail: strings.TrimSpace(v)}
		}
	}
	return run.Check{Name: name, Status: run.StatusPass, Detail: "installed, version not reported"}
}

// sigstoreCICheck reports whether gate can sign keyless here. Signing needs GitHub Actions and
// a job with id-token: write. Another known CI fails, because sigstore is not supported there.
// A run outside any CI is skipped, because only verifying works there. A pull request job
// without the token gets a warning, since the default pr pipeline does not run gate.
func (app *App) sigstoreCICheck() run.Check {
	name := "permit.sigstore"
	if other := permit.OtherCI(app.Getenv); other != "" {
		return run.Check{Name: name, Status: run.StatusFail,
			Detail: fmt.Sprintf("sigstore signing works on GitHub Actions only, this run is on %s. Use policy.permit.signing: hmac there", other)}
	}
	if app.Getenv("GITHUB_ACTIONS") != "true" {
		return run.Check{Name: name, Status: run.StatusSkipped,
			Detail: "not on GitHub Actions: gate signs only in a GitHub job with id-token: write, permit verify works here"}
	}
	if app.Getenv(permit.OIDCRequestURLEnv) != "" && app.Getenv(permit.OIDCRequestTokenEnv) != "" {
		return run.Check{Name: name, Status: run.StatusPass, Detail: "the job can request an OIDC token for cosign"}
	}
	msg := "the job has no OIDC token, so gate cannot sign. Add permissions: id-token: write to the job that runs gate"
	if strings.HasPrefix(app.Getenv("GITHUB_EVENT_NAME"), "pull_request") {
		return run.Check{Name: name, Status: run.StatusPass, Detail: "warning: " + msg}
	}
	return run.Check{Name: name, Status: run.StatusFail, Detail: msg}
}

// sigstoreIdentityCheck warns when the configured identity or issuer does not look like a
// GitHub Actions workflow certificate, which every signature would then fail to match. When
// the project calls the reusable deploy workflow at refs, the identity must name that
// workflow, because the certificate carries the called workflow, not the caller.
func sigstoreIdentityCheck(p config.PermitSigning, refs []string) run.Check {
	name := "permit.identity"
	var warn []string
	if p.Issuer != permit.GitHubIssuer {
		warn = append(warn, fmt.Sprintf("issuer is not %s", permit.GitHubIssuer))
	}
	if !strings.HasPrefix(p.Identity, "https://github.com/") || !strings.Contains(p.Identity, "/.github/workflows/") || !strings.Contains(p.Identity, "@refs/") {
		warn = append(warn, "identity is not shaped like https://github.com/<owner>/<repo>/.github/workflows/<file>@refs/<ref>")
	}
	if len(refs) > 0 {
		prefix := "https://github.com/" + ReusableDeployWorkflow + "@"
		calls := ReusableDeployWorkflow + "@" + strings.Join(refs, ", @")
		ref, isReusable := strings.CutPrefix(p.Identity, prefix)
		switch {
		case !isReusable:
			warn = append(warn, fmt.Sprintf("this project calls %s, and a PERMIT signed there carries the identity %s<full ref> (%srefs/tags/%s when %s is a tag), not this repository's own workflow",
				calls, prefix, prefix, refs[0], refs[0]))
		case !refMatches(ref, refs):
			warn = append(warn, fmt.Sprintf("identity names the reusable workflow at %s but this project calls %s", ref, calls))
		}
	}
	if len(warn) > 0 {
		return run.Check{Name: name, Status: run.StatusPass,
			Detail: "warning: " + strings.Join(warn, ", ") + ", so GitHub signatures will not verify"}
	}
	return run.Check{Name: name, Status: run.StatusPass, Detail: p.Identity}
}

// refMatches reports whether ref, the part of an identity after "@", is one of the refs a
// workflow calls, written in full (refs/tags/v1) or short (v1).
func refMatches(ref string, refs []string) bool {
	for _, r := range refs {
		if ref == r || strings.HasSuffix(ref, "/"+r) {
			return true
		}
	}
	return false
}
