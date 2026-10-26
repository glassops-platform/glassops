# Security Policy

## Supported Versions

No version has been released. Only the current `main` branch is maintained.

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please report it responsibly.

### How to Report

1. **Do NOT** open a public GitHub issue for security vulnerabilities
2. Contact the repository maintainer privately through the contact details on the maintainer's GitHub profile
3. Include a detailed description of the vulnerability
4. Provide steps to reproduce if possible

### What to Expect

- **Acknowledgment**: We will acknowledge receipt within 48 hours
- **Assessment**: We will assess the vulnerability and determine its severity
- **Resolution**: Critical vulnerabilities will be addressed as priority
- **Disclosure**: We will coordinate disclosure timing with you

### Scope

This security policy applies to:

- The GitHub Action in `packages/runtime`
- Related configuration and contract files
- Example workflow templates in `examples/`

### Out of Scope

- Vulnerabilities in third-party dependencies (report these upstream)
- Issues in user-provided configuration files
- Theoretical vulnerabilities without proof of concept

## Security Best Practices

When using GlassOps Runtime:

1. **Secrets Management**: Always use GitHub Secrets for sensitive values (`jwt_key`, `client_id`)
2. **Least Privilege**: Use dedicated External Client Apps with minimal permissions
3. **Audit Logs**: Store deployment contracts for compliance audit trails
4. **Version Pinning**: Pin to a full commit SHA, since no release tags exist yet

## Acknowledgments

We appreciate researchers who practice responsible disclosure.
