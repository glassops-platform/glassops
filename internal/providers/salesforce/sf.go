package salesforce

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Cmd is one external program run.
type Cmd struct {
	Name string
	Args []string
	// Dir is the working directory. Empty means the current one.
	Dir string
	// Env is added to the process environment as KEY=VALUE pairs. A key here replaces the
	// same key from the process environment.
	Env []string
	// Stdin, when set, is written to the program's standard input.
	Stdin []byte
}

// Output is what a finished program produced. A non-zero exit code is not an error, because
// sf reports failures as JSON on stdout with exit 1.
type Output struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

// Exec runs c. It returns an error only when the program could not run at all or ctx ended.
type Exec func(ctx context.Context, c Cmd) (Output, error)

func osExec(ctx context.Context, c Cmd) (Output, error) {
	cmd := exec.CommandContext(ctx, c.Name, c.Args...)
	cmd.Dir = c.Dir
	cmd.Env = mergeEnv(os.Environ(), c.Env)
	if c.Stdin != nil {
		cmd.Stdin = bytes.NewReader(c.Stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	out := Output{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}
	var exitErr *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &exitErr) && ctx.Err() == nil:
		out.ExitCode = exitErr.ExitCode()
	default:
		if ctx.Err() != nil {
			return out, ctx.Err()
		}
		return out, fmt.Errorf("run %s: %w", c.Name, err)
	}
	return out, nil
}

// mergeEnv returns base with extra applied. Keys compare case-insensitively on Windows,
// where the environment is case-insensitive.
func mergeEnv(base, extra []string) []string {
	if len(extra) == 0 {
		return base
	}
	key := func(kv string) string {
		k, _, _ := strings.Cut(kv, "=")
		if isWindows {
			return strings.ToUpper(k)
		}
		return k
	}
	override := map[string]bool{}
	for _, kv := range extra {
		override[key(kv)] = true
	}
	out := make([]string, 0, len(base)+len(extra))
	for _, kv := range base {
		if !override[key(kv)] {
			out = append(out, kv)
		}
	}
	return append(out, extra...)
}

// envelope is the JSON sf prints with --json. A success has status 0 and a result. A failure
// has a non-zero status, a name and a message, and some commands put detail in data instead
// of result. For example a failed deploy validate puts only data.deployId there.
type envelope struct {
	Status   int             `json:"status"`
	Name     string          `json:"name"`
	Message  string          `json:"message"`
	Result   json.RawMessage `json:"result"`
	Data     json.RawMessage `json:"data"`
	Warnings []string        `json:"warnings"`
}

// SFError is a failure sf reported in its JSON envelope.
type SFError struct {
	Command string
	Name    string
	Message string
	Status  int
	// Data is the envelope's data field, kept for callers that need detail from it.
	Data json.RawMessage
}

func (e *SFError) Error() string {
	msg := strings.TrimSpace(e.Message)
	if e.Name != "" {
		return fmt.Sprintf("sf %s: %s: %s", e.Command, e.Name, msg)
	}
	return fmt.Sprintf("sf %s: %s", e.Command, msg)
}

// parseEnvelope reads sf's --json output. sf can print a prompt or other text before the
// JSON, so parsing starts at the first brace.
func parseEnvelope(stdout []byte) (envelope, error) {
	var env envelope
	i := bytes.IndexByte(stdout, '{')
	if i < 0 {
		return env, errors.New("sf printed no JSON")
	}
	if err := json.Unmarshal(stdout[i:], &env); err != nil {
		return env, fmt.Errorf("could not read sf JSON output: %w", err)
	}
	return env, nil
}

// sf runs "sf <args> --json" in session s (nil for no session) from dir and parses the
// envelope. A non-zero status comes back as *SFError together with the envelope.
func (p *Provider) sf(ctx context.Context, s *session, dir string, args ...string) (envelope, error) {
	full := append([]string{}, args...)
	if !contains(full, "--json") {
		full = append(full, "--json")
	}
	c := Cmd{Name: "sf", Args: full, Dir: dir, Env: sfBaseEnv()}
	if s != nil {
		c.Env = append(c.Env, s.env()...)
	}
	out, err := p.exec(ctx, c)
	if err != nil {
		return envelope{}, err
	}
	env, perr := parseEnvelope(out.Stdout)
	if perr != nil {
		return env, fmt.Errorf("sf %s (exit %d): %w", commandName(args), out.ExitCode, perr)
	}
	if env.Status != 0 {
		return env, &SFError{Command: commandName(args), Name: env.Name, Message: env.Message, Status: env.Status, Data: env.Data}
	}
	return env, nil
}

// sfBaseEnv keeps sf quiet and predictable in every run.
func sfBaseEnv() []string {
	return []string{
		"SF_DISABLE_TELEMETRY=true",
		"SF_AUTOUPDATE_DISABLE=true",
		"SF_SKIP_NEW_VERSION_CHECK=true",
		"FORCE_COLOR=0",
	}
}

// commandName is the leading words of args before the first flag, such as "org display".
func commandName(args []string) string {
	var words []string
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			break
		}
		words = append(words, a)
	}
	return strings.Join(words, " ")
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
