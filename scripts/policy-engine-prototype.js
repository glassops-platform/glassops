/**
 * GlassOps Policy Engine - Prototype v1
 *
 * Demonstrates the "Simplest Possible Engine" concept:
 * 1. Takes PolicyRules
 * 2. Takes AuditEvents
 * 3. Matches and Returns Actions
 */

// --- 1. MOCK DATA ---

const mockPolicy = {
    apiVersion: 'governance.glassops.io/v1',
    kind: 'PolicyRule',
    metadata: {
        name: 'no-direct-profile-edits',
        severity: 'HIGH'
    },
    spec: {
        match: {
            environment: 'production',
            resource: { type: 'Profile' },
            action: ['update', 'delete']
        },
        exceptions: [
            {
                id: 'ci-bot-allow',
                condition: { actor: 'ci-bot@company.com' }
            }
        ],
        actions: [
            { type: 'block', message: 'Direct profile edits are forbidden.' },
            { type: 'alert', channel: '#sec-ops' }
        ]
    }
};

const mockEventViolation = {
    id: 'evt_001',
    environment: 'production',
    actor: 'john.doe@company.com', // Violation
    action: 'update',
    resource: { type: 'Profile', name: 'Admin' }
};

const mockEventException = {
    id: 'evt_002',
    environment: 'production',
    actor: 'ci-bot@company.com', // Exception
    action: 'update',
    resource: { type: 'Profile', name: 'Admin' }
};

// --- 2. ENGINE LOGIC ---

function evaluate(event, rules) {
    const results = [];

    for (const rule of rules) {
        console.log(`Evaluating Rule: ${rule.metadata.name} against Event: ${event.id}`);

        // A. Check Match Criteria
        const match = rule.spec.match;
        if (match.environment && match.environment !== event.environment) continue;
        if (match.resource.type && match.resource.type !== event.resource.type) continue;
        if (match.action && !match.action.includes(event.action)) continue;

        console.log(`  -> Match Criteria Met.`);

        // B. Check Exceptions
        let isExcepted = false;
        if (rule.spec.exceptions) {
            for (const exception of rule.spec.exceptions) {
                // Simple equality check for prototype
                if (exception.condition.actor === event.actor) {
                    console.log(`  -> Exception Met: ${exception.id}`);
                    isExcepted = true;
                    break;
                }
            }
        }

        if (isExcepted) {
            results.push({ rule: rule.metadata.name, result: 'ALLOWED_BY_EXCEPTION' });
            continue;
        }

        // C. Trigger Actions
        console.log(`  -> VIOLATION! Triggering Actions:`);
        results.push({
            rule: rule.metadata.name,
            result: 'VIOLATION',
            severity: rule.metadata.severity,
            actions: rule.spec.actions
        });
    }

    return results;
}

// --- 3. RUNTIME ---

console.log('--- TEST 1: User Edit (Should Violate) ---');
const result1 = evaluate(mockEventViolation, [mockPolicy]);
console.log(JSON.stringify(result1, null, 2));

console.log('\n--- TEST 2: CI Bot Edit (Should Apply Exception) ---');
const result2 = evaluate(mockEventException, [mockPolicy]);
console.log(JSON.stringify(result2, null, 2));
