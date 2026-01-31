---
type: Documentation
domain: platform
origin: packages/platform/sfdx-project.json
last_modified: 2026-01-31
generated: true
source: packages/platform/sfdx-project.json
generated_at: 2026-01-31T11:05:07.783357
hash: 06f30cefba9ff14ee0e89759a269061d3ef6688ba2cfef642e806bc82c07207d
---

# SFDX Project Configuration

This document details the structure of the `sfdx-project.json` file. This file is the central configuration for projects managed with the Salesforce DX (SFDX) command-line interface. It defines the project’s structure, Salesforce organization connection details, and package definitions.

## Overview

The `sfdx-project.json` file instructs the SFDX CLI how to interact with a Salesforce project. It specifies which directories contain source code, the default package for commands, and other project-level settings. We use this file to manage metadata-driven development for Salesforce.

## Schema Details

The `sfdx-project.json` file is a JSON object with the following key attributes:

### `packageDirectories` (Array of Objects, Required)

This array defines the packages within the project. Each object in the array represents a single package.

*   **`path`** (String, Required): The relative path to the directory containing the package’s source code.  Example: `"glassops-platform"`
*   **`default`** (Boolean, Required):  Indicates whether this package is the default package for SFDX commands. Only one package can be designated as the default.  Example: `true` or `false`
*   **`package`** (String, Required): The name of the package. This is a human-readable identifier. Example: `"GlassOps"`
*   **`versionName`** (String, Required): A user-friendly version name for the package. Example: `"ver 1.0"`
*   **`versionNumber`** (String, Required): The version number of the package, following semantic versioning principles. The `.NEXT` suffix indicates an unreleased version. Example: `"1.0.0.NEXT"`

### `name` (String, Required)

The name of the project. This is a human-readable identifier. Example: `"glassops"`

### `namespace` (String, Optional)

The namespace for the project. If not specified, the project operates in the default namespace. Example: `""` (empty string for no namespace)

### `sfdcLoginUrl` (String, Optional)

The URL used to log in to a Salesforce organization. Defaults to `https://login.salesforce.com`. You may need to change this for sandboxes or other specific environments. Example: `"https://login.salesforce.com"`

### `sourceApiVersion` (String, Required)

The Salesforce API version used for source migration and deployment.  This value must be a valid Salesforce API version string (e.g., `"60.0"`). Example: `"60.0"`

## Common Use Cases

*   **Package Development:** Defining multiple packages within a single project allows for modular development and version control of different components.
*   **Default Package Selection:** Setting the `default` flag simplifies command execution by automatically targeting the most frequently used package.
*   **Source Management:**  The `packageDirectories` array enables the SFDX CLI to correctly identify and manage source code within the project.
*   **Environment Configuration:** The `sfdcLoginUrl` allows you to specify the correct login URL for different Salesforce environments.
*   **API Version Control:**  The `sourceApiVersion` ensures compatibility between your project’s source code and the target Salesforce organization.

## Example

The provided `sfdx-project.json` file defines two packages:

1.  `GlassOps`: Located in the `glassops-platform` directory, not the default package. Version `1.0.0.NEXT`.
2.  `GlassOps Implementation`: Located in the `force-app` directory, designated as the default package. Version `0.1.0.NEXT`.

The project is named `glassops`, uses no namespace, connects to the production Salesforce instance, and targets API version 60.0.

You should modify this file to reflect the specific structure and requirements of your Salesforce project.