//go:build mutation_open

package config

import "testing"

// Finding: a YAML anchor that refers to itself makes Parse recurse until the Go runtime dies
// with "fatal error: stack overflow". The process cannot recover, so `glassops gate`, `check`
// and every other command that loads the config crash on a hostile glassops.yaml. The cause is
// toJSONValue in schema.go, which follows alias nodes with no depth or cycle limit.
//
// Reproduce: go test -tags mutation_open -run TestMutOpenConfigAliasCycle ./internal/core/config
func TestMutOpenConfigAliasCycle(t *testing.T) {
	_, err := Parse([]byte("&a [*a]\n"))
	if err == nil {
		t.Fatal("a cyclic config must be refused")
	}
}
