package salesforce

import (
	"context"
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/providers"
)

func TestConfigureSelectsEngine(t *testing.T) {
	cases := []struct {
		engine string
		want   string
		err    string
	}{
		{"", config.EngineNative, ""},
		{"native", config.EngineNative, ""},
		{"exec", "", "unknown engine"},
		{"external", "", "unknown engine"},
	}
	for _, tc := range cases {
		t.Run(tc.engine, func(t *testing.T) {
			p := New()
			err := p.Configure(&config.Config{Salesforce: config.Salesforce{Engine: tc.engine}})
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("err = %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := p.Engine().Name(); got != tc.want {
				t.Fatalf("engine = %q, want %q", got, tc.want)
			}
		})
	}
	if got := New().Engine().Name(); got != config.EngineNative {
		t.Fatalf("unconfigured engine = %q", got)
	}
	p := New()
	_ = p.Configure(&config.Config{Salesforce: config.Salesforce{Engine: "native"}})
	if err := p.Configure(nil); err != nil || p.Engine().Name() != config.EngineNative {
		t.Fatal("Configure(nil) did not go back to native")
	}
}

// TestNativeEngineNeedsASession: the native engine refuses validate and deploy without a
// session from this provider.
func TestNativeEngineNeedsASession(t *testing.T) {
	p := New()
	if err := p.Configure(&config.Config{Salesforce: config.Salesforce{Engine: "native"}}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := p.Validate(ctx, nil, providers.ChangeSet{}, providers.ValidateOptions{}); err == nil {
		t.Fatal("validate did not fail")
	}
	if _, err := p.Deploy(ctx, nil, providers.ChangeSet{}, providers.DeployOptions{}); err == nil {
		t.Fatal("deploy did not fail")
	}
}
