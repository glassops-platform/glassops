// Package gha provides GitHub Actions integration utilities.
// This replaces the @actions/core package from TypeScript.
package gha

import (
	"fmt"
	"os"
	"strings"
)

// GetInput retrieves an action input by name.
// It checks in order: INPUT_<NAME>, GLASSOPS_<NAME>, then returns empty string.
func GetInput(name string) string {
	// Try Docker-style INPUT_<NAME> first
	envName := "INPUT_" + strings.ToUpper(name)
	if val := os.Getenv(envName); val != "" {
		return val
	}

	// Try GLASSOPS_ prefix
	glassopsEnvName := "GLASSOPS_" + strings.ToUpper(name)
	if val := os.Getenv(glassopsEnvName); val != "" {
		return val
	}

	return ""
}

// GetInputWithDefault retrieves an action input or returns a default value.
func GetInputWithDefault(name, defaultValue string) string {
	if val := GetInput(name); val != "" {
		return val
	}
	return defaultValue
}

// SetOutput sets an action output parameter.
func SetOutput(name, value string) {
	outputFile := os.Getenv("GITHUB_OUTPUT")
	if outputFile != "" {
		f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			fmt.Fprintf(f, "%s=%s\n", name, value)
			return
		}
	}
	// Fallback to old-style ::set-output
	fmt.Printf("::set-output name=%s::%s\n", name, value)
}

// SetSecret masks a value in logs.
func SetSecret(secret string) {
	fmt.Printf("::add-mask::%s\n", secret)
}

// SetFailed sets the action as failed with an error message.
func SetFailed(message string) {
	fmt.Printf("::error::%s\n", message)
}

// Info logs an info message.
func Info(message string) {
	fmt.Println(message)
}

// Warning logs a warning message.
func Warning(message string) {
	fmt.Printf("::warning::%s\n", message)
}

// Error logs an error message.
func Error(message string) {
	fmt.Printf("::error::%s\n", message)
}

// StartGroup starts a log group.
func StartGroup(name string) {
	fmt.Printf("::group::%s\n", name)
}

// EndGroup ends a log group.
func EndGroup() {
	fmt.Println("::endgroup::")
}
