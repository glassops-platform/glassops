# GlassOps Vision: Universal Governance Protocol

> **"GlassOps governs outcomes, not implementations."**

**Version:** 2.0
**Status:** Strategic Authority

---

## 1. The Core Question

Most DevOps systems ask only one question:

> "Did the deployment succeed?"

GlassOps asks the harder question:

> "Should this code exist in production?"

**Execution success does not imply architectural health.**

## 2. The Problem: Governance Fragmentation

Organizations use dozens of tools (SonarQube, Terraform, Salesforce), but they don't talk to each other.

- **No Correlation:** A code change causes an alert, but you can't prove it.
- **Vendor Lock-in:** Switching tools means rewriting compliance rules.
- **Black Boxes:** Deployment logic is hidden in proprietary platforms.

## 3. The Solution: Governance as a Protocol

GlassOps is not a product; it is a **Universal Governance Protocol**.

It decouples **Policy** (The Brain) from **Execution** (The Muscle).

- **Policy** is defined once (Github, Config, CMDT).
- **Execution** is delegated to pluggable Adapters.
- **Audit** is normalized into a standard Deployment Contract.

### The Layered Contract Model

We use the right standard for each signal:

| Layer             | Standard          | Purpose                             |
| ----------------- | ----------------- | ----------------------------------- |
| **1. Governance** | **SARIF**         | Policy violations (Static Analysis) |
| **2. Telemetry**  | **OpenTelemetry** | Runtime metrics (Time series)       |
| **3. Transport**  | **CloudEvents**   | Lifecycle state (Event bus)         |

## 4. The Origin Story

> [!NOTE]
> Born from the pain of "Successful Deployments" that destroyed architecture.

We hired brilliant consultants. They deployed successfully every week. Two years later, we had 5 triggers per object and SOQL in loops. We realized that **we couldn't automate safely because we couldn't enforce invariants.**

GlassOps exists to enforce **Architectural Invariants** (e.g., "One Trigger Per Object") _before_ execution is permitted.

## 5. Strategy: The 8/10 Vision

We are mostly focused on the **8/10 Vision** (Months 0-18):

- **Scope:** Salesforce-first.
- **Authority:** GitHub Actions.
- **Goal:** Prove the protocol works in production.

We effectively ignore the "10/10 Vision" (K8s Controller, Multi-Cloud) until the 8/10 goal is met. This constraint is deliberate.

---

**Next Steps:**

- See [Architecture](architecture/overview.md) for how it works.
