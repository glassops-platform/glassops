# ADR 010: Identity Contract

**Status:** Accepted  
**Date:** 2026-01-24  
**Deciders:** Ryan Bumstead

---

## Context

GlassOps operates across multiple execution substrates (GitHub Actions, Salesforce, Kubernetes), each with their own identity systems. Currently, there is **no unified identity contract** to represent "who" is performing governance actions.

### The Identity Gap

**Current State:**

- **GitHub Actions:** `GITHUB_ACTOR` environment variable
- **Salesforce:** JWT-OAuth with `username` claim
- **Kubernetes:** ServiceAccount with `subjects[].name`
- **Policy Overrides:** Email address in override request

**Problem:** These are incompatible identity representations. The governance layer cannot answer:

- "Who triggered this deployment across all substrates?"
- "Does Alice have permission to override policy in Salesforce AND Kubernetes?"
- "Can we trace this audit trail back to a human?"

### Real-World Scenario

```
Deployment abc123:
  - Triggered by: github:alice (GitHub Actions)
  - Authenticated to: salesforce:alice@company.com (JWT)
  - Approved by: alice@company.com (override request)
  - Service Account: k8s:system:serviceaccount:glassops:deployer
```

**Question:** Are all four the same person? Current system cannot answer.

---

## Decision

**Create a Universal Identity Contract that works across all execution substrates.**

### Core Principle

> **"Governance decisions are meaningless without knowing who made them."**

The Identity Contract is **embedded in every SARIF contract** via `invocations[].properties.glassops.identity`.

---

## Identity Contract Schema

### Minimal Contract

```json
{
  "identity": {
    "subject": "alice@company.com",
    "provider": "github|salesforce|kubernetes|manual",
    "provider_id": "github:alice|sf:005000000...|k8s:sa:deployer|email:alice@...",
    "verified": true|false,
    "timestamp": "2026-01-24T10:00:00Z"
  }
}
```

### Full Contract (with Authorization)

```json
{
    "identity": {
        "subject": "alice@company.com",
        "displayName": "Alice Developer",
        "provider": "github",
        "provider_id": "github:alice",
        "verified": true,
        "timestamp": "2026-01-24T10:00:00Z",

        "roles": ["developer", "deployer"],
        "teams": ["platform-team", "security-team"],

        "authorization": {
            "can_deploy": true,
            "can_override": false,
            "can_approve_overrides": true,
            "max_coverage_reduction": 0
        },

        "context": {
            "ip_address": "192.168.1.100",
            "user_agent": "GitHub-Actions/2.0",
            "session_id": "sess_abc123",
            "mfa_verified": true
        }
    }
}
```

---

## Implementation Strategy

### Phase 1: Identity Detection (Current)

**Goal:** Extract identity from execution context

**GitHub Actions:**

```javascript
function detectIdentity() {
    return {
        subject: process.env.GITHUB_ACTOR_EMAIL || `${process.env.GITHUB_ACTOR}@users.noreply.github.com`,
        provider: 'github',
        provider_id: `github:${process.env.GITHUB_ACTOR}`,
        verified: !!process.env.GITHUB_ACTOR_EMAIL,
        timestamp: new Date().toISOString()
    };
}
```

**Salesforce JWT:**

```javascript
function detectIdentity(jwtPayload) {
    return {
        subject: jwtPayload.sub, // username
        provider: 'salesforce',
        provider_id: `sf:${jwtPayload.user_id}`,
        verified: true, // JWT signature validated
        timestamp: new Date(jwtPayload.iat * 1000).toISOString()
    };
}
```

**Kubernetes ServiceAccount:**

```go
func detectIdentity(ctx context.Context) *Identity {
    sa := ctx.Value("serviceaccount")
    return &Identity{
        Subject:    sa.Annotations["glassops.io/human-user"],
        Provider:   "kubernetes",
        ProviderID: fmt.Sprintf("k8s:sa:%s:%s", sa.Namespace, sa.Name),
        Verified:   true,
        Timestamp:  time.Now(),
    }
}
```

---

### Phase 2: Identity Resolution (Bridge)

**Goal:** Map provider-specific IDs to canonical subjects

**Identity Resolution Service:**

```yaml
apiVersion: identity.glassops.io/v1
kind: IdentityMapping
metadata:
    name: alice-identity
spec:
    canonical_subject: alice@company.com

    mappings:
        - provider: github
          provider_id: github:alice
          verified_at: '2026-01-20T00:00:00Z'

        - provider: salesforce
          provider_id: sf:005000000abcXYZ
          verified_at: '2026-01-20T00:00:00Z'

        - provider: kubernetes
          provider_id: k8s:sa:glassops:alice-deployer
          verified_at: '2026-01-20T00:00:00Z'
```

**Resolution Logic:**

```javascript
async function resolveIdentity(detectedIdentity) {
    // Look up canonical subject from provider_id
    const mapping = await identityStore.findMapping(detectedIdentity.provider_id);

    if (mapping) {
        return {
            ...detectedIdentity,
            subject: mapping.canonical_subject,
            verified: true
        };
    }

    // Fallback: Use detected identity as-is
    return detectedIdentity;
}
```

---

### Phase 3: Identity Authorization (Future)

**Goal:** Enforce permissions based on identity

**Authorization Policy:**

```yaml
apiVersion: policy.glassops.io/v1
kind: AuthorizationPolicy
metadata:
    name: deployment-permissions
spec:
    rules:
        - subjects: ['alice@company.com']
          roles: ['deployer']
          permissions:
              deploy: ['production', 'staging']
              override: false

        - subjects: ['bob@company.com']
          roles: ['platform-lead']
          permissions:
              deploy: ['production', 'staging', 'dev']
              override: true
              approve_overrides: true
```

---

## SARIF Integration

### Where Identity Goes

**Location:** `invocations[].properties.glassops.identity`

**Example SARIF Contract:**

```json
{
    "$schema": "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0-rtm.5.json",
    "version": "2.1.0",
    "runs": [
        {
            "tool": {
                "driver": {
                    "name": "glassops-native-adapter",
                    "version": "1.0.0"
                }
            },
            "invocations": [
                {
                    "executionSuccessful": true,
                    "endTimeUtc": "2026-01-24T10:00:00Z",
                    "properties": {
                        "glassops": {
                            "deploymentId": "0Af...",
                            "identity": {
                                "subject": "alice@company.com",
                                "provider": "github",
                                "provider_id": "github:alice",
                                "verified": true,
                                "timestamp": "2026-01-24T09:55:00Z",
                                "roles": ["developer", "deployer"],
                                "authorization": {
                                    "can_deploy": true,
                                    "can_override": false
                                }
                            }
                        }
                    }
                }
            ],
            "results": []
        }
    ]
}
```

---

## Audit Trail Benefits

### Before Identity Contract

```
Deployment abc123: Who did this?
  └─ GITHUB_ACTOR=alice (Could be anyone named alice)
```

### After Identity Contract

```
Deployment abc123: alice@company.com (verified)
  ├─ Provider: GitHub (github:alice)
  ├─ Roles: [developer, deployer]
  ├─ Authorization: can_deploy=true, can_override=false
  ├─ MFA: verified
  └─ IP: 192.168.1.100
```

**Audit Questions Now Answerable:**

- ✅ "Who deployed this?" → alice@company.com
- ✅ "Were they authorized?" → Yes, can_deploy=true
- ✅ "Was identity verified?" → Yes, MFA verified
- ✅ "Can we correlate across substrates?" → Yes, canonical subject

---

## Cross-Substrate Correlation

### Example: Override Request Flow

**1. Deployment Attempt (GitHub Actions)**

```json
{
    "identity": {
        "subject": "alice@company.com",
        "provider": "github",
        "provider_id": "github:alice",
        "authorization": {
            "can_deploy": true,
            "can_override": false // ← Cannot override
        }
    }
}
```

**2. Override Request (CLI/API)**

```bash
glassops override request \
  --repository acme/salesforce \
  --reason "Production outage" \
  --requester alice@company.com  # ← Same canonical subject
```

**3. Override Approval (Kubernetes Operator)**

```json
{
    "identity": {
        "subject": "bob@company.com",
        "provider": "kubernetes",
        "provider_id": "k8s:sa:glassops:platform-lead",
        "authorization": {
            "can_approve_overrides": true // ← Can approve
        }
    }
}
```

**4. Deployment with Override (GitHub Actions)**

```json
{
    "identity": {
        "subject": "alice@company.com", // ← Same requester
        "provider": "github",
        "authorization": {
            "can_deploy": true,
            "override_approved_by": "bob@company.com", // ← Approver tracked
            "override_id": "override-1234"
        }
    }
}
```

**Result:** Complete audit trail across all substrates.

---

## Consequences

### Positive

- ✅ **Unified Identity** - Single canonical subject across all substrates
- ✅ **Audit Trail** - Can trace actions back to humans
- ✅ **Authorization** - Enforce permissions consistently
- ✅ **Compliance** - SOC 2 / audit-ready identity records
- ✅ **Cross-Substrate** - Correlate actions across GitHub, Salesforce, K8s

### Negative

- ⚠️ **Identity Mapping Maintenance** - Must keep mappings updated
- ⚠️ **Privacy Concerns** - Storing email addresses in SARIF contracts
- ⚠️ **Resolution Overhead** - Extra lookup for canonical subject
- ⚠️ **Migration** - Existing contracts lack identity field

### Neutral

- 📝 Identity is optional in SARIF (can be missing)
- 📝 Verified=false is valid (unverified identity still captured)
- 📝 Provider-specific fields allowed in `context`

---

## Privacy & Security Considerations

### Personal Data in SARIF

**Problem:** Email addresses are PII (Personally Identifiable Information)

**Mitigations:**

1. **Hashing:** Store `subject_hash` instead of email
2. **Pseudonymization:** Use stable pseudonym (e.g., `user_123`)
3. **Retention:** Auto-delete identity after 90 days
4. **Access Control:** Restrict SARIF contract access

**Recommendation:** Start with email, add hashing in Phase 2.

---

## Alternatives Considered

### Alternative 1: No Identity Contract

**Rejected:** Audit trail is meaningless without knowing "who"

### Alternative 2: Provider-Specific Identity Only

**Rejected:** Cannot correlate across substrates (GitHub vs Salesforce vs K8s)

### Alternative 3: External Identity Provider (OAuth2)

**Rejected:** Too complex for 8/10 vision, consider for 10/10

### Alternative 4: Sign SARIF Contracts with Private Keys

**Considered:** Strong non-repudiation, but high complexity. Future enhancement.

---

## Migration Path

### Existing Contracts Without Identity

**Problem:** Contracts emitted before ADR-010 lack identity field

**Solution:** Backfill from audit logs

```javascript
async function backfillIdentity(contract) {
    if (contract.runs[0].invocations[0].properties?.glassops?.identity) {
        return contract; // Already has identity
    }

    // Extract from audit context
    const auditData = contract.runs[0].invocations[0].properties?.glassops;

    const identity = {
        subject: auditData?.triggeredBy || 'unknown',
        provider: 'github', // Assumption for pre-ADR-010 contracts
        provider_id: `github:${auditData?.triggeredBy}`,
        verified: false, // Cannot verify retroactively
        timestamp: contract.runs[0].invocations[0].endTimeUtc,
        _backfilled: true
    };

    contract.runs[0].invocations[0].properties.glassops.identity = identity;
    return contract;
}
```

---

## Related ADRs

- [ADR-007: Protocol Supremacy Enforcement](007-protocol-supremacy-enforcement.md)
- [ADR-009: The 8/10 vs 10/10 Bridge Strategy](009-8-10-vs-10-10-bridge-strategy.md) ← Identified identity gap
- [GlassSpec ADR-001: Layered Contract Model](../../packages/glassspec/adr/001-layered-contract-model.md)

---

## References

- [OASIS SARIF 2.1.0](https://docs.oasis-open.org/sarif/sarif/v2.1.0/)
- [OAuth 2.0 RFC 6749](https://datatracker.ietf.org/doc/html/rfc6749)
- [NIST SP 800-63B: Digital Identity Guidelines](https://pages.nist.gov/800-63-3/sp800-63b.html)

---

**Author:** Ryan Bumstead  
**Implemented:** 2026-01-24  
**Status:** Active - Identity field required in all new contracts

## Alternatives

- None considered.
