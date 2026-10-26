package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const fullConfig = `version: 1
provider: salesforce
targets:
  devhub:
    id: 00D000000000001
    tier: devhub
    auth: { method: jwt, username: $SF_DEVHUB_USERNAME, client_id: $SF_CLIENT_ID, key: $SF_JWT_KEY }
  qa:
    id: 00D000000000002AAA
    tier: qa
    branch: develop
    auth: { method: sfdx-url, url: $SF_QA_AUTH_URL }
  prod:
    id: $SF_PROD_ORG_ID
    tier: prod
    branch: main
    auth: { method: jwt, username: $SF_PROD_USERNAME, client_id: $SF_CLIENT_ID, key: $SF_JWT_KEY }
policy:
  windows:
    prod: [{ days: [tue, thu], from: "18:00", to: "22:00", tz: America/Chicago }]
  blackouts:
    - { from: 2026-12-20, to: 2027-01-03, tiers: [prod] }
  rules:
    coverage_min: 75
    analyzer_max_severity: 2
    separation_of_duties: { tiers: [prod] }
    destructive_needs_label: { label: approved-destructive, tiers: [qa, prod] }
  actors: ./glassops/actors.yaml
salesforce:
  test_level: { default: RunLocalTests, ephemeral: RunLocalTests }
  delta: { enabled: true, from: auto }
  api_version: auto
pipelines:
  pr:
    - plan
    - step: env delete
      when: always
`

func TestParseFullConfig(t *testing.T) {
	cfg, err := Parse([]byte(fullConfig))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cfg.Targets["qa"].Auth.Method != "sfdx-url" || cfg.Targets["qa"].Branch != "develop" {
		t.Fatalf("qa target: %+v", cfg.Targets["qa"])
	}
	if got := cfg.Policy.Blackouts[0].From; got != "2026-12-20" {
		t.Fatalf("blackout from = %q, want the date as written", got)
	}
	if *cfg.Policy.Rules.CoverageMin != 75 {
		t.Fatalf("coverage_min = %d", *cfg.Policy.Rules.CoverageMin)
	}
	want := []Step{{Run: "plan"}, {Run: "env delete", When: "always"}}
	if !reflect.DeepEqual(cfg.Pipelines["pr"], want) {
		t.Fatalf("pr pipeline = %+v", cfg.Pipelines["pr"])
	}
	refs := EnvRefs(cfg)
	wantRefs := []string{"SF_CLIENT_ID", "SF_DEVHUB_USERNAME", "SF_JWT_KEY", "SF_PROD_ORG_ID", "SF_PROD_USERNAME", "SF_QA_AUTH_URL"}
	if !reflect.DeepEqual(refs, wantRefs) {
		t.Fatalf("EnvRefs = %v", refs)
	}
}

func TestParseRejects(t *testing.T) {
	cases := []struct {
		name string
		yaml string
		want string
	}{
		{"missing version", "provider: salesforce\n", "/"},
		{"wrong version", "version: 2\n", "/version"},
		{"unknown key", "version: 1\nfoo: bar\n", "foo"},
		{"bad org id", "version: 1\ntargets:\n  qa: { id: 12345 }\n", "/targets/qa/id"},
		{"literal secret", "version: 1\ntargets:\n  qa:\n    auth: { method: sfdx-url, url: force://secret }\n", "/targets/qa/auth"},
		{"window without tz", "version: 1\npolicy:\n  windows:\n    prod: [{ days: [mon], from: \"18:00\", to: \"20:00\" }]\n", "tz"},
		{"bad day", "version: 1\npolicy:\n  windows:\n    prod: [{ days: [funday], from: \"18:00\", to: \"20:00\", tz: UTC }]\n", "/policy/windows/prod/0/days/0"},
		{"bad when", "version: 1\npipelines:\n  pr:\n    - { step: plan, when: sometimes }\n", "/pipelines/pr/0"},
		{"two prod tiers", "version: 1\ntargets:\n  a: { tier: prod }\n  b: { tier: prod }\n", "at most one target may have tier prod"},
		{"duplicate org id", "version: 1\ntargets:\n  a: { id: 00D000000000001 }\n  b: { id: 00D000000000001AAA }\n", "same org Id"},
		{"unknown tz", "version: 1\npolicy:\n  windows:\n    prod: [{ days: [mon], from: \"18:00\", to: \"20:00\", tz: Mars/Olympus }]\n", "unknown time zone"},
		{"blackout reversed", "version: 1\npolicy:\n  blackouts: [{ from: 2027-01-03, to: 2026-12-20 }]\n", "is before"},
		{"duplicate key", "version: 1\ntargets:\n  a: {}\n  a: {}\n", "already defined"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse([]byte(tc.yaml))
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not mention %q", err, tc.want)
			}
		})
	}
}

func TestLoadDefaultsWhenMissing(t *testing.T) {
	l, err := Load(t.TempDir(), "", "salesforce")
	if err != nil {
		t.Fatal(err)
	}
	if l.Path != "" || l.Config.Version != 1 || l.Config.Provider != "salesforce" {
		t.Fatalf("defaults: %+v path=%q", l.Config, l.Path)
	}
	if len(l.Config.Pipelines["pr"]) == 0 || len(l.Config.Pipelines["deploy"]) == 0 {
		t.Fatal("default pipelines missing")
	}
}

func TestLoadExplicitMissingFails(t *testing.T) {
	_, err := Load(t.TempDir(), filepath.Join(t.TempDir(), "nope.yaml"), "")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestLoadReadsFileAndFillsProvider(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte("version: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	l, err := Load(dir, "", "salesforce")
	if err != nil {
		t.Fatal(err)
	}
	if l.Path == "" || l.Config.Provider != "salesforce" {
		t.Fatalf("loaded: %+v path=%q", l.Config, l.Path)
	}
}

func TestDefaultConfigPassesSchema(t *testing.T) {
	// The defaults must be expressible as a valid file, because init writes them.
	if problems := Check(Default("salesforce")); len(problems) > 0 {
		t.Fatalf("default config problems: %v", problems)
	}
}
