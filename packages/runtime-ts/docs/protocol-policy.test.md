---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/protocol/policy.test.ts
last_modified: 2026-01-29
generated: true
source: packages/runtime-ts/src/protocol/policy.test.ts
generated_at: 2026-01-29T20:57:23.412597
hash: 3f4533598237bae7d4dbe66cd3c23671496fb9bde5d1b088263469c91d0aeb2e
---

## Protocol Policy Documentation

This document details the functionality of the Protocol Policy component, responsible for enforcing governance and security constraints during runtime operations. It outlines the policy’s features, configuration, and behavior.

**Overview**

The Protocol Policy provides mechanisms to control deployment timing and permitted plugins. It reads configuration from a `devops-config.json` file, falling back to a default, permissive policy if the file is absent.  The policy is designed to enhance operational safety and adherence to organizational standards.

**Key Features**

*   **Deployment Freeze Windows:**  The policy can prevent deployments during specified days and times. This feature is configurable via the `freeze_windows` array within the `devops-config.json` file.
*   **Plugin Whitelisting:** The policy enforces the use of approved plugins.  A `plugin_whitelist` array in the configuration defines the permitted plugins, optionally including version constraints.
*   **Configuration Loading:** The policy attempts to load configuration from a `devops-config.json` file located in the `config/` directory.
*   **Validation:** The policy validates the loaded configuration against a predefined schema, ensuring data integrity.

**Configuration**

The policy’s behavior is governed by the `devops-config.json` file. The following sections are relevant:

*   **`governance.enabled` (boolean):** Enables or disables governance checks. If disabled, all policy checks are bypassed.
*   **`governance.freeze_windows` (array of objects):** Defines periods when deployments are prohibited. Each object requires:
    *   `day` (string): The day of the week (e.g., "Monday", "Friday").
    *   `start` (string): The start time in HH:MM format (e.g., "09:00").
    *   `end` (string): The end time in HH:MM format (e.g., "17:00").
*   **`governance.plugin_whitelist` (array of strings):**  Lists allowed plugins.  Plugins can be specified with or without version constraints using the `@version` syntax (e.g., "sfdx-hardis@^6.0.0").
*   **`runtime` (object):** Contains runtime-specific information, but does not directly affect policy enforcement.

**Methods**

*   **`checkFreeze(config)`:**  Determines if the current time falls within a defined freeze window. Throws an error if a freeze is active.
*   **`validatePluginWhitelist(config, pluginName)`:** Checks if a given plugin is present in the configured whitelist. Returns `true` if the plugin is allowed, `false` otherwise.
*   **`getPluginVersionConstraint(config, pluginName)`:** Retrieves the version constraint for a specified plugin from the whitelist. Returns `null` if the plugin is not whitelisted or has no version constraint.
*   **`load()`:** Loads the configuration from `config/devops-config.json`. If the file is missing or invalid, it defaults to an unsafe policy (governance disabled) and issues a warning.

**Error Handling**

The policy handles the following error conditions:

*   **Invalid JSON:** Throws an error if the `devops-config.json` file contains invalid JSON.
*   **Invalid Schema:** Throws an error if the loaded configuration does not conform to the expected schema.
*   **File Not Found:**  Defaults to an unsafe policy and logs a warning if the `devops-config.json` file is not found.
*   **Unexpected Errors:** Catches and re-throws unexpected errors during configuration loading as a generic "Invalid Governance Policy" error.

**Usage**

You should call the `load()` method to initialize the policy with the configuration. Subsequently, use `checkFreeze()` and `validatePluginWhitelist()` to enforce governance rules before performing sensitive operations.  The `getPluginVersionConstraint()` method can be used to determine the allowed version range for a plugin.