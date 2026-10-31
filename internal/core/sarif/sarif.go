// Package sarif holds the parts of SARIF 2.1.0 that glassops writes and merges. A run read
// from another tool keeps every field it had, so merging never loses the tool's own data.
package sarif

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// Version is the SARIF version written.
const Version = "2.1.0"

// SchemaURI is the schema identifier written into every log.
const SchemaURI = "https://json.schemastore.org/sarif-2.1.0.json"

// Levels SARIF allows on a result.
const (
	LevelError   = "error"
	LevelWarning = "warning"
	LevelNote    = "note"
	LevelNone    = "none"
)

// Log is one SARIF file.
type Log struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []Run  `json:"runs"`
}

// Run is one tool's results. A Run decoded from JSON re-encodes as it was read.
type Run struct {
	Tool    Tool     `json:"tool"`
	Results []Result `json:"results"`

	raw json.RawMessage
}

// UnmarshalJSON decodes the typed fields and keeps the original bytes.
func (r *Run) UnmarshalJSON(data []byte) error {
	type plain Run
	var p plain
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	*r = Run(p)
	r.raw = append(json.RawMessage(nil), data...)
	return nil
}

// MarshalJSON writes the original bytes when the run was read from a file.
func (r Run) MarshalJSON() ([]byte, error) {
	if r.raw != nil {
		return r.raw, nil
	}
	type plain Run
	if r.Results == nil {
		r.Results = []Result{}
	}
	return json.Marshal(plain(r))
}

// Tool names the program that produced a run.
type Tool struct {
	Driver Driver `json:"driver"`
}

// Driver is the tool's main component.
type Driver struct {
	Name           string `json:"name"`
	Version        string `json:"version,omitempty"`
	InformationURI string `json:"informationUri,omitempty"`
	Rules          []Rule `json:"rules,omitempty"`
}

// Rule describes one rule id used by the results.
type Rule struct {
	ID               string `json:"id"`
	ShortDescription *Text  `json:"shortDescription,omitempty"`
}

// Text is a SARIF message string.
type Text struct {
	Text string `json:"text"`
}

// Result is one finding.
type Result struct {
	RuleID    string     `json:"ruleId"`
	Level     string     `json:"level"`
	Message   Text       `json:"message"`
	Locations []Location `json:"locations,omitempty"`
}

// Location points at a file and optionally a line.
type Location struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

// PhysicalLocation is a file region.
type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation"`
	Region           *Region          `json:"region,omitempty"`
}

// ArtifactLocation is a file path or URI.
type ArtifactLocation struct {
	URI string `json:"uri"`
}

// Region is a line range. Only the start line is used.
type Region struct {
	StartLine int `json:"startLine"`
}

// New returns an empty log.
func New() *Log {
	return &Log{Schema: SchemaURI, Version: Version, Runs: []Run{}}
}

// Merge returns one log holding every run of every input, in order. Each run keeps its tool.
func Merge(logs ...*Log) *Log {
	out := New()
	for _, l := range logs {
		if l == nil {
			continue
		}
		out.Runs = append(out.Runs, l.Runs...)
	}
	return out
}

// Marshal encodes the log as indented JSON with a trailing newline.
func Marshal(l *Log) ([]byte, error) {
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode sarif: %w", err)
	}
	return append(data, '\n'), nil
}

// Read reads a SARIF file. It checks the version and that runs is present.
func Read(path string) (*Log, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read sarif: %w", err)
	}
	var l Log
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, fmt.Errorf("parse sarif %s: %w", path, err)
	}
	if l.Version != Version {
		return nil, fmt.Errorf("sarif %s: version %q, want %s", path, l.Version, Version)
	}
	if l.Runs == nil {
		return nil, fmt.Errorf("sarif %s: no runs array", path)
	}
	return &l, nil
}

// Validate checks encoded SARIF against the fields of the 2.1.0 schema that glassops relies
// on: version, $schema, runs[].tool.driver.name, and for every result a ruleId, a level and
// message.text. It is strict about glassops output and is not a full schema validator.
func Validate(data []byte) error {
	var doc struct {
		Schema  string `json:"$schema"`
		Version string `json:"version"`
		Runs    []struct {
			Tool struct {
				Driver struct {
					Name string `json:"name"`
				} `json:"driver"`
			} `json:"tool"`
			Results []struct {
				RuleID  string `json:"ruleId"`
				Level   string `json:"level"`
				Message struct {
					Text string `json:"text"`
				} `json:"message"`
				Locations []struct {
					PhysicalLocation struct {
						ArtifactLocation struct {
							URI string `json:"uri"`
						} `json:"artifactLocation"`
						Region *struct {
							StartLine int `json:"startLine"`
						} `json:"region"`
					} `json:"physicalLocation"`
				} `json:"locations"`
			} `json:"results"`
		} `json:"runs"`
	}
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&doc); err != nil {
		return fmt.Errorf("sarif is not valid JSON: %w", err)
	}
	var errs []error
	if doc.Version != Version {
		errs = append(errs, fmt.Errorf("version is %q, want %s", doc.Version, Version))
	}
	if doc.Schema == "" {
		errs = append(errs, errors.New("$schema is missing"))
	}
	if doc.Runs == nil {
		errs = append(errs, errors.New("runs is missing"))
	}
	for i, r := range doc.Runs {
		if r.Tool.Driver.Name == "" {
			errs = append(errs, fmt.Errorf("runs[%d].tool.driver.name is missing", i))
		}
		for j, res := range r.Results {
			at := fmt.Sprintf("runs[%d].results[%d]", i, j)
			if res.RuleID == "" {
				errs = append(errs, fmt.Errorf("%s.ruleId is missing", at))
			}
			switch res.Level {
			case LevelError, LevelWarning, LevelNote, LevelNone:
			default:
				errs = append(errs, fmt.Errorf("%s.level %q is not a SARIF level", at, res.Level))
			}
			if res.Message.Text == "" {
				errs = append(errs, fmt.Errorf("%s.message.text is missing", at))
			}
			for k, loc := range res.Locations {
				if loc.PhysicalLocation.ArtifactLocation.URI == "" {
					errs = append(errs, fmt.Errorf("%s.locations[%d] has no artifactLocation.uri", at, k))
				}
				if rg := loc.PhysicalLocation.Region; rg != nil && rg.StartLine < 1 {
					errs = append(errs, fmt.Errorf("%s.locations[%d].region.startLine must be at least 1", at, k))
				}
			}
		}
	}
	return errors.Join(errs...)
}
