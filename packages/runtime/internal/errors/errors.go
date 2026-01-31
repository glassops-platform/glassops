// Package errors provides structured error types for the GlassOps runtime.
// This enables better error categorization for telemetry and GitHub Actions.
package errors

import (
	"fmt"
)

// GlassOpsError is the base error type for all runtime errors.
type GlassOpsError struct {
	Message string
	Phase   string
	Code    string
	Cause   error
}

func (e *GlassOpsError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *GlassOpsError) Unwrap() error {
	return e.Cause
}

// PolicyError indicates a governance policy violation.
type PolicyError struct {
	GlassOpsError
}

// NewPolicyError creates a new policy violation error.
func NewPolicyError(message string, cause error) *PolicyError {
	return &PolicyError{
		GlassOpsError: GlassOpsError{
			Message: message,
			Phase:   "Policy",
			Code:    "POLICY_VIOLATION",
			Cause:   cause,
		},
	}
}

// BootstrapError indicates a CLI bootstrap failure.
type BootstrapError struct {
	GlassOpsError
}

// NewBootstrapError creates a new bootstrap error.
func NewBootstrapError(message string, cause error) *BootstrapError {
	return &BootstrapError{
		GlassOpsError: GlassOpsError{
			Message: message,
			Phase:   "Bootstrap",
			Code:    "BOOTSTRAP_FAILED",
			Cause:   cause,
		},
	}
}

// IdentityError indicates an authentication failure.
type IdentityError struct {
	GlassOpsError
}

// NewIdentityError creates a new identity/authentication error.
func NewIdentityError(message string, cause error) *IdentityError {
	return &IdentityError{
		GlassOpsError: GlassOpsError{
			Message: message,
			Phase:   "Identity",
			Code:    "AUTHENTICATION_FAILED",
			Cause:   cause,
		},
	}
}

// ContractError indicates a contract generation or validation failure.
type ContractError struct {
	GlassOpsError
}

// NewContractError creates a new contract error.
func NewContractError(message string, cause error) *ContractError {
	return &ContractError{
		GlassOpsError: GlassOpsError{
			Message: message,
			Phase:   "Contract",
			Code:    "CONTRACT_GENERATION_FAILED",
			Cause:   cause,
		},
	}
}

// AnalyzerError indicates a code analysis failure.
type AnalyzerError struct {
	GlassOpsError
}

// NewAnalyzerError creates a new analyzer error.
func NewAnalyzerError(message string, cause error) *AnalyzerError {
	return &AnalyzerError{
		GlassOpsError: GlassOpsError{
			Message: message,
			Phase:   "Analyzer",
			Code:    "ANALYSIS_FAILED",
			Cause:   cause,
		},
	}
}

// FreezeError indicates a deployment is blocked by a freeze window.
type FreezeError struct {
	GlassOpsError
	Day   string
	Start string
	End   string
}

// NewFreezeError creates a new freeze window error.
func NewFreezeError(day, start, end string) *FreezeError {
	return &FreezeError{
		GlassOpsError: GlassOpsError{
			Message: fmt.Sprintf("Deployment blocked by governance window (%s %s-%s)", day, start, end),
			Phase:   "Policy",
			Code:    "FROZEN",
			Cause:   nil,
		},
		Day:   day,
		Start: start,
		End:   end,
	}
}

// IsGlassOpsError checks if an error is a GlassOps error.
func IsGlassOpsError(err error) bool {
	_, ok := err.(*GlassOpsError)
	if ok {
		return true
	}
	// Check embedded types
	switch err.(type) {
	case *PolicyError, *BootstrapError, *IdentityError, *ContractError, *AnalyzerError, *FreezeError:
		return true
	}
	return false
}

// GetPhase extracts the phase from a GlassOps error.
func GetPhase(err error) string {
	if e, ok := err.(*GlassOpsError); ok {
		return e.Phase
	}
	switch e := err.(type) {
	case *PolicyError:
		return e.Phase
	case *BootstrapError:
		return e.Phase
	case *IdentityError:
		return e.Phase
	case *ContractError:
		return e.Phase
	case *AnalyzerError:
		return e.Phase
	case *FreezeError:
		return e.Phase
	}
	return "Unknown"
}

// GetCode extracts the error code from a GlassOps error.
func GetCode(err error) string {
	if e, ok := err.(*GlassOpsError); ok {
		return e.Code
	}
	switch e := err.(type) {
	case *PolicyError:
		return e.Code
	case *BootstrapError:
		return e.Code
	case *IdentityError:
		return e.Code
	case *ContractError:
		return e.Code
	case *AnalyzerError:
		return e.Code
	case *FreezeError:
		return e.Code
	}
	return "UNKNOWN_ERROR"
}
