---
type: Documentation
domain: native
origin: packages/adapters/native/cmd/main.go
last_modified: 2026-01-31
generated: true
source: packages/adapters/native/cmd/main.go
generated_at: 2026-01-31T08:50:50.685943
hash: 2c4417b1c1c5b20a2809807e6d189a0b7a365808c5821a7565335171823efb5f
---

## Adapter Command-Line Interface Documentation

This document describes the functionality and design of the adapter command-line interface (CLI). This CLI provides a means to interact with a native adapter, currently focused on Salesforce, for security analysis tasks.

**Package Purpose**

The primary purpose of this package is to provide a command-line interface for interacting with the adapter. It handles parsing arguments, invoking adapter-specific logic, and formatting output.  It acts as the entry point for external tools or systems to request analysis from the adapter.

**Key Types**

The following data structures are defined to represent the adapter’s interactions and responses:

*   **InfoResponse:** Represents the adapter’s identification and capabilities. It includes the adapter’s `Name`, `Version`, and a list of supported `Capabilities`.
*   **TransformResponse:**  Encapsulates the result of a transformation operation. It contains a `Status` indicating success or failure, `Adapter` metadata detailing the adapter used, and `Output` metadata describing the transformation result.
*   **AdapterMetadata:**  Provides information about the adapter itself, including its `Name`, `Version`, and the `Substrate` it operates on (e.g., "salesforce").
*   **OutputMetadata:** Contains details about the output of a transformation, such as the path to the generated `SarifPath` file, a `TraceID` for tracking, and arbitrary `Metadata` key-value pairs.

**Important Functions**

*   **main():** The entry point of the application. It parses command-line arguments and dispatches execution to the appropriate handler function based on the provided command.  If no command or an unrecognized command is given, it prints usage instructions and exits.
*   **handleInfo():**  Implements the `info` command. It constructs an `InfoResponse` object with the adapter’s details and prints it as a JSON string to standard output.
*   **handleTransform():** Implements the `transform` command. It parses flags for input file path (`--input`), output SARIF file path (`--output`), and policy reference (`--policy-ref`). It validates that `--input` and `--output` are provided. Currently, it provides a mock implementation that prints a message indicating the transformation process and then constructs and prints a `TransformResponse` as a JSON string.
*   **handleValidate():** Implements the `validate` command. It parses a flag for the input file path (`--input`). It validates that `--input` is provided. Currently, it provides a mock implementation that prints a message indicating the validation process and then prints a simple JSON string indicating validation success with an empty list of violations.

**Error Handling**

The CLI employs basic error handling:

*   Missing or invalid command-line arguments result in error messages printed to standard output, followed by a program exit with a non-zero status code (1).
*   Flag parsing errors are handled by the `flag` package, which automatically prints error messages and exits.
*   The `json.MarshalIndent` function's error return is currently ignored. In a production environment, this should be handled appropriately.

**Concurrency**

This version of the adapter CLI does not employ concurrency (goroutines or channels). All operations are performed sequentially within the `main` function and its handler functions.

**Design Decisions**

*   **Command-Based Structure:** The CLI is structured around commands (`info`, `transform`, `validate`) to provide a clear and organized interface.
*   **Flag-Based Argument Parsing:** The `flag` package is used for parsing command-line arguments, providing a standard and reliable mechanism.
*   **JSON Output:**  Responses are formatted as JSON to facilitate easy integration with other tools and systems.
*   **Mock Implementations:** The `transform` and `validate` commands currently contain mock implementations. This allows for initial development and testing of the CLI structure without requiring the full adapter logic to be implemented. You will need to replace these with actual adapter calls.
*   **Clear Error Messages:**  The CLI provides informative error messages to help users diagnose and resolve issues.