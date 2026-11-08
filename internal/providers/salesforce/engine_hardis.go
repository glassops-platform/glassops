package salesforce

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/providers"
)

// HardisTestedVersion is the sfdx-hardis version the hardis engine was tested with. It is also
// the oldest version accepted when salesforce.hardis.min_version is not set.
const HardisTestedVersion = "8.20.1"

// hardisPlugin is the sf plugin name of sfdx-hardis.
const hardisPlugin = "sfdx-hardis"

// HardisInstallSteps tells the user how to add sfdx-hardis. GlassOps never installs it.
const HardisInstallSteps = "add it to the image you run GlassOps in: " +
	"sf plugins install " + hardisPlugin + "@" + HardisTestedVersion +
	" (see examples/hardis/Dockerfile). GlassOps never installs it"

// ErrHardisMissing means sfdx-hardis is not installed as an sf plugin.
var ErrHardisMissing = errors.New("sfdx-hardis is not installed")

// ErrHardisTooOld means the installed sfdx-hardis is older than the minimum version.
var ErrHardisTooOld = errors.New("sfdx-hardis is older than the minimum version")

// HardisCheck is what CheckHardisPlugins found.
type HardisCheck struct {
	// Version is the installed version.
	Version string
	// Warning is set when the version is newer than HardisTestedVersion.
	Warning string
}

// CheckHardisPlugins reads the output of "sf plugins --json" and checks that sfdx-hardis is
// installed at min or later. An empty min means HardisTestedVersion. A missing plugin is
// ErrHardisMissing and an old one ErrHardisTooOld, both with the install steps.
func CheckHardisPlugins(out []byte, min string) (HardisCheck, error) {
	if min == "" {
		min = HardisTestedVersion
	}
	plugins, err := parsePlugins(out)
	if err != nil {
		return HardisCheck{}, fmt.Errorf("could not read sf plugins output: %w", err)
	}
	for _, pl := range plugins {
		if pl.Name != hardisPlugin {
			continue
		}
		c := HardisCheck{Version: pl.Version}
		older, err := versionLess(pl.Version, min)
		if err != nil {
			return c, fmt.Errorf("sfdx-hardis version: %w", err)
		}
		if older {
			return c, fmt.Errorf("%w: %s is installed, %s or later is needed, %s", ErrHardisTooOld, pl.Version, min, HardisInstallSteps)
		}
		if newer, _ := versionLess(HardisTestedVersion, pl.Version); newer {
			c.Warning = fmt.Sprintf("sfdx-hardis %s is newer than the tested version %s", pl.Version, HardisTestedVersion)
		}
		return c, nil
	}
	return HardisCheck{}, fmt.Errorf("%w, %s", ErrHardisMissing, HardisInstallSteps)
}

// version is a parsed major.minor.patch with an optional prerelease suffix.
type version struct {
	nums [3]int
	pre  string
}

// parseVersion reads "8", "8.20", "8.20.1" or "8.20.1-beta.1". Build metadata after "+" is
// ignored.
func parseVersion(s string) (version, error) {
	var v version
	s = strings.TrimPrefix(strings.TrimSpace(s), "v")
	s, _, _ = strings.Cut(s, "+")
	core, pre, _ := strings.Cut(s, "-")
	v.pre = pre
	parts := strings.Split(core, ".")
	if core == "" || len(parts) > 3 {
		return v, fmt.Errorf("not a version: %q", s)
	}
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return v, fmt.Errorf("not a version: %q", s)
		}
		v.nums[i] = n
	}
	return v, nil
}

// versionLess reports whether a is older than b. A prerelease is older than its release.
func versionLess(a, b string) (bool, error) {
	va, err := parseVersion(a)
	if err != nil {
		return false, err
	}
	vb, err := parseVersion(b)
	if err != nil {
		return false, err
	}
	for i := range va.nums {
		if va.nums[i] != vb.nums[i] {
			return va.nums[i] < vb.nums[i], nil
		}
	}
	switch {
	case va.pre == vb.pre, va.pre == "":
		return false, nil
	case vb.pre == "":
		return true, nil
	}
	return va.pre < vb.pre, nil
}

// hardisDeployCommand is the only sfdx-hardis command the engine runs. It wraps sf project
// deploy start with the same flags.
//
// The validate and quick wrappers are never run. In sfdx-hardis 8.20.1 both run sf project
// deploy start underneath: the validate wrapper made a real deploy that was not check-only,
// and the quick wrapper refused --job-id and -i. Validate and quick deploys go to sf directly.
const hardisDeployCommand = "hardis:project:deploy:start"

// hardisBlankedEnv are the tokens sfdx-hardis uses to post pull request comments and
// notifications. They are set empty for the hardis process, so GlassOps' report stays the only
// writer to the pull request.
var hardisBlankedEnv = []string{
	"GITHUB_TOKEN",
	"SYSTEM_ACCESSTOKEN",
	"CI_SFDX_HARDIS_GITLAB_TOKEN",
	"CI_SFDX_HARDIS_BITBUCKET_TOKEN",
}

// hardisCommandKeys are the top-level keys of an sfdx-hardis project config that run commands
// around a deploy. They sit outside the plan a PERMIT binds, so the engine refuses them.
var hardisCommandKeys = []string{"commandsPreDeploy", "commandsPostDeploy"}

// hardisEngine plans with GlassOps' own plan and deploys with sfdx-hardis, which the user
// installs. GlassOps hands it the plan's package.xml and destructive manifests unchanged, so
// the PERMIT binds what is deployed. Validate, quick deploy and resume use sf directly.
type hardisEngine struct {
	p   *Provider
	cfg config.Hardis
}

func newHardisEngine(p *Provider, cfg config.Hardis) Engine {
	return hardisEngine{p: p, cfg: cfg}
}

func (e hardisEngine) Name() string { return config.EngineHardis }

// Plan is the native plan. sfdx-hardis computes no change set here: GlassOps' plan is the
// source of truth.
func (e hardisEngine) Plan(ctx context.Context, req providers.PlanRequest) (providers.ChangeSet, error) {
	return e.p.nativePlan(ctx, req)
}

// Validate runs the native check-only validate after the hardis checks. A project config that
// defines deploy commands fails it before any sf call.
func (e hardisEngine) Validate(ctx context.Context, s providers.Session, cs providers.ChangeSet, opts providers.ValidateOptions) (providers.ValidateResult, error) {
	if _, ok := asSession(s); !ok {
		return providers.ValidateResult{}, errors.New("salesforce: validate needs a session from this provider")
	}
	if problems, err := HardisCommandProblems(opts.Dir); err != nil || len(problems) > 0 {
		if err != nil {
			return providers.ValidateResult{}, err
		}
		return providers.ValidateResult{Problems: problems, Facts: hardisRefusedFacts()}, nil
	}
	version, err := e.checkInstalled(ctx, opts.Dir)
	if err != nil {
		return providers.ValidateResult{}, err
	}
	res, err := e.p.nativeValidate(ctx, s, cs, opts)
	if err != nil {
		return res, err
	}
	res.Facts = withHardisFacts(res.Facts, version)
	res.Facts["salesforce.hardis.validate_path"] = config.EngineNative
	return res, nil
}

// Deploy deploys cs with sfdx-hardis. A resume and a quick deploy go to sf directly, and a
// refused quick deploy falls back to a hardis deploy unless opts.QuickOnly is set.
func (e hardisEngine) Deploy(ctx context.Context, s providers.Session, cs providers.ChangeSet, opts providers.DeployOptions) (providers.DeployResult, error) {
	sess, ok := asSession(s)
	if !ok {
		return providers.DeployResult{}, errors.New("salesforce: deploy needs a session from this provider")
	}
	dir, err := filepath.Abs(opts.Dir)
	if err != nil {
		return providers.DeployResult{}, fmt.Errorf("salesforce: resolve project dir: %w", err)
	}
	if problems, err := HardisCommandProblems(dir); err != nil || len(problems) > 0 {
		if err != nil {
			return providers.DeployResult{}, err
		}
		return providers.DeployResult{Problems: problems, Facts: hardisRefusedFacts()}, nil
	}
	if err := checkPlanFiles(dir, cs); err != nil {
		return providers.DeployResult{}, err
	}
	version, err := e.checkInstalled(ctx, dir)
	if err != nil {
		return providers.DeployResult{}, err
	}

	if opts.ResumeID != "" {
		res, err := e.p.nativeDeploy(ctx, s, cs, opts)
		if err != nil {
			return res, err
		}
		res.Facts = withHardisFacts(res.Facts, version)
		res.Facts["salesforce.hardis.resume_path"] = config.EngineNative
		return res, nil
	}

	var fallback map[string]any
	if opts.ValidatedID != "" {
		quickOpts := opts
		quickOpts.QuickOnly = true
		res, err := e.p.nativeDeploy(ctx, s, cs, quickOpts)
		if err != nil {
			return res, err
		}
		refused, _ := res.Facts["salesforce.quick_refused"].(bool)
		if !refused || opts.QuickOnly {
			res.Facts = withHardisFacts(res.Facts, version)
			res.Facts["salesforce.hardis.quick_path"] = config.EngineNative
			return res, nil
		}
		reason := ""
		if len(res.Problems) > 0 {
			reason = strings.TrimSuffix(res.Problems[0], " A quick deploy is required, so no full deploy was run.")
		}
		fallback = map[string]any{
			"salesforce.quick_refused":        true,
			"salesforce.quick_fallback":       reason,
			"salesforce.hardis.quick_path":    config.EngineNative,
			"salesforce.hardis.fallback_path": config.EngineHardis,
		}
	}

	args := []string{hardisDeployCommand,
		"--manifest", filepath.Join(dir, filepath.FromSlash(cs.Manifest)),
		"--target-org", sess.alias,
		"--wait", waitMinutes(opts.Wait),
	}
	for _, d := range cs.Destructive {
		flag := "--post-destructive-changes"
		if path.Base(d) == preDestructiveName {
			flag = "--pre-destructive-changes"
		}
		args = append(args, flag, filepath.Join(dir, filepath.FromSlash(d)))
	}
	if opts.TestLevel != "" {
		args = append(args, "--test-level", opts.TestLevel)
		if opts.TestLevel == "RunSpecifiedTests" {
			for _, t := range opts.Tests {
				args = append(args, "--tests", t)
			}
		}
	}
	scratch, err := e.p.isScratch(ctx, sess)
	if err != nil {
		return providers.DeployResult{}, err
	}
	if scratch {
		// The same reason as the native deploy: a scratch org tracks source, and the project
		// is the source of truth for it.
		args = append(args, "--ignore-conflicts")
	}
	env, runErr := e.hardis(ctx, sess, dir, args)
	res, err := e.p.readDeploy(ctx, sess, dir, env, runErr, false, fallback)
	if err != nil {
		return res, err
	}
	res.Facts = withHardisFacts(res.Facts, version)
	res.Facts["salesforce.hardis.deploy_path"] = config.EngineHardis
	return res, nil
}

// checkInstalled reads "sf plugins --json" and checks sfdx-hardis against min_version. A
// missing or old sfdx-hardis is a tool error that carries the install steps.
func (e hardisEngine) checkInstalled(ctx context.Context, dir string) (string, error) {
	out, err := e.p.exec(ctx, Cmd{Name: "sf", Args: []string{"plugins", "--json"}, Dir: dir, Env: sfBaseEnv()})
	if err != nil {
		return "", fmt.Errorf("salesforce: engine hardis could not list sf plugins: %w", err)
	}
	if out.ExitCode != 0 {
		return "", fmt.Errorf("salesforce: engine hardis could not list sf plugins (exit %d), %s", out.ExitCode, HardisInstallSteps)
	}
	got, err := CheckHardisPlugins(out.Stdout, e.cfg.MinVersion)
	if err != nil {
		return "", fmt.Errorf("salesforce: engine hardis: %w", err)
	}
	return got.Version, nil
}

// hardis runs one sfdx-hardis command with --json in session sess, with the pull request
// tokens blanked, and returns the sf envelope it wraps.
func (e hardisEngine) hardis(ctx context.Context, sess *session, dir string, args []string) (envelope, error) {
	full := append(append([]string{}, args...), "--json")
	c := Cmd{Name: "sf", Args: full, Dir: dir, Env: append(append(sfBaseEnv(), sess.env()...), hardisEnv()...)}
	out, err := e.p.exec(ctx, c)
	if err != nil {
		return envelope{}, err
	}
	return readHardisOutput(commandName(args), out.Stdout, out.ExitCode)
}

// hardisEnv blanks every token in hardisBlankedEnv.
func hardisEnv() []string {
	env := make([]string, 0, len(hardisBlankedEnv))
	for _, k := range hardisBlankedEnv {
		env = append(env, k+"=")
	}
	return env
}

// hardisOutput is the envelope sfdx-hardis prints with --json. Its result holds outputstring:
// the whole sf envelope as an object when the deploy succeeded, or text when it failed.
type hardisOutput struct {
	Status  int             `json:"status"`
	Name    string          `json:"name"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Result  *struct {
		OutputString json.RawMessage `json:"outputstring"`
	} `json:"result"`
}

// hardisDeployID matches the "Deployment ID: <Id>" line of an sfdx-hardis failure summary.
var hardisDeployID = regexp.MustCompile(`(?m)^Deployment ID:\s*(0Af[0-9A-Za-z]{12}(?:[0-9A-Za-z]{3})?)\s*$`)

// readHardisOutput turns sfdx-hardis output into the sf envelope it wraps, and an *SFError for
// a failure, the same as Provider.sf returns, so the native deploy reading applies.
//
// A failure's outputstring is a summary followed by the sf error JSON, or by nothing. The
// deploy Id is read from that JSON, or else from the summary's "Deployment ID" line, and put in
// the error's data so readDeploy follows it with deploy report. Output that fits neither shape
// is an error, never a pass.
func readHardisOutput(command string, stdout []byte, exitCode int) (envelope, error) {
	i := bytes.IndexByte(stdout, '{')
	if i < 0 {
		return envelope{}, fmt.Errorf("sf %s (exit %d): printed no JSON", command, exitCode)
	}
	var outer hardisOutput
	if err := json.NewDecoder(bytes.NewReader(stdout[i:])).Decode(&outer); err != nil {
		return envelope{}, fmt.Errorf("sf %s (exit %d): could not read the JSON output: %w", command, exitCode, err)
	}
	if outer.Result == nil || len(outer.Result.OutputString) == 0 || string(outer.Result.OutputString) == "null" {
		// sf refused the command before sfdx-hardis ran, so this is an sf envelope.
		if outer.Status != 0 {
			return envelope{}, &SFError{Command: command, Name: outer.Name, Message: outer.Message, Status: outer.Status, Data: outer.Data}
		}
		return envelope{}, fmt.Errorf("sf %s (exit %d): the output has no outputstring", command, exitCode)
	}
	raw := bytes.TrimSpace(outer.Result.OutputString)

	if raw[0] == '{' {
		var inner envelope
		if err := json.Unmarshal(raw, &inner); err != nil {
			return envelope{}, fmt.Errorf("sf %s: could not read the sf result inside the output: %w", command, err)
		}
		if inner.Status != 0 {
			return inner, &SFError{Command: command, Name: inner.Name, Message: inner.Message, Status: inner.Status, Data: inner.Data}
		}
		if outer.Status != 0 || exitCode != 0 {
			return envelope{}, fmt.Errorf("sf %s (exit %d): a failure status around a passing sf result", command, exitCode)
		}
		return inner, nil
	}

	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return envelope{}, fmt.Errorf("sf %s: outputstring is neither an object nor text: %w", command, err)
	}
	status := outer.Status
	if status == 0 {
		status = 1
	}
	if inner, ok := embeddedEnvelope(text); ok {
		if inner.Status == 0 {
			inner.Status = status
		}
		return envelope{}, &SFError{Command: command, Name: inner.Name, Message: inner.Message, Status: inner.Status, Data: inner.Data}
	}
	summary := hardisSummary(text)
	if m := hardisDeployID.FindStringSubmatch(text); m != nil {
		data, _ := json.Marshal(map[string]string{"id": m[1]})
		return envelope{}, &SFError{Command: command, Name: "HardisDeployFailed", Message: summary, Status: status, Data: data}
	}
	return envelope{}, &SFError{Command: command, Name: "HardisDeployFailed", Message: summary, Status: status}
}

// embeddedEnvelope finds the sf error JSON inside a failure's outputstring. It starts at a
// line that opens with a brace.
func embeddedEnvelope(text string) (envelope, bool) {
	for k := 0; k < len(text); k++ {
		if text[k] != '{' || (k > 0 && text[k-1] != '\n') {
			continue
		}
		var env envelope
		if json.NewDecoder(strings.NewReader(text[k:])).Decode(&env) == nil && (env.Name != "" || env.Status != 0) {
			return env, true
		}
	}
	return envelope{}, false
}

// hardisSummary keeps the status, components and tests lines of a failure summary, or the
// first line when it has none of them.
func hardisSummary(text string) string {
	var keep []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		for _, prefix := range []string{"Status:", "Components:", "Apex tests:"} {
			if strings.HasPrefix(line, prefix) {
				keep = append(keep, line)
			}
		}
	}
	if len(keep) == 0 {
		first, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
		return oneLine(first)
	}
	return "sfdx-hardis deploy failed: " + strings.Join(keep, ", ")
}

// HardisCommandProblems lists the sfdx-hardis project config files under dir that define
// commandsPreDeploy or commandsPostDeploy, one line per file and key. Those commands run
// around a deploy and are not part of the plan a PERMIT binds, so engine hardis refuses them.
// A config file that cannot be read is listed too, so it never passes unread.
func HardisCommandProblems(dir string) ([]string, error) {
	var files []string
	for _, pattern := range []string{
		filepath.Join(dir, "config", ".sfdx-hardis*.yml"),
		filepath.Join(dir, "config", ".sfdx-hardis*.yaml"),
		filepath.Join(dir, "config", "branches", ".sfdx-hardis*.yml"),
		filepath.Join(dir, "config", "branches", ".sfdx-hardis*.yaml"),
	} {
		m, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("salesforce: look for sfdx-hardis config: %w", err)
		}
		files = append(files, m...)
	}
	sort.Strings(files)
	var problems []string
	for _, f := range files {
		rel, err := filepath.Rel(dir, f)
		if err != nil {
			rel = f
		}
		rel = filepath.ToSlash(rel)
		data, err := os.ReadFile(f)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: could not read the sfdx-hardis config: %v", rel, err))
			continue
		}
		var top map[string]any
		if err := yaml.Unmarshal(data, &top); err != nil {
			problems = append(problems, fmt.Sprintf("%s: could not read the sfdx-hardis config as YAML, so engine hardis cannot check it for deploy commands", rel))
			continue
		}
		for _, key := range hardisCommandKeys {
			if _, ok := top[key]; ok {
				problems = append(problems, fmt.Sprintf("%s defines %s. Engine hardis refuses deploy commands from the sfdx-hardis config, because they run outside the plan the PERMIT binds. Remove the key, or run those steps as their own pipeline steps", rel, key))
			}
		}
	}
	return problems, nil
}

// withHardisFacts adds the engine and the sfdx-hardis version to facts.
func withHardisFacts(facts map[string]any, version string) map[string]any {
	if facts == nil {
		facts = map[string]any{}
	}
	facts["salesforce.engine"] = config.EngineHardis
	facts["salesforce.hardis.version"] = version
	return facts
}

// hardisRefusedFacts are the facts of a run refused for sfdx-hardis deploy commands.
func hardisRefusedFacts() map[string]any {
	return map[string]any{
		"salesforce.engine":                  config.EngineHardis,
		"salesforce.hardis.commands_refused": true,
	}
}
