---
type: Documentation
domain: native
last_modified: 2026-02-02
generated: true
source: packages/adapters/native/cmd/main.go
generated_at: 2026-02-02T22:22:02.327197
hash: 2c4417b1c1c5b20a2809807e6d189a0b7a365808c5821a7565335171823efb5f
---

## Adapter Command-Line Interface Documentation

This document describes the functionality and design of the adapter command-line interface (CLI). This CLI serves as an entry point for interacting with a native adapter, currently focused on Salesforce data. It provides commands to retrieve adapter information, transform data, and validate data.

### Package Purpose

The primary responsibility of this package is to provide a command-line interface for interacting with the adapter. It parses command-line arguments, invokes the appropriate adapter logic, and formats the output. This CLI acts as a bridge between external systems and the core adapter functionality.

### Key Types

The following types define the data structures used for communication:

*   **`InfoResponse`**: Represents the adapter's information, including its name, version, and supported capabilities.
    ```go
    type InfoResponse struct {
    	Name         string   `json:"name"`
    	Version      string   `json:"version"`
    	Capabilities []string `json:"capabilities"`
    }
    ```
*   **`TransformResponse`**:  Encapsulates the result of a transformation operation, including status, adapter metadata, and output details.
    ```go
    type TransformResponse struct {
    	Status  string          `json:"status"`
    	Adapter AdapterMetadata `json:"adapter"`
    	Output  OutputMetadata  `json:"output"`
    }
    ```
*   **`AdapterMetadata`**: Contains metadata about the adapter itself, such as its name, version, and the underlying data substrate.
    ```go
    type AdapterMetadata struct {
    	Name      string `json:"name"`
    	Version   string `json:"version"`
    	Substrate string `json:"substrate"`
    }
    ```
*   **`OutputMetadata`**:  Provides details about the transformation output, including the path to the generated SARIF file, a trace ID, and any additional metadata.
    ```go
    type OutputMetadata struct {
    	SarifPath string                 `json:"sarif_path"`
    	TraceID   string                 `json:"trace_id"`
    	Metadata  map[string]interface{} `json:"metadata"`
    }
    ```

### Important Functions

*   **`main()`**: This is the entry point of the application. It parses command-line arguments and dispatches execution to the appropriate handler function based on the provided command (`info`, `transform`, `validate`). If no command or an unknown command is provided, it prints usage instructions and exits.
*   **`handleInfo()`**:  This function handles the `info` command. It creates an `InfoResponse` object with the adapter's details and prints it as a JSON string to standard output.
*   **`handleTransform()`**: This function handles the `transform` command. It uses the `flag` package to parse command-line arguments specific to the transform operation (input path, output path, policy reference). It validates that the required `--input` and `--output` flags are provided. Currently, it includes a mock implementation that prints a message indicating the transformation process and then constructs and prints a `TransformResponse` object as a JSON string.
*   **`handleValidate()`**: This function handles the `validate` command. It uses the `flag` package to parse the `--input` flag. It validates that the `--input` flag is provided. Currently, it includes a mock implementation that prints a message indicating the validation process and then prints a simple JSON string indicating validation success with an empty violations list.

### Error Handling

The CLI employs basic error handling:

*   If the user provides insufficient arguments or an invalid command, an error message is printed to standard output, and the program exits with a non-zero exit code (1).
*   Within `handleTransform` and `handleValidate`, missing required flags result in an error message and program exit.
*   JSON marshaling errors are currently ignored (represented by the blank assignment `_` in `json.MarshalIndent`). In a production environment, these errors should be handled more robustly.

### Concurrency

This CLI does not currently employ any concurrency patterns (goroutines or channels). All operations are performed sequentially within the `main` function and its handler functions.

### Design Decisions

*   **Command-Based Structure**: The CLI is structured around a command-based approach, allowing for easy extension with new functionalities in the future.
*   **Flag Parsing**: The `flag` package is used for parsing command-line arguments, providing a standard and convenient way to handle options and parameters.
*   **JSON Output**:  The CLI outputs data in JSON format, making it easy to integrate with other systems and tools.
*   **Mock Implementations**: The `transform` and `validate` commands currently contain mock implementations. This allows for initial development and testing without requiring the full adapter logic to be implemented. You will need to replace these with actual adapter calls.