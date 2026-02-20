# Phloem Grafts - Shareable Memory Bundles

This directory contains seed grafts for the Canopy ecosystem.

## Available Grafts

### canopy-engineering-standards.graft

Core patterns for Crown, Phloem, and Cambium development. Includes:
- Architecture decisions
- Best practices
- Engineering standards
- Development patterns

**Usage:**
```bash
canopy graft import grafts/canopy-engineering-standards.graft
```

## Creating New Grafts

Export memories as grafts:
```bash
canopy graft export --tags "architecture,patterns" --output my-graft.graft --name "My Graft" --desc "Description"
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
canopy graft inspect my-graft.graft
```
