package exec

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/glassops-platform/glassops/internal/providers"
)

// The variables every exec command gets.
const (
	EnvPermit     = "GLASSOPS_PERMIT"
	EnvPlan       = "GLASSOPS_PLAN"
	EnvPlanDir    = "GLASSOPS_PLAN_DIR"
	EnvTarget     = "GLASSOPS_TARGET"
	EnvTargetID   = "GLASSOPS_TARGET_ID"
	EnvResultsDir = "GLASSOPS_RESULTS_DIR"
	// EnvFrom and EnvTo are the plan's git refs, set for exec.plan only.
	EnvFrom = "GLASSOPS_FROM"
	EnvTo   = "GLASSOPS_TO"
	// EnvDeployID is the deploy's Id from its result, set for exec.verify only.
	EnvDeployID = "GLASSOPS_DEPLOY_ID"
)

// permitKeyEnv is the PERMIT signing key. It is removed from every command's environment,
// because a deploy tool that can read it could sign its own PERMIT. It equals permit.KeyEnv.
const permitKeyEnv = "GLASSOPS_PERMIT_KEY"

// ResultSchema is the schema value of an exec-<step>.json result file.
const ResultSchema = "glassops.engine-result/v1"

// The steps an exec command runs for, as the result file names them.
const (
	stepPlan     = "plan"
	stepValidate = "validate"
	stepDeploy   = "deploy"
	stepVerify   = "verify"
)

// ResultFile is the name of the file a command for step may write in the results dir.
func ResultFile(step string) string { return "exec-" + step + ".json" }

// engineResult is schemas/engine-result.json.
type engineResult struct {
	Schema             string         `json:"schema"`
	Step               string         `json:"step"`
	Passed             *bool          `json:"passed"`
	ID                 string         `json:"id"`
	ComponentsDeployed *int           `json:"components_deployed"`
	Problems           []string       `json:"problems"`
	Facts              map[string]any `json:"facts"`
}

// outcome is what one command run decided.
type outcome struct {
	passed     bool
	id         string
	components int
	problems   []string
	facts      map[string]any
	summary    string
}

// resultsDir is the absolute results dir, created if needed.
func (p *Provider) resultsDir(dir string) (string, error) {
	rd := p.ws.ResultsDir
	if rd == "" {
		// The same default as the CLI's --results-dir.
		rd = filepath.Join(dir, ".glassops", "results")
	}
	abs, err := filepath.Abs(rd)
	if err != nil {
		return "", fmt.Errorf("results dir: %w", err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return "", fmt.Errorf("results dir: %w", err)
	}
	return abs, nil
}

// commandEnv is the environment a command runs with: the process environment without the
// PERMIT key, then the GLASSOPS_* variables. t is nil when there is no target, as for plan.
// planDir is the change set's dir relative to dir, or "" when there is none yet.
func (p *Provider) commandEnv(dir string, t *providers.Target, planDir string, extra []string) ([]string, error) {
	results, err := p.resultsDir(dir)
	if err != nil {
		return nil, err
	}
	set := map[string]string{EnvResultsDir: results}
	if t != nil {
		set[EnvTarget] = t.Name
		set[EnvTargetID] = t.ID
	}
	if planDir != "" {
		abs, err := filepath.Abs(filepath.Join(dir, filepath.FromSlash(planDir)))
		if err != nil {
			return nil, fmt.Errorf("plan dir: %w", err)
		}
		set[EnvPlanDir] = abs
	}
	for name, pair := range map[string]struct{ bound, def string }{
		EnvPlan:   {p.ws.PlanPath, filepath.Join(dir, ".glassops", "plan.json")},
		EnvPermit: {p.ws.PermitPath, filepath.Join(dir, ".glassops", "permit.json")},
	} {
		path := pair.bound
		if path == "" {
			if _, err := os.Stat(pair.def); err != nil {
				continue
			}
			path = pair.def
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		set[name] = abs
	}
	for _, kv := range extra {
		if k, v, ok := strings.Cut(kv, "="); ok {
			set[k] = v
		}
	}

	var env []string
	for _, kv := range p.environ() {
		k, _, _ := strings.Cut(kv, "=")
		if _, ours := set[k]; ours || strings.EqualFold(k, permitKeyEnv) {
			continue
		}
		env = append(env, kv)
	}
	for _, k := range sortedKeys(set) {
		env = append(env, k+"="+set[k])
	}
	return env, nil
}

// step runs command for step with sh -c in dir and reads its outcome: the exit code and,
// when the command wrote one, its result file. A result file left by an earlier run is
// removed first, so it can never stand in for this run. Only a zero exit with a result file
// that does not say otherwise passes. A missing command, a command that cannot run and a
// result file that does not match the schema are errors.
func (p *Provider) step(ctx context.Context, step, command, dir string, t *providers.Target, planDir string, extra []string) (outcome, error) {
	if strings.TrimSpace(command) == "" {
		return outcome{}, fmt.Errorf("exec.%s is not set, so there is nothing to run for %s", step, step)
	}
	if dir == "" {
		dir = p.dir()
	}
	if dir == "" {
		return outcome{}, fmt.Errorf("exec provider: the project dir is not bound, so exec.%s cannot run", step)
	}
	env, err := p.commandEnv(dir, t, planDir, extra)
	if err != nil {
		return outcome{}, err
	}
	results, err := p.resultsDir(dir)
	if err != nil {
		return outcome{}, err
	}
	path := filepath.Join(results, ResultFile(step))
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return outcome{}, fmt.Errorf("remove the earlier %s: %w", ResultFile(step), err)
	}

	code, err := p.shell(ctx, dir, env, command, p.stdout(), p.stderr())
	if err != nil {
		return outcome{}, fmt.Errorf("exec.%s: %w", step, err)
	}

	out := outcome{facts: map[string]any{}}
	r, found, err := readResult(path, step)
	if err != nil {
		return outcome{}, err
	}
	for k, v := range r.Facts {
		out.facts[k] = v
	}
	out.facts["exec.exit_code"] = code
	out.facts["exec.result_file"] = found
	out.passed = code == 0
	out.summary = fmt.Sprintf("exec.%s exit %d", step, code)
	if !found {
		out.summary += ", no result file"
		if code != 0 {
			out.problems = append(out.problems, fmt.Sprintf("exec.%s exited %d", step, code))
		}
		return out, nil
	}
	out.id = r.ID
	if r.ComponentsDeployed != nil {
		out.components = *r.ComponentsDeployed
	}
	out.problems = append(out.problems, r.Problems...)
	out.summary += fmt.Sprintf(", result file passed %t", *r.Passed)
	switch {
	case code != 0 && *r.Passed:
		out.problems = append(out.problems, fmt.Sprintf("exec.%s exited %d but its result file says passed, so it counts as failed", step, code))
	case code != 0 && len(r.Problems) == 0:
		out.problems = append(out.problems, fmt.Sprintf("exec.%s exited %d", step, code))
	case code == 0 && !*r.Passed && len(r.Problems) == 0:
		out.problems = append(out.problems, fmt.Sprintf("exec.%s exited 0 but its result file says not passed", step))
	}
	out.passed = code == 0 && *r.Passed
	return out, nil
}

// readResult reads and checks the result file at path. found is false when there is none.
func readResult(path, step string) (engineResult, bool, error) {
	var r engineResult
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return r, false, nil
	}
	name := ResultFile(step)
	if err != nil {
		return r, false, fmt.Errorf("read %s: %w", name, err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&r); err != nil {
		return r, false, fmt.Errorf("%s does not match schemas/engine-result.json: %w", name, err)
	}
	if dec.More() {
		return r, false, fmt.Errorf("%s holds more than one JSON value", name)
	}
	switch {
	case r.Schema != ResultSchema:
		return r, false, fmt.Errorf("%s: schema is %q, want %q", name, r.Schema, ResultSchema)
	case r.Step != step:
		return r, false, fmt.Errorf("%s: step is %q, want %q", name, r.Step, step)
	case r.Passed == nil:
		return r, false, fmt.Errorf("%s: passed is missing", name)
	case r.ComponentsDeployed != nil && *r.ComponentsDeployed < 0:
		return r, false, fmt.Errorf("%s: components_deployed is negative", name)
	}
	for k := range r.Facts {
		if !strings.HasPrefix(k, "exec.") {
			return r, false, fmt.Errorf("%s: fact %q does not start with exec.", name, k)
		}
	}
	return r, true, nil
}
