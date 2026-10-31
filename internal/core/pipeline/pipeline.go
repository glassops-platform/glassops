// Package pipeline turns a named pipeline from glassops.yaml into steps and runs them in
// order. A step is either a glassops command written as one string, such as
// "validate --target ephemeral", or an exec step that runs a shell command. Each step has a
// when condition that decides whether it runs after an earlier step failed.
package pipeline

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/glassops-platform/glassops/internal/core/config"
)

// When decides whether a step runs, given what happened before it.
type When string

const (
	// Success runs the step only when no earlier step failed. It is the default.
	Success When = "success"
	// Failure runs the step only when an earlier step failed.
	Failure When = "failure"
	// Always runs the step whatever happened before, including after a cancel.
	Always When = "always"
)

// Step is one resolved pipeline entry.
type Step struct {
	// Index is the step's 1-based position in the pipeline.
	Index int
	// Name is the command name, such as "plan" or "env", or "exec" for an exec step.
	Name string
	// Argv is the glassops command line without the program name. It is nil for an exec step.
	Argv []string
	// Exec is the shell command of an exec step.
	Exec string
	// ExecN is the exec step's 1-based count among the pipeline's exec steps, used to name
	// its result exec.<n>.json. It is 0 for a glassops step.
	ExecN int
	// When is the step's condition, never empty.
	When When
	// Raw is the step as written in the config.
	Raw string
}

// IsExec reports whether the step runs a shell command instead of a glassops command.
func (s Step) IsExec() bool { return s.Name == config.ExecStep }

// ErrUnknownPipeline means the config has no pipeline with the requested name.
var ErrUnknownPipeline = errors.New("unknown pipeline")

// Resolve returns the steps of the named pipeline.
func Resolve(cfg *config.Config, name string) ([]Step, error) {
	steps, ok := cfg.Pipelines[name]
	if !ok {
		names := make([]string, 0, len(cfg.Pipelines))
		for n := range cfg.Pipelines {
			names = append(names, n)
		}
		sort.Strings(names)
		known := "none are configured"
		if len(names) > 0 {
			known = "configured: " + strings.Join(names, ", ")
		}
		return nil, fmt.Errorf("%w %q, %s", ErrUnknownPipeline, name, known)
	}
	return Parse(steps)
}

// Parse turns config steps into resolved steps. It returns every problem it finds, so a bad
// pipeline is reported in one go.
func Parse(steps []config.Step) ([]Step, error) {
	var out []Step
	var problems []string
	execN := 0
	for i, cs := range steps {
		s, err := parseStep(cs)
		if err != nil {
			problems = append(problems, fmt.Sprintf("step %d (%q): %v", i+1, cs.Run, err))
			continue
		}
		s.Index = i + 1
		if s.IsExec() {
			execN++
			s.ExecN = execN
		}
		out = append(out, s)
	}
	if len(problems) > 0 {
		return nil, errors.New(strings.Join(problems, "; "))
	}
	if len(out) == 0 {
		return nil, errors.New("the pipeline has no steps")
	}
	return out, nil
}

func parseStep(cs config.Step) (Step, error) {
	when, err := parseWhen(cs.When)
	if err != nil {
		return Step{}, err
	}
	argv, err := Split(cs.Run)
	if err != nil {
		return Step{}, err
	}
	if len(argv) == 0 {
		return Step{}, errors.New("empty step")
	}
	s := Step{Name: argv[0], When: when, Raw: cs.Run}
	switch {
	case argv[0] == config.ExecStep:
		if len(argv) > 1 {
			return Step{}, errors.New("an exec step takes its command from run, write {step: exec, run: <command>}")
		}
		if strings.TrimSpace(cs.Exec) == "" {
			return Step{}, errors.New("an exec step needs run")
		}
		s.Exec = cs.Exec
		return s, nil
	case cs.Exec != "":
		return Step{}, errors.New("run is only for exec steps")
	case strings.HasPrefix(argv[0], "-"):
		return Step{}, errors.New("a step starts with a command name, not a flag")
	case argv[0] == "run":
		// A pipeline that runs a pipeline could run itself forever.
		return Step{}, errors.New("run cannot be a pipeline step")
	}
	s.Argv = argv
	return s, nil
}

func parseWhen(w string) (When, error) {
	switch When(w) {
	case "", Success:
		return Success, nil
	case Failure, Always:
		return When(w), nil
	}
	return "", fmt.Errorf("when %q is not success, failure or always", w)
}

// Split breaks a step string into arguments the way a POSIX shell would for plain words and
// quotes, without running a shell and without expanding anything. Whitespace separates
// arguments. Single quotes keep everything literal. Double quotes keep everything literal
// except \" and \\. Outside quotes a backslash escapes a following quote, backslash or
// space and is kept as written before anything else, so a Windows path stays intact.
func Split(s string) ([]string, error) {
	var args []string
	var cur strings.Builder
	inArg := false
	rs := []rune(s)
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		switch {
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
			if inArg {
				args = append(args, cur.String())
				cur.Reset()
				inArg = false
			}
		case r == '\'':
			inArg = true
			end := indexRune(rs, i+1, '\'')
			if end < 0 {
				return nil, errors.New("unterminated single quote")
			}
			cur.WriteString(string(rs[i+1 : end]))
			i = end
		case r == '"':
			inArg = true
			closed := false
			for i++; i < len(rs); i++ {
				c := rs[i]
				if c == '\\' && i+1 < len(rs) && (rs[i+1] == '"' || rs[i+1] == '\\') {
					cur.WriteRune(rs[i+1])
					i++
					continue
				}
				if c == '"' {
					closed = true
					break
				}
				cur.WriteRune(c)
			}
			if !closed {
				return nil, errors.New("unterminated double quote")
			}
		case r == '\\':
			inArg = true
			if i+1 < len(rs) && strings.ContainsRune("\"'\\ \t", rs[i+1]) {
				cur.WriteRune(rs[i+1])
				i++
				continue
			}
			cur.WriteRune(r)
		default:
			inArg = true
			cur.WriteRune(r)
		}
	}
	if inArg {
		args = append(args, cur.String())
	}
	return args, nil
}

func indexRune(rs []rune, from int, r rune) int {
	for i := from; i < len(rs); i++ {
		if rs[i] == r {
			return i
		}
	}
	return -1
}
