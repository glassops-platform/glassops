/**
 * GlassOps Policy Resolution Script
 *
 * Implements "Phase 1" of the Governance Protocol:
 * 1. Reads GitHub Environment Variables (Floor)
 * 2. Reads devops-config.json (Team Standard)
 * 3. (Mock) Reads Salesforce CMDT (Compliance Requirement) - logic included but mocked for now
 * 4. Merges to find "High Water Mark" (Additive Governance)
 * 5. Outputs .glassops/policy.json
 */

const fs = require('fs');
const path = require('path');

// --- Configuration ---
const CONFIG_PATH = path.join(process.cwd(), 'config', 'devops-config.json');
const OUTPUT_DIR = path.join(process.cwd(), '.glassops');
const OUTPUT_FILE = path.join(OUTPUT_DIR, 'policy.json');

// --- Helpers ---
function ensureDir(dir) {
    if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
    }
}

function readConfig() {
    if (fs.existsSync(CONFIG_PATH)) {
        try {
            return JSON.parse(fs.readFileSync(CONFIG_PATH, 'utf8'));
        } catch (e) {
            console.warn('⚠️  Could not parse devops-config.json, using defaults.');
        }
    } else {
        console.warn('⚠️  devops-config.json not found, using defaults.');
    }
    return {};
}

// --- Main Logic ---
async function main() {
    console.log('🔍 Resolving Effective Governance Policy...');

    // 1. GitHub Floor (Absolute Minimum)
    const githubFloor = parseInt(process.env.GLASSOPS_MIN_COVERAGE || '75', 10);
    console.log(`   - GitHub Floor: ${githubFloor}%`);

    // 2. Repo Config (Team Standard)
    const config = readConfig();
    const configStandard = config.governance?.minCoverage || 75;
    const engine = config.execution?.engine || 'native';
    console.log(`   - Repo Config: ${configStandard}% (Engine: ${engine})`);

    // 3. Salesforce CMDT (Compliance Additive)
    // TODO: In real implementation, this would query Salesforce via SFDX
    const sfRequired = 0;
    console.log(`   - Salesforce CMDT: ${sfRequired}% (Mocked)`);

    // 4. Validation: Config cannot lower GitHub floor
    if (configStandard < githubFloor) {
        console.error(
            `❌ Policy Violation: Repo config (${configStandard}%) cannot be lower than GitHub floor (${githubFloor}%).`
        );
        process.exit(1);
    }

    // 5. Additive Merge (High Water Mark)
    const effectiveCoverage = Math.max(githubFloor, configStandard, sfRequired);
    console.log(`✅ Effective Policy: Minimum Coverage = ${effectiveCoverage}%`);

    // 6. Generate Policy Object
    const policy = {
        required_coverage: effectiveCoverage,
        require_tests: config.governance?.requireTests !== false,
        engine: engine,
        sources: {
            github_floor: githubFloor,
            config_standard: configStandard,
            salesforce_cmdt: sfRequired
        },
        timestamp: new Date().toISOString()
    };

    // 7. Write Output
    ensureDir(OUTPUT_DIR);
    fs.writeFileSync(OUTPUT_FILE, JSON.stringify(policy, null, 2));
    console.log(`📄 Policy written to ${OUTPUT_FILE}`);

    // 8. Output for GitHub Actions
    if (process.env.GITHUB_OUTPUT) {
        fs.appendFileSync(process.env.GITHUB_OUTPUT, `required_coverage=${effectiveCoverage}\n`);
        fs.appendFileSync(process.env.GITHUB_OUTPUT, `engine=${engine}\n`);
    }
}

main().catch((err) => {
    console.error('❌ Fatal Error:', err);
    process.exit(1);
});
