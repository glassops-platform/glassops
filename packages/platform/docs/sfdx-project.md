---
type: Documentation
domain: platform
origin: packages/platform/sfdx-project.json
last_modified: 2026-01-26
generated: true
source: packages/platform/sfdx-project.json
generated_at: 2026-01-26T05:02:16.444Z
hash: 06f30cefba9ff14ee0e89759a269061d3ef6688ba2cfef642e806bc82c07207d
---

# sfdx-project.json Documentation

This document details the structure and purpose of the `sfdx-project.json` file. This file is the core configuration file for a Salesforce DX project, defining the project's structure, metadata organization, and connection settings. It is used by the Salesforce CLI (SFDX) to manage and deploy metadata to a Salesforce org.

## Data Representation

The `sfdx-project.json` file represents the overall configuration of a Salesforce development project. It defines how the project is organized into packages, which Salesforce org it connects to, and the API version used for metadata operations.  It essentially acts as a blueprint for the project, guiding the SFDX CLI in its operations.

## Schema Details

### Top-Level Properties

*   **`packageDirectories` (Array, Required):**  An array of objects, each defining a package within the project.  Packages are used to group related metadata for deployment and version control.
*   **`name` (String, Required):** The name of the project. This is a human-readable identifier.
*   **`namespace` (String, Optional):** The namespace prefix for the project. If left blank, no namespace is applied.  This is important for managing metadata in environments with multiple applications.
*   **`sfdcLoginUrl` (String, Optional):** The URL used to log in to a Salesforce org. Defaults to `https://login.salesforce.com`.  Can be used to specify a sandbox or other custom login URL.
*   **`sourceApiVersion` (String, Required):** The Salesforce API version used for metadata operations.  This ensures compatibility between the project and the target Salesforce org.

### `packageDirectories` Object Properties

Each object within the `packageDirectories` array represents a single package and has the following properties:

*   **`path` (String, Required):** The relative path to the directory containing the package's metadata. This is the root directory for the package.
*   **`default` (Boolean, Required):**  A flag indicating whether this package is the default package for SFDX commands. Only one package can be designated as the default.  The default package is used when commands don't explicitly specify a package.
*   **`package` (String, Required):** The name of the package. This is the identifier used for the package in Salesforce.
*   **`versionName` (String, Optional):** A human-readable version name for the package (e.g., "ver 1.0").
*   **`versionNumber` (String, Required):** The version number of the package, following semantic versioning (e.g., "1.0.0.NEXT"). The `.NEXT` suffix indicates a version that is still in development.

## Common Use Cases

*   **Project Setup:**  Creating a new Salesforce DX project with a defined structure and configuration.
*   **Metadata Management:**  Organizing metadata into logical packages for easier deployment and version control.
*   **Deployment:**  Deploying specific packages to Salesforce orgs.
*   **Source Control:**  Integrating the project with a source control system (e.g., Git) to track changes and collaborate with other developers.
*   **Continuous Integration/Continuous Delivery (CI/CD):** Automating the build, test, and deployment process.
*   **Multi-Package Development:** Managing projects that contain multiple independent packages.
*   **Namespace Management:** Defining and applying a namespace to the project's metadata.