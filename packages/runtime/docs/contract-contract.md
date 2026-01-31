---
type: Documentation
domain: runtime
origin: packages/runtime/internal/contract/contract.go
last_modified: 2026-01-31
generated: true
source: packages/runtime/internal/contract/contract.go
generated_at: 2026-01-31T09:59:05.077763
hash: a3f26726bb5f1246580a0587bbde1927f7284f9f574ef6261e317958a6f2bb26
---

## Deployment Contract Package Documentation

This package defines the schema for a Deployment Contract, representing the outcome of a deployment governance process. It provides a structured way to record details about the deployment, its quality, and an audit trail.

**Key Types**

*   **DeploymentContract**: The central type, encapsulating all information about a deployment’s contract. It includes metadata, status, quality metrics, and audit details.
    ```go
    type DeploymentContract struct {
    	SchemaVersion string  `json:"schemaVersion"`
    	Meta          Meta    `json:"meta"`
    	Status        string  `json:"status"`
    	Quality       Quality `json:"quality"`
    	Audit         Audit   `json:"audit"`
    }
    ```

*   **Meta**: Contains metadata about the deployment execution environment.
    ```go
    type Meta struct {
    	Adapter   string `json:"adapter"`
    	Engine    string `json:"engine"`
    	Timestamp string `json:"timestamp"`
    	Trigger   string `json:"trigger"`
    }
    ```

*   **Quality**: Holds code quality metrics associated with the deployment.
    ```go
    type Quality struct {
    	Coverage Coverage    `json:"coverage"`
    	Tests    TestResults `json:"tests"`
    }
    ```

*   **Coverage**: Represents code coverage information.
    ```go
    type Coverage struct {
    	Actual   float64 `json:"actual"`
    	Required float64 `json:"required"`
    	Met      bool    `json:"met"`
    }
    ```

*   **TestResults**: Stores the results of test execution.
    ```go
    type TestResults struct {
    	Total  int `json:"total"`
    	Passed int `json:"passed"`
    	Failed int `json:"failed"`
    }
    ```

*   **Audit**: Contains information for auditing the deployment.
    ```go
    type Audit struct {
    	TriggeredBy string `json:"triggeredBy"`
    	OrgID       string `json:"orgId"`
    	Repository  string `json:"repository"`
    	Commit      string `json:"commit"`
    }
    ```

*   **ValidationError**: A custom error type used to signal validation failures within the contract.
    ```go
    type ValidationError struct {
    	Field   string
    	Message string
    }
    ```

**Important Functions**

*   **New()**: Creates a new `DeploymentContract` instance with default values. The default status is set to "Succeeded", the adapter and engine are set to "native", and the timestamp is set to the current UTC time. The coverage requirement is initialized to 80.
    ```go
    func New() *DeploymentContract {
    	return &DeploymentContract{
    		SchemaVersion: "1.0",
    		Meta: Meta{
    			Adapter:   "native",
    			Engine:    "native",
    			Timestamp: time.Now().UTC().Format(time.RFC3339),
    		},
    		Status: "Succeeded",
    		Quality: Quality{
    			Coverage: Coverage{Required: 80},
    			Tests:    TestResults{},
    		},
    	}
    }
    ```

*   **ToJSON()**: Serializes a `DeploymentContract` instance into a JSON byte slice with indentation for readability.
    ```go
    func (c *DeploymentContract) ToJSON() ([]byte, error) {
    	return json.MarshalIndent(c, "", "  ")
    }
    ```

*   **Validate()**: Validates the `DeploymentContract` to ensure data integrity. It checks the validity of the `Status` and `Engine` fields against predefined allowed values, and verifies that `Coverage.Actual` and `Coverage.Required` fall within the range of 0 to 100. Returns a `ValidationError` if any validation fails; otherwise, returns nil.
    ```go
    func (c *DeploymentContract) Validate() error {
    	// ... validation logic ...
    	return nil
    }
    ```

**Error Handling**

The package employs a custom error type, `ValidationError`, to provide specific information about validation failures. This allows callers to easily identify which field failed validation and the reason for the failure. The `Error()` method on `ValidationError` provides a human-readable error message.

**Design Decisions**

*   **JSON Serialization**: The use of JSON for serialization allows for easy integration with other systems and services.
*   **Explicit Validation**: The `Validate()` method provides a clear and explicit way to ensure the integrity of the contract data.
*   **Default Values**: The `New()` function provides sensible default values, simplifying contract creation.
*   **Schema Versioning**: The `SchemaVersion` field allows for future evolution of the contract schema without breaking compatibility. You should increment this value when making changes to the contract structure.