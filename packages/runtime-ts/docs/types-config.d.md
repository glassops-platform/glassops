---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/types/config.d.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/types/config.d.ts
generated_at: 2026-01-31T10:15:22.709139
hash: 10e0f67f12da9b5de1b7aeac604b2417e83996caf4f4cf3d95d537e6db4127ee
---

# GlassOps Configuration Reference

This document details the structure of the GlassOps configuration file. It provides a comprehensive overview of available settings and their purpose, intended for both developers and operators.

## Overview

The configuration file defines the behavior of the GlassOps system. It controls aspects such as execution environment, governance policies, environment-specific settings, and notification preferences. You can extend the configuration with custom properties as needed.

## Configuration Structure

The top-level configuration is a JSON object conforming to the `GlassOpsConfiguration` interface.

### `version`

*Type:* `"1.0"`
*Description:* Specifies the configuration file version. Currently, only `"1.0"` is supported.

### `metadata` (Optional)

*Type:* `object`
*Description:* Contains metadata about the configuration itself.

*   `last_updated` (Optional): *Type:* `string` - Timestamp of the last configuration update.
*   `schema_version` (Optional): *Type:* `string` - Version of the configuration schema used.

### `execution`

*Type:* `object`
*Description:* Defines the execution environment for GlassOps operations.

*   `engine` *Required:* *Type:* `"native" | "hardis"` - Specifies the execution engine. `"native"` indicates the primary execution environment, while `"hardis"` designates an alternative.
*   `fallback` (Optional): *Type:* `"native" | "none"` - Defines a fallback engine in case the primary engine is unavailable. `"native"` will attempt to fall back to the primary engine, while `"none"` disables fallback behavior.

### `governance`

*Type:* `object`
*Description:* Sets governance policies for operations.

*   `minCoverage` *Required:* *Type:* `number` - Minimum code coverage percentage required for operations.
*   `requireTests` *Required:* *Type:* `boolean` - Indicates whether tests are required before operations can proceed.
*   `enforcedBy` (Optional): *Type:* `string` - Identifies the entity responsible for enforcing these policies. This is primarily for documentation and audit purposes.

### `environments`

*Type:* `object`
*Description:* Defines configurations for different environments (e.g., development, staging, production). Each key represents an environment name.

*   `display_name` (Optional): *Type:* `string` - A user-friendly name for the environment.
*   `branch_mapping` (Optional): *Type:* `string` - Maps a branch name to this environment.
*   `deployment_policy` (Optional): *Type:* `object` - Defines policies governing deployments to this environment.
    *   `test_level` (Optional): *Type:* `string` - The required test level for deployment.
    *   `wait_time` (Optional): *Type:* `string` - Time to wait after deployment before considering it successful.
    *   `use_delta` (Optional): *Type:* `boolean` - Whether to deploy only changes (delta) or the entire application.
    *   `validation_required` (Optional): *Type:* `boolean` - Indicates whether validation is required before deployment.
    *   `auto_deploy_on_merge` (Optional): *Type:* `boolean` - Enables automatic deployment upon merging to the mapped branch.
*   `quality_gates` (Optional): *Type:* `object` - Defines quality gate criteria for this environment.
    *   `minCoverage` (Optional): *Type:* `number` - Minimum code coverage required for this environment.
    *   `security_severity_threshold` (Optional): *Type:* `number` - Maximum allowed security severity level.
    *   `block_on_test_failure` (Optional): *Type:* `boolean` - Whether to block deployment if tests fail.
*   `github_environment` (Optional): *Type:* `string` - The corresponding GitHub environment name.
*   `notes` (Optional): *Type:* `string` -  Additional notes or documentation for the environment.

### `glassops` (Optional)

*Type:* `object`
*Description:* GlassOps specific settings.

*   `enablePlatformEvents` (Optional): *Type:* `boolean` - Enables or disables platform events.

### `notifications` (Optional)

*Type:* `object`
*Description:* Configures notification settings.

*   `enabled_by_default` (Optional): *Type:* `boolean` - Enables notifications by default.
*   `channels` (Optional): *Type:* `object` - Defines notification channels.
    *   `slack` (Optional): *Type:* `object`
        *   `enabled` (Optional): *Type:* `boolean` - Enables Slack notifications.
        *   `mention_on_failure` (Optional): *Type:* `boolean` - Mentions a designated user on failure.
    *   `email` (Optional): *Type:* `object`
        *   `enabled` (Optional): *Type:* `boolean` - Enables email notifications.

### Extensions

The configuration allows for arbitrary extensions. You can add any additional key-value pairs to the top-level object. These extensions will be preserved and can be accessed by the system or custom logic.