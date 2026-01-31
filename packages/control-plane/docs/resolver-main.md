---
type: Documentation
domain: control-plane
origin: packages/control-plane/cmd/resolver/main.go
last_modified: 2026-01-31
generated: true
source: packages/control-plane/cmd/resolver/main.go
generated_at: 2026-01-31T08:51:06.282208
hash: e53a85af9c11df19f7caa52717a03e4fe1f512958c414471e02e77ad2ada38bc
---

GlassOps Control Plane Resolver Documentation

This application serves as a governance checkpoint within a deployment pipeline. It verifies that a deployed application’s characteristics, as defined in a deployment contract, meet the standards set by a defined policy. This ensures quality and adherence to architectural guidelines before proceeding further in the deployment process.

Key Components:

DeploymentContract: This type, defined in the `github.com/glassops-platform/glassops-control-plane/internal/contract` package, represents the state of the deployed application. It contains information about quality metrics, such as code coverage and static analysis results. Specifically, it includes:
    - Quality.Coverage.Actual: The actual code coverage percentage achieved by the deployment.
    - Quality.StaticAnalysis.Met: A boolean indicating whether static analysis checks passed.
    - Quality.StaticAnalysis.Tool: The name of the static analysis tool used.

Policy: Represented by the types within the `github.com/glassops-platform/glassops-control-plane/internal/policy` package, this defines the acceptable standards for deployments. The `ResolvePolicy` function loads and interprets a policy configuration file. Key attributes include:
    - Governance.MinCoverage: The minimum acceptable code coverage percentage.
    - Governance.StaticAnalysis.Enabled: A boolean indicating whether static analysis is required.

Functions:

main(): This is the entry point of the application. It performs the following steps:
    1. Contract Loading: Reads the deployment contract from the file `.glassops/deployment-contract.json`. If the file is not found, the application terminates with a fatal error.
    2. Policy Resolution: Loads and resolves a policy from the file `devops-config.json`, applying an 80% floor to the policy’s governance settings.
    3. Coverage Enforcement: Compares the actual code coverage reported in the deployment contract against the minimum coverage required by the resolved policy. If the actual coverage is below the minimum, the application prints an error message and exits with a non-zero status code (1).
    4. Static Analysis Enforcement: Checks if static analysis is enabled in the policy and, if so, verifies that the deployment contract indicates static analysis passed. If static analysis is enabled and has not passed, the application prints an error message and exits with a non-zero status code (1).
    5. Success: If all checks pass, the application prints a success message.

Error Handling:

The application employs basic error handling. Critical errors, such as the inability to read the deployment contract, result in a fatal error message printed to the console and application termination.  Other failures, specifically those related to policy resolution, are currently ignored (represented by the blank `_` in `effPolicy, _ := policy.ResolvePolicy(...)`). You should implement more robust error handling in a production environment.

Concurrency:

This application is single-threaded and does not employ goroutines or channels.

Design Decisions:

The application uses a gatekeeper approach to governance. It defines clear pass/fail criteria based on the deployment contract and policy. The use of a JSON-based contract and policy allows for easy configuration and extensibility. The location of the deployment contract (`.glassops/deployment-contract.json`) is a fixed configuration. You may want to make this configurable.