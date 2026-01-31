# Enabling Observability with Grafana Cloud

GlassOps Runtime supports OpenTelemetry (OTLP) natively, allowing you to stream governance traces directly to Grafana Cloud.

## Prerequisites

1.  **Grafana Cloud Account**: [Sign up for free](https://grafana.com/auth/sign-up/create-user) if you don't have one.
2.  **GlassOps Runtime v1**: Make sure you are using `v1` or later.

## Setup Steps

### 1. Get Credentials from Grafana

1.  Log in to your Grafana Cloud Portal.
2.  Navigate to **OpenTelemetry** -> **Configure**.
3.  Note down the following values:
    *   **Endpoint URL**: e.g., `https://otlp-gateway-prod-us-central-0.grafana.net/otlp`
    *   **Instance ID / User**: e.g., `123456`
    *   **API Token**: Generate a token with metrics/traces write permissions.

### 2. Configure GitHub Secrets

Store your API token securely in your GitHub repository secrets:

1.  Go to your repo **Settings** -> **Secrets and variables** -> **Actions**.
2.  Create a new secret:
    *   **Name**: `GRAFANA_OTEL_TOKEN`
    *   **Value**: `<your-api-token>`

### 3. Update Workflow

Add the telemetry configuration to your GlassOps workflow:

```yaml
- name: GlassOps Governance Check
  uses: nobleforge/glassops/runtime@v1
  with:
    # ... other inputs ...
    otel_endpoint: 'https://otlp-gateway-prod-us-central-0.grafana.net/otlp'
    otel_headers: 'Authorization=Basic <base64-encoded-auth>' 
    # OR simpler approach using raw inputs if your OTLP endpoint accepts token directly in header:
    # otel_headers: 'Authorization=Bearer ${{ secrets.GRAFANA_OTEL_TOKEN }}'
```

#### NOTE: Constructing the Authorization Header for Grafana

Grafana Cloud uses Basic Auth where the username is your **Instance ID** and the password is your **API Token**.

You can construct the header value in a separate step or pass it directly if you pre-encode it.

**Recommended Pattern:**

```yaml
steps:
  - name: Setup Telemetry Auth
    id: auth
    run: |
      # Create Base64 encoded credentials (User:Token)
      CRED=$(echo -n "${{ vars.GRAFANA_USER }}:${{ secrets.GRAFANA_TOKEN }}" | base64)
      echo "header=Authorization=Basic $CRED" >> $GITHUB_OUTPUT

  - name: GlassOps Governance Check
    uses: nobleforge/glassops/runtime@v1
    with:
      # ... other inputs ...
      otel_endpoint: 'https://otlp-gateway-prod-us-central-0.grafana.net/otlp'
      otel_headers: ${{ steps.auth.outputs.header }}
```

## Verifying Traces

1.  Run your GitHub Action.
2.  Go to Grafana -> **Explore**.
3.  Select your **Tempo** (Traces) datasource.
4.  Query for `service_name="glassops-runtime"`.
5.  You should see traces for:
    *   `glassops.policy`
    *   `glassops.bootstrap`
    *   `glassops.identity`
