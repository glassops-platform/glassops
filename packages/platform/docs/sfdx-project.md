---
type: Documentation
domain: platform
origin: packages/platform/sfdx-project.json
last_modified: 2026-02-01
generated: true
source: packages/platform/sfdx-project.json
generated_at: 2026-02-01T19:37:09.228996
hash: 06f30cefba9ff14ee0e89759a269061d3ef6688ba2cfef642e806bc82c07207d
---

# SFDX Project Configuration

This document details the structure of the `sfdx-project.json` file. This file is central to managing a project within the platform and defines how the Salesforce DX CLI interacts with your source code and Salesforce organization. It provides metadata about the project, including package definitions, source API version, and login URL.

## Overview

The `sfdx-project.json` file configures a Salesforce DX project. It tells the Salesforce DX CLI where to find your source code, how to package it, and which Salesforce organization to connect to.  We use this file to manage multiple packages within a single project. You should place this file at the root of your project.

## Schema Details

The `sfdx-project.json` file is a JSON object with the following key-value pairs:

### `packageDirectories` (Array of Objects, Required)

This array defines the packages within your project. Each object in the array represents a single package.

*   **`path`** (String, Required): The relative path to the directory containing the package's source code.
*   **`default`** (Boolean, Required):  Indicates whether this package is the default package for certain DX commands. Only one package can be designated as the default.
*   **`package`** (String, Required): The name of the package. This is the identifier used when managing the package.
*   **`versionName`** (String, Required): A human-readable version name for the package (e.g., "ver 1.0").
*   **`versionNumber`** (String, Required): The version number of the package, following semantic versioning (e.g., "1.0.0.NEXT"). The `.NEXT` suffix indicates an unreleased version.

### `name` (String, Required)

The name of the project. This is a descriptive name for your overall development effort.

### `namespace` (String, Optional)

The namespace for your Salesforce organization. If not specified, it defaults to an empty string.  This is important for managing customizations in a multi-tenant environment.

### `sfdcLoginUrl` (String, Optional)

The login URL for your Salesforce organization. Defaults to `https://login.salesforce.com`. You may need to change this for sandboxes or other specific environments.

### `sourceApiVersion` (String, Required)

The Salesforce API version that your project is designed for.  This ensures compatibility between your code and the Salesforce platform.  We recommend keeping this up to date with the latest generally available API version.

## Common Use Cases

*   **Managing Multiple Packages:**  The `packageDirectories` array allows you to organize your project into logical packages, such as a managed package and an unmanaged package for implementation details.
*   **Defining Default Package:** Setting `default` to `true` for a package simplifies commands by making that package the target for operations like source:pull and source:push.
*   **Version Control:** The `versionName` and `versionNumber` fields enable you to track and manage different versions of your packages.
*   **Sandbox/Production Configuration:**  The `sfdcLoginUrl` field allows you to easily switch between different Salesforce environments.
*   **API Compatibility:** The `sourceApiVersion` field ensures that your project remains compatible with the Salesforce platform as it evolves.

## Example

The provided example `sfdx-project.json` file defines two packages:

*   `GlassOps`: Located in the `glassops-platform` directory, this package is not the default.
*   `GlassOps Implementation`: Located in the `force-app` directory, this package *is* the default.

The project is named `glassops`, uses an empty namespace, connects to the standard Salesforce login URL, and is designed for API version 60.0.