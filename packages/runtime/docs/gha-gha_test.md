---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/internal/gha/gha_test.go
generated_at: 2026-02-02T22:36:53.272166
hash: 78792b077f78dcf3f17aa320e2dd3e7fb2cc6d08b94f25a00d6add71f4b7b6b3
---

## Package gha Documentation

This package provides functions for interacting with the GitHub Actions environment. It allows reading input parameters, setting output variables, and emitting diagnostic messages. The package is designed to facilitate the creation of tools and actions that run within the GitHub Actions workflow.

**Key Concepts:**

The package centers around the idea of interacting with the environment variables and standard output streams that GitHub Actions provides to its workflows. It abstracts away the specific formatting required by GitHub Actions for setting outputs, secrets, errors, and warnings.

**Types and Interfaces:**

This package does not define any custom types or interfaces. It operates directly on environment variables and standard output.

**Functions:**

*   **`GetInput(name string) string`**: This function retrieves the value of an input parameter from the GitHub Actions environment. It first checks for an environment variable prefixed with `INPUT_`. If not found, it falls back to checking for a variable prefixed with `GLASSOPS_`.  If neither prefix is found, an empty string is returned. `INPUT_` prefixed variables take precedence over `GLASSOPS_` prefixed variables.

*   **`GetInputWithDefault(name string, defaultValue string) string`**: This function retrieves the value of an input parameter, similar to `GetInput`. However, if the parameter is not found in the environment, it returns the provided `defaultValue`.

*   **`SetOutput(name string, value string)`**: This function sets an output variable in the GitHub Actions environment. It formats the output string according to the GitHub Actions standard (`::set-output name=<name>::<value>`) and writes it to standard output. If the `GITHUB_OUTPUT` environment variable is not set, the output is written directly to standard output.

*   **`SetSecret(value string)`**: This function sets a secret value in the GitHub Actions environment. It formats the output string according to the GitHub Actions standard (`::add-mask::<value>`) and writes it to standard output, masking the value.

*   **`SetFailed(message string)`**: This function marks the current step as failed in the GitHub Actions workflow. It formats the output string according to the GitHub Actions standard (`::error::<message>`) and writes it to standard output.

*   **`Warning(message string)`**: This function emits a warning message in the GitHub Actions workflow. It formats the output string according to the GitHub Actions standard (`::warning::<message>`) and writes it to standard output.

*   **`StartGroup(name string)`**: This function starts a named group in the GitHub Actions log. It formats the output string according to the GitHub Actions standard (`::group::<name>`) and writes it to standard output.

*   **`EndGroup()`**: This function ends the current group in the GitHub Actions log. It formats the output string according to the GitHub Actions standard (`::endgroup::`) and writes it to standard output.

**Error Handling:**

The functions in this package do not return explicit error values. Instead, they rely on the behavior of the GitHub Actions environment.  For example, `GetInput` returns an empty string if the input is not found. `SetFailed` signals a workflow failure through standard output.

**Concurrency:**

This package does not explicitly use goroutines or channels. The functions are designed to be called sequentially within a single workflow step.

**Design Decisions:**

*   **Environment Variable Prefixes:** The package supports two environment variable prefixes (`INPUT_` and `GLASSOPS_`) for input parameters. This allows for flexibility and potential compatibility with existing tools. The `INPUT_` prefix is given priority.
*   **Standard Output for Outputs/Diagnostics:** The package uses standard output to communicate outputs, secrets, errors, and warnings to the GitHub Actions environment. This is the standard mechanism for these types of interactions.
*   **No Explicit Error Returns:** The decision to not return explicit errors simplifies the API and aligns with the typical usage pattern in GitHub Actions, where failures are signaled through the workflow status.