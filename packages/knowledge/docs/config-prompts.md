---
type: Documentation
domain: knowledge
last_modified: 2026-03-08
generated: true
source: packages/knowledge/config/prompts.yml
generated_at: 2026-03-08T22:45:05.045080
hash: 592636b8dd0caae33b20d2c8b81cf201f64a132bee8eba86a55956a6a780fee7
---

# Prompts Configuration Documentation

This YAML file configures the prompts used by the documentation agent within the system. It defines different system and user prompts for various adapter types, enabling tailored documentation generation for different file types and technologies.

## Structure

The file is structured around a top-level key `prompts`. This key contains a dictionary where each key represents an adapter type (e.g., `go`, `py`, `yml`).  A special key `_shared_rules` defines common instructions applied to all adapters. A `default` adapter is also provided as a fallback.

## `_shared_rules`

This key defines a multi-line string containing a set of strict rules that all adapters must adhere to during documentation generation. These rules govern the style, tone, and content of the generated documentation.  Specifically, it prohibits the use of emojis and certain words, mandates the use of "We" or "I" when referring to the project, and forbids mentioning specific names. It also emphasizes the importance of valid Markdown syntax and clarifies the meaning of "IT" in model names.

## Adapter-Specific Configurations

Each adapter type (e.g., `go`, `py`, `ts`, `yml`, `json`, `dockerfile`, `tf`, `apex`, `lwc`, `xml`, `default`, `aura`, `visualforce`, `flow`, `omniscript`, `dataraptor`, `integration_procedure`, `flexcard`, `profile`, `application`, `object`, `field`) has two keys:

*   `system`: A multi-line string defining the instructions given to the language model to guide its documentation generation process. This includes the role the model should assume (e.g., "principal architect", "DevOps engineer"), the desired output format, and specific areas to focus on.
*   `user`: A multi-line string defining the prompt sent to the language model, including the content to be documented. This typically includes a placeholder `{{content}}` which will be replaced with the actual file content during runtime.

## Adapter Details

Below is a summary of each adapter's purpose and focus:

*   **`go`**: Generates documentation for Go source code, focusing on package purpose, types, functions, error handling, and concurrency.
*   **`py`**: Generates documentation for Python source code, focusing on module purpose, classes, functions, type hints, and design patterns.
*   **`ts`**: Generates documentation for TypeScript/JavaScript files, aiming for a concise and comprehensive overview.
*   **`yml`**: Generates documentation for YAML configurations, explaining purpose, structure, and key controls.
*   **`json`**: Generates documentation for JSON schemas or data structures, focusing on data representation, required/optional fields, and use cases.
*   **`dockerfile`**: Generates documentation for Dockerfiles, explaining base images, stages, instructions, security, and build/run processes.
*   **`tf`**: Generates documentation for Terraform configurations, explaining resources, variables, outputs, dependencies, and security.
*   **`apex`**: Generates documentation for Salesforce Apex code, focusing on class/trigger purpose, methods, governor limits, and integration points.
*   **`lwc`**: Generates documentation for Salesforce Lightning Web Components, focusing on component purpose, properties, wire adapters, and event handling.
*   **`xml`**: Generates documentation for XML content, adapting the approach based on the detected type (automation, UI, security, configuration, deployment).
*   **`default`**: A fallback adapter for unknown file types, providing a general documentation approach.
*   **`aura`**: Generates documentation for Salesforce Aura components, focusing on purpose, attributes, handlers, and dependencies.
*   **`visualforce`**: Generates documentation for Salesforce Visualforce pages/components, focusing on purpose, controllers, fields, and logic flow.
*   **`flow`**: Generates documentation for Salesforce Flows, focusing on business logic, triggers, and data interactions.
*   **`omniscript`**: Generates documentation for Salesforce OmniScripts, focusing on user experience, process flow, and data collection.
*   **`dataraptor`**: Generates documentation for Salesforce DataRaptors, focusing on data transformation, input/output, and mapping logic.
*   **`integration_procedure`**: Generates documentation for Salesforce Integration Procedures, focusing on server-side logic, inputs/outputs, and steps executed.
*   **`flexcard`**: Generates documentation for Salesforce FlexCards, focusing on UI/UX, data display, and user actions.
*   **`profile`**: Generates documentation for Salesforce Profiles, focusing on access control, object permissions, and system permissions.
*   **`application`**: Generates documentation for Salesforce Applications, focusing on navigation, purpose, and assigned profiles.
*   **`object`**: Generates documentation for Salesforce Custom Objects, focusing on data model, options, sharing model, and relationships.
*   **`field`**: Generates documentation for Salesforce Custom Fields, focusing on field definition, data type, default values, and attributes.

All adapters share the `{{shared_rules}}` directive within their `system` prompt, ensuring consistent documentation style and quality.