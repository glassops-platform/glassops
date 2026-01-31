package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// --- Contract Definitions (Matching Protobuf) ---

type InfoResponse struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
}

type TransformResponse struct {
	Status  string          `json:"status"`
	Adapter AdapterMetadata `json:"adapter"`
	Output  OutputMetadata  `json:"output"`
}

type AdapterMetadata struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Substrate string `json:"substrate"`
}

type OutputMetadata struct {
	SarifPath string                 `json:"sarif_path"`
	TraceID   string                 `json:"trace_id"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// --- Main CLI Logic ---

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: adapter <command> [args]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "info":
		handleInfo()
	case "transform":
		handleTransform()
	case "validate":
		handleValidate()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func handleInfo() {
	resp := InfoResponse{
		Name:         "salesforce-native-adapter",
		Version:      "1.0.0",
		Capabilities: []string{"transform", "validate"},
	}
	jsonBytes, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(jsonBytes))
}

func handleTransform() {
	transformCmd := flag.NewFlagSet("transform", flag.ExitOnError)
	input := transformCmd.String("input", "", "Path to input source")
	output := transformCmd.String("output", "", "Path to output SARIF")
	policyRef := transformCmd.String("policy-ref", "", "Policy reference")

	transformCmd.Parse(os.Args[2:])

	if *input == "" || *output == "" {
		fmt.Println("Error: --input and --output are required")
		os.Exit(1)
	}

	// Mock Implementation for Week 1 Foundation
	fmt.Printf("Transforming input: %s using policy: %s\n", *input, *policyRef)

	resp := TransformResponse{
		Status: "success",
		Adapter: AdapterMetadata{
			Name:      "salesforce-native-adapter",
			Version:   "1.0.0",
			Substrate: "salesforce",
		},
		Output: OutputMetadata{
			SarifPath: *output,
			TraceID:   "mock-trace-id-123",
			Metadata: map[string]interface{}{
				"components_processed": 10,
				"mock_mode":            true,
			},
		},
	}
	jsonBytes, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(jsonBytes))
}

func handleValidate() {
	validateCmd := flag.NewFlagSet("validate", flag.ExitOnError)
	input := validateCmd.String("input", "", "Path to input source")
	validateCmd.Parse(os.Args[2:])

	if *input == "" {
		fmt.Println("Error: --input is required")
		os.Exit(1)
	}

	fmt.Printf("Validating input: %s\n", *input)
	// Mock response
	fmt.Println(`{"valid": true, "violations": []}`)
}
