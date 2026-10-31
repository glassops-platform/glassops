package exec

import (
	"context"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/providers"
)

func TestNeverDetected(t *testing.T) {
	p := New()
	if p.Name() != "exec" {
		t.Fatalf("name = %q", p.Name())
	}
	if ok, err := p.Detect(t.TempDir()); ok || err != nil {
		t.Fatalf("Detect = %v, %v", ok, err)
	}
}

func TestConfigureKeepsExecSection(t *testing.T) {
	p := New()
	if err := p.Configure(&config.Config{Exec: config.Exec{Paths: []string{"app"}, Deploy: "./d.sh"}}); err != nil {
		t.Fatal(err)
	}
	if p.cfg.Deploy != "./d.sh" || len(p.cfg.Paths) != 1 {
		t.Fatalf("cfg = %+v", p.cfg)
	}
}

func TestOperationsAreNotImplemented(t *testing.T) {
	p := New()
	ctx := context.Background()
	if _, err := p.Authenticate(ctx, providers.Target{}); err == nil {
		t.Fatal("authenticate")
	}
	if _, err := p.Plan(ctx, providers.PlanRequest{}); err == nil {
		t.Fatal("plan")
	}
	if _, err := p.Deploy(ctx, nil, providers.ChangeSet{}, providers.DeployOptions{}); err == nil {
		t.Fatal("deploy")
	}
	if _, err := p.Verify(ctx, nil, providers.VerifyRequest{}); err == nil {
		t.Fatal("verify")
	}
}
