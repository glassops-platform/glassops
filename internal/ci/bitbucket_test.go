package ci

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// The variables below are from docs, not captured: Atlassian's "Variables and secrets" page
// (predefined variables). No Bitbucket runner has been observed, so everything here is
// [unverified] against a real pipeline. The values are synthetic.

var bitbucketPush = map[string]string{
	"BITBUCKET_BUILD_NUMBER":        "42",
	"BITBUCKET_COMMIT":              "0123456789abcdef0123456789abcdef01234567",
	"BITBUCKET_BRANCH":              "main",
	"BITBUCKET_REPO_FULL_NAME":      "acme/widgets",
	"BITBUCKET_STEP_TRIGGERER_UUID": "{11111111-2222-3333-4444-555555555555}",
}

var bitbucketPR = map[string]string{
	"BITBUCKET_BUILD_NUMBER":          "43",
	"BITBUCKET_COMMIT":                "fedcba9876543210fedcba9876543210fedcba98",
	"BITBUCKET_BRANCH":                "feature/x",
	"BITBUCKET_PR_ID":                 "7",
	"BITBUCKET_PR_DESTINATION_BRANCH": "main",
	"BITBUCKET_REPO_FULL_NAME":        "acme/widgets",
	"BITBUCKET_STEP_TRIGGERER_UUID":   "{11111111-2222-3333-4444-555555555555}",
}

var bitbucketTag = map[string]string{
	"BITBUCKET_BUILD_NUMBER":   "44",
	"BITBUCKET_COMMIT":         "0123456789abcdef0123456789abcdef01234567",
	"BITBUCKET_TAG":            "v1.2.3",
	"BITBUCKET_REPO_FULL_NAME": "acme/widgets",
}

func TestBitbucketContext(t *testing.T) {
	uuid := "{11111111-2222-3333-4444-555555555555}"
	cases := []struct {
		name string
		env  map[string]string
		want Context
	}{
		{"push", bitbucketPush, Context{
			CI: "bitbucket", Repo: "acme/widgets", Ref: "refs/heads/main", SHA: "0123456789abcdef0123456789abcdef01234567",
			Event: "push", Actor: uuid, TriggeringActor: uuid, RunID: "42",
		}},
		{"pull request", bitbucketPR, Context{
			CI: "bitbucket", Repo: "acme/widgets", Ref: "refs/heads/feature/x", SHA: "fedcba9876543210fedcba9876543210fedcba98",
			Event: "pull_request", Actor: uuid, TriggeringActor: uuid, RunID: "43", PRNumber: "7", BaseRef: "main",
		}},
		{"tag", bitbucketTag, Context{
			CI: "bitbucket", Repo: "acme/widgets", Ref: "refs/tags/v1.2.3", SHA: "0123456789abcdef0123456789abcdef01234567",
			Event: "push", RunID: "44",
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := newBitbucket(mapEnv(tc.env), &bytes.Buffer{}, "").Context()
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("context\n got %+v\nwant %+v", got, tc.want)
			}
		})
	}
}

func lfOnly(b []byte) string { return string(bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))) }

func TestBitbucketOutputAndSummaryFiles(t *testing.T) {
	dir := t.TempDir()
	b := newBitbucket(mapEnv(bitbucketPush), &bytes.Buffer{}, dir)
	if err := b.SetOutput("status", "pass"); err != nil {
		t.Fatal(err)
	}
	if err := b.SetOutput("result", ".glassops/gate.json"); err != nil {
		t.Fatal(err)
	}
	if err := b.Summary("# one"); err != nil {
		t.Fatal(err)
	}
	if err := b.Summary("# two"); err != nil {
		t.Fatal(err)
	}
	read := func(rel string) string {
		raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		return lfOnly(raw)
	}
	if got := read(DefaultOutputFile); got != "status=pass\nresult=.glassops/gate.json\n" {
		t.Fatalf("outputs = %q", got)
	}
	if got := read(DefaultSummaryFile); got != "# one\n# two\n" {
		t.Fatalf("summary = %q", got)
	}
}

func TestBitbucketFileOverridesAndBadOutputs(t *testing.T) {
	dir := t.TempDir()
	env := map[string]string{OutputFileEnv: "out/o.env", SummaryFileEnv: filepath.Join(dir, "abs", "s.md")}
	b := newBitbucket(mapEnv(env), &bytes.Buffer{}, dir)
	if err := b.SetOutput("a", "1"); err != nil {
		t.Fatal(err)
	}
	if err := b.Summary("x"); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{filepath.Join(dir, "out", "o.env"), filepath.Join(dir, "abs", "s.md")} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected %s: %v", p, err)
		}
	}
	if err := b.SetOutput("bad-name", "1"); err == nil {
		t.Fatal("a name with a hyphen was accepted")
	}
	if err := b.SetOutput("multi", "a\nb"); err == nil {
		t.Fatal("a multi-line value was accepted")
	}
}

func TestBitbucketAnnotateIsAPlainLine(t *testing.T) {
	var out bytes.Buffer
	newBitbucket(mapEnv(nil), &out, "").Annotate(Warning, "careful")
	if out.String() != "warning: careful\n" {
		t.Fatalf("annotate = %q", out.String())
	}
}

// A write that cannot happen is an error the caller sees, never silently dropped.
func TestBitbucketWriteErrorsSurface(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".glassops"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	b := newBitbucket(mapEnv(nil), &bytes.Buffer{}, dir)
	if err := b.SetOutput("a", "1"); err == nil {
		t.Fatal("SetOutput did not fail")
	}
	if err := b.Summary("x"); err == nil {
		t.Fatal("Summary did not fail")
	}
}
