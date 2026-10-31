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
}

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

// Salesforce holds provider settings.
type Salesforce struct {
	Engine     string            `yaml:"engine,omitempty" json:"engine,omitempty"`
	APIVersion string            `yaml:"api_version,omitempty" json:"api_version,omitempty"`
	TestLevel  map[string]string `yaml:"test_level,omitempty" json:"test_level,omitempty"`
	Delta      Delta             `yaml:"delta,omitempty" json:"delta,omitempty"`
	Deploy     SalesforceDeploy  `yaml:"deploy,omitempty" json:"deploy,omitempty"`
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

// Step is one pipeline entry. In YAML it is either a plain command string or an object
// with a step and a when condition.
type Step struct {
	Run  string `yaml:"step" json:"step"`
	When string `yaml:"when,omitempty" json:"when,omitempty"`
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

// SHA256 returns the hex SHA-256 of the raw file bytes. With no file it hashes Config encoded
// as JSON, which encoding/json writes the same way every time for the same value.
func (l *Loaded) SHA256() string {
	data := l.Raw
	if data == nil {
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

// DefaultPipelines are the pr and deploy pipelines from the design.
func DefaultPipelines() map[string][]Step {
	return map[string][]Step{
		"pr": {
			{Run: "plan"},
			{Run: "check"},
			{Run: "analyze"},
			{Run: "env create --from devhub --ttl 1"},
			{Run: "validate --target ephemeral --fallback-target prod --check-only"},
			{Run: "test --target ephemeral"},
			{Run: "report"},
			{Run: "env delete", When: "always"},
		},
		"deploy": {
			{Run: "plan"},
			{Run: "analyze"},
			{Run: "gate --target prod"},
			{Run: "deploy --target prod"},
			{Run: "verify --target prod"},
			{Run: "report"},
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
