package policy

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
	"github.com/glassops-platform/glassops/internal/core/redact"
)

// RegoQuery is the query gate evaluates over the Rego files. Each entry of the set is one
// denial.
const RegoQuery = "data.glassops.deny"

// RegoPrefix starts the name of every check a Rego denial adds, as in "rego.protected_profile".
const RegoPrefix = "rego."

// regoTimeout bounds one opa run, so a policy that never finishes cannot hold gate open.
const regoTimeout = 2 * time.Minute

// maxRegoDetail is the longest detail kept from one denial, in bytes. The rest is cut.
const maxRegoDetail = 500

// regoRuleName is what a denial's rule may be called. It becomes part of a check name.
var regoRuleName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,99}$`)

// OPARunner runs the opa program with args in dir and returns its standard output. It returns
// an error when opa cannot be started or exits non-zero, along with the output it wrote.
type OPARunner func(ctx context.Context, dir string, args ...string) ([]byte, error)

var (
	// RunOPA is how the Rego evaluator runs opa. Tests replace it with a stand-in.
	RunOPA OPARunner = runOPA
	// RegoDir returns the directory relative policy.rego files resolve against for a config,
	// which is the project dir. When it is nil or returns "", paths resolve against the
	// working directory.
	RegoDir func(*config.Config) string
)

// opaProgram is the program runOPA starts. Tests point it at a stand-in.
var opaProgram = "opa"

// ErrOPANotFound is returned when the opa program is not on PATH.
var ErrOPANotFound = errors.New("opa is not installed or not on PATH, and policy.rego needs it")

// runOPA starts opa with exec. Standard error goes into the error, never into the output.
func runOPA(ctx context.Context, dir string, args ...string) ([]byte, error) {
	path, err := exec.LookPath(opaProgram)
	if err != nil {
		return nil, ErrOPANotFound
	}
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if msg := firstLine(stderr.String()); msg != "" {
			return stdout.Bytes(), fmt.Errorf("opa: %w: %s", err, msg)
		}
		return stdout.Bytes(), fmt.Errorf("opa: %w", err)
	}
	return stdout.Bytes(), nil
}

// regoEvaluator evaluates the Rego files in policy.rego with opa eval over the facts.
type regoEvaluator struct {
	cfg config.Rego
}

// NewRego returns the evaluator for the Rego files cfg names.
func NewRego(cfg config.Rego) Evaluator { return regoEvaluator{cfg: cfg} }

// Name returns "rego".
func (regoEvaluator) Name() string { return "rego" }

// Evaluate runs "opa eval" for RegoQuery with the facts document as input. Each entry in the
// deny set becomes a failed check named rego.<rule>. An empty set gives one passing check named
// rego. Rego can only add denials: nothing it returns can pass a rule the Go rules failed.
//
// It fails closed. A missing opa, a missing or unreadable file, an opa error, an undefined
// deny, a deny that is not a set and an entry without a valid rule name are all errors, and
// gate denies on any error.
func (e regoEvaluator) Evaluate(ctx context.Context, in Input) ([]Result, error) {
	if !e.cfg.Enabled() {
		return nil, errors.New("policy.rego names no files")
	}
	if in.Facts == nil {
		return nil, errors.New("no facts to evaluate the Rego rules over")
	}
	dir := ""
	if RegoDir != nil {
		dir = RegoDir(in.Config)
	}
	digest, err := regoDigest(dir, e.cfg.Files)
	if err != nil {
		return nil, err
	}

	input, err := json.Marshal(in.Facts)
	if err != nil {
		return nil, fmt.Errorf("encode the facts for opa: %w", err)
	}
	tmp, err := os.CreateTemp("", "glassops-rego-input-*.json")
	if err != nil {
		return nil, fmt.Errorf("write the opa input: %w", err)
	}
	defer os.Remove(tmp.Name())
	_, werr := tmp.Write(input)
	if cerr := tmp.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		return nil, fmt.Errorf("write the opa input: %w", werr)
	}

	ctx, cancel := context.WithTimeout(ctx, regoTimeout)
	defer cancel()
	out, runErr := RunOPA(ctx, dir, regoArgs(tmp.Name(), e.cfg.Files)...)
	if runErr != nil {
		if msg := opaErrorMessage(out); msg != "" {
			return nil, fmt.Errorf("opa eval failed: %s", redact.Text(msg))
		}
		return nil, fmt.Errorf("opa eval failed: %s", redact.Text(runErr.Error()))
	}
	entries, err := denySet(out)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return []Result{{
			Rule:    "rego",
			Outcome: Pass,
			Detail:  fmt.Sprintf("no denials from %s (sha256 %s)", strings.Join(e.cfg.Files, ", "), digest),
		}}, nil
	}
	return denials(entries)
}

// regoArgs returns the opa arguments that evaluate RegoQuery over files with input as input.
// An empty input evaluates with no input at all, which doctor uses to check the files.
func regoArgs(input string, files []string) []string {
	args := []string{"eval", "--format", "json"}
	if input != "" {
		args = append(args, "--input", input)
	}
	for _, f := range files {
		args = append(args, "--data", f)
	}
	return append(args, RegoQuery)
}

// CheckRego evaluates RegoQuery over files with no input, running opa through run, and reports
// an error unless deny is a set. A missing file, a syntax error or a wrong package name is
// found before gate. dir is where relative files resolve.
func CheckRego(ctx context.Context, run OPARunner, dir string, files []string) error {
	if len(files) == 0 {
		return errors.New("policy.rego names no files")
	}
	if _, err := regoDigest(dir, files); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, regoTimeout)
	defer cancel()
	out, err := run(ctx, dir, regoArgs("", files)...)
	if err != nil {
		if msg := opaErrorMessage(out); msg != "" {
			return fmt.Errorf("opa eval failed: %s", redact.Text(msg))
		}
		return fmt.Errorf("opa eval failed: %s", redact.Text(err.Error()))
	}
	_, err = denySet(out)
	return err
}

// regoDigest reads every file and returns a short sha256 over their names and contents, so the
// PERMIT records which policy passed. A file that cannot be read is an error.
func regoDigest(dir string, files []string) (string, error) {
	h := sha256.New()
	for _, f := range files {
		path := f
		if !filepath.IsAbs(path) && dir != "" {
			path = filepath.Join(dir, path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return "", fmt.Errorf("policy.rego file %s does not exist", f)
			}
			return "", fmt.Errorf("read policy.rego file %s: %w", f, err)
		}
		fmt.Fprintf(h, "%s\x00%d\x00", filepath.ToSlash(f), len(data))
		h.Write(data)
	}
	return hex.EncodeToString(h.Sum(nil))[:12], nil
}

// opaOutput is the part of "opa eval --format json" output the evaluator reads.
type opaOutput struct {
	Result []struct {
		Expressions []struct {
			Value json.RawMessage `json:"value"`
		} `json:"expressions"`
	} `json:"result"`
	Errors []struct {
		Message  string `json:"message"`
		Location *struct {
			File string `json:"file"`
			Row  int    `json:"row"`
		} `json:"location"`
	} `json:"errors"`
}

// opaErrorMessage returns the first error opa reported in its JSON output, or "".
func opaErrorMessage(out []byte) string {
	var o opaOutput
	if json.Unmarshal(out, &o) != nil || len(o.Errors) == 0 {
		return ""
	}
	e := o.Errors[0]
	msg := e.Message
	if e.Location != nil && e.Location.File != "" {
		msg = fmt.Sprintf("%s:%d: %s", filepath.Base(e.Location.File), e.Location.Row, msg)
	}
	if n := len(o.Errors); n > 1 {
		msg = fmt.Sprintf("%s (and %d more)", msg, n-1)
	}
	return msg
}

// denySet reads the deny set from opa's output. An undefined deny, which is what opa reports
// when no file declares package glassops with a deny rule, is an error, because it would
// otherwise pass silently on a typo.
func denySet(out []byte) ([]json.RawMessage, error) {
	var o opaOutput
	if err := json.Unmarshal(out, &o); err != nil {
		return nil, fmt.Errorf("read the opa eval output: %w", err)
	}
	if len(o.Errors) > 0 {
		return nil, fmt.Errorf("opa eval failed: %s", redact.Text(opaErrorMessage(out)))
	}
	if len(o.Result) == 0 {
		return nil, fmt.Errorf("%s is undefined: the Rego files must declare package glassops with a deny rule", RegoQuery)
	}
	if len(o.Result) != 1 || len(o.Result[0].Expressions) != 1 {
		return nil, fmt.Errorf("opa eval returned %d results, want one", len(o.Result))
	}
	var entries []json.RawMessage
	if err := json.Unmarshal(o.Result[0].Expressions[0].Value, &entries); err != nil {
		return nil, fmt.Errorf("%s is not a set of {\"rule\", \"detail\"} objects", RegoQuery)
	}
	return entries, nil
}

// denyEntry is one entry of the deny set.
type denyEntry struct {
	Rule   *string `json:"rule"`
	Detail any     `json:"detail"`
}

// denials turns deny entries into failed checks. An entry without a usable rule name adds no
// check of its own and makes the evaluation an error, which denies too.
func denials(entries []json.RawMessage) ([]Result, error) {
	results := make([]Result, 0, len(entries))
	var bad []string
	for i, raw := range entries {
		var d denyEntry
		if err := json.Unmarshal(raw, &d); err != nil || d.Rule == nil || !regoRuleName.MatchString(*d.Rule) {
			bad = append(bad, fmt.Sprintf("entry %d", i+1))
			continue
		}
		results = append(results, Result{Rule: RegoPrefix + *d.Rule, Outcome: Fail, Detail: denyDetail(d.Detail)})
	}
	if len(bad) > 0 {
		return results, fmt.Errorf("%s has entries without a valid rule name (%s): each entry must be {\"rule\": \"<name>\", \"detail\": \"<text>\"}",
			RegoQuery, strings.Join(bad, ", "))
	}
	return results, nil
}

// denyDetail returns the detail as text, with emails masked and the length capped. A detail
// that is not a string is kept as its JSON.
func denyDetail(v any) string {
	var s string
	switch d := v.(type) {
	case nil:
	case string:
		s = d
	default:
		if b, err := json.Marshal(d); err == nil {
			s = string(b)
		}
	}
	s = redact.Text(strings.TrimSpace(s))
	if s == "" {
		return "denied by Rego"
	}
	if len(s) > maxRegoDetail {
		s = strings.ToValidUTF8(s[:maxRegoDetail], "") + "..."
	}
	return s
}

// firstLine returns the first non-empty line of s, trimmed.
func firstLine(s string) string {
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			return l
		}
	}
	return ""
}
