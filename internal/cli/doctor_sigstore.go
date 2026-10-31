package cli

import (
	"fmt"
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
// and an identity and issuer shaped like a GitHub workflow's. It is silent for hmac. It reads
// only whether the OIDC variables are set, never their values.
func doctorSigstore(app *App, cfg *config.Config) []run.Check {
	if permit.Method(cfg) != config.SigningSigstore {
		return nil
	}
	return []run.Check{app.cosignCheck(), app.sigstoreCICheck(), sigstoreIdentityCheck(cfg.Policy.Permit)}
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
// GitHub Actions workflow certificate, which every signature would then fail to match.
func sigstoreIdentityCheck(p config.PermitSigning) run.Check {
	name := "permit.identity"
	var warn []string
	if p.Issuer != permit.GitHubIssuer {
		warn = append(warn, fmt.Sprintf("issuer is not %s", permit.GitHubIssuer))
	}
	if !strings.HasPrefix(p.Identity, "https://github.com/") || !strings.Contains(p.Identity, "/.github/workflows/") || !strings.Contains(p.Identity, "@refs/") {
		warn = append(warn, "identity is not shaped like https://github.com/<owner>/<repo>/.github/workflows/<file>@refs/<ref>")
	}
	if len(warn) > 0 {
		return run.Check{Name: name, Status: run.StatusPass,
			Detail: "warning: " + strings.Join(warn, ", ") + ", so GitHub signatures will not verify"}
	}
	return run.Check{Name: name, Status: run.StatusPass, Detail: p.Identity}
}
