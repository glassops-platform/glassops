---
type: Documentation
domain: control-plane
origin: packages/control-plane/internal/policy/policy.go
last_modified: 2026-02-01
generated: true
source: packages/control-plane/internal/policy/policy.go
generated_at: 2026-02-01T19:26:36.705439
hash: 83153ee161935cb1582b33943db390eb171887a814847d212dc3e156bca88149
---

## Policy Package Documentation

This package manages the resolution of security and governance policies. It defines structures for representing policy configurations and provides a function to merge a local policy file with a baseline GitHub floor for minimum code coverage.

**Package Responsibilities:**

- Define data structures for representing governance policies, including static analysis settings and minimum code coverage requirements.
- Load and parse policy configurations from a JSON file.
- Merge local policy configurations with a baseline GitHub floor, ensuring that the minimum coverage requirement is never lowered.

**Key Types:**

- `ProtocolConfig`: The top-level structure representing the complete policy configuration. It contains a single field, `Governance`.
- `Governance`:  Defines the core governance rules.
    - `Enabled`: A boolean indicating whether governance checks are active.
    - `MinCoverage`: A float64 representing the minimum required code coverage percentage. This value is often referred to as the “GitHub Floor”.
    - `StaticAnalysis`: A `StaticAnalysis` struct containing settings for static analysis tools.
    - `PluginWhitelist`: A slice of strings representing a list of allowed plugins.
- `StaticAnalysis`: Configures static analysis behavior.
    - `Enabled`: A boolean indicating whether static analysis is enabled.
    - `BlockOn`: A slice of strings specifying the severity levels that should trigger a build failure (e.g., “critical”, “high”).

**Important Functions:**

- `ResolvePolicy(localPath string, githubFloor float64) (ProtocolConfig, error)`: This function resolves the final policy configuration.
    - `localPath`: The path to a local `devops-config.json` file containing policy overrides.
    - `githubFloor`: The baseline minimum code coverage requirement enforced by the GitHub repository.
    - Behavior:
        1.  It attempts to read and parse the JSON file at `localPath`.
        2.  If the file does not exist or parsing fails, it returns a `ProtocolConfig` with only the `githubFloor` set as the `MinCoverage`, effectively using the baseline.
        3.  If the file is successfully parsed, it merges the loaded configuration with the `githubFloor`. The `MinCoverage` in the loaded configuration is updated to the `githubFloor` if it is lower.
        4.  It returns the resolved `ProtocolConfig` and a potential error.

**Error Handling:**

- The `ResolvePolicy` function returns an error if it fails to parse the JSON file at the specified `localPath`. The error is wrapped using `fmt.Errorf` to provide context.
- If the file is missing, no error is returned; instead, a default configuration with the `githubFloor` is used.

**Concurrency:**

- This package does not currently employ goroutines or channels. It operates synchronously.

**Design Decisions:**

- **Highest Value Wins Merge:** The `ResolvePolicy` function implements a “Highest Value Wins” merge strategy. Local configuration values override the baseline `githubFloor` except for `MinCoverage`, which is guaranteed to be at least the `githubFloor`.
- **Fallback Mechanism:** The package provides a fallback mechanism to use the `githubFloor` if the local policy file is missing or invalid. This ensures that a minimum level of governance is always enforced.
- **Additive Merge:** The merging of the local configuration and the GitHub floor is designed to be additive, preventing local settings from reducing the baseline security requirements.