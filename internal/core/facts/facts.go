// Package facts holds the facts document gate evaluates: one flat map of namespaced keys per run.
package facts

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Schema identifies the facts file.
const Schema = "glassops.facts/v1"

// Key names. Provider facts keep their own prefix, such as "salesforce.".
const (
	TargetName = "target.name"
	TargetTier = "target.tier"
	TargetID   = "target.id"
	TargetKind = "target.kind"

	ChangePlanSHA256       = "change.plan_sha256"
	ChangeEmpty            = "change.empty"
	ChangeDestructiveCount = "change.destructive_count"

	CIRepo      = "ci.repo"
	CICommit    = "ci.commit"
	CIRef       = "ci.ref"
	CIEvent     = "ci.event"
	CIActor     = "ci.actor" // who triggered the run, on a push the pusher or merger
	CIPRNumber  = "ci.pr_number"
	CIPRAuthor  = "ci.pr_author"
	CILabels    = "ci.labels"    // []string
	CIApprovals = "ci.approvals" // []Approval
)

// Approval is one piece of approval evidence the CI reported.
type Approval struct {
	Mechanism   string    `json:"mechanism"` // "environment" or "label"
	Approver    string    `json:"approver"`
	State       string    `json:"state,omitempty"` // "approved", "rejected"
	Environment string    `json:"environment,omitempty"`
	Label       string    `json:"label,omitempty"`
	At          time.Time `json:"at,omitempty"`
	Comment     string    `json:"comment,omitempty"`
}

// Document is the facts for one run.
type Document struct {
	Schema string         `json:"schema"`
	Facts  map[string]any `json:"facts"`
}

// New returns an empty document with the schema set.
func New() *Document {
	return &Document{Schema: Schema, Facts: map[string]any{}}
}

// Set stores v under key.
func (d *Document) Set(key string, v any) {
	if d.Facts == nil {
		d.Facts = map[string]any{}
	}
	d.Facts[key] = v
}

// Get returns the value under key and whether it is present.
func (d *Document) Get(key string) (any, bool) {
	v, ok := d.Facts[key]
	return v, ok
}

// String returns the value under key when it is a string.
func (d *Document) String(key string) (string, bool) {
	v, ok := d.Facts[key].(string)
	return v, ok
}

// Int returns the value under key when it is a whole number. It accepts int, int64 and the
// float64 that JSON decoding produces, as long as it has no fraction.
func (d *Document) Int(key string) (int, bool) {
	switch v := d.Facts[key].(type) {
	case int:
		return v, true
	case int64:
		if v < math.MinInt || v > math.MaxInt {
			return 0, false
		}
		return int(v), true
	case float64:
		// float64(math.MaxInt) rounds up to 2^63 on 64-bit, so the upper bound is exclusive.
		if v != math.Trunc(v) || v < math.MinInt || v >= math.MaxInt {
			return 0, false
		}
		return int(v), true
	}
	return 0, false
}

// Bool returns the value under key when it is a bool.
func (d *Document) Bool(key string) (bool, bool) {
	v, ok := d.Facts[key].(bool)
	return v, ok
}

// Strings returns the value under key when it is a list of strings. It accepts []string and
// the []any of strings that JSON decoding produces.
func (d *Document) Strings(key string) ([]string, bool) {
	switch v := d.Facts[key].(type) {
	case []string:
		return v, true
	case []any:
		out := make([]string, 0, len(v))
		for _, e := range v {
			s, ok := e.(string)
			if !ok {
				return nil, false
			}
			out = append(out, s)
		}
		return out, true
	}
	return nil, false
}

// Approvals returns the approvals under CIApprovals. It accepts []Approval and the JSON form
// that decoding produces. A missing or malformed value gives nil.
func (d *Document) Approvals() []Approval {
	switch v := d.Facts[CIApprovals].(type) {
	case []Approval:
		return v
	case []any:
		data, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		var out []Approval
		if err := json.Unmarshal(data, &out); err != nil {
			return nil
		}
		return out
	}
	return nil
}

// Merge copies every key in m that starts with prefix. An empty prefix copies all of them.
func (d *Document) Merge(prefix string, m map[string]any) {
	for k, v := range m {
		if strings.HasPrefix(k, prefix) {
			d.Set(k, v)
		}
	}
}

// Write writes d to path as indented JSON, creating the parent directory.
func Write(path string, d *Document) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create facts dir: %w", err)
	}
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return fmt.Errorf("encode facts: %w", err)
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// Read reads a document written by Write and checks its schema.
func Read(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read facts: %w", err)
	}
	var d Document
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("parse facts %s: %w", path, err)
	}
	if d.Schema != Schema {
		return nil, fmt.Errorf("%s: schema is %q, want %q", path, d.Schema, Schema)
	}
	if d.Facts == nil {
		d.Facts = map[string]any{}
	}
	return &d, nil
}
