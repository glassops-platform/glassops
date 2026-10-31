package pipeline

import (
	"context"
	"time"

	"github.com/glassops-platform/glassops/internal/core/exitcode"
)

// DefaultCleanupTimeout is how long a when: always step may run after the pipeline was
// cancelled. It is short because the CI is about to kill the job.
const DefaultCleanupTimeout = 2 * time.Minute

// Skip reasons recorded for steps that did not run.
const (
	SkipEarlierFailed = "an earlier step failed"
	SkipNoFailure     = "no earlier step failed"
	SkipEmptyPlan     = "the plan is empty"
	SkipCancelled     = "the run was cancelled"
)

// Record is what happened to one step. It is written to run.json.
type Record struct {
	Index int    `json:"index"`
	Step  string `json:"step"`
	When  When   `json:"when"`
	// Argv is the full glassops command line the step ran with, global flags included.
	Argv []string `json:"argv,omitempty"`
	// Run is an exec step's shell command.
	Run string `json:"run,omitempty"`
	// Result is the result file the step wrote, when the runner knows it.
	Result string `json:"result,omitempty"`
	// ExitCode is nil for a skipped step.
	ExitCode   *int      `json:"exit_code,omitempty"`
	Started    time.Time `json:"started,omitempty"`
	DurationMS int64     `json:"duration_ms"`
	Skipped    bool      `json:"skipped,omitempty"`
	SkipReason string    `json:"skip_reason,omitempty"`
	// Error is a problem the runner itself hit after the step, such as an unreadable plan.
	Error string `json:"error,omitempty"`
}

// Runner runs resolved steps. Command, Exec and PlanEmpty are required.
type Runner struct {
	// Command runs a glassops command line and returns its exit code.
	Command func(ctx context.Context, argv []string) int
	// Globals are prepended to every glassops step's argv, so every step reads the same
	// project and writes to the same results dir as the run.
	Globals []string
	// Exec runs an exec step and returns its exit code and the result file it wrote.
	Exec func(ctx context.Context, s Step) (code int, result string)
	// PlanEmpty reports whether the plan the plan step just wrote is empty. It is called
	// only after a plan step that exited 0.
	PlanEmpty func() (bool, error)
	// Before is called before each step runs, and Skipped for each step that does not.
	// Both may be nil.
	Before  func(s Step)
	Skipped func(s Step, reason string)
	// Now is the clock. Nil means time.Now.
	Now func() time.Time
	// CleanupTimeout bounds each when: always step that runs after a cancel. Zero means
	// DefaultCleanupTimeout.
	CleanupTimeout time.Duration
}

// Outcome is the whole run.
type Outcome struct {
	Records []Record
	// Exit is the worst exit code of every step: 2 over 1 over 0. A cancelled run is at
	// least 2, and so is a plan the runner could not read.
	Exit int
	// Cancelled is true when ctx was cancelled before the last step.
	Cancelled bool
	// EmptyPlan is true when a plan step wrote an empty plan.
	EmptyPlan bool
}

// Run runs steps in order. A step whose condition does not hold is recorded as skipped.
// After a plan step writes an empty plan, every later step without when: always is skipped.
// When ctx is cancelled the current step stops through ctx, and only later when: always
// steps run, each with a fresh context bounded by CleanupTimeout, so cleanup such as
// env delete still happens.
func (r *Runner) Run(ctx context.Context, steps []Step) Outcome {
	now := r.Now
	if now == nil {
		now = time.Now
	}
	cleanup := r.CleanupTimeout
	if cleanup <= 0 {
		cleanup = DefaultCleanupTimeout
	}

	var out Outcome
	failed := false
	for _, s := range steps {
		rec := Record{Index: s.Index, Step: s.Name, When: s.When, Run: s.Exec}
		if ctx.Err() != nil {
			out.Cancelled = true
		}
		if reason := skipReason(s.When, failed, out.EmptyPlan, out.Cancelled); reason != "" {
			rec.Skipped = true
			rec.SkipReason = reason
			if r.Skipped != nil {
				r.Skipped(s, reason)
			}
			out.Records = append(out.Records, rec)
			continue
		}

		stepCtx, cancel := ctx, context.CancelFunc(func() {})
		if out.Cancelled {
			stepCtx, cancel = context.WithTimeout(context.Background(), cleanup)
		}
		if r.Before != nil {
			r.Before(s)
		}
		started := now()
		var code int
		if s.IsExec() {
			code, rec.Result = r.Exec(stepCtx, s)
		} else {
			rec.Argv = append(append([]string{}, r.Globals...), s.Argv...)
			code = r.Command(stepCtx, rec.Argv)
		}
		cancel()
		rec.Started = started.UTC()
		rec.DurationMS = now().Sub(started).Milliseconds()
		rec.ExitCode = &code

		if code != exitcode.OK {
			failed = true
		}
		out.Exit = worst(out.Exit, code)

		if s.Name == "plan" && code == exitcode.OK && !out.Cancelled {
			empty, err := r.PlanEmpty()
			switch {
			case err != nil:
				// Without a readable plan nothing after it can be trusted, so this counts
				// as a failed step, never as an empty plan.
				rec.Error = err.Error()
				failed = true
				out.Exit = worst(out.Exit, exitcode.Error)
			case empty:
				out.EmptyPlan = true
			}
		}
		out.Records = append(out.Records, rec)
	}
	if ctx.Err() != nil {
		out.Cancelled = true
	}
	if out.Cancelled {
		out.Exit = worst(out.Exit, exitcode.Error)
	}
	return out
}

func skipReason(w When, failed, emptyPlan, cancelled bool) string {
	if w == Always {
		return ""
	}
	switch {
	case cancelled:
		return SkipCancelled
	case emptyPlan:
		return SkipEmptyPlan
	case w == Failure && !failed:
		return SkipNoFailure
	case w == Success && failed:
		return SkipEarlierFailed
	}
	return ""
}

// worst ranks exit codes 2 over 1 over 0. Any other non-zero code counts as 2.
func worst(a, b int) int {
	return max(rank(a), rank(b))
}

func rank(code int) int {
	switch code {
	case exitcode.OK, exitcode.Fail:
		return code
	}
	return exitcode.Error
}
