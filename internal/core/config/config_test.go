package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
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
    separation_of_duties:
      tiers: [prod]
      approval: { environment: glassops-deploy, label: approved-deploy, approvers: [octocat] }
    destructive_needs_label: { label: approved-destructive, tiers: [qa, prod] }
  actors: ./glassops/actors.yaml
  permit_ttl: 90m
salesforce:
  test_level: { default: RunLocalTests, ephemeral: RunLocalTests }
  delta: { enabled: true }
  deploy: { quick: require }
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
	sod := cfg.Policy.Rules.SeparationOfDuties
	if sod == nil || sod.Approval == nil || sod.Approval.Environment != "glassops-deploy" || !reflect.DeepEqual(sod.Approval.Approvers, []string{"octocat"}) {
		t.Fatalf("separation_of_duties = %+v", sod)
	}
	if got := cfg.Policy.TTL(); got != 90*time.Minute {
		t.Fatalf("TTL = %s, want 90m", got)
	}
	if got := cfg.Salesforce.Deploy.QuickMode(); got != QuickRequire {
		t.Fatalf("QuickMode = %q, want require", got)
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
		{"delta from is gone", "version: 1\nsalesforce:\n  delta: { from: auto }\n", "/salesforce/delta"},
		{"bad quick", "version: 1\nsalesforce:\n  deploy: { quick: sometimes }\n", "/salesforce/deploy/quick"},
		{"permit ttl not a duration", "version: 1\npolicy:\n  permit_ttl: soon\n", "is not a duration"},
		{"permit ttl too short", "version: 1\npolicy:\n  permit_ttl: 1m\n", "is outside"},
		{"permit ttl too long", "version: 1\npolicy:\n  permit_ttl: 25h\n", "is outside"},
		{"approval without mechanism", "version: 1\npolicy:\n  rules:\n    separation_of_duties: { tiers: [prod], approval: { approvers: [octocat] } }\n", "set environment or label"},
		{"label without approvers", "version: 1\npolicy:\n  rules:\n    separation_of_duties: { approval: { label: ok-to-deploy } }\n", "needs approvers"},
		{"unknown approval key", "version: 1\npolicy:\n  rules:\n    separation_of_duties: { approval: { team: ops } }\n", "/policy/rules/separation_of_duties/approval"},
		{"exec without run", "version: 1\npipelines:\n  pr:\n    - { step: exec }\n", "/pipelines/pr/0"},
		{"run on a glassops step", "version: 1\npipelines:\n  pr:\n    - { step: plan, run: echo hi }\n", "/pipelines/pr/0"},
		{"devhub not a target", "version: 1\nenvironments:\n  ephemeral: { devhub: hub }\n", "environments.ephemeral.devhub"},
		{"environment without devhub", "version: 1\nenvironments:\n  ephemeral: { ttl_days: 1 }\n", "/environments/ephemeral"},
		{"ttl days zero", "version: 1\ntargets:\n  devhub: {}\nenvironments:\n  ephemeral: { devhub: devhub, ttl_days: 0 }\n", "/environments/ephemeral/ttl_days"},
		{"bad shape", "version: 1\ntargets:\n  devhub: {}\nenvironments:\n  ephemeral: { devhub: devhub, shape: org }\n", "/environments/ephemeral/shape"},
		{"bad seed", "version: 1\nsalesforce:\n  ephemeral: { seed: some }\n", "/salesforce/ephemeral/seed"},
		{"coverage over 100", "version: 1\ntest:\n  coverage: { org_min: 101 }\n", "/test/coverage/org_min"},
		{"coverage below 0", "version: 1\ntest:\n  coverage: { class_min: -1 }\n", "/test/coverage/class_min"},
		{"max severity 0", "version: 1\nanalyze: { max_severity: 0 }\n", "/analyze/max_severity"},
		{"max severity 6", "version: 1\nanalyze: { max_severity: 6 }\n", "/analyze/max_severity"},
		{"script without run", "version: 1\ncheck:\n  scripts: [{ name: lint }]\n", "/check/scripts/0"},
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

func TestDefaultsForNewSettings(t *testing.T) {
	var p Policy
	if got := p.TTL(); got != 2*time.Hour {
		t.Fatalf("default TTL = %s", got)
	}
	var d SalesforceDeploy
	if got := d.QuickMode(); got != QuickPrefer {
		t.Fatalf("default QuickMode = %q", got)
	}
	cfg, err := Parse([]byte("version: 1\npolicy:\n  permit_ttl: 5m\n  rules:\n    separation_of_duties: { approval: { environment: glassops-deploy } }\n"))
	if err != nil {
		t.Fatalf("a 5m TTL and an environment approval should be valid: %v", err)
	}
	if got := cfg.Policy.TTL(); got != 5*time.Minute {
		t.Fatalf("TTL = %s", got)
	}
}

func TestLoadedSHA256(t *testing.T) {
	dir := t.TempDir()
	body := []byte("version: 1\n")
	if err := os.WriteFile(filepath.Join(dir, FileName), body, 0o644); err != nil {
		t.Fatal(err)
	}
	l, err := Load(dir, "", "salesforce")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(l.Raw, body) {
		t.Fatalf("Raw = %q", l.Raw)
	}
	// The SHA-256 of the file bytes "version: 1\n".
	const want = "09bfcc6a14b83e2192b8673677725c84883ee9cd0c70e45c9ec09daa8f2b2847"
	if got := l.SHA256(); got != want {
		t.Fatalf("SHA256 = %s, want %s", got, want)
	}

	d1, err := Load(t.TempDir(), "", "salesforce")
	if err != nil {
		t.Fatal(err)
	}
	d2, err := Load(t.TempDir(), "", "salesforce")
	if err != nil {
		t.Fatal(err)
	}
	if d1.Raw != nil || d1.SHA256() != d2.SHA256() {
		t.Fatal("the default config hash is not stable")
	}
	if d1.SHA256() == want {
		t.Fatal("the default and file hashes collide")
	}
	other, err := Load(t.TempDir(), "", "other")
	if err != nil {
		t.Fatal(err)
	}
	if other.SHA256() == d1.SHA256() {
		t.Fatal("the default hash ignores the config")
	}

	crlf := &Loaded{Raw: []byte("version: 1\r\n")}
	if got := crlf.SHA256(); got != want {
		t.Fatalf("CRLF SHA256 = %s, want the LF hash %s", got, want)
	}
}

const phase4Config = `version: 1
targets:
  devhub: { id: 00D000000000001, tier: devhub }
environments:
  ephemeral:
    devhub: devhub
    definition: config/project-scratch-def.json
    ttl_days: 2
    shape: $SHAPE_ORG_ID
    prefix: ci
salesforce:
  ephemeral:
    seed: full
test:
  level: { default: RunLocalTests }
  coverage: { org_min: 75, class_min: 70 }
check:
  scripts:
    - name: lint
      run: npm run lint
analyze:
  enabled: true
  rule_selector: Security
  max_severity: 2
pipelines:
  pr:
    - plan
    - step: exec
      run: ./scripts/publish.sh
      when: success
`

func TestParsePhase4Config(t *testing.T) {
	cfg, err := Parse([]byte(phase4Config))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	e := cfg.Environments["ephemeral"]
	want := Environment{DevHub: "devhub", Definition: "config/project-scratch-def.json", TTLDays: 2, Shape: "$SHAPE_ORG_ID", Prefix: "ci"}
	if e != want || e.Days() != 2 || e.NamePrefix() != "ci" {
		t.Fatalf("environment = %+v", e)
	}
	if got := cfg.Salesforce.Ephemeral.SeedMode(); got != SeedFull {
		t.Fatalf("SeedMode = %q", got)
	}
	if cfg.Test.Level["default"] != "RunLocalTests" || *cfg.Test.Coverage.OrgMin != 75 || *cfg.Test.Coverage.ClassMin != 70 {
		t.Fatalf("test = %+v", cfg.Test)
	}
	if !reflect.DeepEqual(cfg.Check.Scripts, []Script{{Name: "lint", Run: "npm run lint"}}) {
		t.Fatalf("check = %+v", cfg.Check)
	}
	if !*cfg.Analyze.Enabled || cfg.Analyze.Selector() != "Security" {
		t.Fatalf("analyze = %+v", cfg.Analyze)
	}
	if sev, ok := cfg.AnalyzerMaxSeverity(); !ok || sev != 2 {
		t.Fatalf("AnalyzerMaxSeverity = %d, %v", sev, ok)
	}
	steps := []Step{{Run: "plan"}, {Run: ExecStep, Exec: "./scripts/publish.sh", When: "success"}}
	if !reflect.DeepEqual(cfg.Pipelines["pr"], steps) {
		t.Fatalf("pr pipeline = %+v", cfg.Pipelines["pr"])
	}
}

func TestPhase4Defaults(t *testing.T) {
	var e Environment
	if e.Days() != 1 || e.NamePrefix() != "glassops" {
		t.Fatalf("environment defaults: %d, %q", e.Days(), e.NamePrefix())
	}
	var se SalesforceEphemeral
	if se.SeedMode() != SeedMinimal {
		t.Fatalf("seed default = %q", se.SeedMode())
	}
	var a Analyze
	if a.Selector() != "Recommended" {
		t.Fatalf("rule selector default = %q", a.Selector())
	}
}

func TestAnalyzerMaxSeverityPrefersPolicy(t *testing.T) {
	one, three := 1, 3
	cases := []struct {
		name   string
		rules  *int
		global *int
		want   int
		ok     bool
	}{
		{"neither", nil, nil, 0, false},
		{"analyze only", nil, &three, 3, true},
		{"policy only", &one, nil, 1, true},
		{"both, policy wins", &one, &three, 1, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &Config{Policy: Policy{Rules: Rules{AnalyzerMaxSeverity: tc.rules}}, Analyze: Analyze{MaxSeverity: tc.global}}
			got, ok := c.AnalyzerMaxSeverity()
			if got != tc.want || ok != tc.ok {
				t.Fatalf("got %d, %v", got, ok)
			}
		})
	}
}

func TestDefaultPipelines(t *testing.T) {
	want := map[string][]Step{
		"pr": {
			{Run: "plan"},
			{Run: "check"},
			{Run: "analyze"},
			{Run: "env create"},
			{Run: "validate --target ephemeral --fallback-target prod"},
			{Run: "test --target ephemeral"},
			{Run: "report", When: "always"},
			{Run: "env delete", When: "always"},
		},
		"deploy": {
			{Run: "plan --target prod"},
			{Run: "validate --target prod"},
			{Run: "gate --target prod"},
			{Run: "deploy --target prod"},
			{Run: "report", When: "always"},
		},
	}
	got := DefaultPipelines()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DefaultPipelines =\n%+v\nwant\n%+v", got, want)
	}
	if !reflect.DeepEqual(Default("salesforce").Pipelines, want) {
		t.Fatal("Default does not use DefaultPipelines")
	}
	// Each call returns its own copy, so a caller that edits one cannot change the defaults.
	got["pr"][0].Run = "changed"
	if DefaultPipelines()["pr"][0].Run != "plan" {
		t.Fatal("DefaultPipelines shares state between calls")
	}
}
