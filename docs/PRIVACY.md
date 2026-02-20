# Phloem Privacy Policy

**Last Updated: February 2026**

## Our Commitment

Your data belongs to you. Phloem is designed so we *can't* access your data — everything runs locally on your machine with zero network connections.

## How Phloem Works

- All data stored locally in `~/.phloem/memories.db` (SQLite)
- MCP server communicates with your IDE via stdio (local pipes, not network)
- No accounts, no registration, no personal information collected

## What We Don't Collect

- No memory content ever leaves your machine
- No telemetry, analytics, or crash reporting
- No network connections of any kind
- No usage tracking or install pings

## Verify It Yourself

This is not just a promise — you can verify it:

```bash
# Inspect your data
phloem audit

# Monitor network activity while using Phloem
sudo lsof -i -P | grep phloem    # Should show nothing

# Automated verification
make verify-privacy
```

## Data Storage

- **Location:** `~/.phloem/memories.db`
- **Format:** SQLite with sqlite-vec extension
- **Contents:** Your memories, embeddings, citations, causal graph edges

## Data Export

```bash
phloem export json my-memories.json
phloem export markdown my-memories.md
```

Your data is yours. Take it anywhere.

## Data Deletion

```bash
rm -rf ~/.phloem
```

No account to delete, no support ticket, no waiting period.

## Source Code

Audit the code yourself: https://github.com/CanopyHQ/phloem

## Third Parties

We don't share data because we don't have your data.

## Changes

Updates announced in GitHub releases. We will never reduce privacy retroactively.

## Contact

Questions about privacy? Email: privacy@canopyhq.io

---

**TL;DR**: Everything stays on your machine. We collect nothing. Verify it yourself with `phloem audit`.
