package salesforce

import (
	"archive/tar"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/providers"
)

const (
	cmdSeedMinimal = "sf project deploy start --metadata"
	cmdSeedFull    = "sf project deploy start --source-dir"
)

// seedRepo is a synthetic project in a real git repo. The base commit has an Account field
// and a custom object with a field. HEAD deletes both fields. It returns the dir and the base
// commit.
func seedRepo(t *testing.T) (string, string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on the PATH")
	}
	dir := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "core.autocrlf=false"}, args...)...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	write := func(rel, content string) {
		t.Helper()
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git("init", "-q")
	write(ProjectFile, `{"packageDirectories":[{"path":"force-app","default":true}],"sourceApiVersion":"62.0"}`)
	write(".forceignore", "**/jsconfig.json\n")
	write("force-app/main/default/objects/Account/fields/Tier__c.field-meta.xml", "<CustomField/>\n")
	write("force-app/main/default/objects/Widget__c/Widget__c.object-meta.xml", "<CustomObject/>\n")
	write("force-app/main/default/objects/Widget__c/fields/Size__c.field-meta.xml", "<CustomField/>\n")
	write("notes/outside.txt", "not in a package dir\n")
	git("add", "-A")
	git("commit", "-q", "-m", "base")
	base := git("rev-parse", "HEAD")
	for _, rel := range []string{"force-app/main/default/objects/Account/fields/Tier__c.field-meta.xml", "force-app/main/default/objects/Widget__c/fields/Size__c.field-meta.xml"} {
		if err := os.Remove(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Fatal(err)
		}
	}
	git("add", "-A")
	git("commit", "-q", "-m", "delete fields")
	return dir, base
}

// writeDestructive writes a destructive manifest with members under the plan dir and returns
// a change set naming it.
func writeDestructive(t *testing.T, dir string, members map[string][]string) providers.ChangeSet {
	t.Helper()
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n" + `<Package xmlns="http://soap.sforce.com/2006/04/metadata">` + "\n")
	for typ, names := range members {
		b.WriteString("  <types>\n")
		for _, n := range names {
			b.WriteString("    <members>" + n + "</members>\n")
		}
		b.WriteString("    <name>" + typ + "</name>\n  </types>\n")
	}
	b.WriteString("  <version>62.0</version>\n</Package>\n")
	rel := ".glassops/plan/" + planDestructive
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return providers.ChangeSet{Schema: providers.PlanSchema, Destructive: []string{rel}}
}

// seedExec runs git for real and answers sf from replies. It records, for every sf deploy,
// the files the temp project held at that moment.
type seedExec struct {
	*seqExec
	projectFiles [][]string
	projectDirs  []string
}

func newSeedExec(replies map[string][]fakeReply) *seedExec {
	s := &seedExec{seqExec: &seqExec{replies: replies}}
	return s
}

func (s *seedExec) run(ctx context.Context, c Cmd) (Output, error) {
	if c.Name == "git" {
		return osExec(ctx, c)
	}
	if c.Name == "sf" && len(c.Args) > 2 && c.Args[1] == "deploy" {
		var files []string
		_ = filepath.Walk(c.Dir, func(p string, fi os.FileInfo, err error) error {
			if err == nil && !fi.IsDir() {
				files = append(files, relSlash(c.Dir, p))
			}
			return nil
		})
		s.projectFiles = append(s.projectFiles, files)
		s.projectDirs = append(s.projectDirs, c.Dir)
	}
	return s.seqExec.run(ctx, c)
}

func seedProvider(t *testing.T, s *seedExec) (*Provider, *session) {
	t.Helper()
	p := NewWithOptions(Options{Exec: s.run, TempDir: t.TempDir()})
	return p, newTestSession(t, providers.Target{Name: "ephemeral", Tier: "ephemeral"})
}

func TestSeedMinimalFieldOnStandardObject(t *testing.T) {
	dir, base := seedRepo(t)
	cs := writeDestructive(t, dir, map[string][]string{"CustomField": {"Account.Tier__c"}})
	s := newSeedExec(map[string][]fakeReply{cmdSeedMinimal: {{Stdout: fixture(t, "deploy-start-seed-field.json")}}})
	p, sess := seedProvider(t, s)

	res, err := p.Seed(context.Background(), sess, providers.SeedRequest{Dir: dir, BaseRef: base, Plan: cs, Mode: config.SeedMinimal})
	if err != nil {
		t.Fatal(err)
	}
	if res.Mode != "minimal" || res.Fallback || !reflect.DeepEqual(res.Members, []string{"CustomField:Account.Tier__c"}) {
		t.Fatalf("result %+v", res)
	}
	if res.Facts["salesforce.seed_deploy_id"] != "0Af000000000008AAA" || res.Facts["salesforce.seed_components"] != 1 {
		t.Fatalf("facts %v", res.Facts)
	}
	deploys := s.commands("sf project deploy")
	if len(deploys) != 1 {
		t.Fatalf("deploys %v", deploys)
	}
	want := "sf project deploy start --metadata CustomField:Account.Tier__c --test-level NoTestRun --ignore-conflicts --target-org glassops-target --wait 60 --json"
	if deploys[0] != want {
		t.Fatalf("deploy\n got %s\nwant %s", deploys[0], want)
	}
	// The temp project holds the base ref's project file, .forceignore and package dirs, and
	// nothing outside them. The deleted field is there, because it comes from the base ref.
	files := strings.Join(s.projectFiles[0], " ")
	for _, w := range []string{ProjectFile, ".forceignore", "force-app/main/default/objects/Account/fields/Tier__c.field-meta.xml"} {
		if !strings.Contains(files, w) {
			t.Fatalf("temp project %v lacks %s", s.projectFiles[0], w)
		}
	}
	if strings.Contains(files, "notes/") || strings.Contains(files, ".glassops") {
		t.Fatalf("temp project holds files outside the package dirs: %v", s.projectFiles[0])
	}
	if _, err := os.Stat(s.projectDirs[0]); !os.IsNotExist(err) {
		t.Fatalf("temp project %s was not removed", s.projectDirs[0])
	}
}

func TestSeedMinimalFieldOnCustomObjectAddsTheObject(t *testing.T) {
	dir, base := seedRepo(t)
	cs := writeDestructive(t, dir, map[string][]string{"CustomField": {"Widget__c.Size__c"}})
	s := newSeedExec(map[string][]fakeReply{cmdSeedMinimal: {{Stdout: fixture(t, "deploy-start-seed-field.json")}}})
	p, sess := seedProvider(t, s)

	res, err := p.Seed(context.Background(), sess, providers.SeedRequest{Dir: dir, BaseRef: base, Plan: cs, Mode: config.SeedMinimal})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"CustomField:Widget__c.Size__c", "CustomObject:Widget__c"}
	if !reflect.DeepEqual(res.Members, want) {
		t.Fatalf("members %v, want %v", res.Members, want)
	}
	d := s.commands("sf project deploy")[0]
	if !strings.Contains(d, "--metadata CustomField:Widget__c.Size__c --metadata CustomObject:Widget__c ") {
		t.Fatalf("deploy %s", d)
	}
}

func TestSeedFallsBackToFullOnMissingParent(t *testing.T) {
	dir, base := seedRepo(t)
	cs := writeDestructive(t, dir, map[string][]string{"CustomField": {"Widget__c.Size__c"}})
	s := newSeedExec(map[string][]fakeReply{
		cmdSeedMinimal: {{Stdout: fixture(t, "deploy-start-seed-missing-parent.json"), ExitCode: 1}},
		cmdSeedFull:    {{Stdout: fixture(t, "deploy-start-seed-field.json")}},
	})
	p, sess := seedProvider(t, s)

	res, err := p.Seed(context.Background(), sess, providers.SeedRequest{Dir: dir, BaseRef: base, Plan: cs, Mode: config.SeedMinimal})
	if err != nil {
		t.Fatal(err)
	}
	if res.Mode != "full" || !res.Fallback || res.Facts["salesforce.seed_missing"] != "Widget__c" {
		t.Fatalf("result %+v", res)
	}
	// Full mode reports what the deploy reported, without package.xml.
	if !reflect.DeepEqual(res.Members, []string{"CustomField:Account.Tier__c"}) {
		t.Fatalf("members %v", res.Members)
	}
	deploys := s.commands("sf project deploy")
	if len(deploys) != 2 || !strings.HasPrefix(deploys[1], "sf project deploy start --source-dir force-app --test-level NoTestRun --ignore-conflicts") {
		t.Fatalf("deploys %v", deploys)
	}
}

func TestSeedFullMode(t *testing.T) {
	dir, base := seedRepo(t)
	cs := writeDestructive(t, dir, map[string][]string{"CustomField": {"Account.Tier__c"}})
	s := newSeedExec(map[string][]fakeReply{cmdSeedFull: {{Stdout: fixture(t, "deploy-start-seed-field.json")}}})
	p, sess := seedProvider(t, s)
	res, err := p.Seed(context.Background(), sess, providers.SeedRequest{Dir: dir, BaseRef: base, Plan: cs, Mode: config.SeedFull})
	if err != nil {
		t.Fatal(err)
	}
	if res.Mode != "full" || res.Fallback || len(s.commands(cmdSeedMinimal)) != 0 {
		t.Fatalf("result %+v, deploys %v", res, s.commands("sf project deploy"))
	}
}

func TestSeedFailures(t *testing.T) {
	cases := []struct {
		name    string
		replies map[string][]fakeReply
		want    string
		deploys int
	}{
		{"compile error does not fall back", map[string][]fakeReply{
			cmdSeedMinimal: {{Stdout: mustFixture("deploy-start-compile-fail.json"), ExitCode: 1}}}, "seed deploy Failed", 1},
		{"full seed after the fallback fails too", map[string][]fakeReply{
			cmdSeedMinimal: {{Stdout: mustFixture("deploy-start-seed-missing-parent.json"), ExitCode: 1}},
			cmdSeedFull:    {{Stdout: mustFixture("deploy-start-compile-fail.json"), ExitCode: 1}}}, "needed Widget__c, and the full seed failed", 2},
		{"sf error with no result", map[string][]fakeReply{
			cmdSeedMinimal: {{Stdout: []byte(`{"status":1,"name":"NoOrgFound","message":"No authorization information found"}`), ExitCode: 1}}}, "NoOrgFound", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir, base := seedRepo(t)
			cs := writeDestructive(t, dir, map[string][]string{"CustomField": {"Widget__c.Size__c"}})
			s := newSeedExec(tc.replies)
			p, sess := seedProvider(t, s)
			_, err := p.Seed(context.Background(), sess, providers.SeedRequest{Dir: dir, BaseRef: base, Plan: cs, Mode: config.SeedMinimal})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err %v, want %q", err, tc.want)
			}
			if n := len(s.commands("sf project deploy")); n != tc.deploys {
				t.Fatalf("deploys %d, want %d", n, tc.deploys)
			}
		})
	}
}

func TestSeedDoesNothing(t *testing.T) {
	dir, base := seedRepo(t)
	withDest := writeDestructive(t, dir, map[string][]string{"CustomField": {"Account.Tier__c"}})
	cases := []struct {
		name string
		req  providers.SeedRequest
	}{
		{"off", providers.SeedRequest{Dir: dir, BaseRef: base, Plan: withDest, Mode: config.SeedOff}},
		{"no destructive manifest", providers.SeedRequest{Dir: dir, BaseRef: base, Plan: providers.ChangeSet{Destructive: []string{}}, Mode: config.SeedMinimal}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newSeedExec(map[string][]fakeReply{})
			p, sess := seedProvider(t, s)
			res, err := p.Seed(context.Background(), sess, tc.req)
			if err != nil || res.Mode != "none" || len(res.Members) != 0 || len(s.calls) != 0 {
				t.Fatalf("res %+v, err %v, calls %v", res, err, s.commands(""))
			}
		})
	}
}

func TestSeedRefusesBadInput(t *testing.T) {
	dir, base := seedRepo(t)
	cs := writeDestructive(t, dir, map[string][]string{"CustomField": {"Account.Tier__c"}})
	cases := []struct {
		name string
		req  providers.SeedRequest
		want string
	}{
		{"unknown mode", providers.SeedRequest{Dir: dir, BaseRef: base, Plan: cs, Mode: "some"}, "not minimal, full or off"},
		{"no base ref", providers.SeedRequest{Dir: dir, Plan: cs, Mode: config.SeedMinimal}, "needs the base ref"},
		{"base ref that is a flag", providers.SeedRequest{Dir: dir, BaseRef: "--output=/tmp/x", Plan: cs, Mode: config.SeedMinimal}, "needs the base ref"},
		{"base ref that does not exist", providers.SeedRequest{Dir: dir, BaseRef: "0000000000000000000000000000000000000000", Plan: cs, Mode: config.SeedMinimal}, "at the base ref"},
		{"missing destructive manifest", providers.SeedRequest{Dir: dir, BaseRef: base, Plan: providers.ChangeSet{Destructive: []string{".glassops/plan/nope.xml"}}, Mode: config.SeedMinimal}, "read manifest"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newSeedExec(map[string][]fakeReply{})
			p, sess := seedProvider(t, s)
			if _, err := p.Seed(context.Background(), sess, tc.req); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err %v, want %q", err, tc.want)
			}
			if n := len(s.commands("sf")); n != 0 {
				t.Fatalf("sf ran %d times", n)
			}
		})
	}
}

func TestWithParents(t *testing.T) {
	cases := []struct {
		in, want []string
	}{
		{[]string{"CustomField:Account.Tier__c"}, []string{"CustomField:Account.Tier__c"}},
		{[]string{"CustomField:Widget__c.Size__c"}, []string{"CustomField:Widget__c.Size__c", "CustomObject:Widget__c"}},
		{[]string{"CustomField:Setting__mdt.Value__c", "RecordType:Event__e.X"}, []string{"CustomField:Setting__mdt.Value__c", "CustomObject:Event__e", "CustomObject:Setting__mdt", "RecordType:Event__e.X"}},
		{[]string{"ValidationRule:Widget__c.Rule", "ListView:Widget__c.All", "FieldSet:Widget__c.F", "CompactLayout:Widget__c.C", "WebLink:Widget__c.W", "BusinessProcess:Widget__c.B"},
			[]string{"BusinessProcess:Widget__c.B", "CompactLayout:Widget__c.C", "CustomObject:Widget__c", "FieldSet:Widget__c.F", "ListView:Widget__c.All", "ValidationRule:Widget__c.Rule", "WebLink:Widget__c.W"}},
		{[]string{"CustomObject:Widget__c", "CustomField:Widget__c.Size__c"}, []string{"CustomField:Widget__c.Size__c", "CustomObject:Widget__c"}},
		{[]string{"ApexClass:Widget__c.Thing", "Layout:Widget__c-Layout"}, []string{"ApexClass:Widget__c.Thing", "Layout:Widget__c-Layout"}},
		{[]string{"CustomField:ns__Widget__c.ns__Size__c"}, []string{"CustomField:ns__Widget__c.ns__Size__c", "CustomObject:ns__Widget__c"}},
	}
	for _, tc := range cases {
		if got := withParents(tc.in); !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("withParents(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestExtractTarRefusesEscapes(t *testing.T) {
	for _, name := range []string{"../evil.txt", "/abs.txt", "a/../../evil.txt", "C:/evil.txt"} {
		var buf bytes.Buffer
		tw := tar.NewWriter(&buf)
		_ = tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: 1, Typeflag: tar.TypeReg})
		_, _ = tw.Write([]byte("x"))
		_ = tw.Close()
		root := t.TempDir()
		if err := extractTar(buf.Bytes(), root); err == nil {
			t.Fatalf("%q was extracted", name)
		}
	}
}

func TestExtractTarSkipsLinks(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	_ = tw.WriteHeader(&tar.Header{Name: "link", Linkname: "/etc/passwd", Typeflag: tar.TypeSymlink})
	_ = tw.WriteHeader(&tar.Header{Name: "dir/file.txt", Mode: 0o644, Size: 2, Typeflag: tar.TypeReg})
	_, _ = tw.Write([]byte("ok"))
	_ = tw.Close()
	root := t.TempDir()
	if err := extractTar(buf.Bytes(), root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(root, "link")); !os.IsNotExist(err) {
		t.Fatalf("link was created: %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(root, "dir", "file.txt")); err != nil || string(data) != "ok" {
		t.Fatalf("file %q %v", data, err)
	}
}
