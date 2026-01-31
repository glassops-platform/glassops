---
type: Documentation
domain: control-plane
origin: packages/control-plane/internal/policy/policy.go
last_modified: 2026-01-31
generated: true
source: packages/control-plane/internal/policy/policy.go
generated_at: 2026-01-31T08:51:39.163872
hash: 83153ee161935cb1582b33943db390eb171887a814847d212dc3e156bca88149
---

## Policy Package Documentation

This package manages the resolution and merging of security and governance policies. It provides a mechanism to combine locally defined policies with a baseline established by a remote source (represented by the “githubFloor” value). The core principle is “Highest Value Wins,” meaning more restrictive settings take precedence, but the baseline coverage cannot be reduced.

**Key Types:**

*   **ProtocolConfig:** The top-level structure representing the complete policy configuration. It contains a single field, `Governance`.
*   **Governance:** Defines the core governance rules.
    *   `Enabled`: A boolean indicating whether governance checks are active.
    *   `MinCoverage`: A float64 representing the minimum acceptable code coverage (e.g., for tests). This acts as a floor, enforced by the system.
    *   `StaticAnalysis`: Contains settings related to static code analysis.
    *   `PluginWhitelist`: A slice of strings listing allowed plugins.
*   **StaticAnalysis:** Configures static code analysis behavior.
    *   `Enabled`: A boolean indicating whether static analysis is enabled.
    *   `BlockOn`: A slice of strings specifying severity levels that, if found during static analysis, should block further processing.

**Important Functions:**

*   **ResolvePolicy(localPath string, githubFloor float64) (ProtocolConfig, error):** This function is the primary entry point for obtaining the effective policy configuration.
    1.  It attempts to read a policy file from the provided `localPath` (expected to be a JSON file).
    2.  If the file is not found or cannot be read, it returns a `ProtocolConfig` with only the `MinCoverage` set to the `githubFloor` value, effectively using the remote baseline.
    3.  If the file is successfully read and parsed into a `ProtocolConfig`, the function performs an additive merge. Specifically, it ensures that the `MinCoverage` in the loaded configuration is *not* less than the provided `githubFloor` value. If it is, the `MinCoverage` is updated to match the `githubFloor`.
    4.  The function returns the resolved `ProtocolConfig` and a potential error.

**Error Handling:**

The `ResolvePolicy` function handles errors during file reading and JSON parsing. If an error occurs during parsing, it returns the original error wrapped with a more informative message indicating a policy parsing failure.  If the file is missing, no error is returned; instead, a default configuration based on the `githubFloor` is provided.

**Design Decisions:**

*   **Additive Merge:** The merge strategy prioritizes local configurations while guaranteeing a minimum level of security/coverage defined by the `githubFloor`. This allows for customization without compromising baseline standards.
*   **File Fallback:** The system gracefully handles missing local policy files by falling back to the remote baseline. This ensures that the system remains functional even without a local configuration.
*   **JSON Format:** The use of JSON for policy configuration provides a human-readable and easily parsable format.