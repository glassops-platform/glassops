package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/glassops-platform/glassops/internal/ci"
	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/exitcode"
	"github.com/glassops-platform/glassops/internal/core/facts"
	"github.com/glassops-platform/glassops/internal/core/guard"
	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/core/policy"
	"github.com/glassops-platform/glassops/internal/core/run"
	"github.com/glassops-platform/glassops/internal/providers"
)

// sodRule is the rule whose result fills the PERMIT's separation_of_duties block.
const sodRule = "separation_of_duties"

// evidencePrefix names the checks gate adds when it could not gather a piece of evidence.
const evidencePrefix = "evidence."

// The policy and signing seams gate calls. They are variables so tests can supply stand-ins.
var (
	gateRules  = policy.Default
	gateSigner = permit.NewHMAC
	gateSign   = permit.Sign
	gateWrite  = permit.Write
)

type gateOptions struct {
	target     string
	planPath   string
	permitPath string
	noConnect  bool
}

func newGateCommand(app *App, g *globals) *cobra.Command {
	o := &gateOptions{}
	cmd := &cobra.Command{
		Use:   "gate",
		Short: "Evaluate policy for a deploy and write a signed PERMIT",
		Long: "gate reads the change set, earlier results and the CI context, builds the facts\n" +
			"for one target and runs the policy rules over them. It always writes a signed\n" +
			"PERMIT to .glassops/permit.json, which records allow or deny and every rule result.\n" +
			"It logs in to the target to run the org Id and tier guard unless --no-connect is\n" +
			"set. It exits 1 when policy denies the deploy, and 2 when GLASSOPS_PERMIT_KEY is not\n" +
			"set.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.gate(cmd.Context(), g, o)
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.target, "target", "", "target name from glassops.yaml (required)")
	f.StringVar(&o.planPath, "plan", "", "change set file (default <dir>/.glassops/plan.json)")
	f.StringVar(&o.permitPath, "permit", "", "where the PERMIT is written (default <dir>/.glassops/permit.json)")
	f.BoolVar(&o.noConnect, "no-connect", false, "do not log in to the target, and use its configured Id")
	_ = cmd.MarkFlagRequired("target")
	return cmd
}

// gateRun is the state one gate run builds up.
type gateRun struct {
	app  *App
	cfg  *config.Config
	proj *project
	doc  *facts.Document
	// evidence holds a failing check for each piece of evidence gate could not gather because
	// of a tool or infrastructure error. Any one of them denies the PERMIT and makes the step
	// exit 2.
	evidence []policy.Result
}

func (r *gateRun) evidenceFail(name, detail string) {
	r.evidence = append(r.evidence, policy.Result{Rule: evidencePrefix + name, Outcome: policy.Fail, Detail: detail})
}

// gate builds the facts for one target, evaluates the rules and writes a signed PERMIT. A
// missing key, plan or target stops it with exit 2 before any PERMIT is written. Every later
// problem still produces a signed PERMIT, one that denies.
func (app *App) gate(ctx context.Context, g *globals, o *gateOptions) error {
	c, err := app.detectCI(g)
	if err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	proj, err := app.loadProject(g)
	if err != nil {
		return err
	}
	// loadProject keeps only the config, and the PERMIT needs the hash of the file itself.
	loaded, err := config.Load(g.dir, g.configPath, proj.provider.Name())
	if err != nil {
		return &exitcode.ExitError{Code: exitcode.Error, Err: err}
	}
	cfg := loaded.Config
	rememberProjectDir(cfg, g.dir)

	res := run.New("gate", app.Now)
	res.Provider = proj.provider.Name()
	res.Target = o.target

	planPath := o.planPath
	if planPath == "" {
		planPath = filepath.Join(g.dir, ".glassops", "plan.json")
	}
	permitPath := o.permitPath
	if permitPath == "" {
		permitPath = filepath.Join(g.dir, ".glassops", "permit.json")
	}

	// A PERMIT left by an earlier run must never stand in for this one, whatever happens next.
	if err := os.Remove(permitPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		res.AddCheck("permit", run.StatusError, fmt.Sprintf("remove the earlier PERMIT: %v", err))
		return app.finish(res, c, g)
	}

	signer, err := permitSigner(app.Getenv(permit.KeyEnv))
	if err != nil {
		res.AddCheck("permit.key", run.StatusError, err.Error())
		return app.finish(res, c, g)
	}
	cs, err := providers.ReadChangeSet(planPath)
	if err != nil {
		res.AddCheck("plan", run.StatusError, err.Error())
		return app.finish(res, c, g)
	}
	t, err := app.target(cfg, o.target)
	if err != nil {
		res.AddCheck("target", run.StatusError, err.Error())
		return app.finish(res, c, g)
	}
	res.Facts["target.source"] = targetSource(t)

	gr := &gateRun{app: app, cfg: cfg, proj: proj, doc: facts.New()}
	cictx := c.Context()
	if cictx.SHA == "" {
		gr.evidenceFail("commit", "the CI reported no commit, so the PERMIT cannot be bound to one")
	}
	if cs.SHA256 == "" {
		gr.evidenceFail("plan", "the change set has no plan_sha256, so the PERMIT cannot be bound to it")
	} else if err := providers.CheckDigest(g.dir, cs); err != nil {
		// The PERMIT names the plan by its hash, so plan files edited after plan must never be
		// signed for.
		gr.evidenceFail("plan.digest", err.Error())
	}

	// Earlier results go in first, so the facts gate owns (target, change, ci) always win.
	validated, bound := gr.resultFacts(g, t, cs)
	gr.doc.Merge("", facts.FromChangeSet(cs).Facts)
	tgt, guardCheck, connected := gr.targetFacts(ctx, t, o.noConnect)
	// Design 7a: an ephemeral target gets an automatic PERMIT with only the org Id guard
	// applied. It is issued only when the guard really ran and saw a scratch target.
	ephemeral := connected && tgt.Kind == string(providers.KindScratch)
	gr.ciFacts(ctx, cictx, !ephemeral)

	factsPath := filepath.Join(g.dir, ".glassops", "facts.json")
	if err := facts.Write(factsPath, gr.doc); err != nil {
		gr.evidenceFail("facts", err.Error())
	} else {
		res.Artifacts = append(res.Artifacts, factsPath)
	}

	now := app.Now().UTC()
	// validation.bound only says whether deploy may quick deploy. It never denies.
	checks := []policy.Result{guardCheck, bound}
	allow := guardCheck.Outcome != policy.Fail
	if ephemeral {
		checks = append(checks, policy.Result{
			Rule:    "ephemeral",
			Outcome: policy.Pass,
			Detail:  "ephemeral target, automatic PERMIT with only the org Id guard applied",
		})
	} else {
		rules := gateRules()
		if len(rules) == 0 {
			// An empty rule set would allow everything, so it is a defect, never a pass.
			gr.evidenceFail("policy", "no policy rules are loaded")
		}
		ok, results := policy.Evaluate(rules, policy.Input{Facts: gr.doc, Config: cfg, Now: now})
		allow = allow && ok
		checks = append(checks, results...)
	}
	if len(gr.evidence) > 0 {
		allow = false
		checks = append(checks, gr.evidence...)
	}

	p := permit.Permit{
		Schema:       permit.Schema,
		ID:           newPermitID(),
		Allow:        allow,
		IssuedAt:     now,
		ExpiresAt:    now.Add(cfg.Policy.TTL()),
		Target:       tgt,
		Commit:       cictx.SHA,
		PlanSHA256:   cs.SHA256,
		ConfigSHA256: loaded.SHA256(),
		Validation:   validated,
		Checks:       checks,
	}
	if p.ID == "" {
		res.AddCheck("permit", run.StatusError, "could not generate a PERMIT Id")
		return app.finish(res, c, g)
	}
	if ephemeral {
		p.SeparationOfDuties = &permit.SoD{Mode: "not_applicable"}
	} else {
		p.SeparationOfDuties = sodBlock(checks, gr.doc)
	}

	for _, ch := range checks {
		res.AddCheck(ch.Rule, ruleStatus(ch), ch.Detail)
	}
	if err := gateSign(&p, signer); err != nil {
		res.AddCheck("permit", run.StatusError, fmt.Sprintf("sign: %v", err))
		return app.finish(res, c, g)
	}
	if err := gateWrite(permitPath, p); err != nil {
		res.AddCheck("permit", run.StatusError, fmt.Sprintf("write: %v", err))
		return app.finish(res, c, g)
	}
	res.Artifacts = append(res.Artifacts, permitPath)
	res.Facts["gate.allow"] = allow
	res.Facts["gate.ephemeral"] = ephemeral
	res.Facts["gate.permit_id"] = p.ID
	res.Facts["gate.expires_at"] = p.ExpiresAt.Format(time.RFC3339)
	verdict := "deny"
	if allow {
		verdict = "allow"
	}
	res.AddCheck("permit", run.StatusPass, fmt.Sprintf("%s, signed, written to %s, expires %s", verdict, permitPath, p.ExpiresAt.Format(time.RFC3339)))
	if !allow {
		res.Degrade(run.StatusFail)
	}
	return app.finish(res, c, g)
}

// permitSigner reads the key from GLASSOPS_PERMIT_KEY with permit.ParseKey, so gate accepts
// exactly the keys permit verify accepts, and returns the signer for it. No error includes the
// value.
func permitSigner(value string) (permit.Signer, error) {
	if strings.TrimSpace(value) == "" {
		return nil, fmt.Errorf("%s is not set, so no PERMIT can be signed", permit.KeyEnv)
	}
	key, err := permit.ParseKey(value)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", permit.KeyEnv, err)
	}
	return gateSigner(key), nil
}

// newPermitID returns 16 random bytes in hex, or "" when the system has no randomness.
func newPermitID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// resultFacts merges the facts from results/plan.json and results/validate.json when they
// exist. It returns the validation the PERMIT binds to, or nil, and the validation.bound check
// that says which and why. A result that exists but cannot be read is an evidence failure.
func (r *gateRun) resultFacts(g *globals, t providers.Target, cs providers.ChangeSet) (*permit.Validation, policy.Result) {
	bound := policy.Result{Rule: "validation.bound", Outcome: policy.Skipped, Detail: "no validate result, so deploy runs a full deploy"}
	var validated *permit.Validation
	for _, step := range []string{"plan", "validate"} {
		rr, ok, err := readResult(filepath.Join(g.results(), step+".json"))
		if err != nil {
			r.evidenceFail("results."+step, err.Error())
			if step == "validate" {
				bound.Detail = "the validate result cannot be read, so deploy runs a full deploy"
			}
			continue
		}
		if !ok {
			continue
		}
		r.doc.Merge("", facts.FromResult(rr).Facts)
		if step == "validate" {
			validated, bound = r.validation(rr, t, cs)
		}
	}
	return validated, bound
}

// validation returns the validation a validate result proves for target t and plan cs, or nil
// with the reason. It binds only a passed check-only validate of the same target and the same
// plan, at the test level validate recorded.
func (r *gateRun) validation(rr run.Result, t providers.Target, cs providers.ChangeSet) (*permit.Validation, policy.Result) {
	skip := func(why string) (*permit.Validation, policy.Result) {
		return nil, policy.Result{Rule: "validation.bound", Outcome: policy.Skipped, Detail: why + ", so deploy runs a full deploy"}
	}
	prefix := r.proj.provider.Name() + "."
	id, _ := rr.Facts[prefix+"deploy_id"].(string)
	checkOnly, _ := rr.Facts[prefix+"check_only"].(bool)
	planHash, _ := rr.Facts["validate.plan_sha256"].(string)
	level, _ := rr.Facts["validate.test_level"].(string)
	switch {
	case rr.Status != run.StatusPass:
		return skip(fmt.Sprintf("the validate result is %s, not pass", rr.Status))
	case rr.Target != t.Name:
		return skip(fmt.Sprintf("the validate result is for target %q, not %q", rr.Target, t.Name))
	case !checkOnly:
		return skip("the validate result is not a check-only run")
	case id == "":
		return skip("the validate result has no deploy Id")
	case planHash == "":
		return skip("the validate result does not record which plan it validated")
	case planHash != cs.SHA256:
		return skip("the validate result is for a different plan")
	}
	return &permit.Validation{DeployID: id, TestLevel: level, Passed: true},
		policy.Result{Rule: "validation.bound", Outcome: policy.Pass, Detail: "check-only validation " + id + " of this plan is bound for a quick deploy"}
}

// readResult reads a result envelope. A missing file is not an error.
func readResult(path string) (run.Result, bool, error) {
	var rr run.Result
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return rr, false, nil
	}
	if err != nil {
		return rr, false, fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &rr); err != nil {
		return rr, false, fmt.Errorf("parse %s: %w", path, err)
	}
	if rr.Schema != run.Schema {
		return rr, false, fmt.Errorf("%s: schema is %q, want %q", path, rr.Schema, run.Schema)
	}
	return rr, true, nil
}

// targetFacts sets the target.* facts. Unless noConnect is set it logs in and runs the guard,
// and the Id and kind come from what the target reports. It returns the PERMIT target, the
// guard's check, and whether the guard ran and passed.
func (r *gateRun) targetFacts(ctx context.Context, t providers.Target, noConnect bool) (permit.Target, policy.Result, bool) {
	pt := permit.Target{Name: t.Name, Tier: t.Tier, ID: t.ID, Kind: configuredKind(t.Tier)}
	check := policy.Result{Rule: "guard", Outcome: policy.Skipped, Detail: "--no-connect, so the guard did not run and the configured Id and kind are recorded"}
	connected := false
	if !noConnect {
		s, info, err := r.app.connect(ctx, r.proj.provider, t)
		var mm *guard.MismatchError
		switch {
		case errors.As(err, &mm):
			check = policy.Result{Rule: "guard", Outcome: policy.Fail, Detail: err.Error()}
			r.evidenceFail("target", "the guard refused the target")
		case err != nil:
			check = policy.Result{Rule: "guard", Outcome: policy.Fail, Detail: "the guard could not run"}
			r.evidenceFail("target", err.Error())
		default:
			_ = s.Close()
			r.doc.Merge(r.proj.provider.Name()+".", info.Facts)
			pt.ID = info.ID
			pt.Kind = string(info.Kind)
			check = policy.Result{Rule: "guard", Outcome: policy.Pass, Detail: fmt.Sprintf("%s %s", info.Kind, info.ID)}
			connected = true
		}
	}
	r.doc.Set(facts.TargetName, pt.Name)
	r.doc.Set(facts.TargetTier, pt.Tier)
	r.doc.Set(facts.TargetID, pt.ID)
	r.doc.Set(facts.TargetKind, pt.Kind)
	return pt, check, connected
}

// configuredKind is the kind the guard requires for a tier, or unknown for a tier with no
// such rule. It is the kind --no-connect records.
func configuredKind(tier string) string {
	switch tier {
	case guard.TierProd, guard.TierDevHub:
		return string(providers.KindProduction)
	case guard.TierEphemeral:
		return string(providers.KindScratch)
	}
	return string(providers.KindUnknown)
}

// ciFacts sets the ci.* facts from the CI context and, when useAPI is set and the CI gave a
// token, from the GitHub API. A fact nobody reported stays absent, so a rule that needs it
// fails.
func (r *gateRun) ciFacts(ctx context.Context, c ci.Context, useAPI bool) {
	set := func(key, v string) {
		if v != "" {
			r.doc.Set(key, v)
		}
	}
	set(facts.CIRepo, c.Repo)
	set(facts.CICommit, c.SHA)
	set(facts.CIRef, c.Ref)
	set(facts.CIEvent, c.Event)
	// On a push this is GITHUB_ACTOR, the pusher or merger.
	set(facts.CIActor, c.Actor)

	prNumber, author, labels := c.PRNumber, c.PRAuthor, c.Labels
	labelsKnown := len(c.Labels) > 0
	approvals := append([]facts.Approval(nil), c.Approvals...)
	approvalsKnown := len(c.Approvals) > 0

	var api *ci.GitHubAPI
	if useAPI && c.Token != "" && c.Repo != "" {
		api = ci.NewGitHubAPI(r.app.Getenv("GITHUB_API_URL"), c.Token, nil)
	}
	if api != nil {
		switch {
		case prNumber != "":
			// A pull request event: GITHUB_SHA is the test merge commit, so read the pull
			// request itself.
			n, err := strconv.Atoi(prNumber)
			if err != nil {
				r.evidenceFail("ci.pull_request", fmt.Sprintf("pull request number %q is not a number", prNumber))
				break
			}
			a, l, err := api.Pull(ctx, c.Repo, n)
			if err != nil {
				r.evidenceFail("ci.pull_request", err.Error())
				break
			}
			if author == "" {
				author = a
			}
			if !labelsKnown {
				labels, labelsKnown = l, true
			}
		case c.SHA != "":
			// A push: find the pull request the commit was merged from.
			n, a, l, err := api.PullForCommit(ctx, c.Repo, c.SHA)
			switch {
			case err != nil:
				r.evidenceFail("ci.pull_request", err.Error())
			case n != 0:
				prNumber = strconv.Itoa(n)
				if author == "" {
					author = a
				}
				if !labelsKnown {
					labels, labelsKnown = l, true
				}
			}
		}
		if c.RunID != "" {
			got, err := api.RunApprovals(ctx, c.Repo, c.RunID)
			if err != nil {
				r.evidenceFail("ci.approvals", err.Error())
			} else {
				approvals = append(approvals, got...)
				approvalsKnown = true
			}
		}
	}

	set(facts.CIPRNumber, prNumber)
	set(facts.CIPRAuthor, author)
	if labelsKnown {
		r.doc.Set(facts.CILabels, labels)
		// The configured approval label on the pull request is label approval evidence. The
		// approver is whoever added the label, read from the issue events. Without the API it
		// stays empty, and the separation of duties rule fails it.
		if sod := r.cfg.Policy.Rules.SeparationOfDuties; sod != nil && sod.Approval != nil && sod.Approval.Label != "" {
			for _, l := range labels {
				if l != sod.Approval.Label {
					continue
				}
				ap := facts.Approval{Mechanism: "label", Label: l}
				if n, err := strconv.Atoi(prNumber); api != nil && err == nil {
					login, err := api.LabelAdder(ctx, c.Repo, n, l)
					if err != nil {
						r.evidenceFail("ci.label_approver", err.Error())
					}
					ap.Approver = login
				}
				approvals = append(approvals, ap)
				approvalsKnown = true
			}
		}
	}
	if approvalsKnown {
		r.doc.Set(facts.CIApprovals, approvals)
	}
}

// sodBlock builds the PERMIT's separation_of_duties record from the rule's result.
func sodBlock(checks []policy.Result, doc *facts.Document) *permit.SoD {
	author, _ := doc.String(facts.CIPRAuthor)
	deployer, _ := doc.String(facts.CIActor)
	s := &permit.SoD{Mode: "not_applicable", Author: author, Deployer: deployer}
	for _, ch := range checks {
		if ch.Rule != sodRule {
			continue
		}
		switch ch.Outcome {
		case policy.WaivedByConfig:
			s.Mode = "waived_by_config"
		case policy.SatisfiedByApproval:
			s.Mode = "satisfied_by_approval"
			s.Mechanism = ch.Mechanism
			s.Approver = ch.Approver
		case policy.Pass, policy.Fail:
			s.Mode = "enforced"
		}
		return s
	}
	return s
}

// ruleStatus maps a rule outcome to a result check status. An evidence failure is an error,
// because a tool or infrastructure problem stopped gate from seeing the evidence.
func ruleStatus(r policy.Result) run.Status {
	switch r.Outcome {
	case policy.Fail:
		if strings.HasPrefix(r.Rule, evidencePrefix) {
			return run.StatusError
		}
		return run.StatusFail
	case policy.Skipped:
		return run.StatusSkipped
	default:
		return run.StatusPass
	}
}
