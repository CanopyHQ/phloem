# Phloem Grafts - Shareable Memory Bundles

This directory contains seed grafts for the Phloem ecosystem.

## Available Grafts

### phloem-engineering-standards.graft

Core patterns for Phloem development. Includes:
- Architecture decisions
- Best practices
- Engineering standards
- Development patterns

**Usage:**
```bash
phloem graft import grafts/phloem-engineering-standards.graft
```

## Creating New Grafts

Export memories as grafts:
```bash
phloem graft export --tags "architecture,patterns" --output my-graft.graft --name "My Graft" --desc "Description"
```

Share grafts:
- Upload to GitHub releases
- Share via Twitter/Slack
- Include in documentation

## Graft Format

Grafts use a binary format with:
- Magic bytes: `PHLO`
- Version: 1
- Gzip-compressed JSON payload

Inspect without importing:
```bash
phloem graft inspect my-graft.graft
```
