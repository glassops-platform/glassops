# Adapter Development Guide (Draft)

> [!CAUTION]
> This specification is currently in **DRAFT** status. The Adapter Interface is subject to change before v1.0.0 finalization. Contributors are encouraged to experiment, but please pin dependencies and expect breaking changes.

> **Building execution adapters for GlassOps**

This guide walks you through creating a new adapter that integrates your preferred deployment tool with the GlassOps governance protocol.

---

## What is an Adapter?

An adapter is a **stateless worker** that conforms to the [GlassOps Protocol](../../glassspec/README.md).
It translates tool-specific output into the **Universal SARIF Contract** defined in [Adapter Interface](../../glassspec/README.md).

1. Consumes deployment intent.
2. Executes using your chosen tool.
3. Emits a standard **SARIF 2.1.0** contract.

Think of it as a translator between your tool and the governance layer.

> [!IMPORTANT]
> **Protocol Supremacy:** All adapters MUST emit [OASIS SARIF 2.1.0](https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html) format as defined in [Adapter Interface](../../glassspec/README.md). Do NOT invent custom schemas. See the Protocol's [Anti-Patterns section](../../glassspec/README.md) for scope boundaries.

---

## The Adapter Contract

Every adapter MUST implement this interface:

### Required Functions

```javascript
/**
 * Validate deployment without execution (check-only/dry-run)
 * @returns {DeploymentContract} Draft contract with validation results
 */
async function simulate(options) {
    // Run check-only deployment
    // Calculate coverage
    // Return draft contract
}

/**
 * Execute validated deployment using Quick Deploy
 * @param {string} validationId - ID from successful simulation
 * @returns {DeploymentContract} Final contract with deployment results
 */
async function execute(validationId, options) {
    // Use quick deploy with validation ID
    // Return final contract
}

/**
 * Normalize tool-specific errors to standard format
 * @param {Error} rawError - Tool-specific error object
 * @returns {NormalizedError} Standard error format
 */
function normalizeError(rawError) {
    // Convert tool errors to standard schema
}
```

---

## The Deployment Contract Schema

Your adapter must emit a **SARIF 2.1.0** contract as defined in the [GlassOps Adapter Interface](../../glassspec/README.md).

### Minimal Compliant SARIF Contract

```json
{
    "$schema": "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0-rtm.5.json",
    "version": "2.1.0",
    "runs": [
        {
            "tool": {
                "driver": {
                    "name": "glassops-mytool-adapter",
                    "version": "1.0.0",
                    "informationUri": "https://github.com/yourorg/glassops-mytool-adapter"
                }
            },
            "invocations": [
                {
                    "executionSuccessful": true,
                    "endTimeUtc": "2026-01-24T12:00:00Z",
                    "properties": {
                        "glassops": {
                            "deploymentId": "0Af5e000000abcXYZ",
                            "targetOrg": "production",
                            "testLevel": "RunLocalTests",
                            "componentsDeployed": 42,
                            "testsRun": 120
                        }
                    }
                }
            ],
            "results": [
                {
                    "ruleId": "COVERAGE_THRESHOLD",
                    "level": "error",
                    "message": {
                        "text": "Code Coverage (72%) is below threshold (75%)"
                    },
                    "locations": [
                        {
                            "physicalLocation": {
                                "artifactLocation": {
                                    "uri": "force-app/main/default/classes/AccountService.cls"
                                }
                            }
                        }
                    ]
                }
            ]
        }
    ]
}
```

### Key Mapping Rules

Adapters translate tool-specific concepts to SARIF fields:

| Tool Concept            | SARIF Field                      | Notes                                         |
| :---------------------- | :------------------------------- | :-------------------------------------------- |
| **Violation ID**        | `result.ruleId`                  | Must be stable and queryable.                 |
| **Severity**            | `result.level`                   | Map to `error`, `warning`, `note`, or `none`. |
| **Error Message**       | `result.message.text`            | Human-readable explanation.                   |
| **File Path**           | `location.physicalLocation.uri`  | Relative path from repo root.                 |
| **Line Number**         | `region.startLine`               | 1-based line number.                          |
| **Deployment Metadata** | `invocation.properties.glassops` | Custom metadata in property bag.              |

### What Goes in SARIF vs. What Doesn't

**✅ DO normalize into SARIF:**

- Policy violations (coverage failures, test failures)
- Code quality findings (static analysis results)
- Deployment decisions (approved, rejected, override)
- Governance outcomes

**❌ DO NOT normalize into SARIF:**

- CPU/Memory metrics → Use OpenTelemetry
- Live logs → Use OpenTelemetry Logs
- Trace IDs → Use CloudEvents Context
- Raw configuration state → Link to native store

See the Protocol's [Anti-Patterns section](../../glassspec/README.md) for complete guidance.

---

## Step-by-Step: Building Your First Adapter

### Step 1: Set Up the Project Structure

```bash
mkdir glassops-mytool-adapter
cd glassops-mytool-adapter

# Initialize package.json
npm init -y

# Install dependencies
npm install @actions/core @actions/exec
```

### Step 2: Create the Adapter Script

```javascript
// src/adapter.js
const core = require('@actions/core');
const exec = require('@actions/exec');
const fs = require('fs').promises;
const path = require('path');

class MyToolAdapter {
    constructor(options) {
        this.options = options;
    }

    async simulate() {
        core.info('Running check-only deployment...');

        let output = '';
        const options = {
            listeners: {
                stdout: (data) => {
                    output += data.toString();
                }
            }
        };

        try {
            // Execute your tool in check-only mode
            await exec.exec(
                'mytool',
                [
                    'deploy',
                    '--check-only',
                    '--source-dir',
                    this.options.sourceDir,
                    '--test-level',
                    this.options.testLevel
                ],
                options
            );

            // Parse output and build contract
            const results = this.parseOutput(output);
            return this.buildContract('draft', results);
        } catch (error) {
            return this.buildContract('failed', null, error);
        }
    }

    async execute(validationId) {
        core.info(`Executing quick deploy with validation ID: ${validationId}`);

        let output = '';
        const options = {
            listeners: {
                stdout: (data) => {
                    output += data.toString();
                }
            }
        };

        try {
            await exec.exec('mytool', ['deploy', '--quick-deploy', '--validation-id', validationId], options);

            const results = this.parseOutput(output);
            return this.buildContract('final', results);
        } catch (error) {
            return this.buildContract('failed', null, error);
        }
    }

    parseOutput(output) {
        // Parse your tool's output format
        // Extract: deployment ID, coverage, test results, etc.

        // Example for JSON output:
        const data = JSON.parse(output);

        return {
            deploymentId: data.id,
            coverage: data.coverage,
            tests: data.testResults,
            componentsDeployed: data.components.length
        };
    }

    buildContract(mode, results, error = null) {
        // Build SARIF 2.1.0 compliant contract
        const contract = {
            $schema: 'https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0-rtm.5.json',
            version: '2.1.0',
            runs: [
                {
                    tool: {
                        driver: {
                            name: 'glassops-mytool-adapter',
                            version: '1.0.0',
                            informationUri: 'https://github.com/yourorg/glassops-mytool-adapter'
                        }
                    },
                    invocations: [
                        {
                            executionSuccessful: !error,
                            endTimeUtc: new Date().toISOString(),
                            properties: {
                                glassops: {
                                    deploymentId: results?.deploymentId || 'unknown',
                                    mode: mode === 'draft' ? 'validate' : 'deploy',
                                    componentsDeployed: results?.componentsDeployed || 0,
                                    testsRun: results?.tests?.total || 0,
                                    triggeredBy: process.env.GITHUB_ACTOR,
                                    repository: process.env.GITHUB_REPOSITORY,
                                    ref: process.env.GITHUB_REF,
                                    commit: process.env.GITHUB_SHA,
                                    runUrl: `https://github.com/${process.env.GITHUB_REPOSITORY}/actions/runs/${process.env.GITHUB_RUN_ID}`
                                }
                            }
                        }
                    ],
                    results: []
                }
            ]
        };

        // Add coverage violation if below threshold
        if (results && results.coverage < this.options.minCoverage) {
            contract.runs[0].results.push({
                ruleId: 'COVERAGE_THRESHOLD',
                level: 'error',
                message: {
                    text: `Code Coverage (${results.coverage}%) is below threshold (${this.options.minCoverage}%)`
                },
                properties: {
                    coverage: {
                        actual: results.coverage,
                        required: this.options.minCoverage
                    }
                }
            });
        }

        // Add test failures
        if (results?.tests?.failed > 0) {
            contract.runs[0].results.push({
                ruleId: 'TEST_FAILURE',
                level: 'error',
                message: {
                    text: `${results.tests.failed} test(s) failed out of ${results.tests.total}`
                },
                properties: {
                    tests: results.tests
                }
            });
        }

        // Add general error if present
        if (error) {
            contract.runs[0].results.push({
                ruleId: 'DEPLOYMENT_FAILURE',
                level: 'error',
                message: {
                    text: error.message
                },
                properties: {
                    errorCode: error.code || 'UNKNOWN',
                    errorType: 'deployment_failure'
                }
            });
        }

        return contract;
    }

    normalizeError(error) {
        return {
            code: error.code || 'UNKNOWN',
            message: error.message,
            type: 'deployment_failure'
        };
    }

    async writeContract(contract) {
        const contractPath = '.glassops/glassops-contract.sarif.json';
        await fs.mkdir('.glassops', { recursive: true });

        // Write atomically (temp file + rename)
        const tempPath = `${contractPath}.tmp`;
        await fs.writeFile(tempPath, JSON.stringify(contract, null, 2));
        await fs.rename(tempPath, contractPath);

        core.info(`Contract written to ${contractPath}`);
    }
}

module.exports = MyToolAdapter;
```

### Step 3: Create the GitHub Action Wrapper

```javascript
// src/index.js
const core = require('@actions/core');
const MyToolAdapter = require('./adapter');

async function run() {
    try {
        const adapter = new MyToolAdapter({
            sourceDir: core.getInput('source-dir', { required: true }),
            testLevel: core.getInput('test-level') || 'RunLocalTests',
            minCoverage: parseInt(core.getInput('min-coverage') || '75')
        });

        // Phase 1: Simulate
        core.info('Phase 1: Simulation (Check-Only)');
        const draftContract = await adapter.simulate();
        await adapter.writeContract(draftContract);

        // Check if execution was successful
        const invocation = draftContract.runs[0].invocations[0];
        if (!invocation.executionSuccessful) {
            core.setFailed('Simulation failed');
            return;
        }

        // Phase 2: Check governance - look for blocking violations
        const blockingResults = draftContract.runs[0].results.filter((r) => r.level === 'error');

        if (blockingResults.length > 0) {
            const messages = blockingResults.map((r) => r.message.text).join('; ');
            core.setFailed(`Governance violations: ${messages}`);
            return;
        }

        // Phase 3: Execute
        core.info('Phase 2: Execution (Quick Deploy)');
        const deploymentId = draftContract.runs[0].invocations[0].properties.glassops.deploymentId;
        const finalContract = await adapter.execute(deploymentId);
        await adapter.writeContract(finalContract);

        const finalInvocation = finalContract.runs[0].invocations[0];
        if (!finalInvocation.executionSuccessful) {
            core.setFailed('Deployment failed');
            return;
        }

        core.info('✅ Deployment successful');
        core.setOutput('deployment-id', finalInvocation.properties.glassops.deploymentId);

        // Extract coverage from results if present
        const coverageResult = finalContract.runs[0].results.find((r) => r.ruleId === 'COVERAGE_THRESHOLD');
        if (coverageResult?.properties?.coverage) {
            core.setOutput('coverage', coverageResult.properties.coverage.actual);
        }
    } catch (error) {
        core.setFailed(error.message);
    }
}

run();
```

### Step 4: Create action.yml

```yaml
name: 'GlassOps MyTool Adapter'
description: 'Execute governed deployments using MyTool'
author: 'Your Name'

inputs:
    source-dir:
        description: 'Source directory containing metadata'
        required: true
    test-level:
        description: 'Test level (NoTestRun, RunSpecifiedTests, RunLocalTests, RunAllTestsInOrg)'
        required: false
        default: 'RunLocalTests'
    min-coverage:
        description: 'Minimum code coverage required'
        required: false
        default: '75'

outputs:
    deployment-id:
        description: 'Salesforce deployment ID'
    coverage:
        description: 'Actual code coverage percentage'

runs:
    using: 'node20'
    main: 'dist/index.js'
```

---

## Testing Your Adapter

### Unit Tests

```javascript
// test/adapter.test.js
const MyToolAdapter = require('../src/adapter');

describe('MyToolAdapter', () => {
    it('should parse output correctly', () => {
        const adapter = new MyToolAdapter({ sourceDir: 'force-app' });
        const output = JSON.stringify({
            id: '0Af123',
            coverage: 85,
            testResults: { total: 10, passed: 10, failed: 0 }
        });

        const result = adapter.parseOutput(output);
        expect(result.coverage).toBe(85);
    });

    it('should build valid SARIF contract', () => {
        const adapter = new MyToolAdapter({
            sourceDir: 'force-app',
            minCoverage: 75
        });

        const contract = adapter.buildContract('draft', {
            deploymentId: '0Af123',
            coverage: 85,
            tests: { total: 10, passed: 10, failed: 0 },
            componentsDeployed: 5
        });

        // Validate SARIF structure
        expect(contract.version).toBe('2.1.0');
        expect(contract.$schema).toContain('sarif-2.1.0');
        expect(contract.runs).toHaveLength(1);
        expect(contract.runs[0].tool.driver.name).toBe('glassops-mytool-adapter');

        // Coverage above threshold should have no error results
        expect(contract.runs[0].results).toHaveLength(0);
    });

    it('should add coverage violation when below threshold', () => {
        const adapter = new MyToolAdapter({
            sourceDir: 'force-app',
            minCoverage: 75
        });

        const contract = adapter.buildContract('draft', {
            deploymentId: '0Af123',
            coverage: 65, // Below threshold
            tests: { total: 10, passed: 10, failed: 0 },
            componentsDeployed: 5
        });

        // Should have coverage violation result
        expect(contract.runs[0].results).toHaveLength(1);
        expect(contract.runs[0].results[0].ruleId).toBe('COVERAGE_THRESHOLD');
        expect(contract.runs[0].results[0].level).toBe('error');
    });
});
```

### Integration Test

```bash
#!/bin/bash
# test/integration-test.sh

set -e

echo "Running integration test..."

# Set up test environment
export GITHUB_ACTOR="test-user"
export GITHUB_REPOSITORY="test/repo"
export GITHUB_REF="refs/heads/main"
export GITHUB_SHA="abc123"
export GITHUB_RUN_ID="123"

# Run adapter
node src/index.js <<EOF
source-dir=force-app
test-level=RunLocalTests
min-coverage=75
EOF

# Validate contract exists
if [ ! -f ".glassops/glassops-contract.sarif.json" ]; then
  echo "❌ Contract not generated"
  exit 1
fi

# Validate SARIF schema
if ! jq -e '.version == "2.1.0"' .glassops/glassops-contract.sarif.json; then
  echo "❌ Invalid SARIF version"
  exit 1
fi

if ! jq -e '.$schema | contains("sarif-2.1.0")' .glassops/glassops-contract.sarif.json; then
  echo "❌ Invalid SARIF schema reference"
  exit 1
fi

echo "✅ Integration test passed"
```

---

## Adapter Development Checklist

Use this checklist to ensure your adapter is production-ready:

### Core Functionality

- [ ] Implements `simulate()` function
- [ ] Implements `execute()` function
- [ ] Implements `normalizeError()` function
- [ ] Emits valid SARIF 2.1.0 JSON (validate against schema)
- [ ] Writes contract to `.glassops/glassops-contract.sarif.json`
- [ ] Writes contract atomically (temp file + rename)

### Error Handling

- [ ] Handles tool crashes gracefully
- [ ] Handles timeout scenarios
- [ ] Handles network failures
- [ ] Handles invalid input
- [ ] Returns contract even on failure

### Contract Accuracy

- [ ] Governance findings mapped to `results[]` array
- [ ] Deployment metadata in `invocations[].properties.glassops`
- [ ] Tool information in `tool.driver` fields
- [ ] Error-level violations use `level: "error"`
- [ ] File locations use relative paths from repo root
- [ ] Conforms to [Adapter Interface](../../glassspec/README.md)

### Testing

- [ ] Unit tests for parsing logic
- [ ] Unit tests for contract building
- [ ] Integration test with real tool
- [ ] Tests failure scenarios
- [ ] Tests edge cases (zero coverage, no tests, etc.)

### Documentation

- [ ] README with usage examples
- [ ] Error code reference
- [ ] Troubleshooting guide
- [ ] Comparison with other adapters
- [ ] Migration guide (if replacing existing tool)

### Distribution

- [ ] Published to GitHub
- [ ] Tagged release with version
- [ ] action.yml correctly configured
- [ ] Dependencies properly declared
- [ ] License file included (Apache 2.0)

---

## Common Pitfalls

### 1. Partial Contract on Failure

**Problem:** Adapter crashes and leaves no contract.

**Solution:** Always emit a contract, even on failure:

```javascript
try {
    const results = await deploy();
    return buildContract('success', results);
} catch (error) {
    return buildContract('failed', null, error);
}
```

### 2. Non-Atomic Writes

**Problem:** Governance layer reads partial JSON.

**Solution:** Use temp file + rename:

```javascript
const tempPath = `${contractPath}.tmp`;
await fs.writeFile(tempPath, JSON.stringify(contract));
await fs.rename(tempPath, contractPath); // Atomic
```

### 3. Tool-Specific Error Codes

**Problem:** Different adapters use different error formats.

**Solution:** Normalize to standard codes:

```javascript
const ERROR_CODES = {
    COVERAGE_FAILURE: 'COVERAGE_BELOW_THRESHOLD',
    TEST_FAILURE: 'TEST_EXECUTION_FAILED',
    AUTH_FAILURE: 'AUTHENTICATION_FAILED'
};
```

### 4. Missing Audit Trail

**Problem:** Can't trace back to source commit.

**Solution:** Always include GitHub context:

```javascript
audit: {
  triggeredBy: process.env.GITHUB_ACTOR,
  repository: process.env.GITHUB_REPOSITORY,
  commit: process.env.GITHUB_SHA
}
```

---

## Example Adapters

### Minimal Adapter (Shell Script)

```bash
#!/bin/bash
# Simplest possible adapter using sf CLI

set -e

SOURCE_DIR=$1
TEST_LEVEL=$2
MIN_COVERAGE=${3:-75}

# Simulate
sf project deploy start \
  --source-dir "$SOURCE_DIR" \
  --test-level "$TEST_LEVEL" \
  --dry-run \
  --json > deploy-result.json

# Parse and build SARIF 2.1.0 contract
jq --arg minCov "$MIN_COVERAGE" '{
  "$schema": "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0-rtm.5.json",
  "version": "2.1.0",
  "runs": [{
    "tool": {
      "driver": {
        "name": "glassops-sf-adapter",
        "version": "1.0.0"
      }
    },
    "invocations": [{
      "executionSuccessful": (.result.success == true),
      "endTimeUtc": (now | strftime("%Y-%m-%dT%H:%M:%SZ")),
      "properties": {
        "glassops": {
          "deploymentId": .result.id,
          "mode": "validate"
        }
      }
    }],
    "results": (
      if (.result.coverage < ($minCov | tonumber)) then
        [{
          "ruleId": "COVERAGE_THRESHOLD",
          "level": "error",
          "message": {
            "text": "Coverage \(.result.coverage)% below threshold \($minCov)%"
          }
        }]
      else
        []
      end
    )
  }]
}' deploy-result.json > .glassops/glassops-contract.sarif.json
```

### Full-Featured Adapter (Node.js)

See the complete example in the Step-by-Step section above.

---

## Publishing Your Adapter

### 1. Create Repository

```bash
gh repo create glassops-mytool-adapter --public
git init
git add .
git commit -m "feat: initial adapter implementation"
git branch -M main
git remote add origin https://github.com/yourusername/glassops-mytool-adapter.git
git push -u origin main
```

### 2. Tag a Release

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### 3. Create Release Notes

```markdown
## v1.0.0

### Features

- Initial implementation of MyTool adapter
- Support for check-only and quick deploy
- Normalized deployment contract

### Usage

\`\`\`yaml

- uses: yourusername/glassops-mytool-adapter@v1
  with:
  source-dir: force-app
  \`\`\`
```

### 4. Submit to GlassOps Registry

Create a PR to `glassops-platform/adapters-registry`:

```yaml
# adapters/mytool.yml
name: glassops-mytool-adapter
author: Your Name
repository: yourusername/glassops-mytool-adapter
version: 1.0.0
tool: MyTool
description: Execute governed deployments using MyTool
```

---

## Getting Help

- 💬 [Adapter Development Discussions](https://github.com/glassops-platform/glassops/discussions/categories/adapter-development)
- 📖 [Contract Schema Reference](../../../docs/reference/platform-reference.md)
- 🐛 [Report Adapter Issues](https://github.com/glassops-platform/glassops/issues)
- 📧 Direct questions: [@rdbumstead](https://github.com/rdbumstead)

---

## Next Steps

1. Review existing adapters for patterns
2. Build a minimal proof-of-concept
3. Test with real Salesforce org
4. Share in Discussions for feedback
5. Submit to registry when ready

Thank you for contributing to the GlassOps ecosystem!
