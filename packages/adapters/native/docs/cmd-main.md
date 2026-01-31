---
type: Documentation
domain: native
origin: packages/adapters/native/cmd/main.go
last_modified: 2026-01-31
generated: true
source: packages/adapters/native/cmd/main.go
generated_at: 2026-01-31T09:45:42.125394
hash: 2c4417b1c1c5b20a2809807e6d189a0b7a365808c5821a7565335171823efb5f
---

## Adapter Command-Line Interface Documentation

This document describes the functionality and design of the adapter command-line interface (CLI). This CLI provides a means to interact with a native adapter, currently focused on Salesforce, for security analysis tasks.

**Package Purpose**

The primary responsibility of this package is to provide a command-line interface for interacting with the adapter. It handles parsing arguments, invoking adapter-specific logic, and formatting output.  It acts as the entry point for external tools or systems to request security analysis operations.

**Key Types**

The following types define the data structures used for communication:

*   **`InfoResponse`**: Represents the adapter's information, including its name, version, and supported capabilities.
    ```go
    type InfoResponse struct {
    	Name         string   `json:"name"`
    	Version      string   `json:"version"`
    	Capabilities []string `json:"capabilities"`
    }
    ```
*   **`TransformResponse`**:  Encapsulates the result of a transformation operation. It includes the status, adapter metadata, and output metadata.
    ```go
    type TransformResponse struct {
    	Status  string          `json:"status"`
    	Adapter AdapterMetadata `json:"adapter"`
    	Output  OutputMetadata  `json:"output"`
    }
    ```
*   **`AdapterMetadata`**: Contains information about the adapter itself, such as its name, version, and the underlying platform (substrate) it operates on.
    ```go
    type AdapterMetadata struct {
    	Name      string `json:"name"`
    	Version   string `json:"version"`
    	Substrate string `json:"substrate"`
    }
    ```
*   **`OutputMetadata`**: Provides details about the output of an operation, including the path to the generated SARIF file, a trace ID, and any additional metadata.
    ```go
    type OutputMetadata struct {
    	SarifPath string                 `json:"sarif_path"`
    	TraceID   string                 `json:"trace_id"`
    	Metadata  map[string]interface{} `json:"metadata"`
    }
    ```

**Important Functions**

*   **`main()`**: This is the entry point of the application. It parses command-line arguments and dispatches execution to the appropriate handler function based on the provided command (`info`, `transform`, `validate`). If an unknown command is given, it prints an error message and exits.
*   **`handleInfo()`**:  This function handles the `info` command. It creates an `InfoResponse` object with the adapter's details and prints it as a JSON string to standard output.
*   **`handleTransform()`**: This function handles the `transform` command. It uses the `flag` package to parse command-line arguments specific to the transform operation, including the input file path, output SARIF file path, and policy reference. It validates that the required `--input` and `--output` flags are provided. Currently, it provides a mock implementation that prints a message indicating the transformation process and then constructs and prints a `TransformResponse` object as a JSON string.
*   **`handleValidate()`**: This function handles the `validate` command. It uses the `flag` package to parse the `--input` flag. It validates that the `--input` flag is provided. Currently, it provides a mock implementation that prints a message indicating the validation process and then prints a simple JSON string indicating a successful validation with no violations.

**Error Handling**

The CLI employs basic error handling:

*   If insufficient command-line arguments are provided, an error message is printed to standard output, and the program exits with a non-zero exit code (1).
*   If required flags for specific commands (e.g., `--input` and `--output` for `transform`) are missing, an error message is printed, and the program exits.
*   The `flag` package's `ExitOnError` behavior is used to automatically exit the program if there are issues parsing flags.

**Concurrency**

This initial version of the adapter CLI does not employ concurrency (goroutines or channels). Future iterations may introduce concurrency to improve performance, particularly for the `transform` and `validate` operations.

**Design Decisions**

*   **Command-Based Structure**: The CLI is structured around commands (`info`, `transform`, `validate`) to provide a clear and organized interface.
*   **Flag Parsing**: The `flag` package is used for parsing command-line arguments, providing a standard and reliable mechanism for handling options.
*   **JSON Output**:  All responses are formatted as JSON strings for easy parsing by other tools and systems.
*   **Mock Implementations**: The `transform` and `validate` commands currently use mock implementations to provide a foundation for future development.  These will be replaced with actual adapter logic.
*   **Data Structures**: The data structures (`InfoResponse`, `TransformResponse`, etc.) are designed to align with potential protobuf definitions, facilitating future integration with a more robust communication layer.