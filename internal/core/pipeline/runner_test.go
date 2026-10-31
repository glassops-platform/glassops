package pipeline

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
)

// fakeRun is a synthetic runner: each step's exit code comes from codes by step name, and
// every call is recorded with whether its context was already cancelled.
type fakeRun struct {
	codes     map[string]int
	empty     bool
	emptyErr  error
	calls     []string
	argvs     [][]string
	cancelled []bool
	onCall    func(name string)
}

func (f *fakeRun) runner() *Runner {
	return &Runner{
		Globals: []string{"--dir=/p"},
		Command: func(ctx context.Context, argv []string) int {
			name := argv[1]
			f.calls = append(f.calls, name)
			f.argvs = append(f.argvs, argv)
			f.cancelled = append(f.cancelled, ctx.Err() != nil)
			if f.onCall != nil {
				f.onCall(name)
			}
			return f.codes[name]
		},
		Exec: func(ctx context.Context, s Step) (int, string) {
			name := "exec:" + s.Exec
			f.calls = append(f.calls, name)
			f.cancelled = append(f.cancelled, ctx.Err() != nil)
			if f.onCall != nil {
				f.onCall(name)
			}
			return f.codes[name], "/r/exec.json"
		},
		PlanEmpty: func() (bool, error) { return f.empty, f.emptyErr },
	}
}

func mustParse(t *testing.T, steps ...config.Step) []Step {
	t.Helper()
	s, err := Parse(steps)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestRunWhenSemantics(t *testing.T) {
	steps := []config.Step{
		{Run: "plan"},
		{Run: "check"},
		{Run: "analyze"},
		{Run: "notify", When: "failure"},
		{Run: "report", When: "always"},
	}
	cases := []struct {
		name      string
		codes     map[string]int
		wantCalls []string
		wantSkip  map[string]string
		wantExit  int
	}{
		{
			name:      "all pass",
			codes:     map[string]int{},
			wantCalls: []string{"plan", "check", "analyze", "report"},
			wantSkip:  map[string]string{"notify": SkipNoFailure},
			wantExit:  0,
		},
		{
			name:      "check fails",
			codes:     map[string]int{"check": 1},
			wantCalls: []string{"plan", "check", "notify", "report"},
			wantSkip:  map[string]string{"analyze": SkipEarlierFailed},
			wantExit:  1,
		},
		{
			name:      "error beats a later fail",
			codes:     map[string]int{"plan": 2, "notify": 1},
			wantCalls: []string{"plan", "notify", "report"},
			wantSkip:  map[string]string{"check": SkipEarlierFailed, "analyze": SkipEarlierFailed},
			wantExit:  2,
		},
		{
			name:      "fail then error in always",
			codes:     map[string]int{"analyze": 1, "report": 2},
			wantCalls: []string{"plan", "check", "analyze", "notify", "report"},
			wantSkip:  map[string]string{},
			wantExit:  2,
		},
		{
			name:      "odd exit code counts as error",
			codes:     map[string]int{"check": 3},
			wantCalls: []string{"plan", "check", "notify", "report"},
			wantSkip:  map[string]string{"analyze": SkipEarlierFailed},
			wantExit:  2,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakeRun{codes: tc.codes}
			out := f.runner().Run(context.Background(), mustParse(t, steps...))
			if !reflect.DeepEqual(f.calls, tc.wantCalls) {
				t.Fatalf("calls = %v, want %v", f.calls, tc.wantCalls)
			}
			if out.Exit != tc.wantExit {
				t.Fatalf("exit = %d, want %d", out.Exit, tc.wantExit)
			}
			skips := map[string]string{}
			for _, r := range out.Records {
				if r.Skipped {
					skips[r.Step] = r.SkipReason
					if r.ExitCode != nil {
						t.Fatalf("skipped %s has an exit code", r.Step)
					}
				} else if r.ExitCode == nil || *r.ExitCode != tc.codes[r.Step] {
					t.Fatalf("record %+v", r)
				}
			}
			if !reflect.DeepEqual(skips, tc.wantSkip) {
				t.Fatalf("skips = %v, want %v", skips, tc.wantSkip)
			}
			if len(out.Records) != len(steps) || out.Cancelled || out.EmptyPlan {
				t.Fatalf("outcome %+v", out)
			}
		})
	}
}

func TestRunPrependsGlobals(t *testing.T) {
	f := &fakeRun{codes: map[string]int{}}
	out := f.runner().Run(context.Background(), mustParse(t, config.Step{Run: "validate --target 'my org'"}))
	want := []string{"--dir=/p", "validate", "--target", "my org"}
	if !reflect.DeepEqual(f.argvs[0], want) || !reflect.DeepEqual(out.Records[0].Argv, want) {
		t.Fatalf("argv = %q, record %q", f.argvs[0], out.Records[0].Argv)
	}
}

func TestRunEmptyPlanSkipsAllButAlways(t *testing.T) {
	f := &fakeRun{codes: map[string]int{}, empty: true}
	out := f.runner().Run(context.Background(), mustParse(t,
		config.Step{Run: "plan"},
		config.Step{Run: "validate"},
		config.Step{Run: "notify", When: "failure"},
		config.Step{Run: config.ExecStep, Exec: "publish"},
		config.Step{Run: "env delete", When: "always"},
	))
	if !reflect.DeepEqual(f.calls, []string{"plan", "env"}) {
		t.Fatalf("calls = %v", f.calls)
	}
	if out.Exit != 0 || !out.EmptyPlan {
		t.Fatalf("outcome %+v", out)
	}
	for _, i := range []int{1, 2, 3} {
		if r := out.Records[i]; !r.Skipped || r.SkipReason != SkipEmptyPlan {
			t.Fatalf("record %d = %+v", i, r)
		}
	}
}

func TestRunPlanCheckedOnlyAfterPass(t *testing.T) {
	f := &fakeRun{codes: map[string]int{"plan": 2}, empty: true}
	out := f.runner().Run(context.Background(), mustParse(t, config.Step{Run: "plan"}, config.Step{Run: "notify", When: "failure"}))
	if out.EmptyPlan || !reflect.DeepEqual(f.calls, []string{"plan", "notify"}) || out.Exit != 2 {
		t.Fatalf("a failed plan was read as empty: %+v, calls %v", out, f.calls)
	}
}

func TestRunUnreadablePlanIsAnError(t *testing.T) {
	f := &fakeRun{codes: map[string]int{}, emptyErr: errors.New("no plan.json")}
	out := f.runner().Run(context.Background(), mustParse(t,
		config.Step{Run: "plan"},
		config.Step{Run: "deploy"},
		config.Step{Run: "report", When: "always"},
	))
	if out.Exit != 2 || out.EmptyPlan {
		t.Fatalf("outcome %+v", out)
	}
	if !reflect.DeepEqual(f.calls, []string{"plan", "report"}) {
		t.Fatalf("calls = %v", f.calls)
	}
	if !strings.Contains(out.Records[0].Error, "no plan.json") || out.Records[1].SkipReason != SkipEarlierFailed {
		t.Fatalf("records %+v", out.Records)
	}
}

func TestRunCancelStillRunsAlways(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f := &fakeRun{codes: map[string]int{"validate": 2}}
	// validate is the step running when SIGTERM arrives. It sees the cancel and stops.
	f.onCall = func(name string) {
		if name == "validate" {
			cancel()
		}
	}
	r := f.runner()
	r.CleanupTimeout = time.Second
	out := r.Run(ctx, mustParse(t,
		config.Step{Run: "plan"},
		config.Step{Run: "validate"},
		config.Step{Run: "test"},
		config.Step{Run: "notify", When: "failure"},
		config.Step{Run: config.ExecStep, Exec: "cleanup", When: "always"},
		config.Step{Run: "env delete", When: "always"},
	))
	if !reflect.DeepEqual(f.calls, []string{"plan", "validate", "exec:cleanup", "env"}) {
		t.Fatalf("calls = %v", f.calls)
	}
	// The always steps got a live context, not the cancelled one.
	if !reflect.DeepEqual(f.cancelled, []bool{false, false, false, false}) {
		t.Fatalf("cancelled = %v", f.cancelled)
	}
	if !out.Cancelled || out.Exit != 2 {
		t.Fatalf("outcome %+v", out)
	}
	for _, i := range []int{2, 3} {
		if out.Records[i].SkipReason != SkipCancelled {
			t.Fatalf("record %d = %+v", i, out.Records[i])
		}
	}
	if out.Records[4].Result != "/r/exec.json" {
		t.Fatalf("exec record %+v", out.Records[4])
	}
}

func TestRunCancelBetweenPassingStepsIsAnError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f := &fakeRun{codes: map[string]int{}}
	f.onCall = func(string) { cancel() }
	out := f.runner().Run(ctx, mustParse(t, config.Step{Run: "plan"}, config.Step{Run: "deploy"}))
	if out.Exit != 2 || !out.Cancelled || !reflect.DeepEqual(f.calls, []string{"plan"}) {
		t.Fatalf("outcome %+v, calls %v", out, f.calls)
	}
}

func TestRunCleanupContextIsBounded(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var deadline time.Time
	var ok bool
	r := &Runner{
		Command: func(ctx context.Context, argv []string) int {
			deadline, ok = ctx.Deadline()
			return 0
		},
		PlanEmpty:      func() (bool, error) { return false, nil },
		CleanupTimeout: 5 * time.Second,
	}
	start := time.Now()
	r.Run(ctx, mustParse(t, config.Step{Run: "env delete", When: "always"}))
	if !ok || deadline.Sub(start) > 6*time.Second || deadline.Before(start) {
		t.Fatalf("cleanup deadline %v (set %v)", deadline, ok)
	}
}

func TestRunRecordsDuration(t *testing.T) {
	clock := time.Date(2026, 10, 31, 9, 0, 0, 0, time.UTC)
	r := &Runner{
		Command: func(context.Context, []string) int {
			clock = clock.Add(1500 * time.Millisecond)
			return 0
		},
		PlanEmpty: func() (bool, error) { return false, nil },
		Now:       func() time.Time { return clock },
	}
	out := r.Run(context.Background(), mustParse(t, config.Step{Run: "check"}))
	if out.Records[0].DurationMS != 1500 || !out.Records[0].Started.Equal(time.Date(2026, 10, 31, 9, 0, 0, 0, time.UTC)) {
		t.Fatalf("record %+v", out.Records[0])
	}
}
