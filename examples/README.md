# GlassOps Examples

Example configuration and workflow templates for the runtime in `packages/runtime`.

## Current Examples

- **[Deployment Workflow](workflows/glassops-deploy.yml)**: A GitHub Actions workflow that runs the runtime and uploads its contract. It does not deploy. Add your own authenticated deploy step.
- **[Configuration Template](config/devops-config.json)**: A starter `devops-config.json` using the keys the runtime reads: freeze windows (compared in UTC) and the analyzer. Save it as `config/devops-config.json`.
- **[MegaLinter sketch](workflows/glassops-scanner-megalinter.yml)**: A conceptual sketch only. No scanner adapter exists.
