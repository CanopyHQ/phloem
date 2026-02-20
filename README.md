# 🧠 Canopy

**Your AI's memory, synchronized across every tool you use.**

Canopy is a local-first memory layer that gives AI tools persistent context. Chat with Claude Code, switch to Cursor, then continue in Windsurf — your AI remembers everything. No more repeating yourself. No more lost context.

[![Version](https://img.shields.io/badge/version-0.6.1--beta-blue)](https://github.com/CanopyHQ/canopy/releases)
[![License](https://img.shields.io/badge/license-Proprietary-red)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-macOS-lightgrey)](https://github.com/CanopyHQ/canopy)

> **Early Access**: Canopy is in public beta. The first 100 Pro users get **Founding Member** pricing — $5/mo locked for up to 36 months (reasonable terms apply). Core features are free.

## The Problem

Every AI tool today lives in a silo:
- Cursor doesn't know what you told Claude Code
- Your IDE agent forgets when you start a new session
- Context gets lost when you switch between tools
- You waste time re-explaining the same things

**Your workflows span multiple AI tools. Your memory shouldn't be trapped in any single one.**

## What Canopy Does

Canopy creates a **unified memory layer** that works across all your AI tools. It's like having a shared brain for all your AI assistants.

```bash
# Remember something from the terminal
$ canopy remember "Our API uses JWT tokens with 24h expiry" --tags api,auth

# Your AI tools recall it automatically via MCP
# In Cursor, Claude Code, or Windsurf — just ask about auth tokens
# and Canopy provides the context.
```

Your AI tools connect to Canopy via [MCP](https://modelcontextprotocol.io) (Model Context Protocol), an open standard for AI tool integration. When you chat with AI in Cursor, it automatically has access to everything you've stored.

### Key Features

**🔒 Privacy-First**
- All data stored locally on your device (`~/.phloem/memories.db`)
- No cloud required, no account needed
- You own your data completely

**🌐 Cross-Tool Memory**
- Works with Cursor, Windsurf, and Claude Code via MCP
- Memories available everywhere, instantly
- One source of truth for all your AI context

**🔍 Semantic Search**
- Find memories by meaning, not exact words
- Built-in embeddings (no external API needed)
- Fast retrieval even with thousands of memories

**📎 Citation Tracking**
- Link memories to specific code locations (file + line range)
- Verify citations are still accurate after refactoring
- Trace which memories are affected by code changes

**📤 Portable**
- Export as JSON or Markdown anytime
- Import existing conversations from ChatGPT or Claude
- Memory Grafts: shareable context bundles for team onboarding

**⚡ Lightweight**
- Native macOS binary, minimal resources
- Works offline — no internet dependency
- Starts in milliseconds

## Installation

**macOS (Homebrew):**
```bash
brew tap canopyhq/tap
brew install canopy
```

**Verify Installation:**
```bash
canopy version
canopy doctor    # diagnose common setup issues
```

## Quick Start

### 1. Set Up IDE Integration

The fastest way — auto-detects your installed IDEs:

```bash
canopy setup
```

Or configure a specific IDE:

```bash
canopy setup cursor
canopy setup windsurf
canopy setup claude-code
```

That's it. Restart your IDE, and Canopy's MCP tools are available.

<details>
<summary>Manual configuration (if you prefer)</summary>

**Cursor** (`~/.cursor/mcp.json`):
```json
{
  "mcpServers": {
    "canopy": {
      "command": "canopy",
      "args": ["serve"]
    }
  }
}
```

**Windsurf** (`~/.codeium/windsurf/mcp_config.json`):
```json
{
  "mcpServers": {
    "canopy": {
      "command": "canopy",
      "args": ["serve"]
    }
  }
}
```

**Claude Code:**
```bash
claude mcp add canopy canopy serve
```
</details>

### 2. Store Your First Memory

```bash
# From terminal
canopy remember "Our database schema uses UUID primary keys" --tags db,schema

# Or from your IDE — just chat naturally:
# "Remember that we're using Tailwind CSS for styling"
```

### 3. AI Tools Automatically Recall

When you chat with AI in Cursor, Claude Code, or Windsurf:
```
You: "Show me how to create a new user"

AI: [Automatically recalls your database schema via Canopy]
    "Based on your UUID primary keys, here's how to create a user..."
```

No explicit recall needed — your AI tools fetch relevant context automatically via MCP.

## Command Line Usage

### Core Commands

```bash
# Store information (with optional tags)
canopy remember "Important context about your project"
canopy remember "API endpoints use /api/v2 prefix" --tags api,versioning

# View memory statistics
canopy status

# Start MCP server (for IDE integration)
canopy serve

# Diagnose setup issues
canopy doctor
```

> **Note**: `recall`, `forget`, and `list` are MCP tools, not CLI commands. They're available to your AI tools when connected via MCP. Use the CLI for `remember`, `status`, and `serve`.

### Import Existing Conversations

```bash
# Import ChatGPT export
canopy import chatgpt ~/Downloads/conversations.json

# Import Claude conversations
canopy import claude ~/Downloads/claude-conversations/
```

### Export Your Data

```bash
# Export as JSON
canopy export json memories-backup.json

# Export as Markdown
canopy export markdown memories.md
```

### Memory Grafts

Share curated memory collections with your team:

```bash
# Create a graft from specific tags
canopy graft export --tags "python,best-practices" --output python-guide.graft

# Share with team — teammate runs:
canopy graft import python-guide.graft

# Inspect a graft before importing
canopy graft inspect python-guide.graft
```

## MCP Tools Reference

When connected via MCP, your AI tools get these capabilities:

| Tool | Description |
|------|-------------|
| `remember` | Store memories with tags and citations |
| `recall` | Semantic search across memories |
| `forget` | Delete a specific memory |
| `list_memories` | Browse recent memories |
| `session_context` | Preload relevant context at session start |
| `compose` | Combine memories from multiple topics |
| `add_citation` | Link a memory to a specific code location |
| `verify_citation` | Check if cited code has changed |
| `causal_query` | Trace which memories are affected by changes |
| `prefetch` | Predictive memory loading for current context |

## Use Cases

### Software Development
```bash
# Store project conventions
canopy remember "We use React Query for data fetching" --tags frontend
canopy remember "Error handling uses custom ErrorBoundary" --tags frontend,errors

# Your IDE agent now knows your codebase patterns
```

### Team Onboarding
```bash
# Export project context for new teammates
canopy graft export --tags "architecture,conventions" --output onboarding.graft

# New teammate imports it and hits the ground running
canopy graft import onboarding.graft
```

### Cross-Tool Workflows
```bash
# Morning: Research in Claude Code
# Afternoon: Code in Cursor
# Evening: Quick fix in Windsurf
# All sessions share the same memory
```

## How It Works

```
┌─────────────────────────────────────────────────┐
│  Your AI Tools                                  │
│  • Cursor (MCP)                                │
│  • Windsurf (MCP)                              │
│  • Claude Code (MCP)                           │
│  • Terminal CLI                                 │
└─────────────────┬───────────────────────────────┘
                  │
                  ▼
         ┌─────────────────┐
         │  Canopy Memory  │
         │  Layer (MCP)    │
         └────────┬─────────┘
                  │
                  ▼
         ┌─────────────────┐
         │  Local Storage  │
         │  ~/.phloem/     │
         │  memories.db    │
         └─────────────────┘
```

**MCP Protocol**: Canopy implements the [Model Context Protocol](https://modelcontextprotocol.io), an open standard for AI tool integration. Any MCP-compatible tool can connect to Canopy.

**Local Embeddings**: Semantic search uses TF-IDF embeddings computed locally. No external API calls, no data leaves your machine.

**Storage**: SQLite database at `~/.phloem/memories.db`. Plain SQL schema you can inspect or export anytime.

## Cloud Sync (Optional)

```bash
# Sign in to enable multi-device sync
canopy signin

# Or connect manually
canopy connect
```

All data stays local by default. Cloud sync is opt-in for multi-device scenarios.

## Privacy & Security

**Local-First Philosophy**
- All memories stored on your device
- No data transmitted without explicit action
- No account required for core features

**Telemetry**
- Opt-in only
- Only version, platform, and anonymous device ID
- No content, queries, or personal data

**Disable Telemetry:**
```bash
canopy telemetry disable
```

**Delete Everything:**
```bash
rm -rf ~/.phloem
```

## Storage & Performance

**Storage Location**: `~/.phloem/memories.db`

**Database Size**: ~1KB per memory (text + metadata + embeddings)
- 1,000 memories ≈ 1 MB
- 10,000 memories ≈ 10 MB
- 100,000 memories ≈ 100 MB

**Search Performance**:
- <10ms for 10K memories
- <50ms for 100K memories
- Uses SQLite FTS5 + TF-IDF embeddings

## Troubleshooting

Run the built-in diagnostics first:
```bash
canopy doctor
```

**Cursor not showing Canopy tools:**
```bash
# Re-run setup
canopy setup cursor

# Verify MCP config exists
cat ~/.cursor/mcp.json

# Restart Cursor
```

**Canopy command not found:**
```bash
# If installed via Homebrew
brew link canopy

# Check it's in PATH
which canopy
```

## Roadmap

- [x] MCP server for IDE integration (Cursor, Windsurf, Claude Code)
- [x] CLI for terminal workflows
- [x] Local embeddings (no external dependencies)
- [x] Memory Grafts (shareable context bundles)
- [x] Citation tracking and verification
- [x] Auto-setup for IDEs (`canopy setup`)
- [x] Cloud sync for multi-device
- [ ] Browser extension (ChatGPT/Claude capture)
- [ ] Windows support
- [ ] Linux support
- [ ] VS Code extension
- [ ] Team collaboration features

## FAQ

**Q: Does Canopy replace my AI tool?**
No. Canopy adds memory to your existing AI tools. You still use Claude Code, Cursor, Windsurf, etc. — they just remember more.

**Q: What's the difference between Canopy and [Mem0](https://github.com/mem0ai/mem0) / [Zep](https://github.com/getzep/zep)?**
Canopy is local-first (no server required), uses the MCP standard (works with any compatible tool), and is designed for individual developers. Mem0/Zep are server-based memory for AI apps.

**Q: Is this free?**
Core features are free and always will be. Pro tier ($9/mo) adds cloud sync, full memory history, and priority support. Founding Members (first 100 Pro users) get $5/mo locked for up to 36 months.

**Q: Is my data encrypted?**
Data is stored in plaintext SQLite on your device. Your disk encryption (FileVault on macOS) protects it at rest.

**Q: Does this work offline?**
Yes. Core functionality works 100% offline. Only cloud sync requires internet.

**Q: Why beta?**
The core memory layer is stable and production-tested. We're in beta to gather feedback before committing to a stable API.

## Support

- **Email**: [support@canopyhq.io](mailto:support@canopyhq.io)
- **CLI**: Run `canopy support` to send an email with system info pre-filled
- **Web**: [canopyhq.io/support](https://canopyhq.io/support)
- **Self-diagnose**: Run `canopy doctor` to check your setup

## License

Proprietary. Free for individual use. See [LICENSE](LICENSE).

---

**Built by [Canopy](https://canopyhq.io)**

*Part of the Canopy AI Infrastructure Platform*
