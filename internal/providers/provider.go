// Package providers holds the seam between the neutral core and each delivery target type.
//
// The core speaks in targets, sessions, change sets and facts. Anything specific to one kind
// of target, such as a Salesforce org type, goes into facts under the provider's name, for
// example "salesforce.org_type". Deploy joins the interface in a later phase.
package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/glassops-platform/glassops/internal/core/config"
)

// Tool is an external program a provider needs on the PATH.
type Tool struct {
	// Name is how doctor reports the tool.
	Name string
	// Version runs the tool and returns its version, or an error when it is missing or broken.
	Version func() (string, error)
}

// Target is one configured target, resolved by name.
type Target struct {
	// Name is the key under targets in glassops.yaml.
	Name string
	// ID is the expected target Id, with any $VAR already expanded.
	ID     string
	Tier   string
	Branch string
	// Auth still holds $VAR references. The provider resolves them when it logs in, so
	// secret values never pass through the core.
	Auth config.Auth
}

// Session is a logged-in connection to one target. Close removes everything the login left
// on disk. It is safe to call more than once.
type Session interface {
	Target() Target
	Close() error
}

// Kind is the class of a target as the guard sees it.
type Kind string

const (
	KindProduction Kind = "production"
	KindSandbox    Kind = "sandbox"
	KindScratch    Kind = "scratch"
	KindUnknown    Kind = "unknown"
)

// TargetInfo is what a target reports about itself after login.
type TargetInfo struct {
	// ID is the target's own Id, as the target reports it.
	ID   string `json:"id"`
	Kind Kind   `json:"kind"`
	// Facts holds provider detail under keys prefixed with the provider name. It never holds
	// a company or organization name, and usernames only in masked form.
	Facts map[string]any `json:"facts"`
}

// PlanRequest asks a provider for the changes between two git refs.
type PlanRequest struct {
	// Dir is the project root.
	Dir string
	// From and To are resolved git refs. Full ignores them and plans the whole source.
	From string
	To   string
	Full bool
	// OutDir is where the provider writes its manifests, normally <Dir>/.glassops/plan.
	OutDir string
}

// PlanSchema identifies the change set file.
const PlanSchema = "glassops.plan/v1"

// ChangeSet is the result of a plan, written to .glassops/plan.json. Paths are relative to
// the project root, with forward slashes.
type ChangeSet struct {
	Schema   string `json:"schema"`
	Provider string `json:"provider"`
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	Full     bool   `json:"full"`
	// Dir holds every file the provider wrote for this plan.
	Dir string `json:"dir"`
	// Manifest is what a validate or deploy sends.
	Manifest string `json:"manifest"`
	// Destructive lists the destructive manifests, if any.
	Destructive []string `json:"destructive"`
	Empty       bool     `json:"empty"`
	// SHA256 is computed over the manifest files, so a later step can tell the plan changed.
	SHA256 string         `json:"plan_sha256"`
	Facts  map[string]any `json:"facts"`
}

// WriteChangeSet writes cs as indented JSON.
func WriteChangeSet(path string, cs ChangeSet) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create plan dir: %w", err)
	}
	data, err := json.MarshalIndent(cs, "", "  ")
	if err != nil {
		return fmt.Errorf("encode plan: %w", err)
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// ReadChangeSet reads a change set written by WriteChangeSet.
func ReadChangeSet(path string) (ChangeSet, error) {
	var cs ChangeSet
	data, err := os.ReadFile(path)
	if err != nil {
		return cs, fmt.Errorf("read plan: %w", err)
	}
	if err := json.Unmarshal(data, &cs); err != nil {
		return cs, fmt.Errorf("parse plan %s: %w", path, err)
	}
	if cs.Schema != PlanSchema {
		return cs, fmt.Errorf("%s: schema is %q, want %q", path, cs.Schema, PlanSchema)
	}
	return cs, nil
}

// ValidateOptions controls a check-only validate.
type ValidateOptions struct {
	// Dir is the project root. Change set paths are relative to it.
	Dir string
	// TestLevel is the provider's test level name, such as RunLocalTests. Empty means the
	// provider default.
	TestLevel string
	// Tests names the tests to run when the test level needs them.
	Tests []string
}

// ValidateResult is the outcome of a check-only validate.
type ValidateResult struct {
	Passed bool
	// ID is the target's Id for the validated change, kept so a later deploy can refer to it.
	ID string
	// Problems has one line per component or test failure.
	Problems []string
	Facts    map[string]any
}

// Provider is one kind of delivery target, such as Salesforce.
type Provider interface {
	// Name is the value used for provider in glassops.yaml.
	Name() string
	// Detect reports whether dir holds a project for this provider.
	Detect(dir string) (bool, error)
	// Tools lists the external programs the provider needs.
	Tools() []Tool
	// CanonicalID returns the form of id that two Ids are compared by.
	CanonicalID(id string) string
	// Authenticate logs in to t. The caller must Close the session.
	Authenticate(ctx context.Context, t Target) (Session, error)
	// Describe reads what the logged-in target reports about itself.
	Describe(ctx context.Context, s Session) (TargetInfo, error)
	// Plan works out the change set. It needs no login.
	Plan(ctx context.Context, req PlanRequest) (ChangeSet, error)
	// Validate runs a check-only validate of cs. It never changes the target.
	Validate(ctx context.Context, s Session, cs ChangeSet, opts ValidateOptions) (ValidateResult, error)
}

// Registry lists every known provider in detection order.
type Registry []Provider

// Detect returns the first provider that recognises dir, or nil when none does.
func (r Registry) Detect(dir string) (Provider, error) {
	for _, p := range r {
		ok, err := p.Detect(dir)
		if err != nil {
			return nil, err
		}
		if ok {
			return p, nil
		}
	}
	return nil, nil
}

// Get returns the provider with the given name, or nil.
func (r Registry) Get(name string) Provider {
	for _, p := range r {
		if p.Name() == name {
			return p
		}
	}
	return nil
}
