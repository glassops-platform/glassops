// Package exitcode defines the process exit codes every glassops command uses.
package exitcode

const (
	// OK means the command passed.
	OK = 0
	// Fail means a policy or check failed. The tool itself worked.
	Fail = 1
	// Error means a tool or infrastructure problem stopped the command.
	Error = 2
)

// Worst returns the most severe of the given codes: Error over Fail over OK.
func Worst(codes ...int) int {
	worst := OK
	for _, c := range codes {
		if c > worst {
			worst = c
		}
	}
	return worst
}

// ExitError carries an exit code through cobra's error return.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error { return e.Err }
