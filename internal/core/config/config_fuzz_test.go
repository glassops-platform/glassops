package config

import (
	"testing"
)

// FuzzMutConfigParse feeds arbitrary YAML to Parse. It must never panic, and a config it
// accepts must pass its own semantic check and survive the accessors the commands call.
func FuzzMutConfigParse(f *testing.F) {
	f.Add([]byte("version: 1\n"))
	f.Add([]byte(`version: 1
targets:
  prod:
    id: 00D000000000001AAA
    tier: prod
    auth: { method: sfdx-url, url: $SF_PROD_AUTH_URL }
policy:
  permit_ttl: 90m
  windows:
    prod:
      - { days: [mon, tue], from: "09:00", to: "17:00", tz: America/New_York }
  blackouts:
    - { from: 2027-12-20, to: 2028-01-02, tiers: [prod], reason: freeze }
  rules:
    coverage_min: 75
    separation_of_duties:
      tiers: [prod]
      approval: { environment: glassops-deploy }
salesforce:
  test_level: { prod: RunLocalTests }
`))
	f.Add([]byte("version: 2\n"))
	f.Add([]byte("version: 1\ntargets: []\n"))
	f.Add([]byte("version: 1\npolicy: { permit_ttl: -5m }\n"))
	f.Add([]byte("version: 1\npolicy: { permit: { signing: sigstore, identity: x, issuer: y } }\n"))
	f.Add([]byte("version: 1\ntargets:\n  a: &x { id: 00D000000000001AAA, tier: prod }\n  b: *x\n"))
	f.Add([]byte("\x00\x01\x02"))
	f.Add([]byte(""))
	f.Fuzz(func(t *testing.T, data []byte) {
		cfg, err := Parse(data)
		if err != nil {
			return
		}
		if problems := Check(cfg); len(problems) > 0 {
			t.Fatalf("Parse accepted a config that Check refuses: %v", problems)
		}
		_ = EnvRefs(cfg)
		_ = cfg.Policy.TTL()
		_ = cfg.Policy.Permit.Method()
		_, _ = cfg.AnalyzerMaxSeverity()
		_ = cfg.Salesforce.EngineName()
		_ = cfg.Salesforce.Deploy.QuickMode()
		l := &Loaded{Config: cfg, Raw: data}
		if l.SHA256() == "" {
			t.Fatal("no hash for a loaded config")
		}
	})
}
