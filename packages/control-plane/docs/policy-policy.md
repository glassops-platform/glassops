---
type: Documentation
domain: control-plane
origin: packages/control-plane/internal/policy/policy.go
last_modified: 2026-01-31
generated: true
source: packages/control-plane/internal/policy/policy.go
generated_at: 2026-01-31T09:46:42.329340
hash: 83153ee161935cb1582b33943db390eb171887a814847d212dc3e156bca88149
---

## Policy Package Documentation

This package manages the resolution of security and governance policies. It provides a mechanism to load, merge, and apply policies that define acceptable risk levels and security configurations.

**Package Responsibilities:**

The primary responsibility of this package is to provide the `ResolvePolicy` function, which consolidates policy settings from a local file with a baseline "floor" defined by a remote source (represented by `githubFloor`). This ensures that local configurations can enhance security but never reduce it below an acceptable minimum.

**Key Types:**

*   **`ProtocolConfig`**: This is the top-level structure representing the complete policy configuration. It contains a single field, `Governance`.
*   **`Governance`**: This structure defines the core governance rules.
    *   `Enabled`: A boolean indicating whether governance checks are active.
    *   `MinCoverage`: A float64 representing the minimum acceptable code coverage (e.g., for tests). This acts as a safety net, preventing policies from lowering security standards.
    *   `StaticAnalysis`: A `StaticAnalysis` struct containing settings for static code analysis.
    *   `PluginWhitelist`: A slice of strings listing allowed plugins.
*   **`StaticAnalysis`**: This structure configures static code analysis behavior.
    *   `Enabled`: A boolean indicating whether static analysis is enabled.
    *   `BlockOn`: A slice of strings specifying severity levels that, if found during static analysis, should block further processing. For example, `["critical", "high"]` would block on critical and high severity issues.

**Important Functions:**

*   **`ResolvePolicy(localPath string, githubFloor float64) (ProtocolConfig, error)`**: This function is the core of the policy resolution process.
    1.  It attempts to read a policy configuration from the file specified by `localPath`.
    2.  If the file does not exist or cannot be read, it returns a `ProtocolConfig` with only the `MinCoverage` set to the provided `githubFloor` value. This provides a default policy.
    3.  If the file is successfully read, it attempts to parse the JSON content into a `ProtocolConfig` struct. If parsing fails, it returns the partially parsed config and an error.
    4.  It then performs an additive merge. Specifically, it ensures that the `MinCoverage` in the loaded configuration is *not* less than the `githubFloor` value. If it is, the `MinCoverage` is updated to the `githubFloor` value.
    5.  Finally, it returns the resolved `ProtocolConfig` and a `nil` error if successful.

**Error Handling:**

The `ResolvePolicy` function handles errors primarily during file reading and JSON parsing. Errors are wrapped using `fmt.Errorf` to provide context. If an error occurs during parsing, the function returns the partially parsed configuration along with the error. If the file is missing or unreadable, a default configuration is returned with a `nil` error.

**Concurrency:**

This package does not currently employ goroutines or channels. It operates synchronously.

**Design Decisions:**

*   **"Highest Value Wins" Merge Strategy:** The policy resolution uses a simple additive merge strategy. Local configurations can enhance security, but they cannot weaken it below the baseline defined by the `githubFloor`.
*   **Default Policy:** The package provides a sensible default policy (minimum coverage enforced) if a local configuration file is unavailable.
*   **JSON Format:** The policy configuration is stored in JSON format for ease of readability and modification.
*   **Error Wrapping:** Errors are wrapped to provide more informative error messages and maintain context.