package guard

import (
	"errors"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/providers"
)

func canon15(id string) string {
	if len(id) > 15 {
		return id[:15]
	}
	return id
}

func TestCheck(t *testing.T) {
	const id18, id15 = "00D000000000001AAA", "00D000000000001"
	cases := []struct {
		name  string
		tier  string
		want  string
		got   string
		kind  providers.Kind
		fails string
	}{
		{"same 18", "uat", id18, id18, providers.KindSandbox, ""},
		{"15 against 18", "uat", id15, id18, providers.KindSandbox, ""},
		{"18 against 15", "uat", id18, id15, providers.KindSandbox, ""},
		{"case differs", "uat", id15, "00d000000000001", providers.KindSandbox, "Id does not match"},
		{"other org", "uat", id18, "00D000000000002AAA", providers.KindSandbox, "Id does not match"},
		{"nothing configured", "uat", "", id18, providers.KindSandbox, "no Id is configured"},
		{"nothing reported", "uat", id18, "", providers.KindSandbox, "Id does not match"},
		{"prod is production", "prod", id18, id18, providers.KindProduction, ""},
		{"prod on a sandbox", "prod", id18, id18, providers.KindSandbox, "tier prod needs a production target"},
		{"prod on a scratch org", "prod", id18, id18, providers.KindScratch, "tier prod needs a production target"},
		{"devhub is production", "devhub", id18, id18, providers.KindProduction, ""},
		{"devhub on a sandbox", "devhub", id18, id18, providers.KindSandbox, "tier devhub needs a production target"},
		{"ephemeral is scratch", "ephemeral", id18, id18, providers.KindScratch, ""},
		{"ephemeral on a sandbox", "ephemeral", id18, id18, providers.KindSandbox, "tier ephemeral needs a scratch target"},
		{"ephemeral on production", "ephemeral", id18, id18, providers.KindProduction, "tier ephemeral needs a scratch target"},
		{"other tier, any kind", "qa", id18, id18, providers.KindUnknown, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := Check(providers.Target{Name: "t", ID: c.want, Tier: c.tier}, providers.TargetInfo{ID: c.got, Kind: c.kind}, canon15)
			if c.fails == "" {
				if err != nil {
					t.Fatalf("unexpected: %v", err)
				}
				return
			}
			var m *MismatchError
			if !errors.As(err, &m) || !strings.Contains(err.Error(), c.fails) {
				t.Fatalf("err = %v, want %q", err, c.fails)
			}
			if c.want != "" && c.got != "" && (!strings.Contains(err.Error(), c.want) || !strings.Contains(err.Error(), c.got)) {
				t.Fatalf("error must name both Ids: %v", err)
			}
		})
	}
}
