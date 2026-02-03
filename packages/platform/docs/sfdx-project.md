---
type: Documentation
domain: platform
last_modified: 2026-02-02
generated: true
source: packages/platform/sfdx-project.json
generated_at: 2026-02-02T22:33:29.945104
hash: 06f30cefba9ff14ee0e89759a269061d3ef6688ba2cfef642e806bc82c07207d
---

# SFDX Project Configuration

This document details the structure and purpose of the `sfdx-project.json` file. This file is the central configuration for projects managed with the Salesforce DX (SFDX) command-line interface. It defines how SFDX interacts with your Salesforce organization and manages your source code.

## Overview

The `sfdx-project.json` file provides metadata about your SFDX project. It specifies the package directories, project name, namespace, Salesforce login URL, and the API version of your target Salesforce organization. We use this file to understand the project's structure and dependencies when performing operations like source retrieval, deployment, and testing.

## Schema Details

The `sfdx-project.json` file is a JSON object with the following key attributes:

### `packageDirectories` (Array of Objects, Required)

This array defines the packages within your project. Each object in the array represents a single package.

*   **`path`** (String, Required): The relative path to the directory containing the package's source code.
*   **`default`** (Boolean, Required):  Indicates whether this package is the default package for SFDX commands. Only one package can be designated as the default.
*   **`package`** (String, Required): The name of the package. This is the logical name used to identify the package.
*   **`versionName`** (String, Required): A human-readable version name for the package (e.g., "ver 1.0").
*   **`versionNumber`** (String, Required): The version number of the package, following semantic versioning (e.g., "1.0.0.NEXT"). The `.NEXT` suffix indicates an unreleased version.

In the provided example, two package directories are defined:

1.  `glassops-platform`: Contains the core "GlassOps" package. It is *not* the default package.
2.  `force-app`: Contains the "GlassOps Implementation" package. This *is* the default package.

### `name` (String, Required)

The name of the SFDX project. This is a descriptive name for your overall project. In the example, the project name is "glassops".

### `namespace` (String, Required)

The namespace for your Salesforce organization. If you are working in a Developer Edition organization or a scratch org without a namespace, this value will be an empty string ("").  You should update this if you are deploying to an organization with a custom namespace.

### `sfdcLoginUrl` (String, Required)

The URL used to log in to your Salesforce organization. The default value is `https://login.salesforce.com` for production environments. You may need to change this for sandboxes or other specific environments.

### `sourceApiVersion` (String, Required)

The API version of the target Salesforce organization. This ensures compatibility between your source code and the organization.  In the example, the source API version is "60.0". You should update this to match the API version of the organization you are deploying to.

## Common Use Cases

*   **Project Setup:** When you initialize a new SFDX project, we create this file to define the project's initial configuration.
*   **Source Management:**  SFDX uses this file to determine which directories contain Salesforce source code and how to manage those sources.
*   **Deployment:** During deployment, SFDX uses the `packageDirectories` information to package and deploy your code to the target organization.
*   **Scratch Org Creation:** When creating scratch orgs, SFDX uses the `sourceApiVersion` to ensure the scratch org is created with the correct API version.
*   **Package Versioning:** The `versionName` and `versionNumber` attributes allow you to track and manage different versions of your packages.

## Updating the Configuration

You can modify this file directly using a text editor. However, it is recommended to use the SFDX CLI commands to manage the configuration. For example, you can use the `sfdx project update` command to update the `sourceApiVersion`.  Always validate the JSON syntax after making changes.