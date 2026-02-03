---
type: Documentation
domain: control-plane
last_modified: 2026-02-02
generated: true
source: packages/control-plane/internal/policy/policy.go
generated_at: 2026-02-02T22:23:16.603306
hash: 83153ee161935cb1582b33943db390eb171887a814847d212dc3e156bca88149
---

## Policy Package Documentation

This package manages the resolution and merging of security and governance policies. It provides a mechanism to combine locally defined policies with a baseline established by a remote source (represented by the GitHub Floor). The core principle is “Highest Value Wins,” meaning more restrictive settings take precedence.

**Package Responsibilities:**

- Loading policy configurations from a file.
- Merging a local policy with a remote baseline (GitHub Floor).
- Ensuring minimum governance coverage requirements are met.
- Providing a structured representation of governance rules.

**Key Types:**

- `ProtocolConfig`: This is the top-level structure representing the complete policy configuration. It contains a single field, `Governance`.
- `Governance`: This structure defines the core governance rules.
    - `Enabled`: A boolean indicating whether governance checks are active.
    - `MinCoverage`: A float64 representing the minimum acceptable code coverage percentage. This value is influenced by the GitHub Floor.
    - `StaticAnalysis`: A `StaticAnalysis` struct containing settings for static code analysis.
    - `PluginWhitelist`: A slice of strings listing allowed plugins.
- `StaticAnalysis`: This structure configures static code analysis behavior.
    - `Enabled`: A boolean indicating whether static analysis is enabled.
    - `BlockOn`: A slice of strings specifying severity levels that should block deployments (e.g., “critical”, “high”).

**Important Functions:**

- `ResolvePolicy(localPath string, githubFloor float64) (ProtocolConfig, error)`: This function is the primary entry point for resolving the final policy configuration.
    - `localPath`: The file path to the local policy configuration file (e.g., “devops-config.json”).
    - `githubFloor`: The baseline minimum coverage value enforced by the remote source.
    - Behavior:
        1.  It attempts to read and parse the policy from the specified `localPath`.
        2.  If the file does not exist or parsing fails, it returns a `ProtocolConfig` with only the `MinCoverage` set to the `githubFloor` value. This provides a default policy.
        3.  If the file is successfully parsed, it merges the local configuration with the `githubFloor`. Specifically, it ensures that the `MinCoverage` in the local configuration is *not* lower than the `githubFloor`. The `githubFloor` value always takes precedence.
        4.  It returns the resolved `ProtocolConfig` and a potential error.

**Error Handling:**

- The `ResolvePolicy` function returns an error if it fails to parse the JSON policy file. The error is wrapped using `fmt.Errorf` to provide context.
- If the policy file is missing, no error is returned; instead, a default configuration based on the `githubFloor` is returned.

**Design Decisions:**

- **Highest Value Wins:** The merging strategy prioritizes more restrictive settings. This ensures that the final policy is at least as secure as the baseline.
- **GitHub Floor as a Minimum:** The `githubFloor` acts as a lower bound for `MinCoverage`. This prevents local configurations from relaxing security requirements below an acceptable level.
- **Fallback Mechanism:** The package provides a fallback to a default policy if the local configuration file is unavailable or invalid. This ensures that the system continues to operate with reasonable defaults.