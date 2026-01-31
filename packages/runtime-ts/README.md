# GlassOps Runtime

> [!CAUTION]
> **DEPRECATED**: This TypeScript-based runtime is being replaced by the high-performance Go runtime located in `packages/runtime`. Please migrate your workflows.

> A governed Salesforce execution environment for GitHub Actions

GlassOps Runtime bootstraps a secure, policy-compliant environment for Salesforce deployments, ensuring governance controls are enforced before any CLI operation.

## Features

- **Policy Enforcement**: Freeze windows, test coverage requirements, and plugin whitelisting
- **Deployment Contracts**: Cryptographically signed audit trail for every deployment
- **Static Code Analysis**: Optional PMD/ESLint integration for code quality gates
- **Plugin Management**: Controlled installation with version constraints
- **Secure Authentication**: JWT-based authentication with secure key handling
- **Docker Container**: Consistent, reproducible execution environment

## Quick Start

### Basic Usage

```yaml
- name: Setup GlassOps Runtime
  uses: glassops/runtime@v1
  with:
      jwt_key: ${{ secrets.SALESFORCE_JWT_KEY }}
      client_id: ${{ secrets.SALESFORCE_CLIENT_ID }}
      username: ${{ secrets.SALESFORCE_USERNAME }}
```

### Advanced Usage with Policy Enforcement

```yaml
- name: Setup Governed Runtime
  id: glassops
  uses: glassops/runtime@v1
  with:
      jwt_key: ${{ secrets.SALESFORCE_JWT_KEY }}
      client_id: ${{ secrets.SALESFORCE_CLIENT_ID }}
      username: ${{ secrets.SALESFORCE_USERNAME }}
      instance_url: 'https://test.salesforce.com'
      enforce_policy: 'true'
      plugins: 'sfdx-hardis@^4.0.0,@salesforce/plugin-deploy-retrieve'
      coverage_required: '85'
      test_results: '{"total": 150, "passed": 148, "failed": 2}'

- name: Deploy if Governance Passes
  if: steps.glassops.outputs.glassops_ready == 'true'
  run: sf project deploy start
```

## Inputs

| Input                 | Description                            | Required | Default                        |
| --------------------- | -------------------------------------- | -------- | ------------------------------ |
| `jwt_key`             | Private key for JWT authentication     | Yes      | -                              |
| `client_id`           | Salesforce Connected App consumer key  | Yes      | -                              |
| `username`            | Target org username                    | Yes      | -                              |
| `instance_url`        | Salesforce login URL                   | No       | `https://login.salesforce.com` |
| `enforce_policy`      | Enable governance policy checks        | No       | `true`                         |
| `test_results`        | JSON with test results                 | No       | -                              |
| `coverage_percentage` | Actual code coverage percentage        | No       | `0`                            |
| `coverage_required`   | Required coverage threshold            | No       | `80`                           |
| `plugins`             | Comma-separated CLI plugins to install | No       | -                              |
| `skip_auth`           | Skip authentication (for testing)      | No       | `false`                        |
| `config_path`         | Path to devops-config.json             | No       | `config/devops-config.json`    |

## Outputs

| Output           | Description                             |
| ---------------- | --------------------------------------- |
| `runtime_id`     | Unique session identifier               |
| `org_id`         | Authenticated Salesforce org ID         |
| `is_locked`      | Whether deployment is blocked by policy |
| `contract_path`  | Path to deployment contract JSON        |
| `glassops_ready` | Whether runtime is ready for execution  |

## Configuration

Create `config/devops-config.json` in your repository:

```json
{
    "version": "1.0",
    "execution": {
        "engine": "native"
    },
    "governance": {
        "minCoverage": 75,
        "requireTests": true,
        "enforcedBy": "Platform Team",
        "freeze_windows": [
            {
                "day": "Friday",
                "start": "18:00",
                "end": "23:59"
            },
            {
                "day": "Saturday",
                "start": "00:00",
                "end": "23:59"
            }
        ],
        "plugin_whitelist": ["sfdx-hardis@^4.0.0", "@salesforce/plugin-deploy-retrieve", "@salesforce/plugin-data"],
        "analyzer": {
            "enabled": true,
            "severity_threshold": 2,
            "rulesets": ["security", "performance"],
            "opinionated": true
        }
    },
    "runtime": {
        "cli_version": "latest",
        "node_version": "20"
    }
}
```

### Configuration Schema

See [devops-config.schema.json](../../config/devops-config.schema.json) for the complete JSON schema.

## Security

### JWT Key Setup

1. Generate a private key and certificate:

```bash
openssl req -x509 -sha256 -nodes -days 36500 -newkey rsa:2048 \
  -keyout server.key -out server.crt
```

2. Create a Connected App in Salesforce with OAuth settings

3. Store the private key in GitHub Secrets as `SALESFORCE_JWT_KEY`

### Best Practices

- Always use GitHub Secrets for sensitive inputs
- Enable `enforce_policy` in production workflows
- Pin plugin versions with exact constraints
- Use freeze windows for deployment blackouts
- Review deployment contracts regularly
- Never commit JWT keys to version control
- Don't use `skip_auth` in production

## Deployment Contract

Every execution generates a deployment contract:

```json
{
    "schemaVersion": "1.0",
    "meta": {
        "adapter": "native",
        "engine": "native",
        "timestamp": "2026-01-27T10:30:00Z",
        "trigger": "push"
    },
    "status": "Succeeded",
    "quality": {
        "coverage": {
            "actual": 87.5,
            "required": 80,
            "met": true
        },
        "tests": {
            "total": 150,
            "passed": 148,
            "failed": 2
        }
    },
    "audit": {
        "triggeredBy": "github-user",
        "orgId": "00D1234567890ABC",
        "repository": "org/repo",
        "commit": "abc123def456"
    }
}
```

## Static Code Analysis

Enable PMD-based code analysis:

```json
{
    "governance": {
        "analyzer": {
            "enabled": true,
            "severity_threshold": 2,
            "rulesets": ["security"],
            "opinionated": true
        }
    }
}
```

Violations above the severity threshold will block deployment.

## Environment-Specific Policies

```json
{
    "environments": {
        "prod": {
            "quality_gates": {
                "minCoverage": 85,
                "block_on_test_failure": true
            },
            "deployment_policy": {
                "test_level": "RunAllTestsInOrg",
                "validation_required": true
            }
        },
        "dev": {
            "quality_gates": {
                "minCoverage": 0
            },
            "deployment_policy": {
                "test_level": "NoTestRun"
            }
        }
    }
}
```

## Troubleshooting

### Authentication Fails

**Error**: `Authentication Failed. Check Client ID and JWT Key.`

**Solution**:

- Verify JWT key format (must include BEGIN/END markers)
- Check Connected App settings in Salesforce
- Ensure username is authorized for JWT flow
- Verify `instance_url` matches org type (login vs. test)

### Policy Violation

**Error**: `FROZEN: Deployment blocked by governance window`

**Solution**:

- Check current UTC time against freeze windows
- Request emergency override (if process exists)
- Wait until freeze window expires

### Plugin Installation Fails

**Error**: `Plugin 'xyz' is not in the whitelist`

**Solution**:

- Add plugin to `plugin_whitelist` in devops-config.json
- Verify spelling and scoping (e.g., `@salesforce/plugin-name`)
- Check for typos in version constraints

## Contributing

Contributions welcome! Please read [CONTRIBUTING.md](../../CONTRIBUTING.md) first.

### Development Setup

```bash
# Clone repository
git clone https://github.com/glassops/runtime.git
cd runtime

# Install dependencies
npm install

# Build
npm run build

# Run tests
npm test

# Lint
npm run lint
```

### Testing Locally

```yaml
# Use local action in workflow
- uses: ./
  with:
      skip_auth: 'true'
      enforce_policy: 'false'
```

## Licenses

Code: Apache-2.0 © Ryan Bumstead
Docs: C.C 4.0 © Ryan Bumstead

## Acknowledgments

- Salesforce DX CLI Team
- GitHub Actions Team
- Open Source Community
