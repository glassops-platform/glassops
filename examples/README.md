# GlassOps Examples

Example workflow templates for the `glassops` CLI.

## Current Examples

- **[Deployment Workflow](workflows/glassops-deploy.yml)**: A GitHub Actions workflow that runs `plan`, `validate`, `gate`, `deploy` and `report` against `prod` on every push to `main`. The job runs in a protected environment named `glassops-deploy` that holds the PERMIT signing key. The [GitHub deploy guide](../docs/guides/deploy-with-github.md) explains each part.
- **[MegaLinter sketch](workflows/glassops-scanner-megalinter.yml)**: A conceptual sketch only. No scanner adapter exists.
