---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/src/types/config.d.ts
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/src/types/config.d.ts
generated_at: 2026-01-31T09:18:38.204468
hash: 10e0f67f12da9b5de1b7aeac604b2417e83996caf4f4cf3d95d537e6db4127ee
---

## GlassOps Configuration Document

This document details the structure of the GlassOps configuration file. This configuration drives the behavior of the system, defining execution parameters, governance rules, environment settings, and notification preferences.

**Configuration Version:**

The `version` field, currently set to "1.0", indicates the configuration schema version. This allows for future compatibility and updates.

**Metadata:**

The optional `metadata` section provides information about the configuration itself.
*   `last_updated`: A timestamp indicating the last modification date.
*   `schema_version`:  Specifies the version of the configuration schema used.

**Execution Settings:**

The `execution` section controls how GlassOps operates.
*   `engine`: Determines the execution engine. Options are "native" or "hardis".
*   `fallback`: Specifies a fallback engine in case the primary engine fails. Options are "native" or "none".

**Governance Policies:**

The `governance` section defines quality and security standards.
*   `minCoverage`:  The minimum acceptable code coverage percentage.
*   `requireTests`: A boolean value indicating whether tests are mandatory.
*   `enforcedBy`:  An optional string identifying the entity responsible for enforcing these policies.

**Environment Configurations:**

The `environments` section defines settings for each deployment environment.  Each environment is identified by a unique key.
*   `display_name`: A user-friendly name for the environment.
*   `branch_mapping`:  Associates the environment with a specific branch.
*   `deployment_policy`: Controls deployment behavior.
    *   `test_level`: Specifies the required test level.
    *   `wait_time`: Defines a waiting period before deployment.
    *   `use_delta`:  Indicates whether to deploy only changes.
    *   `validation_required`:  Requires validation before deployment.
    *   `auto_deploy_on_merge`: Automatically deploys on branch merge.
*   `quality_gates`: Defines quality criteria for the environment.
    *   `minCoverage`: Minimum code coverage for this environment.
    *   `security_severity_threshold`:  Maximum acceptable security severity level.
    *   `block_on_test_failure`:  Prevents deployment if tests fail.
*   `github_environment`:  Links to a corresponding GitHub environment.
*   `notes`:  Allows for adding descriptive notes about the environment.

**GlassOps Platform Features:**

The optional `glassops` section controls platform-level features.
*   `enablePlatformEvents`: Enables or disables platform events.

**Notification Settings:**

The optional `notifications` section configures how GlassOps sends notifications.
*   `enabled_by_default`:  Enables or disables notifications globally.
*   `channels`: Defines notification channels.
    *   `slack`: Configures Slack notifications.
        *   `enabled`: Enables or disables Slack notifications.
        *   `mention_on_failure`: Mentions relevant users on failure.
    *   `email`: Configures email notifications.
        *   `enabled`: Enables or disables email notifications.

**Extensibility:**

The configuration supports extensions. You can add custom key-value pairs at the top level of the configuration object to extend functionality.  These extensions are dynamically handled by the system.