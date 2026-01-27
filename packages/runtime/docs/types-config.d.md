---
type: Documentation
domain: runtime
origin: packages/runtime/src/types/config.d.ts
last_modified: 2026-01-26
generated: true
source: packages/runtime/src/types/config.d.ts
generated_at: 2026-01-26T14:16:31.303Z
hash: c4f39a4628a6623f50e75d37ba5cd6ea99ceb674446975de1c81a0d9b4c222c3
---

## GlassOps Configuration Document

This document details the structure and options available within the GlassOps configuration file. This configuration governs the behavior of the system, defining execution parameters, governance policies, environment settings, and notification preferences.

**1. Core Configuration**

*   **version:** (String, required) Specifies the configuration schema version. Currently set to "1.0".
*   **metadata:** (Object, optional) Contains information about the configuration itself.
    *   **last\_updated:** (String, optional) Timestamp indicating the last modification date.
    *   **schema\_version:** (String, optional) Version of the schema used for this configuration.

**2. Execution Settings**

*   **engine:** (String, required) Determines the execution engine used by GlassOps. Options are:
    *   **native:**  Uses the native execution environment.
    *   **hardis:** Uses the Hardis execution environment.
*   **fallback:** (String, optional) Specifies a fallback engine in case the primary engine fails. Options are:
    *   **native:** Fallback to the native execution environment.
    *   **none:** No fallback engine is used.

**3. Governance Policies**

These settings define the quality and security standards enforced by GlassOps.

*   **minCoverage:** (Number, required) Minimum code coverage percentage required for successful execution.
*   **requireTests:** (Boolean, required)  Indicates whether tests are mandatory for execution.
*   **enforcedBy:** (String, optional) Identifies the entity responsible for enforcing these governance policies. This is primarily for documentation and audit purposes.

**4. Environment Configurations**

This section allows defining configurations for different environments (e.g., development, staging, production).

*   Each key represents a unique environment name.
*   **display\_name:** (String, optional) A user-friendly name for the environment.
*   **branch\_mapping:** (String, optional) Maps a Git branch to this environment.
*   **deployment\_policy:** (Object, optional) Controls the deployment process for this environment.
    *   **test\_level:** (String, optional) Specifies the required test level.
    *   **wait\_time:** (String, optional) Defines a waiting period before deployment.
    *   **use\_delta:** (Boolean, optional) Indicates whether to deploy only changes since the last deployment.
    *   **validation\_required:** (Boolean, optional) Specifies if validation steps are needed before deployment.
    *   **auto\_deploy\_on\_merge:** (Boolean, optional) Enables automatic deployment upon merging into the mapped branch.
*   **quality\_gates:** (Object, optional) Defines quality criteria that must be met before deployment.
    *   **minCoverage:** (Number, optional) Minimum code coverage for this environment.
    *   **security\_severity\_threshold:** (Number, optional) Maximum allowed security severity level.
    *   **block\_on\_test\_failure:** (Boolean, optional) Prevents deployment if tests fail.
*   **github\_environment:** (String, optional) Links this environment to a GitHub environment.
*   **notes:** (String, optional)  Allows adding descriptive notes about the environment.

**5. GlassOps Specific Settings**

*   **glassops:** (Object, optional) Contains settings specific to the GlassOps platform.
    *   **enablePlatformEvents:** (Boolean, optional) Enables or disables platform event emission.

**6. Notification Settings**

*   **notifications:** (Object, optional) Configures notification behavior.
    *   **enabled\_by\_default:** (Boolean, optional) Enables notifications globally by default.
    *   **channels:** (Object, optional) Defines notification channels.
        *   **slack:** (Object, optional) Configures Slack notifications.
            *   **enabled:** (Boolean, optional) Enables Slack notifications.
            *   **mention\_on\_failure:** (Boolean, optional) Mentions specific users in Slack upon failure.
        *   **email:** (Object, optional) Configures email notifications.
            *   **enabled:** (Boolean, optional) Enables email notifications.

**7. Extensibility**

*   The configuration supports arbitrary extensions. You can add custom keys and values to the root object to extend functionality. These extensions are not validated by the core schema.