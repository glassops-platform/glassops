package salesforce

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/providers"
)

const testProject = `{"packageDirectories":[{"path":"force-app","default":true},{"path":"extra"}],"sourceApiVersion":"65.0"}`

// planProject returns a project dir holding sfdx-project.json.
func planProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ProjectFile), []byte(testProject), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// copyTree copies the files under src into dst.
func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// sgdFake answers an sgd call the way the real tool does: it writes the manifests that
// testdata/sgd holds for the scenario into --output-dir and prints the JSON fixture.
func sgdFake(t *testing.T, scenario, jsonFixture string, calls *[]Cmd) Exec {
	t.Helper()
	return func(_ context.Context, c Cmd) (Output, error) {
		*calls = append(*calls, c)
		if jsonFixture != "sgd-delta-bad-ref.json" {
			copyTree(t, filepath.Join("testdata", "sgd", scenario), argValue(c.Args, "--output-dir"))
		}
		return Output{Stdout: fixture(t, jsonFixture)}, nil
	}
}

func argValue(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func planProvider(exec Exec) *Provider {
	return NewWithOptions(Options{Exec: exec})
}

func planReq(dir string) providers.PlanRequest {
	return providers.PlanRequest{Dir: dir, From: "aaa", To: "bbb", OutDir: filepath.Join(dir, ".glassops", "plan")}
}

func TestPlanDeltaParsesManifests(t *testing.T) {
	dir := planProject(t)
	var calls []Cmd
	p := planProvider(sgdFake(t, "delta", "sgd-delta.json", &calls))

	cs, err := p.Plan(context.Background(), planReq(dir))
	if err != nil {
		t.Fatal(err)
	}
	if cs.Schema != providers.PlanSchema || cs.Provider != "salesforce" || cs.From != "aaa" || cs.To != "bbb" || cs.Full {
		t.Fatalf("header = %+v", cs)
	}
	if cs.Dir != ".glassops/plan" || cs.Manifest != ".glassops/plan/package/package.xml" {
		t.Fatalf("paths: dir %q manifest %q", cs.Dir, cs.Manifest)
	}
	if want := []string{".glassops/plan/destructiveChanges/destructiveChanges.xml"}; !reflect.DeepEqual(cs.Destructive, want) {
		t.Fatalf("Destructive = %v, want %v", cs.Destructive, want)
	}
	if cs.Empty {
		t.Fatal("plan is empty, want not empty")
	}
	wantFacts := map[string]any{
		"salesforce.counts":            map[string]int{"ApexClass": 2, "PermissionSet": 1},
		"salesforce.destructive_count": 1,
		"salesforce.apex_classes":      []string{"Greeter", "GreeterTest"},
		"salesforce.permission_sets":   []string{"Greeter_User"},
		"salesforce.profiles":          []string{},
	}
	if !reflect.DeepEqual(cs.Facts, wantFacts) {
		t.Fatalf("Facts = %#v\nwant %#v", cs.Facts, wantFacts)
	}

	if len(calls) != 1 {
		t.Fatalf("%d calls, want 1", len(calls))
	}
	c := calls[0]
	if c.Name != "sf" || c.Dir != dir {
		t.Fatalf("call = %s in %s", c.Name, c.Dir)
	}
	for _, want := range []string{"sgd source delta", "--from aaa", "--to bbb", "--source-dir force-app", "--source-dir extra", "--json"} {
		if !strings.Contains(strings.Join(c.Args, " "), want) {
			t.Fatalf("args %v lack %q", c.Args, want)
		}
	}
	if strings.Contains(strings.Join(c.Args, " "), "--ignore-file") {
		t.Fatalf("args %v pass --ignore-file without a .forceignore", c.Args)
	}
}

func TestPlanDeltaPassesForceignore(t *testing.T) {
	dir := planProject(t)
	if err := os.WriteFile(filepath.Join(dir, ".forceignore"), []byte("**/x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var calls []Cmd
	if _, err := planProvider(sgdFake(t, "delta", "sgd-delta.json", &calls)).Plan(context.Background(), planReq(dir)); err != nil {
		t.Fatal(err)
	}
	if got := argValue(calls[0].Args, "--ignore-file"); got != ".forceignore" {
		t.Fatalf("--ignore-file = %q, want .forceignore (args %v)", got, calls[0].Args)
	}
}

func TestPlanEmpty(t *testing.T) {
	dir := planProject(t)
	var calls []Cmd
	cs, err := planProvider(sgdFake(t, "empty", "sgd-delta-empty.json", &calls)).Plan(context.Background(), planReq(dir))
	if err != nil {
		t.Fatal(err)
	}
	if !cs.Empty || len(cs.Destructive) != 0 || cs.Facts["salesforce.destructive_count"] != 0 {
		t.Fatalf("empty plan = %+v", cs)
	}
	if cs.Destructive == nil {
		t.Fatal("Destructive is nil, want an empty list so the JSON is []")
	}
}

func TestPlanSGDFailureIsAnError(t *testing.T) {
	dir := planProject(t)
	var calls []Cmd
	_, err := planProvider(sgdFake(t, "", "sgd-delta-bad-ref.json", &calls)).Plan(context.Background(), planReq(dir))
	if err == nil || !strings.Contains(err.Error(), "fetch depth") {
		t.Fatalf("error = %v, want one naming the fetch depth", err)
	}
}

func TestPlanRefusesOutDirOutsideGlassops(t *testing.T) {
	dir := planProject(t)
	sentinel := filepath.Join(dir, "keep.txt")
	if err := os.WriteFile(sentinel, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var calls []Cmd
	p := planProvider(sgdFake(t, "delta", "sgd-delta.json", &calls))
	for name, out := range map[string]string{
		"project root":    dir,
		"glassops itself": filepath.Join(dir, ".glassops"),
		"sibling":         filepath.Join(dir, "force-app"),
		"escape":          filepath.Join(dir, ".glassops", "..", "keep.txt"),
		"outside":         filepath.Join(filepath.Dir(dir), "elsewhere"),
		"empty":           "",
	} {
		t.Run(name, func(t *testing.T) {
			req := planReq(dir)
			req.OutDir = out
			if _, err := p.Plan(context.Background(), req); err == nil {
				t.Fatal("expected a refusal")
			}
		})
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("sentinel gone: %v", err)
	}
	if len(calls) != 0 {
		t.Fatalf("sgd ran %d times for a refused plan", len(calls))
	}
}

func TestPlanClearsOutDirFirst(t *testing.T) {
	dir := planProject(t)
	stale := filepath.Join(dir, ".glassops", "plan", "stale", "old.xml")
	if err := os.MkdirAll(filepath.Dir(stale), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stale, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(dir, ".glassops", "results", "doctor.json")
	if err := os.MkdirAll(filepath.Dir(other), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(other, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	var calls []Cmd
	if _, err := planProvider(sgdFake(t, "delta", "sgd-delta.json", &calls)).Plan(context.Background(), planReq(dir)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale file survived: %v", err)
	}
	if _, err := os.Stat(other); err != nil {
		t.Fatalf("a sibling under .glassops was removed: %v", err)
	}
}

func TestPlanDigest(t *testing.T) {
	dir := planProject(t)
	var calls []Cmd
	p := planProvider(sgdFake(t, "delta", "sgd-delta.json", &calls))
	cs, err := p.Plan(context.Background(), planReq(dir))
	if err != nil {
		t.Fatal(err)
	}

	// The documented formula: for each file under the plan dir in sorted relative-path
	// order, the path, a newline and the content.
	h := sha256.New()
	for _, rel := range []string{
		".glassops/plan/destructiveChanges/destructiveChanges.xml",
		".glassops/plan/destructiveChanges/package.xml",
		".glassops/plan/package/package.xml",
	} {
		data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		h.Write([]byte(rel + "\n"))
		h.Write(data)
	}
	if want := hex.EncodeToString(h.Sum(nil)); cs.SHA256 != want {
		t.Fatalf("SHA256 = %s, want %s", cs.SHA256, want)
	}

	again, err := p.Plan(context.Background(), planReq(dir))
	if err != nil {
		t.Fatal(err)
	}
	if again.SHA256 != cs.SHA256 {
		t.Fatal("the digest changed between identical plans")
	}

	empty, err := planProvider(sgdFake(t, "empty", "sgd-delta-empty.json", &calls)).Plan(context.Background(), planReq(dir))
	if err != nil {
		t.Fatal(err)
	}
	if empty.SHA256 == cs.SHA256 {
		t.Fatal("different manifests gave the same digest")
	}
}

func TestPlanFullMode(t *testing.T) {
	dir := planProject(t)
	var calls []Cmd
	exec := func(_ context.Context, c Cmd) (Output, error) {
		calls = append(calls, c)
		out := argValue(c.Args, "--output-dir")
		data, err := os.ReadFile(filepath.Join("testdata", "sgd", "delta", "package", "package.xml"))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, "package.xml"), data, 0o644); err != nil {
			t.Fatal(err)
		}
		return Output{Stdout: []byte(`{"status":0,"result":{"path":"x/package.xml","name":"package.xml"},"warnings":[]}`)}, nil
	}
	req := planReq(dir)
	req.Full, req.From, req.To = true, "", ""
	cs, err := planProvider(exec).Plan(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !cs.Full || cs.From != "" || cs.Empty || len(cs.Destructive) != 0 || cs.Manifest != ".glassops/plan/package/package.xml" {
		t.Fatalf("full plan = %+v", cs)
	}
	args := strings.Join(calls[0].Args, " ")
	for _, want := range []string{"project generate manifest", "--source-dir force-app", "--source-dir extra", "--name package", "--json"} {
		if !strings.Contains(args, want) {
			t.Fatalf("args %q lack %q", args, want)
		}
	}
	if strings.Contains(args, "sgd") {
		t.Fatalf("full mode ran sgd: %q", args)
	}
}

func TestPlanNeedsPackageDirectories(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ProjectFile), []byte(`{"packageDirectories":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var calls []Cmd
	_, err := planProvider(sgdFake(t, "delta", "sgd-delta.json", &calls)).Plan(context.Background(), planReq(dir))
	if err == nil || !strings.Contains(err.Error(), "no package directories") {
		t.Fatalf("error = %v", err)
	}
}

// sgdAvailable reports whether sf can run sgd here. Locally the plugin link can dangle.
func sgdAvailable() bool {
	if _, err := exec.LookPath("sf"); err != nil {
		return false
	}
	if _, err := exec.LookPath("git"); err != nil {
		return false
	}
	return exec.Command("sf", "sgd", "source", "delta", "--help").Run() == nil
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// TestPlanRealSGD builds a repo from testdata/sfdx and runs the real sgd over its two commits.
func TestPlanRealSGD(t *testing.T) {
	if !sgdAvailable() {
		t.Skip("sf with the sfdx-git-delta plugin is not usable here (run this in the glassops image)")
	}
	root := filepath.Join("..", "..", "..", "testdata", "sfdx")
	repo := t.TempDir()
	git(t, repo, "init", "-q", "-b", "main")
	git(t, repo, "config", "user.name", "GlassOps Test")
	git(t, repo, "config", "user.email", "test@example.invalid")
	git(t, repo, "config", "commit.gpgsign", "false")
	copyTree(t, filepath.Join(root, "commit1"), repo)
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-q", "-m", "one")
	copyTree(t, filepath.Join(root, "commit2"), repo)
	removals, err := os.ReadFile(filepath.Join(root, "commit2.remove"))
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range strings.Fields(string(removals)) {
		if err := os.Remove(filepath.Join(repo, filepath.FromSlash(rel))); err != nil {
			t.Fatal(err)
		}
	}
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-q", "-m", "two")

	req := providers.PlanRequest{
		Dir:    repo,
		From:   git(t, repo, "rev-parse", "HEAD~1"),
		To:     git(t, repo, "rev-parse", "HEAD"),
		OutDir: filepath.Join(repo, ".glassops", "plan"),
	}
	cs, err := New().Plan(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if cs.Empty {
		t.Fatal("plan is empty")
	}
	if want := []string{"Greeter", "GreeterTest"}; !reflect.DeepEqual(cs.Facts["salesforce.apex_classes"], want) {
		t.Fatalf("apex_classes = %v, want %v", cs.Facts["salesforce.apex_classes"], want)
	}
	if want := []string{"Greeter_User"}; !reflect.DeepEqual(cs.Facts["salesforce.permission_sets"], want) {
		t.Fatalf("permission_sets = %v, want %v", cs.Facts["salesforce.permission_sets"], want)
	}
	if cs.Facts["salesforce.destructive_count"] != 1 || len(cs.Destructive) != 1 {
		t.Fatalf("destructive = %v, facts %v", cs.Destructive, cs.Facts)
	}
	data, err := os.ReadFile(filepath.Join(repo, filepath.FromSlash(cs.Destructive[0])))
	if err != nil || !strings.Contains(string(data), "Account.Tier__c") {
		t.Fatalf("destructive manifest = %q, %v", data, err)
	}

	// The same project in full mode.
	full := req
	full.Full, full.From, full.To = true, "", ""
	fcs, err := New().Plan(context.Background(), full)
	if err != nil {
		t.Fatal(err)
	}
	if fcs.Empty || len(fcs.Destructive) != 0 {
		t.Fatalf("full plan = %+v", fcs)
	}
}
