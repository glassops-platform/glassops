package config

import (
	"strings"
	"testing"
)

func execConfig(ids ...string) string {
	var b strings.Builder
	b.WriteString("version: 1\nprovider: exec\nexec:\n  paths: [infra]\ntargets:\n")
	for i, id := range ids {
		b.WriteString("  t" + string(rune('a'+i)) + ":\n    id: " + id + "\n    tier: qa\n")
	}
	return b.String()
}

func TestExecTargetTakesAPlainId(t *testing.T) {
	c, err := Parse([]byte(execConfig("prod-eu-1", "acct:123/stage")))
	if err != nil {
		t.Fatal(err)
	}
	if p := Check(c); len(p) != 0 {
		t.Fatalf("plain exec Ids refused: %v", p)
	}
}

func TestExecTargetRefusesABadOrDuplicateId(t *testing.T) {
	for _, in := range []string{execConfig("-starts-with-dash"), execConfig("has space"), execConfig("same", "same")} {
		c, err := Parse([]byte(in))
		if err == nil && len(Check(c)) == 0 {
			t.Errorf("accepted %q", in)
		}
	}
}

func TestSalesforceTargetStillNeedsAnOrgId(t *testing.T) {
	in := "version: 1\nprovider: salesforce\ntargets:\n  prod:\n    id: prod-eu-1\n    tier: prod\n"
	c, err := Parse([]byte(in))
	if err == nil && len(Check(c)) == 0 {
		t.Fatal("a Salesforce target accepted a non org Id")
	}
}
