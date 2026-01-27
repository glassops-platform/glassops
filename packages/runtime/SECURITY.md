# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x     | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please report it responsibly.

### How to Report

1. **Do NOT** open a public GitHub issue for security vulnerabilities
2. Email security concerns to the repository maintainer
3. Include a detailed description of the vulnerability
4. Provide steps to reproduce if possible

### What to Expect

- **Acknowledgment**: We will acknowledge receipt within 48 hours
- **Assessment**: We will assess the vulnerability and determine its severity
- **Resolution**: Critical vulnerabilities will be addressed as priority
- **Disclosure**: We will coordinate disclosure timing with you

### Scope

This security policy applies to:

- The `glassops-runtime` GitHub Action
- Related configuration schemas and contracts
- CI/CD workflow templates

### Out of Scope

- Vulnerabilities in third-party dependencies (report these upstream)
- Issues in user-provided configuration files
- Theoretical vulnerabilities without proof of concept

## Security Best Practices

When using GlassOps Runtime:

1. **Secrets Management**: Always use GitHub Secrets for sensitive values (`jwt_key`, `client_id`)
2. **Least Privilege**: Use dedicated External Client Apps with minimal permissions
3. **Audit Logs**: Store deployment contracts for compliance audit trails
4. **Version Pinning**: Pin to specific versions in production workflows

## Acknowledgments

We appreciate researchers who practice responsible disclosure.
