package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/glassops-platform/glassops-control-plane/internal/contract"
	"github.com/glassops-platform/glassops-control-plane/internal/policy"
)

func main() {
	fmt.Println("GlassOps Control Plane: Initiating Governance Check...")

	// 1. Load the Contract emitted by the Runtime
	contractData, err := os.ReadFile(".glassops/deployment-contract.json")
	if err != nil {
		log.Fatalf("❌ CRITICAL: Deployment Contract not found: %v", err)
	}

	var dc contract.DeploymentContract
	json.Unmarshal(contractData, &dc)

	// 2. Resolve Policy (Using an 80% floor for this example)
	effPolicy, _ := policy.ResolvePolicy("devops-config.json", 80.0)

	// 3. The Gate: Enforce Quality
	if dc.Quality.Coverage.Actual < effPolicy.Governance.MinCoverage {
		fmt.Printf("🔴 BLOCKED: Coverage %v%% is below required %v%%\n", 
			dc.Quality.Coverage.Actual, effPolicy.Governance.MinCoverage)
		os.Exit(1)
	}

	// 4. The Gate: Enforce Architectural Invariants (Static Analysis)
	if effPolicy.Governance.StaticAnalysis.Enabled && !dc.Quality.StaticAnalysis.Met {
		fmt.Printf("🔴 BLOCKED: Architectural Invariants failed via %s\n", 
			dc.Quality.StaticAnalysis.Tool)
		os.Exit(1)
	}

	fmt.Println("✅ GOVERNANCE PASSED: Intent matches Policy.")
}
