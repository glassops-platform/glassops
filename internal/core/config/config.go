// Package config loads glassops.yaml, validates it against schemas/config.json and the
// semantic rules the schema cannot express, and supplies defaults when there is no file.
package config

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// FileName is the config file glassops looks for in the project root.
const FileName = "glassops.yaml"

// Config is the parsed glassops.yaml.
type Config struct {
	Version    int               `yaml:"version" json:"version"`
	Provider   string            `yaml:"provider,omitempty" json:"provider,omitempty"`
	Targets    map[string]Target `yaml:"targets,omitempty" json:"targets,omitempty"`
	Policy     Policy            `yaml:"policy,omitempty" json:"policy,omitempty"`
	Salesforce Salesforce        `yaml:"salesforce,omitempty" json:"salesforce,omitempty"`
	Pipelines  map[string][]Step `yaml:"pipelines,omitempty" json:"pipelines,omitempty"`
	// Environments maps the target name an ephemeral environment registers as to how it is
	// created.
	Environments map[string]Environment `yaml:"environments,omitempty" json:"environments,omitempty"`
	Test         Test                   `yaml:"test,omitempty" json:"test,omitempty"`
	Check        Checks                 `yaml:"check,omitempty" json:"check,omitempty"`
	Analyze      Analyze                `yaml:"analyze,omitempty" json:"analyze,omitempty"`
	Verify       Verify                 `yaml:"verify,omitempty" json:"verify,omitempty"`
	// Exec configures provider exec. Other providers ignore it.
	Exec      Exec      `yaml:"exec,omitempty" json:"exec,omitempty"`
	Telemetry Telemetry `yaml:"telemetry,omitempty" json:"telemetry,omitempty"`
	Report    Report    `yaml:"report,omitempty" json:"report,omitempty"`
}

// Modes for verify.foreign_deploys and verify.flows.
const (
	VerifyFail = "fail"
	VerifyWarn = "warn"
	VerifyOff  = "off"
)

// Verify holds the post-deploy checks verify runs.
type Verify struct {
	// Smoke is an optional shell command verify runs after its own checks, like an exec step.
	Smoke string `yaml:"smoke,omitempty" json:"smoke,omitempty"`
	// ForeignDeploys is what another deploy to the target during the deploy window does: fail,
	// warn or off. Use ForeignDeploysMode to read it with the default, warn.
	ForeignDeploys string `yaml:"foreign_deploys,omitempty" json:"foreign_deploys,omitempty"`
	// Flows is what an active Flow in the plan that the target does not run as its latest
	// version does: fail, warn or off. Use FlowsMode to read it with the default, fail.
	Flows string `yaml:"flows,omitempty" json:"flows,omitempty"`
}

// ForeignDeploysMode returns ForeignDeploys, or VerifyWarn when it is not set.
func (v Verify) ForeignDeploysMode() string {
	if v.ForeignDeploys == "" {
		return VerifyWarn
	}
	return v.ForeignDeploys
}

// FlowsMode returns Flows, or VerifyFail when it is not set.
func (v Verify) FlowsMode() string {
	if v.Flows == "" {
		return VerifyFail
	}
	return v.Flows
}

// ExecProvider is the provider name that runs configured commands instead of a built-in
// provider.
const ExecProvider = "exec"

// Exec holds the commands provider exec runs. Each is a shell command run with sh -c in the
// project dir.
type Exec struct {
	// Paths are the project paths whose changed files make up the plan.
	Paths []string `yaml:"paths,omitempty" json:"paths,omitempty"`
	// Plan may write more plan files into $GLASSOPS_PLAN_DIR.
	Plan string `yaml:"plan,omitempty" json:"plan,omitempty"`
	// Describe prints {"id": "...", "kind": "..."} for the target.
	Describe string `yaml:"describe,omitempty" json:"describe,omitempty"`
	Validate string `yaml:"validate,omitempty" json:"validate,omitempty"`
	// Deploy is the deploy command. deploy -- <cmd> overrides it.
	Deploy string `yaml:"deploy,omitempty" json:"deploy,omitempty"`
	Verify string `yaml:"verify,omitempty" json:"verify,omitempty"`
}

// DefaultServiceName is the OTLP service name when telemetry.otlp.service_name is not set.
const DefaultServiceName = "glassops"

// Telemetry holds the optional export of run records.
type Telemetry struct {
	OTLP OTLP `yaml:"otlp,omitempty" json:"otlp,omitempty"`
}

// Enabled reports whether telemetry export is configured.
func (t Telemetry) Enabled() bool { return t.OTLP.Endpoint != "" }

// OTLP configures export over OTLP/HTTP. Export is on only when Endpoint is set.
type OTLP struct {
	Endpoint string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	// Headers are sent with every export. Each value is a $VAR reference, never a literal.
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	// ServiceName names the service in the spans. Use Service to read it with the default.
	ServiceName string `yaml:"service_name,omitempty" json:"service_name,omitempty"`
}

// Service returns ServiceName, or DefaultServiceName when it is not set.
func (o OTLP) Service() string {
	if o.ServiceName == "" {
		return DefaultServiceName
	}
	return o.ServiceName
}

// Report holds the report step settings.
type Report struct {
	// PRComment makes report post its summary as one pull request comment.
	PRComment bool `yaml:"pr_comment,omitempty" json:"pr_comment,omitempty"`
}

// Defaults for the environments section.
const (
	DefaultEnvPrefix  = "glassops"
	DefaultEnvTTLDays = 1
)

// Environment says how one ephemeral environment is created.
type Environment struct {
	// DevHub is the name of the configured target that creates the environment.
	DevHub string `yaml:"devhub" json:"devhub"`
	// Definition is the definition file, relative to the project root.
	Definition string `yaml:"definition,omitempty" json:"definition,omitempty"`
	// TTLDays is how many days the environment lives. Use Days to read it with the default.
	TTLDays int `yaml:"ttl_days,omitempty" json:"ttl_days,omitempty"`
	// Shape is an optional source org Id for an org shape, or a $VAR holding one.
	Shape string `yaml:"shape,omitempty" json:"shape,omitempty"`
	// Prefix starts every environment name. Use NamePrefix to read it with the default.
	Prefix string `yaml:"prefix,omitempty" json:"prefix,omitempty"`
}

// Days returns TTLDays, or DefaultEnvTTLDays when it is not set.
func (e Environment) Days() int {
	if e.TTLDays == 0 {
		return DefaultEnvTTLDays
	}
	return e.TTLDays
}

// NamePrefix returns Prefix, or DefaultEnvPrefix when it is not set.
func (e Environment) NamePrefix() string {
	if e.Prefix == "" {
		return DefaultEnvPrefix
	}
	return e.Prefix
}

// Test holds the test step settings.
type Test struct {
	// Level maps a tier, or "default", to a test level name.
	Level    map[string]string `yaml:"level,omitempty" json:"level,omitempty"`
	Coverage Coverage          `yaml:"coverage,omitempty" json:"coverage,omitempty"`
}

// Coverage holds the minimum coverage percents the test step enforces. Nil means not set.
type Coverage struct {
	OrgMin   *int `yaml:"org_min,omitempty" json:"org_min,omitempty"`
	ClassMin *int `yaml:"class_min,omitempty" json:"class_min,omitempty"`
	// ClassScope picks which classes class_min applies to when a class has no covered lines:
	// "changed" (the default) counts the classes in the plan, or every untouched class when
	// there is no plan, and "all" counts every class the org reports.
	ClassScope string `yaml:"class_scope,omitempty" json:"class_scope,omitempty"`
}

// Checks holds the scripts the check step runs.
type Checks struct {
	Scripts []Script `yaml:"scripts,omitempty" json:"scripts,omitempty"`
}

// Script is one named command the check step runs.
type Script struct {
	Name string `yaml:"name" json:"name"`
	Run  string `yaml:"run" json:"run"`
}

// DefaultRuleSelector is the analyzer rule selector when analyze.rule_selector is not set.
const DefaultRuleSelector = "Recommended"

// Analyze holds the static analysis settings.
type Analyze struct {
	Enabled      *bool  `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	RuleSelector string `yaml:"rule_selector,omitempty" json:"rule_selector,omitempty"`
	// MaxSeverity is the default for policy.rules.analyzer_max_severity. 1 is critical.
	// Violations at this severity or worse fail gate.
	MaxSeverity *int `yaml:"max_severity,omitempty" json:"max_severity,omitempty"`
}

// Selector returns RuleSelector, or DefaultRuleSelector when it is not set.
func (a Analyze) Selector() string {
	if a.RuleSelector == "" {
		return DefaultRuleSelector
	}
	return a.RuleSelector
}

// AnalyzerMaxSeverity returns policy.rules.analyzer_max_severity when it is set, else
// analyze.max_severity. ok is false when neither is set.
func (c *Config) AnalyzerMaxSeverity() (severity int, ok bool) {
	if v := c.Policy.Rules.AnalyzerMaxSeverity; v != nil {
		return *v, true
	}
	if v := c.Analyze.MaxSeverity; v != nil {
		return *v, true
	}
	return 0, false
}

// Target is one place changes go, such as an org.
type Target struct {
	ID     string `yaml:"id,omitempty" json:"id,omitempty"`
	Tier   string `yaml:"tier,omitempty" json:"tier,omitempty"`
	Branch string `yaml:"branch,omitempty" json:"branch,omitempty"`
	Auth   Auth   `yaml:"auth,omitempty" json:"auth,omitempty"`
}

// Auth names how to log in to a target. Secret values are $VAR references, never literals.
type Auth struct {
	Method      string `yaml:"method,omitempty" json:"method,omitempty"`
	Username    string `yaml:"username,omitempty" json:"username,omitempty"`
	ClientID    string `yaml:"client_id,omitempty" json:"client_id,omitempty"`
	Key         string `yaml:"key,omitempty" json:"key,omitempty"`
	URL         string `yaml:"url,omitempty" json:"url,omitempty"`
	Token       string `yaml:"token,omitempty" json:"token,omitempty"`
	InstanceURL string `yaml:"instance_url,omitempty" json:"instance_url,omitempty"`
}

// Policy holds the rules gate evaluates.
type Policy struct {
	Windows   map[string][]Window `yaml:"windows,omitempty" json:"windows,omitempty"`
	Blackouts []Blackout          `yaml:"blackouts,omitempty" json:"blackouts,omitempty"`
	Rules     Rules               `yaml:"rules,omitempty" json:"rules,omitempty"`
	Actors    string              `yaml:"actors,omitempty" json:"actors,omitempty"`
	// PermitTTL is how long a PERMIT stays valid, as a Go duration such as "2h". Use TTL to
	// read it with the default applied.
	PermitTTL string `yaml:"permit_ttl,omitempty" json:"permit_ttl,omitempty"`
	// Permit chooses how PERMITs are signed.
	Permit PermitSigning `yaml:"permit,omitempty" json:"permit,omitempty"`
	// Rego adds Rego rules that run after the built-in rules.
	Rego Rego `yaml:"rego,omitempty" json:"rego,omitempty"`
}

// PERMIT signing methods for policy.permit.signing.
const (
	SigningHMAC     = "hmac"
	SigningSigstore = "sigstore"
)

// PermitSigning chooses how PERMITs are signed and checked.
type PermitSigning struct {
	// Signing is hmac or sigstore. Use Method to read it with the default, hmac.
	Signing string `yaml:"signing,omitempty" json:"signing,omitempty"`
	// Identity and Issuer are the certificate identity and OIDC issuer a sigstore signature
	// must carry. Both are required for sigstore.
	Identity string `yaml:"identity,omitempty" json:"identity,omitempty"`
	Issuer   string `yaml:"issuer,omitempty" json:"issuer,omitempty"`
	// Repository is the GitHub repository, as owner/name, a sigstore signature must come
	// from. Every repository that calls the same reusable workflow gets the same certificate
	// identity, so verifying pins the repository too. It defaults to GITHUB_REPOSITORY on
	// GitHub Actions and is required anywhere else.
	Repository string `yaml:"repository,omitempty" json:"repository,omitempty"`
}

// Method returns Signing, or SigningHMAC when it is not set.
func (p PermitSigning) Method() string {
	if p.Signing == "" {
		return SigningHMAC
	}
	return p.Signing
}

// Rego lists the Rego policy files gate evaluates.
type Rego struct {
	Files []string `yaml:"files,omitempty" json:"files,omitempty"`
}

// Enabled reports whether any Rego files are configured.
func (r Rego) Enabled() bool { return len(r.Files) > 0 }

// DefaultPermitTTL is the PERMIT lifetime when policy.permit_ttl is not set.
const DefaultPermitTTL = 2 * time.Hour

// Bounds Check enforces on policy.permit_ttl.
const (
	MinPermitTTL = 5 * time.Minute
	MaxPermitTTL = 24 * time.Hour
)

// TTL returns the PERMIT lifetime. An empty or unparsable value gives DefaultPermitTTL.
// Check rejects an unparsable value, so a loaded config never has one.
func (p Policy) TTL() time.Duration {
	if p.PermitTTL == "" {
		return DefaultPermitTTL
	}
	d, err := time.ParseDuration(p.PermitTTL)
	if err != nil {
		return DefaultPermitTTL
	}
	return d
}

// Window is a weekly time range when changes to a tier are allowed.
type Window struct {
	Days []string `yaml:"days" json:"days"`
	From string   `yaml:"from" json:"from"`
	To   string   `yaml:"to" json:"to"`
	TZ   string   `yaml:"tz" json:"tz"`
}

// Blackout is a date range when changes to the listed tiers are refused.
type Blackout struct {
	From   string   `yaml:"from" json:"from"`
	To     string   `yaml:"to" json:"to"`
	Tiers  []string `yaml:"tiers,omitempty" json:"tiers,omitempty"`
	Reason string   `yaml:"reason,omitempty" json:"reason,omitempty"`
}

// Rules are the v1 policy rules.
type Rules struct {
	CoverageMin           *int                  `yaml:"coverage_min,omitempty" json:"coverage_min,omitempty"`
	AnalyzerMaxSeverity   *int                  `yaml:"analyzer_max_severity,omitempty" json:"analyzer_max_severity,omitempty"`
	SeparationOfDuties    *SoDRule              `yaml:"separation_of_duties,omitempty" json:"separation_of_duties,omitempty"`
	DestructiveNeedsLabel *DestructiveLabelRule `yaml:"destructive_needs_label,omitempty" json:"destructive_needs_label,omitempty"`
}

// SoDRule requires that the person who deploys to the listed tiers is not the author of the
// change, unless an approval satisfies it.
type SoDRule struct {
	Tiers    []string     `yaml:"tiers,omitempty" json:"tiers,omitempty"`
	Approval *SoDApproval `yaml:"approval,omitempty" json:"approval,omitempty"`
}

// SoDApproval names the approval that satisfies separation of duties: a CI environment
// approval, or a PR label applied by one of the approvers.
type SoDApproval struct {
	Environment string   `yaml:"environment,omitempty" json:"environment,omitempty"`
	Label       string   `yaml:"label,omitempty" json:"label,omitempty"`
	Approvers   []string `yaml:"approvers,omitempty" json:"approvers,omitempty"`
}

// DestructiveLabelRule requires a PR label before destructive changes reach the listed tiers.
type DestructiveLabelRule struct {
	Label string   `yaml:"label" json:"label"`
	Tiers []string `yaml:"tiers,omitempty" json:"tiers,omitempty"`
}

// Salesforce engines for salesforce.engine.
const (
	EngineNative = "native"
	EngineHardis = "hardis"
)

// Salesforce holds provider settings.
type Salesforce struct {
	// Engine is native or hardis. Use EngineName to read it with the default, native.
	Engine     string              `yaml:"engine,omitempty" json:"engine,omitempty"`
	Hardis     Hardis              `yaml:"hardis,omitempty" json:"hardis,omitempty"`
	APIVersion string              `yaml:"api_version,omitempty" json:"api_version,omitempty"`
	TestLevel  map[string]string   `yaml:"test_level,omitempty" json:"test_level,omitempty"`
	Delta      Delta               `yaml:"delta,omitempty" json:"delta,omitempty"`
	Deploy     SalesforceDeploy    `yaml:"deploy,omitempty" json:"deploy,omitempty"`
	Ephemeral  SalesforceEphemeral `yaml:"ephemeral,omitempty" json:"ephemeral,omitempty"`
}

// EngineName returns Engine, or EngineNative when it is not set.
func (s Salesforce) EngineName() string {
	if s.Engine == "" {
		return EngineNative
	}
	return s.Engine
}

// Hardis holds the settings for engine hardis. The user installs sfdx-hardis. GlassOps never
// ships it.
type Hardis struct {
	// MinVersion is the oldest sfdx-hardis version doctor accepts.
	MinVersion string `yaml:"min_version,omitempty" json:"min_version,omitempty"`
}

// Seed modes for salesforce.ephemeral.seed.
const (
	// SeedMinimal deploys only the components a destructive change deletes, from the base ref.
	SeedMinimal = "minimal"
	// SeedFull deploys the whole source at the base ref.
	SeedFull = "full"
	// SeedOff seeds nothing.
	SeedOff = "off"
)

// SalesforceEphemeral controls ephemeral Salesforce targets.
type SalesforceEphemeral struct {
	// Seed is minimal, full or off. Use SeedMode to read it with the default applied.
	Seed string `yaml:"seed,omitempty" json:"seed,omitempty"`
}

// SeedMode returns Seed, or SeedMinimal when it is not set.
func (e SalesforceEphemeral) SeedMode() string {
	if e.Seed == "" {
		return SeedMinimal
	}
	return e.Seed
}

// Delta controls change-set generation.
type Delta struct {
	Enabled *bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
}

// Quick deploy modes for salesforce.deploy.quick.
const (
	// QuickPrefer quick deploys the validated Id when the org accepts it, else runs a full deploy.
	QuickPrefer = "prefer"
	// QuickRequire fails the deploy when the org refuses the quick deploy.
	QuickRequire = "require"
	// QuickOff always runs a full deploy.
	QuickOff = "off"
)

// SalesforceDeploy controls a real deploy to Salesforce.
type SalesforceDeploy struct {
	// Quick is prefer, require or off. Use QuickMode to read it with the default applied.
	Quick string `yaml:"quick,omitempty" json:"quick,omitempty"`
}

// QuickMode returns Quick, or QuickPrefer when it is not set.
func (d SalesforceDeploy) QuickMode() string {
	if d.Quick == "" {
		return QuickPrefer
	}
	return d.Quick
}

// ExecStep is the step name that runs Exec instead of a glassops command.
const ExecStep = "exec"

// Step is one pipeline entry. In YAML it is either a plain command string or an object
// with a step and a when condition.
type Step struct {
	Run  string `yaml:"step" json:"step"`
	When string `yaml:"when,omitempty" json:"when,omitempty"`
	// Exec is the shell command an exec step runs, with sh -c on Linux and macOS. It is set
	// only when Run is ExecStep.
	Exec string `yaml:"run,omitempty" json:"run,omitempty"`
}

// UnmarshalYAML accepts both "plan" and {step: plan, when: always}.
func (s *Step) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		s.Run = n.Value
		return nil
	}
	type plain Step
	return n.Decode((*plain)(s))
}

// Loaded is a config plus where it came from.
type Loaded struct {
	Config *Config
	// Path is the file read, or "" when defaults were used.
	Path string
	// Raw is the file's bytes as read, or nil when defaults were used.
	Raw []byte
}

// SHA256 returns the hex SHA-256 of the raw file bytes, with CRLF read as LF. With no file it
// hashes Config encoded as JSON, which encoding/json writes the same way every time.
func (l *Loaded) SHA256() string {
	// CRLF and LF checkouts of the same file hash the same, so a PERMIT written on one
	// runner verifies on another.
	data := bytes.ReplaceAll(l.Raw, []byte{'\r', '\n'}, []byte{'\n'})
	if l.Raw == nil {
		// A Config holds only strings, numbers, bools, slices, maps and structs, so encoding
		// cannot fail.
		data, _ = json.Marshal(l.Config)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// ErrNotFound is returned by Load when the file does not exist and required is true.
var ErrNotFound = errors.New("config file not found")

// Load reads path, or dir/glassops.yaml when path is empty. With no file it returns defaults
// for the given provider. Load validates against the schema and the semantic rules and
// returns every problem it finds.
func Load(dir, path, provider string) (*Loaded, error) {
	explicit := path != ""
	if !explicit {
		path = filepath.Join(dir, FileName)
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if explicit {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, path)
		}
		return &Loaded{Config: Default(provider)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	cfg, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if cfg.Provider == "" {
		cfg.Provider = provider
	}
	return &Loaded{Config: cfg, Path: path, Raw: data}, nil
}

// Parse decodes and validates config bytes.
func Parse(data []byte) (*Config, error) {
	var doc yaml.Node
	dec := yaml.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	if err := validateSchema(&doc); err != nil {
		return nil, err
	}
	var cfg Config
	if err := doc.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	if problems := Check(&cfg); len(problems) > 0 {
		return nil, &ValidationError{Problems: problems}
	}
	return &cfg, nil
}

// Default is the config used when a project has no glassops.yaml.
func Default(provider string) *Config {
	return &Config{
		Version:   1,
		Provider:  provider,
		Pipelines: DefaultPipelines(),
	}
}

// DefaultPipelines are the pr and deploy pipelines a project gets when its config names
// none. pr checks a pull request against a scratch org and always deletes the org again.
// deploy plans against what prod last received, validates, gates and deploys. Both always
// report, so a failed run still leaves a summary and SARIF.
func DefaultPipelines() map[string][]Step {
	return map[string][]Step{
		"pr": {
			{Run: "plan"},
			{Run: "check"},
			{Run: "analyze"},
			{Run: "env create"},
			{Run: "validate --target ephemeral --fallback-target prod"},
			{Run: "report", When: "always"},
			{Run: "env delete", When: "always"},
		},
		"deploy": {
			{Run: "plan --target prod"},
			{Run: "validate --target prod"},
			{Run: "gate --target prod"},
			{Run: "deploy --target prod"},
			{Run: "verify --target prod"},
			{Run: "report", When: "always"},
		},
	}
}

// ValidationError lists every problem found in a config.
type ValidationError struct {
	Problems []string
}

func (e *ValidationError) Error() string {
	var b bytes.Buffer
	b.WriteString("invalid config:")
	for _, p := range e.Problems {
		b.WriteString("\n  - ")
		b.WriteString(p)
	}
	return b.String()
}
