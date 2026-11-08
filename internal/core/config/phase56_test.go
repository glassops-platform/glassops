package config

import (
	"reflect"
	"strings"
	"testing"
)

const phase56Config = `version: 1
provider: exec
targets:
  prod: { id: 00D000000000001, tier: prod }
policy:
  permit:
    signing: sigstore
    identity: https://github.com/acme/app/.github/workflows/deploy.yml@refs/heads/main
    issuer: https://token.actions.githubusercontent.com
  rego:
    files: [policy/deploy.rego]
salesforce:
  engine: hardis
  hardis: { min_version: "6.0" }
verify:
  smoke: ./scripts/smoke.sh
  foreign_deploys: fail
  flows: warn
exec:
  paths: [infra, app]
  plan: ./scripts/plan.sh
  describe: ./scripts/describe.sh
  validate: ./scripts/validate.sh
  deploy: ./scripts/deploy.sh
  verify: ./scripts/verify.sh
telemetry:
  otlp:
    endpoint: https://otel.example.com/v1/traces
    headers: { Authorization: $OTEL_AUTH }
    service_name: app-delivery
report:
  pr_comment: true
`

func TestParsePhase56Config(t *testing.T) {
	cfg, err := Parse([]byte(phase56Config))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if p := cfg.Policy.Permit; p.Method() != SigningSigstore || p.Identity == "" || p.Issuer == "" {
		t.Fatalf("permit = %+v", p)
	}
	if !cfg.Policy.Rego.Enabled() || cfg.Policy.Rego.Files[0] != "policy/deploy.rego" {
		t.Fatalf("rego = %+v", cfg.Policy.Rego)
	}
	if cfg.Salesforce.EngineName() != EngineHardis || cfg.Salesforce.Hardis.MinVersion != "6.0" {
		t.Fatalf("salesforce = %+v", cfg.Salesforce)
	}
	v := cfg.Verify
	if v.Smoke != "./scripts/smoke.sh" || v.ForeignDeploysMode() != VerifyFail || v.FlowsMode() != VerifyWarn {
		t.Fatalf("verify = %+v", v)
	}
	wantExec := Exec{Paths: []string{"infra", "app"}, Plan: "./scripts/plan.sh", Describe: "./scripts/describe.sh", Validate: "./scripts/validate.sh", Deploy: "./scripts/deploy.sh", Verify: "./scripts/verify.sh"}
	if !reflect.DeepEqual(cfg.Exec, wantExec) {
		t.Fatalf("exec = %+v", cfg.Exec)
	}
	o := cfg.Telemetry.OTLP
	if !cfg.Telemetry.Enabled() || o.Service() != "app-delivery" || o.Headers["Authorization"] != "$OTEL_AUTH" {
		t.Fatalf("telemetry = %+v", cfg.Telemetry)
	}
	if !cfg.Report.PRComment {
		t.Fatal("report.pr_comment not read")
	}
	if refs := EnvRefs(cfg); !reflect.DeepEqual(refs, []string{"OTEL_AUTH"}) {
		t.Fatalf("EnvRefs = %v", refs)
	}
}

func TestPhase56Defaults(t *testing.T) {
	var c Config
	if c.Policy.Permit.Method() != SigningHMAC {
		t.Fatalf("signing default = %q", c.Policy.Permit.Method())
	}
	if c.Policy.Rego.Enabled() || c.Telemetry.Enabled() || c.Report.PRComment {
		t.Fatal("an opt-in feature is on by default")
	}
	if c.Salesforce.EngineName() != EngineNative {
		t.Fatalf("engine default = %q", c.Salesforce.EngineName())
	}
	if c.Verify.ForeignDeploysMode() != VerifyWarn || c.Verify.FlowsMode() != VerifyFail {
		t.Fatalf("verify defaults = %q, %q", c.Verify.ForeignDeploysMode(), c.Verify.FlowsMode())
	}
	if c.Telemetry.OTLP.Service() != "glassops" {
		t.Fatalf("service default = %q", c.Telemetry.OTLP.Service())
	}
}

func TestParseRejectsPhase56(t *testing.T) {
	cases := []struct {
		name string
		yaml string
		want string
	}{
		{"unknown provider", "version: 1\nprovider: terraform\n", "/provider"},
		{"unknown engine", "version: 1\nsalesforce: { engine: exec }\n", "/salesforce/engine"},
		{"bad hardis version", "version: 1\nsalesforce: { hardis: { min_version: latest } }\n", "/salesforce/hardis/min_version"},
		{"unknown signing", "version: 1\npolicy:\n  permit: { signing: gpg }\n", "/policy/permit/signing"},
		{"sigstore without identity", "version: 1\npolicy:\n  permit: { signing: sigstore, issuer: https://token.actions.githubusercontent.com }\n", "needs both identity and issuer"},
		{"sigstore without issuer", "version: 1\npolicy:\n  permit: { signing: sigstore, identity: me }\n", "needs both identity and issuer"},
		{"empty rego files", "version: 1\npolicy:\n  rego: { files: [] }\n", "/policy/rego/files"},
		{"bad foreign deploys mode", "version: 1\nverify: { foreign_deploys: maybe }\n", "/verify/foreign_deploys"},
		{"bad flows mode", "version: 1\nverify: { flows: maybe }\n", "/verify/flows"},
		{"exec without paths", "version: 1\nprovider: exec\nexec: { deploy: ./d.sh }\n", "exec.paths"},
		{"exec empty command", "version: 1\nexec: { deploy: \"\" }\n", "/exec/deploy"},
		{"literal otlp header", "version: 1\ntelemetry:\n  otlp: { endpoint: https://x.example.com, headers: { Authorization: secret } }\n", "/telemetry/otlp/headers/Authorization"},
		{"otlp without endpoint", "version: 1\ntelemetry:\n  otlp: { service_name: app }\n", "/telemetry/otlp"},
		{"pr comment not bool", "version: 1\nreport: { pr_comment: yes please }\n", "/report/pr_comment"},
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
