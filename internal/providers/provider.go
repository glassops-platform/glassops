// Package providers holds the seam between the neutral core and each delivery target type.
//
// The core speaks in targets, sessions, change sets and facts. Anything specific to one kind
// of target, such as a Salesforce org type, goes into facts under the provider's name, for
// example "salesforce.org_type". Deploy is part of the interface, and callers must verify a
// PERMIT before they call it.
//
// Environments, Seeder and Tester are optional capabilities. A command checks for them with a
// type assertion on the Provider.
package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

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

// DeployOptions controls a real deploy.
type DeployOptions struct {
	// Dir is the project root. Change set paths are relative to it.
	Dir string
	// TestLevel is the provider's test level name. Empty means the provider default.
	TestLevel string
	// Tests names the tests to run when the test level needs them.
	Tests []string
	// ValidatedID is the check-only validation bound into the PERMIT. When set, the provider
	// quick deploys exactly this Id while the target still accepts it.
	ValidatedID string
	// QuickOnly makes a refused quick deploy a failure instead of a fall back to a full deploy.
	QuickOnly bool
	// Wait is how long to wait for the deploy before returning a job Id to resume. Zero means
	// the provider default.
	Wait time.Duration
	// ResumeID resumes watching an earlier deploy instead of starting a new one.
	ResumeID string
}

// DeployResult is the outcome of a real deploy, read from the target's result body.
type DeployResult struct {
	Passed bool
	// ID is the target's Id for the deploy.
	ID string
	// Quick is true when ValidatedID was quick deployed.
	Quick bool
	// TimedOut is true when Wait ran out before the deploy finished. ID can be resumed.
	TimedOut           bool
	ComponentsDeployed int
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
	// Deploy changes the target. Callers must have verified a PERMIT first.
	Deploy(ctx context.Context, s Session, cs ChangeSet, opts DeployOptions) (DeployResult, error)
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

// EnvRequest asks for one ephemeral environment.
type EnvRequest struct {
	// Name is the environment's name, such as glassops-<repo>-<pr>-<run>.
	Name string
	// DefinitionFile is the definition path, relative to Dir.
	DefinitionFile string
	Dir            string
	// TTL is how long the environment lives. Salesforce counts it in whole days.
	TTL time.Duration
	// Shape is an optional source org Id for an org shape.
	Shape string
}

// Env is a created environment. The material a later step logs in with is held only in
// AuthFile, a file with mode 0600, and never in a result.
type Env struct {
	// ID is the environment's own Id, such as an org Id.
	ID string
	// RecordID is the provider's request record, such as ScratchOrgInfo, kept for cleanup.
	RecordID  string
	Name      string
	ExpiresAt time.Time
	// AuthFile is the path of the 0600 file holding the login material.
	AuthFile string
	Facts    map[string]any
}

// Environments is implemented by providers that create ephemeral environments.
type Environments interface {
	// CreateEnv creates one environment from the hub.
	CreateEnv(ctx context.Context, hub Session, req EnvRequest) (Env, error)
	// DeleteEnv deletes env from the hub.
	DeleteEnv(ctx context.Context, hub Session, env Env) error
	// SweepEnvs deletes environments named with prefix that are older than maxAge, and reports them.
	SweepEnvs(ctx context.Context, hub Session, prefix string, maxAge time.Duration) ([]Env, error)
}

// SeedRequest asks a provider to put the base state an ephemeral target needs before a
// destructive change can be validated there.
type SeedRequest struct {
	Dir string
	// BaseRef is the resolved git ref the deleted components come from.
	BaseRef string
	Plan    ChangeSet
	// Mode is "minimal", "full" or "off".
	Mode string
}

// SeedResult is what a seed put on the target.
type SeedResult struct {
	// Mode is the mode actually used: "minimal", "full" or "none".
	Mode string
	// Fallback is true when minimal failed on a missing dependency and full was used.
	Fallback bool
	// Members lists what was seeded, as "Type:Name".
	Members []string
	Facts   map[string]any
}

// Seeder is implemented by providers that can seed an ephemeral target.
type Seeder interface {
	// Seed puts the base state req needs on the target.
	Seed(ctx context.Context, s Session, req SeedRequest) (SeedResult, error)
}

// TestRequest asks for a test run.
type TestRequest struct {
	Dir       string
	TestLevel string
	Tests     []string
	// Wait is how long to wait for the run. Zero means the provider default.
	Wait time.Duration
	// ResumeID resumes watching an earlier run instead of starting a new one.
	ResumeID string
}

// ClassCoverage is the line coverage of one class.
type ClassCoverage struct {
	Name    string
	Lines   int
	Covered int
	Percent int
}

// TestResult is the outcome of a test run.
type TestResult struct {
	Passed   bool
	Finished bool
	RunID    string
	Ran      int
	Failed   int
	// OrgCoverage is the org-wide coverage percent, -1 when unknown.
	OrgCoverage int
	Classes     []ClassCoverage
	Failures    []string
	// JUnit is provider-produced JUnit XML when available.
	JUnit []byte
	Facts map[string]any
}

// Tester is implemented by providers that run tests on a target.
type Tester interface {
	// RunTests runs or resumes a test run.
	RunTests(ctx context.Context, s Session, req TestRequest) (TestResult, error)
}
