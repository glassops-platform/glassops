/* eslint-disable */
const { execSync } = require('child_process');
const fs = require('fs');

console.log('--- 1. Generating Delta ---');
try {
    // Run the existing delta command defined in package.json
    execSync('npm run delta', { stdio: 'inherit' });
} catch (error) {
    console.error('Delta generation failed.');
    process.exit(1);
}

// Check if package.xml exists
const packageXmlPath = 'changed/package/package.xml';
if (!fs.existsSync(packageXmlPath)) {
    console.log('⚠️  No package.xml found. Skipping validation.');
    process.exit(0);
}

// Check if package.xml actually has metadata changes (looks for <types> tag)
const content = fs.readFileSync(packageXmlPath, 'utf8');
if (!content.includes('<types>')) {
    console.log('✅  No metadata changes detected. Skipping validation.');
    process.exit(0);
}

console.log('--- 2. Validating Changes (Dry Run) ---');
try {
    execSync(
        'sf project deploy start --manifest changed/package/package.xml --target-org portfolio --dry-run --test-level RunLocalTests',
        { stdio: 'inherit' }
    );
} catch (error) {
    console.error('❌  Validation Failed.');
    process.exit(1);
}
