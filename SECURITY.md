# Security Policy

## Supported Versions

| Version | Supported          |
|---------|--------------------|
| 0.1.x   | Yes                |

## Reporting a Vulnerability

If you discover a security vulnerability in Phloem, please report it responsibly.

**Do not open a public issue.**

Email: **security@canopyhq.io**

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

We will acknowledge your report within 48 hours and aim to release a fix within 7 days for critical issues.

## Scope

Phloem runs entirely locally with zero network connections. The primary security concerns are:

- **Local file permissions** on `~/.phloem/memories.db`
- **SQLite injection** via MCP tool inputs
- **Path traversal** in citation file paths
- **Denial of service** via malformed MCP requests

## Verification

You can verify Phloem's security posture at any time:

```bash
phloem audit                          # Data inventory, permissions, schema
sudo lsof -i -P | grep phloem        # Should show nothing (no network)
make verify-privacy                   # Automated network isolation test
```
