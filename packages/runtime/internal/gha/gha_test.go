package gha

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestGetInput(t *testing.T) {
	// Test INPUT_ prefix
	os.Setenv("INPUT_TEST_VAR", "input_value")
	defer os.Unsetenv("INPUT_TEST_VAR")

	result := GetInput("test_var")
	if result != "input_value" {
		t.Errorf("expected 'input_value', got '%s'", result)
	}
}

func TestGetInputGlassOpsPrefix(t *testing.T) {
	// Test GLASSOPS_ prefix fallback
	os.Setenv("GLASSOPS_MY_VAR", "glassops_value")
	defer os.Unsetenv("GLASSOPS_MY_VAR")

	result := GetInput("my_var")
	if result != "glassops_value" {
		t.Errorf("expected 'glassops_value', got '%s'", result)
	}
}

func TestGetInputPriority(t *testing.T) {
	// INPUT_ should take priority over GLASSOPS_
	os.Setenv("INPUT_PRIORITY_VAR", "input_wins")
	os.Setenv("GLASSOPS_PRIORITY_VAR", "glassops_loses")
	defer func() {
		os.Unsetenv("INPUT_PRIORITY_VAR")
		os.Unsetenv("GLASSOPS_PRIORITY_VAR")
	}()

	result := GetInput("priority_var")
	if result != "input_wins" {
		t.Errorf("expected 'input_wins', got '%s'", result)
	}
}

func TestGetInputEmpty(t *testing.T) {
	result := GetInput("nonexistent_var")
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestGetInputWithDefault(t *testing.T) {
	// Test with value present
	os.Setenv("INPUT_HAS_VALUE", "actual_value")
	defer os.Unsetenv("INPUT_HAS_VALUE")

	result := GetInputWithDefault("has_value", "default")
	if result != "actual_value" {
		t.Errorf("expected 'actual_value', got '%s'", result)
	}

	// Test with no value
	result = GetInputWithDefault("no_value", "default")
	if result != "default" {
		t.Errorf("expected 'default', got '%s'", result)
	}
}

func TestSetOutputFallback(t *testing.T) {
	// Ensure we use the fallback path (stdout) by unsetting GITHUB_OUTPUT
	oldGithubOutput := os.Getenv("GITHUB_OUTPUT")
	os.Unsetenv("GITHUB_OUTPUT")
	defer func() {
		if oldGithubOutput != "" {
			os.Setenv("GITHUB_OUTPUT", oldGithubOutput)
		}
	}()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	SetOutput("test_output", "test_value")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expected := "::set-output name=test_output::test_value"
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain '%s', got '%s'", expected, output)
	}
}

func TestSetSecret(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	SetSecret("super-secret-value")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expected := "::add-mask::super-secret-value"
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain '%s', got '%s'", expected, output)
	}
}

func TestSetFailed(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	SetFailed("something went wrong")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expected := "::error::something went wrong"
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain '%s', got '%s'", expected, output)
	}
}

func TestWarning(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Warning("this is a warning")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expected := "::warning::this is a warning"
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain '%s', got '%s'", expected, output)
	}
}

func TestStartEndGroup(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	StartGroup("Test Group")
	EndGroup()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "::group::Test Group") {
		t.Error("expected ::group:: in output")
	}

	if !strings.Contains(output, "::endgroup::") {
		t.Error("expected ::endgroup:: in output")
	}
}
