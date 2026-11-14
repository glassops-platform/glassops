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
	execprovider "github.com/glassops-platform/glassops/internal/providers/exec"
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
	// deployCommand is the JSON array of strings given to --deploy-command, the exact argv
	// deploy -- will run for provider exec.
	deployCommand string
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
			"set. When the target's tier has deploy windows, the PERMIT expires when the window\n" +
			"that holds the gate's clock closes. It also expires when the next blackout that\n" +
			"covers the tier begins. Either applies only when it is before the configured TTL.\n" +
			"With provider exec, --deploy-command binds the command deploy -- will run.",
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
	f.StringVar(&o.deployCommand, "deploy-command", "", `provider exec only: the command deploy -- will run, as a JSON array of strings such as '["./deploy.sh","--prod"]'. The PERMIT binds its sha256, never the text`)
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

	commandSHA, err := gateDeployCommand(proj.provider, o.deployCommand)
	if err != nil {
		res.AddCheck("deploy_command", run.StatusError, err.Error())
		return app.finish(res, c, g)
	}

	signer, err := permit.SignerFor(cfg, permit.Keys{
		HMACKey: func() ([]byte, error) { return gateHMACKey(app.Getenv(permit.KeyEnv)) },
		NewHMAC: gateSigner,
	})
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

	// Earlier results go in first, then the plan. Neither may set a fact gate owns (target,
	// change, ci), and gate sets those after both.
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
	expires, capped := permitExpiry(cfg, tgt.Tier, now)
	blackoutCapped := false
	if !ephemeral {
		// The blackout rule does not run for an ephemeral target, so it does not shorten its PERMIT.
		if start, ok := policy.BlackoutStart(cfg, tgt.Tier, now); ok && start.Before(expires) {
			expires, capped, blackoutCapped = start.UTC(), false, true
		}
	}
	var policyErr error
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
		// The Go rules always run first. Rego, when configured, can only add results, and an
		// evaluator that cannot decide denies and makes gate exit 2.
		in := policy.Input{Facts: gr.doc, Config: cfg, Now: now}
		ok, results, err := policy.EvaluateAll(ctx, policy.Evaluators(rules, cfg), in)
		allow = allow && ok
		checks = append(checks, results...)
		policyErr = err
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
		ExpiresAt:    expires,
		Target:       tgt,
		Commit:       cictx.SHA,
		PlanSHA256:   cs.SHA256,
		ConfigSHA256: loaded.SHA256(),
		// The PERMIT binds how the plan files are used too, such as which are destructive.
		ChangeSetSHA256:     permit.ChangeSetDigest(cs),
		DeployCommandSHA256: commandSHA,
		Validation:          validated,
		Checks:              checks,
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
	res.Facts["gate.expires_at_window_end"] = capped
	res.Facts["gate.expires_at_blackout_start"] = blackoutCapped
	if commandSHA != "" {
		res.Facts["gate.deploy_command_sha256"] = commandSHA
	}
	verdict := "deny"
	if allow {
		verdict = "allow"
	}
	expiry := p.ExpiresAt.Format(time.RFC3339)
	if capped {
		expiry += ", when the deploy window closes"
	} else if blackoutCapped {
		expiry += ", when a blackout starts"
	}
	res.AddCheck("permit", run.StatusPass, fmt.Sprintf("%s, signed, written to %s, expires %s", verdict, permitPath, expiry))
	if !allow {
		res.Degrade(run.StatusFail)
	}
	if policyErr != nil {
		// The PERMIT is written as a deny, and the step is an error because policy could not
		// be evaluated.
		res.AddCheck("policy.evaluate", run.StatusError, policyErr.Error())
	}
	return app.finish(res, c, g)
}

// gateHMACKey reads the key from GLASSOPS_PERMIT_KEY with permit.ParseKey, so gate accepts
// exactly the keys permit verify accepts. No error includes the value.
func gateHMACKey(value string) ([]byte, error) {
	if strings.TrimSpace(value) == "" {
		return nil, fmt.Errorf("%s is not set, so no PERMIT can be signed", permit.KeyEnv)
	}
	key, err := permit.ParseKey(value)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", permit.KeyEnv, err)
	}
	return key, nil
}

// permitExpiry is when a PERMIT issued at now expires: after the configured TTL, or when the
// deploy window for tier that holds now closes, whichever is first. capped says the window
// end was used. A tier with no windows, or a clock outside them, keeps the TTL, because the
// window rule then decides, and an unusable window fails that rule.
func permitExpiry(cfg *config.Config, tier string, now time.Time) (time.Time, bool) {
	expires := now.Add(cfg.Policy.TTL())
	end, inside, err := policy.WindowEnd(cfg, tier, now)
	if err == nil && inside && end.Before(expires) {
		return end.UTC(), true
	}
	return expires, false
}

// gateDeployCommand parses --deploy-command and returns the sha256 the PERMIT binds it by, or
// "" when the flag is not set. The value must be a JSON array of at least one string, and the
// provider must take a deploy command, which is provider exec.
func gateDeployCommand(p providers.Provider, value string) (string, error) {
	if value == "" {
		return "", nil
	}
	if _, ok := p.(deployOverrider); !ok {
		return "", fmt.Errorf("--deploy-command needs provider exec, and this project uses %s", p.Name())
	}
	argv, err := parseDeployCommand(value)
	if err != nil {
		return "", err
	}
	return permit.DeployCommandDigest(argv), nil
}

// parseDeployCommand reads a JSON array of strings with at least one element and nothing
// after it. The error never repeats the value.
func parseDeployCommand(value string) ([]string, error) {
	dec := json.NewDecoder(strings.NewReader(value))
	var argv []string
	if err := dec.Decode(&argv); err != nil || dec.More() {
		return nil, errors.New(`--deploy-command must be a JSON array of strings, such as '["./deploy.sh","--prod"]'`)
	}
	if len(argv) == 0 || argv[0] == "" {
		return nil, errors.New("--deploy-command names no command")
	}
	return argv, nil
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
		// Only the allowlisted step and provider facts are copied. The result files are not
		// signed, so a ci., target. or change. fact in one would speak for gate.
		r.doc.MergeOnly(facts.ResultPrefixes(r.proj.provider.Name()), rr.Facts)
		if step == "validate" {
			validated, bound = r.validation(rr, t, cs)
		}
	}
	return validated, bound
}

// validation returns the validation a validate result proves for target t and plan cs, or nil
// with the reason. It binds only a passed check-only validate of the same target and the same
// plan and change set, at the test level validate recorded.
func (r *gateRun) validation(rr run.Result, t providers.Target, cs providers.ChangeSet) (*permit.Validation, policy.Result) {
	skip := func(why string) (*permit.Validation, policy.Result) {
		return nil, policy.Result{Rule: "validation.bound", Outcome: policy.Skipped, Detail: why + ", so deploy runs a full deploy"}
	}
	prefix := r.proj.provider.Name() + "."
	id, _ := rr.Facts[prefix+"deploy_id"].(string)
	checkOnly, _ := rr.Facts[prefix+"check_only"].(bool)
	planHash, _ := rr.Facts["validate.plan_sha256"].(string)
	level, _ := rr.Facts["validate.test_level"].(string)
	changeSet, _ := rr.Facts["validate.change_set_sha256"].(string)
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
	case changeSet == "":
		return skip("the validate result does not record which change set it validated")
	case changeSet != permit.ChangeSetDigest(cs):
		return skip("the validate result is for a different change set")
	case ranNoTests(rr.Facts[prefix+"tests_total"]):
		// [confirmed] A production org refused to quick deploy a RunRelevantTests validation
		// that ran no tests, with no other deploy in between, and quick deployed one that ran
		// tests. Binding it would only add a refused quick deploy before the full deploy.
		return skip("the validation ran no tests, and the target refuses to quick deploy a validation that ran none")
	}
	return &permit.Validation{DeployID: id, TestLevel: level, Passed: true},
		policy.Result{Rule: "validation.bound", Outcome: policy.Pass, Detail: "check-only validation " + id + " of this plan is bound for a quick deploy"}
}

// ranNoTests reports whether a tests_total fact says that no tests ran. A missing or
// non-numeric fact says nothing, so it is false.
func ranNoTests(v any) bool {
	switch n := v.(type) {
	case float64:
		return n == 0
	case int:
		return n == 0
	}
	return false
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
		case errors.Is(err, execprovider.ErrNoDescribe):
			// A provider exec target below prod with no exec.describe cannot report its Id,
			// so the configured Id and kind are recorded, as --no-connect does.
			check = policy.Result{Rule: "guard", Outcome: policy.Skipped, Detail: fmt.Sprintf("tier %s: %s, so the configured Id and kind are recorded", t.Tier, execprovider.ErrNoDescribe)}
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
// because a tool or infrastructure problem stopped gate from seeing the evidence. Any outcome
// that denies, including one gate does not know, is a fail, as it is in the PERMIT.
func ruleStatus(r policy.Result) run.Status {
	switch {
	case r.Outcome.Denies():
		if strings.HasPrefix(r.Rule, evidencePrefix) {
			return run.StatusError
		}
		return run.StatusFail
	case r.Outcome == policy.Skipped:
		return run.StatusSkipped
	default:
		return run.StatusPass
	}
}
