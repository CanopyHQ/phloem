# Phloem

**Local-first AI memory with causal graphs**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/CanopyHQ/phloem)](https://github.com/CanopyHQ/phloem/releases)

Phloem is a persistent memory engine for AI coding tools. It runs as an [MCP](https://modelcontextprotocol.io) (Model Context Protocol) server, giving Claude Code, Cursor, and Windsurf long-term memory that survives across sessions. Your conversations and context are stored locally in SQLite with vector search, fully offline, zero config. Your data never leaves your machine.

## Why Phloem?

- **Causal memory graphs** -- Not just vector search. Phloem builds a directed acyclic graph linking memories by cause and effect, so your AI assistant understands *why* things happened, not just *what*.
- **Citation verification with confidence decay** -- Memories are linked to specific code locations. When code changes, confidence scores decay automatically, so stale context does not mislead your AI.
- **Truly private** -- No accounts, no telemetry, no network calls. SQLite database on your machine. Verify it yourself with `phloem audit`.

## Quick Start

```bash
brew install phloemhq/tap/phloem
phloem setup
```

The `phloem setup` command auto-detects which IDEs you have installed (Claude Code, Cursor, Windsurf) and configures each one to use Phloem as an MCP server. That is all you need to do.

## IDE Setup Guides

### Claude Code

From your terminal:

```bash
phloem setup claude-code
```

Or from within Claude Code, ask: "please run `phloem setup claude-code` in a terminal"

No restart needed. The MCP server auto-starts on first tool use.

Under the hood, this registers Phloem via `claude mcp add`:

```bash
claude mcp add phloem -- phloem serve
```

### Cursor

From your terminal:

```bash
phloem setup cursor
```

Or from within Cursor, ask the AI to run `phloem setup cursor` in a terminal.

**Restart Cursor after setup.**

After setup, your `~/.cursor/mcp.json` will contain:

```json
{
  "mcpServers": {
    "phloem": {
      "command": "/usr/local/bin/phloem",
      "args": ["serve"]
    }
  }
}
```

### Windsurf

From your terminal:

```bash
phloem setup windsurf
```

Or from within Windsurf, ask the AI to run `phloem setup windsurf` in a terminal.

**Restart Windsurf after setup.**

After setup, your `~/.windsurf/mcp_config.json` will contain:

```json
{
  "mcpServers": {
    "phloem": {
      "command": "/usr/local/bin/phloem",
      "args": ["serve"]
    }
  }
}
```

### Auto-detect all IDEs

```bash
phloem setup
```

This finds all installed IDEs and configures them in one step.

## MCP Tools Reference

When connected via MCP, your AI tools get these capabilities:

| Tool | Description |
|------|-------------|
| `remember` | Store a memory with optional tags and context |
| `recall` | Semantic search across memories |
| `forget` | Delete a specific memory by ID |
| `list_memories` | Browse recent memories, optionally filtered by tags |
| `memory_stats` | Get statistics about the memory store |
| `session_context` | Load context from previous sessions |
| `add_citation` | Link a memory to a specific code location |
| `verify_citation` | Check if a citation still matches the code |
| `get_citations` | Get all citations for a memory |
| `verify_memory` | Verify all citations for a memory |
| `causal_query` | Query the causal memory graph (neighbors or affected) |
| `compose` | Run two semantic searches and merge results |
| `prefetch` | Preload relevant memories for current context |
| `prefetch_suggest` | Suggest memories to preload for a given context |

## CLI Commands

| Command | Description |
|---------|-------------|
| `phloem serve` | Start the MCP server (normally started by IDE) |
| `phloem setup` | Auto-detect and configure IDEs |
| `phloem status` | View memory statistics |
| `phloem doctor` | Diagnose and fix configuration issues |
| `phloem audit` | Verify privacy (data inventory, permissions, network) |
| `phloem remember` | Store a memory from the command line |
| `phloem verify ID` | Verify citations for a specific memory |
| `phloem decay` | Apply confidence decay to stale citations |
| `phloem dreams` | View the memory dream log |
| `phloem import SOURCE PATH` | Import AI history (chatgpt or claude) |
| `phloem export [FORMAT] [FILE]` | Export memories (json or markdown) |
| `phloem graft` | Merge memory databases |
| `phloem version` | Print version information |

## How It Works

- **SQLite + sqlite-vec**: Memories stored locally in `~/.phloem/memories.db` with vector embeddings for semantic search.
- **Causal DAG**: Memories are linked in a directed acyclic graph. When you recall a memory, Phloem can traverse the graph to find related causes and effects.
- **Citation verification**: Memories can be linked to specific file:line locations. Phloem checks if the code still matches and decays confidence when it does not.
- **MCP Protocol**: Communicates with IDEs via JSON-RPC over stdio. Any MCP-compatible tool can use Phloem.

## Privacy

- **No network calls**: Phloem makes zero network requests. Ever.
- **No accounts**: No sign-up, no email, no personal information.
- **No telemetry**: No analytics, no crash reporting, no usage tracking.
- **Verify it yourself**: Run `phloem audit` to inspect your data, check permissions, and confirm no network activity.
- **Full details**: See [PRIVACY.md](docs/PRIVACY.md)

## Build from Source

```bash
git clone https://github.com/CanopyHQ/phloem.git
cd phloem
CGO_ENABLED=1 go build -o phloem .
./phloem setup
```

Requires: Go 1.21+, C compiler (for sqlite-vec/CGO).

## Contributing

Phloem is licensed under Apache 2.0. Contributions are welcome.

- Report issues: [GitHub Issues](https://github.com/CanopyHQ/phloem/issues)
- Discuss: [GitHub Discussions](https://github.com/CanopyHQ/phloem/discussions)

---

Built by [Canopy HQ LLC](https://canopyhq.io). Apache 2.0 License.
