# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in toolfoundation, please report it responsibly.

### How to Report

1. **Do NOT open a public GitHub issue** for security vulnerabilities.

2. **Email the maintainer** with details of the vulnerability:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Any suggested fixes (optional)

3. **Allow time for response** - We aim to respond within 48 hours and provide a fix timeline within 7 days.

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your report within 48 hours.
- **Assessment**: We will assess the vulnerability and determine its severity.
- **Fix Timeline**: For confirmed vulnerabilities, we will provide an estimated fix timeline.
- **Disclosure**: We will coordinate with you on public disclosure after a fix is available.
- **Credit**: With your permission, we will credit you in the security advisory.

## Security Measures

This package implements several security measures:

### Schema Validation
- External `$ref` resolution is **disabled** to prevent network access during validation
- JSON Schema validation is deterministic and does not perform I/O

### Input Validation
- Tool names are validated against strict character allowlists
- Tag normalization prevents injection of special characters
- Schema validation rejects malformed input

### Dependencies
- Dependencies are regularly scanned with `govulncheck`
- Security scanning runs in CI via `gosec`

## Scope

This security policy applies to:
- The `model` package (tool definitions, validation)
- The `adapter` package (format conversion)
- The `version` package (version parsing, constraints)

## Out of Scope

The following are out of scope for this security policy:
- Vulnerabilities in dependencies (report to the dependency maintainer)
- Issues in example code that is not meant for production use
- Theoretical attacks without demonstrated impact
