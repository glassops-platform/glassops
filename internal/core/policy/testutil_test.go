package policy

import (
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/facts"
)

// Every fact and config value in these tests is synthetic, built here as Go literals.

func doc(kv map[string]any) *facts.Document {
	d := facts.New()
	for k, v := range kv {
		d.Set(k, v)
	}
	return d
}

func intp(n int) *int { return &n }

func at(t *testing.T, s string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	return ts
}

func one1(t *testing.T, rs []Result) Result {
	t.Helper()
	if len(rs) != 1 {
		t.Fatalf("got %d results, want 1: %+v", len(rs), rs)
	}
	return rs[0]
}

func cfgWith(f func(c *config.Config)) *config.Config {
	c := &config.Config{}
	f(c)
	return c
}
