package errors

import (
	"errors"
	"testing"
)

func TestGlassOpsError(t *testing.T) {
	t.Run("basic error", func(t *testing.T) {
		err := &GlassOpsError{
			Message: "something went wrong",
			Phase:   "Test",
			Code:    "TEST_ERROR",
		}

		expected := "[TEST_ERROR] something went wrong"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("error with cause", func(t *testing.T) {
		cause := errors.New("underlying issue")
		err := &GlassOpsError{
			Message: "something went wrong",
			Phase:   "Test",
			Code:    "TEST_ERROR",
			Cause:   cause,
		}

		if err.Unwrap() != cause {
			t.Error("Unwrap should return the cause")
		}
	})
}

func TestPolicyError(t *testing.T) {
	err := NewPolicyError("whitelist violation", nil)

	if err.Phase != "Policy" {
		t.Errorf("expected phase Policy, got %s", err.Phase)
	}
	if err.Code != "POLICY_VIOLATION" {
		t.Errorf("expected code POLICY_VIOLATION, got %s", err.Code)
	}
}

func TestBootstrapError(t *testing.T) {
	cause := errors.New("sf not found")
	err := NewBootstrapError("CLI bootstrap failed", cause)

	if err.Phase != "Bootstrap" {
		t.Errorf("expected phase Bootstrap, got %s", err.Phase)
	}
	if err.Code != "BOOTSTRAP_FAILED" {
		t.Errorf("expected code BOOTSTRAP_FAILED, got %s", err.Code)
	}
	if err.Unwrap() != cause {
		t.Error("cause should be wrapped")
	}
}

func TestIdentityError(t *testing.T) {
	err := NewIdentityError("invalid SFDX auth URL", nil)

	if err.Phase != "Identity" {
		t.Errorf("expected phase Identity, got %s", err.Phase)
	}
	if err.Code != "AUTHENTICATION_FAILED" {
		t.Errorf("expected code AUTHENTICATION_FAILED, got %s", err.Code)
	}
}

func TestContractError(t *testing.T) {
	err := NewContractError("invalid status", nil)

	if err.Phase != "Contract" {
		t.Errorf("expected phase Contract, got %s", err.Phase)
	}
	if err.Code != "CONTRACT_GENERATION_FAILED" {
		t.Errorf("expected code CONTRACT_GENERATION_FAILED, got %s", err.Code)
	}
}

func TestAnalyzerError(t *testing.T) {
	err := NewAnalyzerError("scan failed", nil)

	if err.Phase != "Analyzer" {
		t.Errorf("expected phase Analyzer, got %s", err.Phase)
	}
	if err.Code != "ANALYSIS_FAILED" {
		t.Errorf("expected code ANALYSIS_FAILED, got %s", err.Code)
	}
}

func TestFreezeError(t *testing.T) {
	err := NewFreezeError("Friday", "18:00", "20:00")

	if err.Phase != "Policy" {
		t.Errorf("expected phase Policy, got %s", err.Phase)
	}
	if err.Code != "FROZEN" {
		t.Errorf("expected code FROZEN, got %s", err.Code)
	}
	if err.Day != "Friday" {
		t.Errorf("expected day Friday, got %s", err.Day)
	}
}

func TestIsGlassOpsError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"PolicyError", NewPolicyError("test", nil), true},
		{"BootstrapError", NewBootstrapError("test", nil), true},
		{"IdentityError", NewIdentityError("test", nil), true},
		{"ContractError", NewContractError("test", nil), true},
		{"AnalyzerError", NewAnalyzerError("test", nil), true},
		{"FreezeError", NewFreezeError("Mon", "09:00", "10:00"), true},
		{"standard error", errors.New("not glassops"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsGlassOpsError(tt.err) != tt.expected {
				t.Errorf("IsGlassOpsError(%s) = %v, want %v", tt.name, !tt.expected, tt.expected)
			}
		})
	}
}

func TestGetPhase(t *testing.T) {
	tests := []struct {
		err      error
		expected string
	}{
		{NewPolicyError("test", nil), "Policy"},
		{NewBootstrapError("test", nil), "Bootstrap"},
		{NewIdentityError("test", nil), "Identity"},
		{NewContractError("test", nil), "Contract"},
		{NewAnalyzerError("test", nil), "Analyzer"},
		{errors.New("unknown"), "Unknown"},
	}

	for _, tt := range tests {
		phase := GetPhase(tt.err)
		if phase != tt.expected {
			t.Errorf("GetPhase() = %s, want %s", phase, tt.expected)
		}
	}
}

func TestGetCode(t *testing.T) {
	tests := []struct {
		err      error
		expected string
	}{
		{NewPolicyError("test", nil), "POLICY_VIOLATION"},
		{NewBootstrapError("test", nil), "BOOTSTRAP_FAILED"},
		{NewIdentityError("test", nil), "AUTHENTICATION_FAILED"},
		{NewContractError("test", nil), "CONTRACT_GENERATION_FAILED"},
		{NewAnalyzerError("test", nil), "ANALYSIS_FAILED"},
		{NewFreezeError("Mon", "09:00", "10:00"), "FROZEN"},
		{errors.New("unknown"), "UNKNOWN_ERROR"},
	}

	for _, tt := range tests {
		code := GetCode(tt.err)
		if code != tt.expected {
			t.Errorf("GetCode() = %s, want %s", code, tt.expected)
		}
	}
}
