// Package main is the entrypoint for the GlassOps runtime.
package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/glassops-platform/glassops/packages/runtime/internal/analyzer"
	"github.com/glassops-platform/glassops/packages/runtime/internal/contract"
	"github.com/glassops-platform/glassops/packages/runtime/internal/gha"
	"github.com/glassops-platform/glassops/packages/runtime/internal/policy"
	"github.com/glassops-platform/glassops/packages/runtime/internal/services"
	"github.com/glassops-platform/glassops/packages/runtime/internal/telemetry"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		gha.SetFailed(err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// LEVEL 1 PRIMITIVE: Core Governance Controls
	// ============================================

	// 1. Environment Context Validation
	requiredEnvVars := []string{
		"GITHUB_WORKSPACE",
		"GITHUB_ACTOR",
		"GITHUB_REPOSITORY",
	}

	var missingEnvVars []string
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			missingEnvVars = append(missingEnvVars, envVar)
		}
	}
	if len(missingEnvVars) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missingEnvVars, ", "))
	}

	// 2. Input Validation & Sanitization
	requiredInputs := []string{"client_id", "jwt_key", "username"}
	var missingInputs []string
	for _, input := range requiredInputs {
		if gha.GetInput(input) == "" {
			missingInputs = append(missingInputs, input)
		}
	}
	if len(missingInputs) > 0 {
		return fmt.Errorf("missing required inputs: %s", strings.Join(missingInputs, ", "))
	}

	// Validate JWT key format (skip if skip_auth is true)
	jwtKey := gha.GetInput("jwt_key")
	if gha.GetInput("skip_auth") != "true" {
		if !strings.Contains(jwtKey, "BEGIN") || !strings.Contains(jwtKey, "END") {
			return fmt.Errorf("invalid JWT key format - must contain BEGIN and END markers")
		}
	}
	if jwtKey != "" {
		gha.SetSecret(jwtKey)
	}

	// Validate Salesforce instance URL
	instanceURL := gha.GetInputWithDefault("instance_url", "https://login.salesforce.com")
	if !isValidURL(instanceURL) {
		return fmt.Errorf("invalid instance URL: %s", instanceURL)
	}

	// 3. Resource Limits Validation
	maxExecutionTime := 30 * time.Minute
	startTime := time.Now()

	// Safety timeout (runs in background)
	go func() {
		time.Sleep(maxExecutionTime)
		if time.Since(startTime) > maxExecutionTime {
			gha.Error("Execution timeout exceeded - terminating session")
			os.Exit(1)
		}
	}()

	// 4. Data Integrity & Compliance Checks
	if os.Getenv("GITHUB_EVENT_NAME") == "pull_request" {
		if os.Getenv("GITHUB_HEAD_REF") == "" {
			return fmt.Errorf("invalid pull request context - missing GITHUB_HEAD_REF")
		}
		if strings.Contains(os.Getenv("GITHUB_HEAD_REF"), ":") {
			gha.Warning("⚠️ Running on forked repository - additional security validations recommended")
		}
	}

	// Validate repository format
	repoPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`)
	if !repoPattern.MatchString(os.Getenv("GITHUB_REPOSITORY")) {
		return fmt.Errorf("invalid repository format: %s", os.Getenv("GITHUB_REPOSITORY"))
	}

	// Generate unique runtime ID
	runtimeID := generateUUID()
	gha.SetOutput("runtime_id", runtimeID)

	// Initialize OpenTelemetry
	serviceName := gha.GetInputWithDefault("otel_service_name", "glassops-runtime")
	if err := telemetry.Init(ctx, serviceName, "1.0.0"); err != nil {
		gha.Warning(fmt.Sprintf("[Telemetry] Failed to initialize: %s", err.Error()))
	}
	defer telemetry.Shutdown(ctx)

	// 0. Cache Retrieval Phase (placeholder - Go doesn't have @actions/cache equivalent)
	gha.StartGroup("📦 Checking Runtime Cache")
	gha.Info("⚠️ Cache operations not yet implemented in Go runtime")
	gha.EndGroup()

	// 1. Policy Phase
	gha.Info("[Policy] Evaluating governance policies...")
	policyEngine := policy.New()
	config, err := policyEngine.Load()
	if err != nil {
		return fmt.Errorf("policy evaluation failed: %w", err)
	}

	// BR-003: Static Analysis Invariants
	if config.Governance.Analyzer != nil && config.Governance.Analyzer.Enabled {
		gha.Info("[Analyzer] Running static code analysis...")
		codeAnalyzer := analyzer.New()

		if config.Governance.Analyzer.Opinionated {
			if err := codeAnalyzer.EnsureCompatibility(); err != nil {
				return err
			}
		}

		ruleset := ""
		if len(config.Governance.Analyzer.Rulesets) > 0 {
			ruleset = config.Governance.Analyzer.Rulesets[0]
		}

		scanResults, err := codeAnalyzer.Scan([]string{"."}, ruleset)
		if err != nil {
			return fmt.Errorf("static analysis failed: %w", err)
		}

		var criticalViolations []analyzer.Violation
		threshold := config.Governance.Analyzer.SeverityThreshold
		for _, v := range scanResults.Violations {
			if v.Severity <= threshold {
				criticalViolations = append(criticalViolations, v)
			}
		}

		if len(criticalViolations) > 0 {
			return fmt.Errorf("static analysis failed: %d critical violations found", len(criticalViolations))
		}
		gha.Info("[Analyzer] ✅ Static analysis passed")
	}

	if gha.GetInput("enforce_policy") == "true" {
		if err := policyEngine.CheckFreeze(config); err != nil {
			gha.SetOutput("is_locked", "true")
			return fmt.Errorf("policy violation: %w", err)
		}
		gha.SetOutput("is_locked", "false")
		gha.Info("[Policy] ✅ Policy check passed - no freeze windows active")
	} else {
		gha.SetOutput("is_locked", "false")
		gha.Warning("[Policy] Policy enforcement disabled")
	}

	// 2. Bootstrap Phase
	gha.Info("[Bootstrap] Bootstrapping Salesforce CLI environment...")
	runtime := services.NewRuntimeEnvironment()
	if err := runtime.Install(config.Runtime.CLIVersion); err != nil {
		return fmt.Errorf("bootstrap phase failed: %w", err)
	}
	gha.Info(fmt.Sprintf("[Bootstrap] ✅ CLI %s installed successfully", config.Runtime.CLIVersion))

	// Health check
	health := services.HealthCheck()
	if !health.Healthy {
		return fmt.Errorf("CLI health check failed: %s", health.Error)
	}

	// Install plugins if specified
	pluginsInput := gha.GetInput("plugins")
	if pluginsInput != "" {
		plugins := splitAndTrim(pluginsInput, ",")
		if len(plugins) > 0 {
			gha.Info(fmt.Sprintf("[Bootstrap] Installing %d plugin(s): %s", len(plugins), strings.Join(plugins, ", ")))
			if err := runtime.InstallPlugins(config, plugins); err != nil {
				return fmt.Errorf("plugin installation failed: %w", err)
			}
			gha.Info("[Bootstrap] ✅ All plugins installed successfully")
		}
	}

	// 3. Identity Phase
	gha.Info("[Identity] Authenticating with Salesforce...")
	var orgID string

	if gha.GetInput("skip_auth") == "true" {
		gha.Warning("[Identity] ⚠️ Skipping authentication as requested by configuration")
		orgID = "00D00000000TEST" // Dummy Org ID for testing
	} else {
		identity := services.NewIdentityResolver()
		authenticatedOrgID, err := identity.Authenticate(services.AuthRequest{
			ClientID:    gha.GetInput("client_id"),
			JWTKey:      gha.GetInput("jwt_key"),
			Username:    gha.GetInput("username"),
			InstanceURL: gha.GetInput("instance_url"),
		})
		if err != nil {
			return fmt.Errorf("Salesforce authentication failed: %w", err)
		}
		orgID = authenticatedOrgID
		gha.Info(fmt.Sprintf("[Identity] ✅ Authenticated with org %s", orgID))
	}

	// 4. Contract Validation Phase
	gha.Info("[Contract] 📄 Generating Deployment Contract v1.0...")

	// Parse test results from input
	testResults := contract.TestResults{Total: 0, Passed: 0, Failed: 0}
	testResultsInput := gha.GetInput("test_results")
	if testResultsInput != "" {
		if err := json.Unmarshal([]byte(testResultsInput), &testResults); err != nil {
			gha.Warning(fmt.Sprintf("[Contract] ⚠️ Invalid test_results JSON, using defaults: %s", err.Error()))
		} else {
			gha.Info(fmt.Sprintf("[Contract] Parsed test results: %d/%d passed", testResults.Passed, testResults.Total))
		}
	}

	// Get coverage data
	coverageActual := parseFloat(gha.GetInput("coverage_percentage"), 0)
	coverageRequired := parseFloat(gha.GetInputWithDefault("coverage_required", "80"), 80)

	deploymentContract := contract.New()
	deploymentContract.Meta.Trigger = os.Getenv("GITHUB_EVENT_NAME")
	if deploymentContract.Meta.Trigger == "" {
		deploymentContract.Meta.Trigger = "manual"
	}
	deploymentContract.Quality.Coverage.Actual = coverageActual
	deploymentContract.Quality.Coverage.Required = coverageRequired
	deploymentContract.Quality.Coverage.Met = coverageActual >= coverageRequired
	deploymentContract.Quality.Tests = testResults
	deploymentContract.Audit.TriggeredBy = os.Getenv("GITHUB_ACTOR")
	if deploymentContract.Audit.TriggeredBy == "" {
		deploymentContract.Audit.TriggeredBy = "unknown"
	}
	deploymentContract.Audit.OrgID = orgID
	deploymentContract.Audit.Repository = os.Getenv("GITHUB_REPOSITORY")
	if deploymentContract.Audit.Repository == "" {
		deploymentContract.Audit.Repository = "unknown"
	}
	deploymentContract.Audit.Commit = os.Getenv("GITHUB_SHA")
	if deploymentContract.Audit.Commit == "" {
		deploymentContract.Audit.Commit = "unknown"
	}

	// Write contract to file
	workspace := os.Getenv("GITHUB_WORKSPACE")
	if workspace == "" {
		workspace = "."
	}
	contractPath := filepath.Join(workspace, "glassops-contract.json")

	contractJSON, err := deploymentContract.ToJSON()
	if err != nil {
		return fmt.Errorf("contract generation failed: %w", err)
	}

	if err := os.WriteFile(contractPath, contractJSON, 0644); err != nil {
		return fmt.Errorf("failed to write contract: %w", err)
	}

	gha.SetOutput("contract_path", contractPath)
	gha.Info(fmt.Sprintf("[Contract] ✅ Contract written to %s", contractPath))

	// 5. Output Session State
	gha.SetOutput("org_id", orgID)
	gha.SetOutput("glassops_ready", "true")
	gha.Info("✅ GlassOps Runtime is ready for governed execution.")

	return nil
}

func isValidURL(s string) bool {
	return len(s) > 8 && (strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://"))
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseFloat(s string, defaultVal float64) float64 {
	if s == "" {
		return defaultVal
	}
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	if err != nil {
		return defaultVal
	}
	return f
}
