# GlassOps Vision

> **"GlassOps governs outcomes, not implementations."**

**Status:** Direction. Most of what follows is not built yet. The only code in this repository is the Go runtime in `packages/runtime`.

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

GlassOps is being rebuilt as a CI-agnostic command line tool for Salesforce delivery. The aim is to decouple **Policy** from **Execution**.

- **Policy** is defined once, in configuration kept with the code.
- **Execution** is delegated to the tools you already use.
- **Audit** is normalized into a standard record of each run.

### The Layered Contract Model

The intended standard for each signal:

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

The focus is the **8/10 Vision**:

- **Scope:** Salesforce first. Other providers may follow later.
- **Runs anywhere CI runs:** GitHub Actions today, other CI systems later.
- **Goal:** Prove the approach works in production.

The "10/10 Vision" (Kubernetes controller, multi-cloud) is set aside until the 8/10 goal is met. This constraint is deliberate.

---

**Next Steps:**

- See [Architecture](architecture/overview.md) for how it works.
