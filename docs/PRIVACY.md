# Phloem Privacy Policy

**Last Updated: January 13, 2026**

## Our Commitment

Phloem is built on a simple principle: **your data belongs to you**. We don't want your data, we don't collect your data, and we've designed Phloem so we *can't* access your data.

## What Phloem Does

Phloem captures and stores your AI conversations from tools like ChatGPT, Claude, and Gemini. This data is stored **locally on your device** in a SQLite database at `~/.phloem/memories.db`.

## Data Collection

### What We Collect

- **No memory content**: We never send your conversation content, memories, prompts, or queries.
- **Minimal diagnostics**: A one-time install ping may be sent with:
  - Version
  - OS
  - CPU architecture
  - Timestamp
- **Controls**: You can disable telemetry via `canopy telemetry disable`.
- **No tracking of who you are**: We do not collect your name, email, or account info.
- **No accounts**: The free version requires no account, no email, no personal information.
- **No network calls**: The core Phloem application makes zero network requests unless you explicitly enable cloud sync.

### What Stays on Your Device

- All captured conversations
- All memories you create
- All search indexes
- All application settings

## Cloud Sync (Optional)

If you choose to enable cloud sync (a paid feature):

- Your memories are encrypted before leaving your device
- You control the encryption keys
- You can delete your cloud data at any time
- You can export all your data at any time
- We cannot read your encrypted memories

Cloud sync is **completely optional**. Phloem works fully offline with all features.

## Browser Extension

The Phloem browser extension:

- Only activates on ChatGPT, Claude, and Gemini websites
- Captures conversation content from these pages
- Sends captured content to your local Phloem installation via Native Messaging
- Does not send any data to external servers
- Does not access any other websites or browser data

## Data Export

You can export all your data at any time:

```bash
canopy export json my-memories.json
canopy export markdown my-memories.md
```

Your data is yours. Take it anywhere.

## Data Deletion

To delete all Phloem data:

```bash
rm -rf ~/.phloem
```

That's it. No account to delete, no support ticket to file, no waiting period.

## Third Parties

We do not share, sell, or provide your data to any third parties because we don't have your data.

## Open Source

Phloem's code is open source. You can audit exactly what it does:
- https://github.com/CanopyHQ/phloem

## Changes to This Policy

If we ever change this policy, we will:
1. Update this document
2. Announce the change prominently
3. Never make changes that reduce your privacy retroactively

## Contact

Questions about privacy? Email: privacy@canopyhq.io

---

**TL;DR**: Your memories stay on your device. We don't collect anything. You can export or delete your data anytime. That's it.
