package exec

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/permit"
	"github.com/glassops-platform/glassops/internal/providers"
)

// fakeCommand stands in for one shell command. It gets the command's environment as a map and
// returns the exit status.
type fakeCommand func(env map[string]string, stdout io.Writer) int

// fakeShell records every command it runs and answers from cmds. An unknown command is an
// error, as a command that cannot start is.
type fakeShell struct {
	cmds map[string]fakeCommand
	runs []fakeRun
}

type fakeRun struct {
	dir     string
	command string
	env     map[string]string
}

func (f *fakeShell) run(_ context.Context, dir string, env []string, command string, stdout, _ io.Writer) (int, error) {
	m := map[string]string{}
	for _, kv := range env {
		k, v, _ := strings.Cut(kv, "=")
		m[k] = v
	}
	f.runs = append(f.runs, fakeRun{dir: dir, command: command, env: m})
	c, ok := f.cmds[command]
	if !ok {
		return -1, errors.New("cannot start")
	}
	return c(m, stdout), nil
}

// fakeGit answers git by its joined arguments.
type fakeGit map[string]string

func (f fakeGit) run(_ context.Context, _ string, args ...string) ([]byte, error) {
	key := strings.Join(args, " ")
	out, ok := f[key]
	if !ok {
		return nil, fmt.Errorf("unexpected git %s", key)
	}
	return []byte(out), nil
}

const (
	fromSHA = "1111111111111111111111111111111111111111"
	toSHA   = "2222222222222222222222222222222222222222"
)

// syntheticDiff is synthetic git diff --name-status -z output: two changed files, one added
// and one deleted under app.
var syntheticDiff = "M\x00app/b.txt\x00A\x00app/a.txt\x00D\x00app/old.txt\x00"

func diffKey(paths ...string) string {
	return strings.Join(append([]string{"--literal-pathspecs", "diff", "--name-status", "--no-renames", "--relative", "-z", fromSHA, toSHA, "--"}, paths...), " ")
}

func newTestProvider(t *testing.T, cfg config.Exec, git fakeGit, cmds map[string]fakeCommand, base ...string) (*Provider, *fakeShell, string) {
	t.Helper()
	sh := &fakeShell{cmds: cmds}
	p := NewWithRunners(sh.run, git.run, base)
	if err := p.Configure(&config.Config{Exec: cfg}); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	p.Bind(Workspace{Dir: dir, Stdout: io.Discard, Stderr: io.Discard})
	return p, sh, dir
}

func planReq(dir string) providers.PlanRequest {
	return providers.PlanRequest{Dir: dir, From: fromSHA, To: toSHA, OutDir: filepath.Join(dir, ".glassops", "plan")}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestNeverDetected(t *testing.T) {
	p := New()
	if p.Name() != "exec" {
		t.Fatalf("name = %q", p.Name())
	}
	if ok, err := p.Detect(t.TempDir()); ok || err != nil {
		t.Fatalf("Detect = %v, %v", ok, err)
	}
}

func TestConfigureKeepsExecSection(t *testing.T) {
	p := New()
	if err := p.Configure(&config.Config{Exec: config.Exec{Paths: []string{"app"}, Deploy: "./d.sh"}}); err != nil {
		t.Fatal(err)
	}
	if p.cfg.Deploy != "./d.sh" || len(p.cfg.Paths) != 1 {
		t.Fatalf("cfg = %+v", p.cfg)
	}
}

func TestPermitKeyEnvMatchesPermitPackage(t *testing.T) {
	if permitKeyEnv != permit.KeyEnv {
		t.Fatalf("permitKeyEnv = %q, permit.KeyEnv = %q", permitKeyEnv, permit.KeyEnv)
	}
}

func TestPlanDeltaListsChangedAndDeletedFiles(t *testing.T) {
	p, _, dir := newTestProvider(t, config.Exec{Paths: []string{"app"}}, fakeGit{diffKey("app"): syntheticDiff}, nil)
	cs, err := p.Plan(context.Background(), planReq(dir))
	if err != nil {
		t.Fatal(err)
	}
	if cs.Provider != "exec" || cs.Dir != ".glassops/plan" || cs.Manifest != ".glassops/plan/changes.txt" || cs.Empty {
		t.Fatalf("cs = %+v", cs)
	}
	if !reflect.DeepEqual(cs.Destructive, []string{".glassops/plan/deleted.txt"}) {
		t.Fatalf("destructive = %v", cs.Destructive)
	}
	if got := readFile(t, filepath.Join(dir, ".glassops", "plan", "changes.txt")); got != "app/a.txt\napp/b.txt\n" {
		t.Fatalf("changes.txt = %q", got)
	}
	if got := readFile(t, filepath.Join(dir, ".glassops", "plan", "deleted.txt")); got != "app/old.txt\n" {
		t.Fatalf("deleted.txt = %q", got)
	}
	if cs.Facts["exec.changed_count"] != 2 || cs.Facts["exec.destructive_count"] != 1 || cs.Facts["exec.plan_command"] != false {
		t.Fatalf("facts = %v", cs.Facts)
	}
	want, err := providers.PlanDigest(dir, filepath.Join(dir, ".glassops", "plan"))
	if err != nil || cs.SHA256 != want || cs.SHA256 == "" {
		t.Fatalf("digest = %q, want %q (%v)", cs.SHA256, want, err)
	}
	if err := providers.CheckDigest(dir, cs); err != nil {
		t.Fatal(err)
	}
}

func TestPlanDigestIsStable(t *testing.T) {
	git := fakeGit{diffKey("app", "lib"): syntheticDiff}
	p, _, dir := newTestProvider(t, config.Exec{Paths: []string{"app", "lib"}}, git, nil)
	first, err := p.Plan(context.Background(), planReq(dir))
	if err != nil {
		t.Fatal(err)
	}
	// The same changes listed in another order give the same plan.
	git[diffKey("app", "lib")] = "D\x00app/old.txt\x00A\x00app/a.txt\x00M\x00app/b.txt\x00"
	second, err := p.Plan(context.Background(), planReq(dir))
	if err != nil {
		t.Fatal(err)
	}
	if first.SHA256 != second.SHA256 {
		t.Fatalf("digest changed: %s then %s", first.SHA256, second.SHA256)
	}
	// A different change set gives a different digest.
	git[diffKey("app", "lib")] = "M\x00app/b.txt\x00"
	third, err := p.Plan(context.Background(), planReq(dir))
	if err != nil {
		t.Fatal(err)
	}
	if third.SHA256 == first.SHA256 {
		t.Fatal("a different change set has the same digest")
	}
	if _, err := os.Stat(filepath.Join(dir, ".glassops", "plan", "deleted.txt")); !os.IsNotExist(err) {
		t.Fatalf("deleted.txt from the earlier plan is still there: %v", err)
	}
}

func TestPlanEmpty(t *testing.T) {
	p, _, dir := newTestProvider(t, config.Exec{Paths: []string{"app"}}, fakeGit{diffKey("app"): ""}, nil)
	cs, err := p.Plan(context.Background(), planReq(dir))
	if err != nil {
		t.Fatal(err)
	}
	if !cs.Empty || len(cs.Destructive) != 0 || cs.SHA256 == "" {
		t.Fatalf("cs = %+v", cs)
	}
}

func TestPlanFullListsTrackedFiles(t *testing.T) {
	key := "--literal-pathspecs ls-tree -r -z --name-only " + toSHA + " -- app"
	p, _, dir := newTestProvider(t, config.Exec{Paths: []string{"./app/"}}, fakeGit{key: "app/z.txt\x00app/a.txt\x00"}, nil)
	req := planReq(dir)
	req.From, req.Full = "", true
	cs, err := p.Plan(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !cs.Full || cs.Empty || len(cs.Destructive) != 0 {
		t.Fatalf("cs = %+v", cs)
	}
	if got := readFile(t, filepath.Join(dir, ".glassops", "plan", "changes.txt")); got != "app/a.txt\napp/z.txt\n" {
		t.Fatalf("changes.txt = %q", got)
	}
}

func TestPlanCommandAddsFilesToTheDigest(t *testing.T) {
	cmds := map[string]fakeCommand{
		"./plan.sh": func(env map[string]string, _ io.Writer) int {
			_ = os.WriteFile(filepath.Join(env[EnvPlanDir], "extra.json"), []byte(`{"x":1}`), 0o644)
			return 0
		},
	}
	git := fakeGit{diffKey("app"): ""}
	p, sh, dir := newTestProvider(t, config.Exec{Paths: []string{"app"}, Plan: "./plan.sh"}, git, cmds)
	cs, err := p.Plan(context.Background(), planReq(dir))
	if err != nil {
		t.Fatal(err)
	}
	if cs.Empty {
		t.Fatal("a plan with a file from exec.plan must not be empty")
	}
	if cs.Facts["exec.plan_files"] != 1 || cs.Facts["exec.plan_command"] != true || cs.Facts["exec.result_file"] != false {
		t.Fatalf("facts = %v", cs.Facts)
	}
	env := sh.runs[0].env
	wantDir, _ := filepath.Abs(filepath.Join(dir, ".glassops", "plan"))
	if env[EnvPlanDir] != wantDir || env[EnvFrom] != fromSHA || env[EnvTo] != toSHA || env[EnvResultsDir] == "" {
		t.Fatalf("env = %v", env)
	}
	if _, ok := env[EnvTarget]; ok {
		t.Fatal("plan has no target")
	}
	bare, _, _ := newTestProvider(t, config.Exec{Paths: []string{"app"}}, git, nil)
	bare.Bind(Workspace{Dir: dir, Stdout: io.Discard, Stderr: io.Discard})
	bareCS, err := bare.Plan(context.Background(), planReq(dir))
	if err != nil {
		t.Fatal(err)
	}
	if bareCS.SHA256 == cs.SHA256 {
		t.Fatal("the file exec.plan wrote is not in the digest")
	}
}

func TestPlanCommandFailureIsAnError(t *testing.T) {
	cmds := map[string]fakeCommand{"./plan.sh": func(map[string]string, io.Writer) int { return 3 }}
	p, _, dir := newTestProvider(t, config.Exec{Paths: []string{"app"}, Plan: "./plan.sh"}, fakeGit{diffKey("app"): ""}, cmds)
	if _, err := p.Plan(context.Background(), planReq(dir)); err == nil || !strings.Contains(err.Error(), "exited 3") {
		t.Fatalf("err = %v", err)
	}
}

func TestPlanRefusals(t *testing.T) {
	cases := []struct {
		name  string
		paths []string
		edit  func(*providers.PlanRequest)
		want  string
	}{
		{"no paths", nil, nil, "exec.paths is empty"},
		{"outside", []string{"../x"}, nil, "inside the project"},
		{"absolute", []string{"/etc"}, nil, "inside the project"},
		{"pathspec magic", []string{":(glob)*"}, nil, "inside the project"},
		{"plan dir outside .glassops", []string{"app"}, func(r *providers.PlanRequest) { r.OutDir = filepath.Join(r.Dir, "app") }, "must be inside"},
		{"delta without refs", []string{"app"}, func(r *providers.PlanRequest) { r.From = "" }, "needs both refs"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, _, dir := newTestProvider(t, config.Exec{Paths: c.paths}, fakeGit{}, nil)
			req := planReq(dir)
			if c.edit != nil {
				c.edit(&req)
			}
			if _, err := p.Plan(context.Background(), req); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want %q", err, c.want)
			}
		})
	}
}

func TestPlanRefusesLineBreakInName(t *testing.T) {
	p, _, dir := newTestProvider(t, config.Exec{Paths: []string{"app"}}, fakeGit{diffKey("app"): "A\x00app/a\nb\x00"}, nil)
	if _, err := p.Plan(context.Background(), planReq(dir)); err == nil || !strings.Contains(err.Error(), "line break") {
		t.Fatalf("err = %v", err)
	}
}

func session4(name, id, tier string) providers.Session {
	return &session{t: providers.Target{Name: name, ID: id, Tier: tier}}
}

func TestDescribe(t *testing.T) {
	print := func(s string, code int) fakeCommand {
		return func(_ map[string]string, w io.Writer) int {
			_, _ = io.WriteString(w, s)
			return code
		}
	}
	cases := []struct {
		name     string
		describe string
		tier     string
		out      fakeCommand
		want     providers.TargetInfo
		err      string
		is       error
	}{
		{"prod needs describe", "", "prod", nil, providers.TargetInfo{}, "", ErrDescribeRequired},
		{"other tiers skip", "", "qa", nil, providers.TargetInfo{}, "", ErrNoDescribe},
		{"reads id and kind", "./describe.sh", "prod", print(`{"id": "acct-1", "kind": "production"}`+"\n", 0), providers.TargetInfo{ID: "acct-1", Kind: providers.KindProduction}, "", nil},
		{"exit code", "./describe.sh", "qa", print(`{"id":"a","kind":"sandbox"}`, 1), providers.TargetInfo{}, "exit 1", nil},
		{"not JSON", "./describe.sh", "qa", print("acct-1", 0), providers.TargetInfo{}, "is not", nil},
		{"unknown field", "./describe.sh", "qa", print(`{"id":"a","kind":"sandbox","name":"x"}`, 0), providers.TargetInfo{}, "is not", nil},
		{"two values", "./describe.sh", "qa", print(`{"id":"a","kind":"sandbox"}{"id":"b","kind":"sandbox"}`, 0), providers.TargetInfo{}, "more than one", nil},
		{"empty id", "./describe.sh", "qa", print(`{"id":"","kind":"sandbox"}`, 0), providers.TargetInfo{}, "id is empty", nil},
		{"unknown kind", "./describe.sh", "qa", print(`{"id":"a","kind":"prod"}`, 0), providers.TargetInfo{}, "kind is", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cmds := map[string]fakeCommand{}
			if c.out != nil {
				cmds[c.describe] = c.out
			}
			p, sh, _ := newTestProvider(t, config.Exec{Paths: []string{"app"}, Describe: c.describe}, fakeGit{}, cmds)
			info, err := p.Describe(context.Background(), session4("t1", "acct-1", c.tier))
			switch {
			case c.is != nil:
				if !errors.Is(err, c.is) || len(sh.runs) != 0 {
					t.Fatalf("err = %v, runs = %d", err, len(sh.runs))
				}
			case c.err != "":
				if err == nil || !strings.Contains(err.Error(), c.err) {
					t.Fatalf("err = %v, want %q", err, c.err)
				}
			default:
				if err != nil || info.ID != c.want.ID || info.Kind != c.want.Kind {
					t.Fatalf("info = %+v, err = %v", info, err)
				}
				if sh.runs[0].env[EnvTarget] != "t1" || sh.runs[0].env[EnvTargetID] != "acct-1" {
					t.Fatalf("env = %v", sh.runs[0].env)
				}
			}
		})
	}
}

// writeChangeSet writes a plan.json and a PERMIT file, so their default paths exist.
func writeChangeSet(t *testing.T, dir string) providers.ChangeSet {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, ".glassops", "plan"), 0o755); err != nil {
		t.Fatal(err)
	}
	cs := providers.ChangeSet{Schema: providers.PlanSchema, Provider: "exec", Dir: ".glassops/plan", Manifest: ".glassops/plan/changes.txt"}
	if err := providers.WriteChangeSet(filepath.Join(dir, ".glassops", "plan.json"), cs); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".glassops", "permit.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	return cs
}

func TestDeployCommandGetsEveryVariableAndNotTheKey(t *testing.T) {
	cmds := map[string]fakeCommand{"./deploy.sh": func(map[string]string, io.Writer) int { return 0 }}
	base := []string{"PATH=/bin", permitKeyEnv + "=c2VjcmV0", "GLASSOPS_TARGET=stale", "TOOL_TOKEN=x"}
	p, sh, dir := newTestProvider(t, config.Exec{Paths: []string{"app"}, Deploy: "./deploy.sh"}, fakeGit{}, cmds, base...)
	cs := writeChangeSet(t, dir)
	out, err := p.Deploy(context.Background(), session4("qa", "acct-2", "qa"), cs, providers.DeployOptions{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Passed || out.Facts["exec.result_file"] != false || out.Facts["exec.exit_code"] != 0 || out.Facts["exec.deploy_override"] != false {
		t.Fatalf("out = %+v", out)
	}
	run := sh.runs[0]
	abs := func(parts ...string) string {
		p, _ := filepath.Abs(filepath.Join(append([]string{dir}, parts...)...))
		return p
	}
	want := map[string]string{
		EnvPermit:     abs(".glassops", "permit.json"),
		EnvPlan:       abs(".glassops", "plan.json"),
		EnvPlanDir:    abs(".glassops", "plan"),
		EnvTarget:     "qa",
		EnvTargetID:   "acct-2",
		EnvResultsDir: abs(".glassops", "results"),
		"PATH":        "/bin",
		"TOOL_TOKEN":  "x",
	}
	for k, v := range want {
		if run.env[k] != v {
			t.Errorf("%s = %q, want %q", k, run.env[k], v)
		}
	}
	if _, ok := run.env[permitKeyEnv]; ok {
		t.Fatal("the PERMIT key reached the deploy command")
	}
	if run.dir != dir {
		t.Fatalf("dir = %q", run.dir)
	}
}

func TestDeployOverride(t *testing.T) {
	cmds := map[string]fakeCommand{"./other.sh 'a b'": func(map[string]string, io.Writer) int { return 0 }}
	p, sh, dir := newTestProvider(t, config.Exec{Paths: []string{"app"}, Deploy: "./deploy.sh"}, fakeGit{}, cmds)
	p.OverrideDeploy("./other.sh 'a b'")
	out, err := p.Deploy(context.Background(), session4("qa", "acct-2", "qa"), writeChangeSet(t, dir), providers.DeployOptions{Dir: dir})
	if err != nil || !out.Passed || sh.runs[0].command != "./other.sh 'a b'" || out.Facts["exec.deploy_override"] != true {
		t.Fatalf("out = %+v, err = %v, runs = %+v", out, err, sh.runs)
	}
}

func TestDeployCannotResume(t *testing.T) {
	p, _, dir := newTestProvider(t, config.Exec{Paths: []string{"app"}, Deploy: "./deploy.sh"}, fakeGit{}, nil)
	if _, err := p.Deploy(context.Background(), session4("qa", "a", "qa"), providers.ChangeSet{}, providers.DeployOptions{Dir: dir, ResumeID: "x"}); err == nil {
		t.Fatal("resume must be refused")
	}
}

func writeResult(body string) fakeCommand {
	return func(env map[string]string, _ io.Writer) int {
		_ = os.WriteFile(filepath.Join(env[EnvResultsDir], "exec-deploy.json"), []byte(body), 0o644)
		return 0
	}
}

func withExit(c fakeCommand, code int) fakeCommand {
	return func(env map[string]string, w io.Writer) int {
		c(env, w)
		return code
	}
}

func TestDeployResultFile(t *testing.T) {
	good := `{"schema":"glassops.engine-result/v1","step":"deploy","passed":true,"id":"dep-9","components_deployed":3,"facts":{"exec.tool":"x"}}`
	failed := `{"schema":"glassops.engine-result/v1","step":"deploy","passed":false,"problems":["thing broke"]}`
	cases := []struct {
		name     string
		cmd      fakeCommand
		passed   bool
		problems []string
		err      string
	}{
		{"passed", writeResult(good), true, nil, ""},
		{"no file, exit 0", func(map[string]string, io.Writer) int { return 0 }, true, nil, ""},
		{"no file, exit 2", func(map[string]string, io.Writer) int { return 2 }, false, []string{"exec.deploy exited 2"}, ""},
		{"file says failed", writeResult(failed), false, []string{"thing broke"}, ""},
		{"file passed, exit 1", withExit(writeResult(good), 1), false, []string{"exec.deploy exited 1 but its result file says passed, so it counts as failed"}, ""},
		{"file failed, no problems", writeResult(`{"schema":"glassops.engine-result/v1","step":"deploy","passed":false}`), false, []string{"exec.deploy exited 0 but its result file says not passed"}, ""},
		{"wrong schema", writeResult(`{"schema":"x","step":"deploy","passed":true}`), false, nil, "schema is"},
		{"wrong step", writeResult(`{"schema":"glassops.engine-result/v1","step":"validate","passed":true}`), false, nil, "step is"},
		{"passed missing", writeResult(`{"schema":"glassops.engine-result/v1","step":"deploy"}`), false, nil, "passed is missing"},
		{"unknown field", writeResult(`{"schema":"glassops.engine-result/v1","step":"deploy","passed":true,"ok":1}`), false, nil, "does not match"},
		{"bad fact key", writeResult(`{"schema":"glassops.engine-result/v1","step":"deploy","passed":true,"facts":{"deploy.id":"x"}}`), false, nil, "does not start with exec."},
		{"negative count", writeResult(`{"schema":"glassops.engine-result/v1","step":"deploy","passed":true,"components_deployed":-1}`), false, nil, "negative"},
		{"not JSON", writeResult(`passed`), false, nil, "does not match"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, _, dir := newTestProvider(t, config.Exec{Paths: []string{"app"}, Deploy: "./deploy.sh"}, fakeGit{}, map[string]fakeCommand{"./deploy.sh": c.cmd})
			out, err := p.Deploy(context.Background(), session4("qa", "a", "qa"), writeChangeSet(t, dir), providers.DeployOptions{Dir: dir})
			if c.err != "" {
				if err == nil || !strings.Contains(err.Error(), c.err) {
					t.Fatalf("err = %v, want %q", err, c.err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if out.Passed != c.passed || !reflect.DeepEqual(out.Problems, c.problems) {
				t.Fatalf("passed = %v, problems = %q", out.Passed, out.Problems)
			}
			if c.name == "passed" && (out.ID != "dep-9" || out.ComponentsDeployed != 3 || out.Facts["exec.tool"] != "x" || out.Facts["exec.result_file"] != true) {
				t.Fatalf("out = %+v", out)
			}
		})
	}
}

func TestStaleResultFileIsRemoved(t *testing.T) {
	cmds := map[string]fakeCommand{"./deploy.sh": func(map[string]string, io.Writer) int { return 1 }}
	p, _, dir := newTestProvider(t, config.Exec{Paths: []string{"app"}, Deploy: "./deploy.sh"}, fakeGit{}, cmds)
	stale := filepath.Join(dir, ".glassops", "results", "exec-deploy.json")
	_ = os.MkdirAll(filepath.Dir(stale), 0o755)
	if err := os.WriteFile(stale, []byte(`{"schema":"glassops.engine-result/v1","step":"deploy","passed":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := p.Deploy(context.Background(), session4("qa", "a", "qa"), writeChangeSet(t, dir), providers.DeployOptions{Dir: dir})
	if err != nil || out.Passed || out.Facts["exec.result_file"] != false {
		t.Fatalf("out = %+v, err = %v", out, err)
	}
}

func TestResultsDirFromWorkspace(t *testing.T) {
	var seen string
	cmds := map[string]fakeCommand{"./deploy.sh": func(env map[string]string, _ io.Writer) int { seen = env[EnvResultsDir]; return 0 }}
	p, _, dir := newTestProvider(t, config.Exec{Paths: []string{"app"}, Deploy: "./deploy.sh"}, fakeGit{}, cmds)
	rd := filepath.Join(t.TempDir(), "out")
	p.Bind(Workspace{Dir: dir, ResultsDir: rd, Stdout: io.Discard, Stderr: io.Discard})
	if _, err := p.Deploy(context.Background(), session4("qa", "a", "qa"), writeChangeSet(t, dir), providers.DeployOptions{Dir: dir}); err != nil {
		t.Fatal(err)
	}
	if seen != rd {
		t.Fatalf("results dir = %q, want %q", seen, rd)
	}
}

func TestMissingCommandsAreErrors(t *testing.T) {
	p, _, dir := newTestProvider(t, config.Exec{Paths: []string{"app"}}, fakeGit{}, nil)
	s := session4("qa", "a", "qa")
	if _, err := p.Validate(context.Background(), s, providers.ChangeSet{}, providers.ValidateOptions{Dir: dir}); err == nil || !strings.Contains(err.Error(), "exec.validate is not set") {
		t.Fatalf("validate err = %v", err)
	}
	if _, err := p.Deploy(context.Background(), s, providers.ChangeSet{}, providers.DeployOptions{Dir: dir}); err == nil || !strings.Contains(err.Error(), "exec.deploy is not set") {
		t.Fatalf("deploy err = %v", err)
	}
	if _, err := p.Verify(context.Background(), s, providers.VerifyRequest{Dir: dir}); err == nil || !strings.Contains(err.Error(), "exec.verify is not set") {
		t.Fatalf("verify err = %v", err)
	}
}

func TestCommandThatCannotStartIsAnError(t *testing.T) {
	p, _, dir := newTestProvider(t, config.Exec{Paths: []string{"app"}, Validate: "./missing.sh"}, fakeGit{}, nil)
	if _, err := p.Validate(context.Background(), session4("qa", "a", "qa"), providers.ChangeSet{}, providers.ValidateOptions{Dir: dir}); err == nil {
		t.Fatal("want an error")
	}
}

func TestValidate(t *testing.T) {
	cmds := map[string]fakeCommand{"./validate.sh": func(env map[string]string, _ io.Writer) int {
		_ = os.WriteFile(filepath.Join(env[EnvResultsDir], "exec-validate.json"), []byte(`{"schema":"glassops.engine-result/v1","step":"validate","passed":true,"id":"val-1"}`), 0o644)
		return 0
	}}
	p, _, dir := newTestProvider(t, config.Exec{Paths: []string{"app"}, Validate: "./validate.sh"}, fakeGit{}, cmds)
	out, err := p.Validate(context.Background(), session4("qa", "a", "qa"), providers.ChangeSet{}, providers.ValidateOptions{Dir: dir})
	if err != nil || !out.Passed || out.ID != "val-1" {
		t.Fatalf("out = %+v, err = %v", out, err)
	}
}

func TestVerify(t *testing.T) {
	for _, code := range []int{0, 4} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			var seen string
			cmds := map[string]fakeCommand{"./verify.sh": func(env map[string]string, _ io.Writer) int { seen = env[EnvDeployID]; return code }}
			p, _, dir := newTestProvider(t, config.Exec{Paths: []string{"app"}, Verify: "./verify.sh"}, fakeGit{}, cmds)
			req := providers.VerifyRequest{Dir: dir, Plan: providers.ChangeSet{Dir: ".glassops/plan"}, Deploy: providers.DeployRecord{ID: "dep-7"}}
			out, err := p.Verify(context.Background(), session4("qa", "a", "qa"), req)
			if err != nil {
				t.Fatal(err)
			}
			if seen != "dep-7" {
				t.Fatalf("%s = %q", EnvDeployID, seen)
			}
			if out.Passed != (code == 0) || out.Checks[0].Name != "verify.exec" {
				t.Fatalf("out = %+v", out)
			}
			if code != 0 && (len(out.Checks) != 2 || out.Checks[1].Status != providers.VerifyFail) {
				t.Fatalf("checks = %+v", out.Checks)
			}
		})
	}
}

func TestAuthenticateTouchesNothing(t *testing.T) {
	p, sh, _ := newTestProvider(t, config.Exec{Paths: []string{"app"}}, fakeGit{}, nil)
	s, err := p.Authenticate(context.Background(), providers.Target{Name: "qa", ID: "a"})
	if err != nil || s.Target().Name != "qa" || s.Close() != nil || len(sh.runs) != 0 {
		t.Fatalf("s = %v, err = %v", s, err)
	}
}
