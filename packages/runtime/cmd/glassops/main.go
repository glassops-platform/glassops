// Package main is the entrypoint for the GlassOps runtime.
package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"time"

	"github.com/glassops-platform/glassops/packages/runtime/internal/analyzer"
	"github.com/glassops-platform/glassops/packages/runtime/internal/contract"
	"github.com/glassops-platform/glassops/packages/runtime/internal/gha"
	"github.com/glassops-platform/glassops/packages/runtime/internal/permit"
	"github.com/glassops-platform/glassops/packages/runtime/internal/policy"
	"github.com/glassops-platform/glassops/packages/runtime/internal/services"
	"github.com/glassops-platform/glassops/packages/runtime/internal/telemetry"
	"github.com/glassops-platform/glassops/packages/runtime/internal/validator"
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
	if err := validator.ValidateEnvironment(); err != nil {
		return err
	}

	// 2. Input Validation & Sanitization
	// 0. Bootstrap: Load secrets from file if specified (Local Debugging Helper)
	if keyFile := gha.GetInput("jwt_key_file"); keyFile != "" {
		gha.Info(fmt.Sprintf("[Bootstrap] Loading JWT key from file: %s", keyFile))
		keyContent, err := os.ReadFile(keyFile)
		if err != nil {
			return fmt.Errorf("failed to load JWT key file: %w", err)
		}
		os.Setenv("INPUT_JWT_KEY", string(keyContent))
	}

	if err := validator.ValidateInputs(); err != nil {
		return err
	}

	// 3. Resource Limits Validation
	maxExecutionTime := 30 * time.Minute
	go enforceTimeout(maxExecutionTime)

	// 4. Data Integrity & Compliance Checks
	if err := validator.ValidateContext(); err != nil {
		return err
	}

	// Validate Salesforce instance URL
	instanceURL := gha.GetInputWithDefault("instance_url", "https://login.salesforce.com")
	if err := validator.EnsureValidInstanceURL(instanceURL); err != nil {
		return err
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

	// 1. Policy Phase
	gha.Info("[Policy] Evaluating governance policies...")
	policyEngine := policy.New()
	config, err := policyEngine.Load()
	if err != nil {
		return fmt.Errorf("policy evaluation failed: %w", err)
	}

	evaluation := permit.PolicyEvaluation{
		Allowed:   true,
		Evaluated: []string{"FreezeCheck"},
	}

	// BR-003: Static Analysis Invariants
	codeAnalyzer := analyzer.New()
	evaluation.Evaluated = append(evaluation.Evaluated, "StaticAnalysis")
	if err := codeAnalyzer.RunIfEnabled(config); err != nil {
		evaluation.Allowed = false
		evaluation.Violations = append(evaluation.Violations, err.Error())
		return err
	}

	if gha.GetInput("enforce_policy") == "true" {
		if err := policyEngine.CheckFreeze(config); err != nil {
			evaluation.Allowed = false
			evaluation.Violations = append(evaluation.Violations, err.Error())
			gha.SetOutput("is_locked", "true")
			return fmt.Errorf("policy violation: %w", err)
		}
		gha.SetOutput("is_locked", "false")
		gha.Info("[Policy] Policy check passed - no freeze windows active")
	} else {
		gha.SetOutput("is_locked", "false")
		gha.Warning("[Policy] Policy enforcement disabled")
	}

	// 2. Identity Phase
	gha.Info("[Identity] Authenticating with Salesforce...")
	var orgID string
	actor := permit.Identity{
		Subject:    os.Getenv("GITHUB_ACTOR"),
		Provider:   "github",
		ProviderID: fmt.Sprintf("github:%s", os.Getenv("GITHUB_ACTOR")),
		Verified:   false,
	}

	if gha.GetInput("skip_auth") == "true" {
		gha.Warning("[Identity] Skipping authentication as requested by configuration")
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
		actor.Subject = gha.GetInput("username")
		actor.Provider = "salesforce"
		actor.ProviderID = fmt.Sprintf("sf:%s", orgID)
		actor.Verified = true
		gha.Info(fmt.Sprintf("[Identity] Authenticated with org %s", orgID))
	}

	// 3. Context Handoff (Permit Generation)
	gha.Info("[Context] Generating GlassOps Permit...")
	permitPath, err := permit.Generate(runtimeID, actor, evaluation, instanceURL)
	if err != nil {
		return err
	}
	gha.Info(fmt.Sprintf("[Context] Permit written to %s", permitPath))

	// 4. Contract Validation Phase
	gha.Info("[Contract] Generating Deployment Contract v1.0...")

	contractPath, err := contract.Generate(orgID)
	if err != nil {
		return fmt.Errorf("contract generation failed: %w", err)
	}

	gha.SetOutput("contract_path", contractPath)
	gha.Info(fmt.Sprintf("[Contract] Contract written to %s", contractPath))

	// 5. Output Session State
	gha.SetOutput("org_id", orgID)
	gha.SetOutput("glassops_ready", "true")
	gha.Info("GlassOps Runtime is ready for governed execution.")

	return nil
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// enforceTimeout monitors execution time and exits if the limit is exceeded.
func enforceTimeout(limit time.Duration) {
	time.Sleep(limit)
	gha.Error("Execution timeout exceeded - terminating session")
	os.Exit(1)
}
