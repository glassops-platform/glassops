---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/services/cli.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/services/cli.ts
generated_at: 2026-01-31T10:12:18.158699
hash: 973b422680240f152da76f14a615c66ecf48871b05f31f1e52fb30e52ddb2199
---

# GlassOps Runtime Environment Documentation

## Overview

This document describes the Runtime Environment service, responsible for managing the Salesforce CLI (sf) and its plugins. It provides functionality to install the CLI, install specified plugins, and verify their successful installation. This service is designed to operate within automated environments, such as CI/CD pipelines, and provides robust error handling and reporting.

## RuntimeEnvironment Class

The `RuntimeEnvironment` class encapsulates the logic for managing the Salesforce CLI and its plugins.

### `install(version)` Method

Installs the Salesforce CLI (sf) if it is not already present in the environment.

*   **Parameters:**
    *   `version` (optional):  The version of the Salesforce CLI to install. Defaults to "latest".
*   **Functionality:**
    1.  Checks if the Salesforce CLI is already installed. If so, the method returns without action.
    2.  If not installed, attempts to install the specified version of the Salesforce CLI using `npm install -g @salesforce/cli@${version}`.  This installation is retried up to three times with a two-second backoff between attempts to handle transient network issues.
    3.  Verifies the installation by running `sf version`.
    4.  Throws an error if the installation fails, indicating a potential issue with the NPM registry.
*   **Output:** None.

### `installPlugins(config, plugins)` Method

Installs a list of Salesforce CLI plugins.

*   **Parameters:**
    *   `config`: A configuration object containing governance settings, including a plugin whitelist.
    *   `plugins`: An array of strings, where each string represents the name of a plugin to install.
*   **Functionality:**
    1.  If the `plugins` array is empty, the method returns without action.
    2.  Iterates through the provided `plugins` array.
    3.  For each plugin:
        *   Validates the plugin against a configured whitelist (if one exists). If no whitelist is configured, the plugin is installed without validation.
        *   If a whitelist is configured, the plugin must be present in the whitelist to proceed.  If not, an error is thrown.
        *   Retrieves any version constraints for the plugin from the configuration.
        *   Constructs the installation command, including the version constraint if specified.
        *   Installs the plugin using `sf plugins install`.
        *   Verifies the installation by listing installed plugins using `sf plugins --json` and confirming the presence of the newly installed plugin.
        *   Throws an error if the installation or verification fails.
*   **Output:** None.

### `execWithAutoConfirm(command, args)` Method

Executes a shell command with automatic confirmation of any prompts.

*   **Parameters:**
    *   `command`: The command to execute (e.g., "sf").
    *   `args`: An array of arguments to pass to the command.
*   **Functionality:**
    1.  Constructs the full command string.
    2.  Prepends `echo y |` to the command on non-Windows platforms and `echo y|` on Windows to automatically answer "yes" to any prompts.
    3.  Executes the command using `exec.exec`.
*   **Output:** None.

## Error Handling

The service incorporates robust error handling:

*   Installation failures are retried.
*   Plugin installation failures result in detailed error messages.
*   Whitelist validation prevents the installation of unauthorized plugins.
*   Installation verification ensures that plugins are installed correctly.

## Configuration

The `ProtocolConfig` object provides configuration options, including:

*   `governance.plugin_whitelist`: An array of strings representing allowed plugin names. If this is empty or not present, plugins are installed without validation.