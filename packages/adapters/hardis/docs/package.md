---
type: Documentation
domain: hardis
origin: packages/adapters/hardis/package.json
last_modified: 2026-01-26
generated: true
source: packages/adapters/hardis/package.json
generated_at: 2026-01-26T05:06:10.005Z
hash: 4a3b5ff12173fb8e2e47de0adb06a41f9846fe418159db048a4071bcd1435d46
---

# `@glassops/hardis-adapter` Package Documentation

This document details the `package.json` file for the Hardis adapter within the GlassOps ecosystem. This adapter facilitates integration with Hardis systems.

## Overview

The `package.json` file is a core component of any Node.js package. It defines metadata about the package, including its name, version, and dependencies.  This specific file describes the Hardis adapter, a module designed to connect and interact with Hardis infrastructure.

## Fields

### `name`

*   **Type:** String
*   **Value:** `@glassops/hardis-adapter`
*   **Required:** Yes
*   **Description:**  The unique identifier for this package within the Node Package Manager (npm) registry and the GlassOps architecture. The `@glassops` scope indicates this package is maintained by the GlassOps organization.

### `version`

*   **Type:** String
*   **Value:** `1.0.0`
*   **Required:** Yes
*   **Description:**  The current version of the Hardis adapter.  Semantic Versioning (SemVer) is employed (Major.Minor.Patch).

### `private`

*   **Type:** Boolean
*   **Value:** `true`
*   **Required:** No
*   **Description:**  Indicates that this package is not intended for public publication to the npm registry.  It is meant to be used internally within the GlassOps system and its dependencies. This prevents accidental or unauthorized external use.

## Use Cases

*   **Integration with Hardis Systems:** The primary use case is to provide a standardized interface for interacting with Hardis environments from within GlassOps workflows.
*   **Automated Infrastructure Management:**  This adapter enables automation of tasks related to Hardis infrastructure, such as deployment, configuration, and monitoring.
*   **Internal Dependency:**  Other GlassOps packages rely on this adapter to provide Hardis-specific functionality.
*   **Development & Testing:** Used during the development and testing phases of GlassOps features that require Hardis integration.