# Phloem

**Long-term memory for AI coding tools — local, private, causal**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/CanopyHQ/phloem)](https://github.com/CanopyHQ/phloem/releases)

Your AI assistant forgets everything when you close the tab. Phloem fixes that.

It runs as an [MCP](https://modelcontextprotocol.io) server that gives Claude Code, Cursor, and Windsurf persistent memory across sessions. Two commands to install, zero config, everything stays on your machine.

```bash
brew install phloemhq/tap/phloem
phloem setup
```

That's it. Your AI now remembers.

---

## See It In Action

### Your AI remembers decisions across sessions

You're in Claude Code working on an auth system:

> **You:** "We decided to use JWT with refresh tokens, not session cookies. The tokens expire in 15 minutes, refresh tokens in 7 days."

Phloem stores this. Two weeks later, in a new session:

> **You:** "Add the logout endpoint"

Your AI already knows the auth architecture — token expiry, refresh strategy, everything. No re-explaining. It calls `recall` behind the scenes and picks up right where you left off.

### Memories link to code — and know when code changes

Phloem doesn't just store text. It links memories to specific lines in your codebase:

```
Memory: "Rate limiter uses sliding window, 100 req/min per API key"
  → src/middleware/rate_limit.go:42-67 (confidence: 0.95)
```

Refactor that file? Phloem notices. The confidence score **decays automatically**, so your AI won't confidently cite stale information. Run `phloem decay` to update scores, or let it happen naturally.

### Understand *why*, not just *what*

Most memory tools are glorified search. Phloem builds a **causal graph** — a directed acyclic graph linking memories by cause and effect:

```
"Switched from REST to gRPC for inter-service calls"
  → caused: "Updated all service clients to use protobuf"
  → caused: "Added buf.gen.yaml to build pipeline"
  → caused: "Rewrote integration tests for gRPC streaming"
```

When your AI recalls the protobuf migration, it can traverse the graph to understand the *full chain of reasoning* that led there — not just the isolated fact.

### Import history you already have

Already been using Claude or ChatGPT? Bring those conversations along:

```bash
phloem import chatgpt ~/Downloads/conversations.json
phloem import claude ~/Downloads/claude-export/
```

### Inspect everything — trust nothing

```bash
$ phloem audit
Data inventory:
  ~/.phloem/memories.db    2.4 MB   (SQLite + sqlite-vec)

Permissions:
  ~/.phloem/               drwx------  (owner only) ✓
  ~/.phloem/memories.db    -rw-------  (owner only) ✓

Network activity:
  No listening sockets found ✓
  Recommendation: run `sudo lsof -i -P | grep phloem` to verify
```

No accounts. No telemetry. No network calls. Ever. [Full privacy details →](docs/PRIVACY.md)

---

## Quick Start

```bash
brew install phloemhq/tap/phloem
phloem setup
```

`phloem setup` auto-detects your IDEs (Claude Code, Cursor, Windsurf) and configures each one. That's the entire setup.

### Claude Code

```bash
phloem setup claude-code
```

Or from within Claude Code, ask: "please run `phloem setup claude-code` in a terminal"

No restart needed. The MCP server auto-starts on first tool use.

### Cursor

```bash
phloem setup cursor
```

Restart Cursor after setup. Your `~/.cursor/mcp.json` will contain:

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

```bash
phloem setup windsurf
```

Restart Windsurf after setup. Your `~/.windsurf/mcp_config.json` will contain:

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

---

## What Phloem Does

### MCP Tools (used by your AI automatically)

| Tool | What it does |
|------|-------------|
| `remember` | Store a memory with tags and context — "always use snake_case in this repo" |
| `recall` | Semantic search — finds relevant memories even with different wording |
| `session_context` | Load previous session context so the AI starts warm, not cold |
| `add_citation` | Link a memory to `file:line` — "this decision is implemented at auth.go:42" |
| `verify_citation` | Check if cited code still matches — catches stale references |
| `verify_memory` | Verify all citations for a memory at once |
| `causal_query` | Traverse the causal graph — "what else was affected by this decision?" |
| `compose` | Merge two semantic searches — "find memories about auth AND deployment" |
| `prefetch` | Preload relevant memories for the current file/context |
| `forget` | Delete a specific memory |
| `list_memories` | Browse recent memories, filter by tags |
| `memory_stats` | How many memories, tags, citations in the store |

### CLI Commands (used by you)

| Command | What it does |
|---------|-------------|
| `phloem setup` | Auto-detect and configure all your IDEs |
| `phloem status` | See how many memories you have, disk usage, last activity |
| `phloem remember "..."` | Store a memory from the terminal — `phloem remember "use bun not npm" --tags "tooling"` |
| `phloem doctor` | Diagnose configuration issues — checks IDE configs, database health |
| `phloem audit` | Privacy verification — data inventory, file permissions, network check |
| `phloem decay` | Apply confidence decay to citations where the code has changed |
| `phloem dreams` | View the memory dream log — see what your AI has been remembering |
| `phloem verify ID` | Check if a specific memory's code citations still match |
| `phloem import` | Import ChatGPT or Claude conversation history |
| `phloem export` | Export all memories as JSON or Markdown |
| `phloem graft` | Share curated memory packs — export your best practices as a `.graft` file |
| `phloem version` | Print version info |

---

## How It Works

**SQLite + sqlite-vec** — Everything stored locally in `~/.phloem/memories.db`. Vector embeddings power semantic search — "find memories about authentication" matches "JWT token refresh logic" even though the words don't overlap.

**Causal DAG** — Memories are linked in a directed acyclic graph. When a decision leads to a change, and that change causes another, Phloem captures the chain. Your AI can ask "what would be affected if we reverted this?" and get a real answer.

**Citation verification** — Memories attach to `file:line` ranges. Phloem reads the file and compares. When code drifts, confidence decays. When code is deleted, the citation is marked invalid. Your AI never confidently cites code that no longer exists.

**MCP Protocol** — JSON-RPC over stdio. Any MCP-compatible tool works. No HTTP server, no ports, no network surface.

---

## Privacy

Phloem makes **zero network requests**. This is not a policy — it's architecture. There is no networking code in the binary. Verify it:

```bash
phloem audit                              # data inventory + permissions
sudo lsof -i -P | grep phloem            # should show nothing
```

No accounts. No telemetry. No analytics. No crash reporting. Your memories are a SQLite file on your disk. Delete it anytime: `rm -rf ~/.phloem`.

[Full privacy policy →](docs/PRIVACY.md)

---

## Build from Source

```bash
git clone https://github.com/CanopyHQ/phloem.git
cd phloem
CGO_ENABLED=1 go build -o phloem .
./phloem setup
```

Requires: Go 1.21+, C compiler (for sqlite-vec/CGO).

## Contributing

Apache 2.0. Contributions welcome.

- [Report issues](https://github.com/CanopyHQ/phloem/issues)
- [Discussions](https://github.com/CanopyHQ/phloem/discussions)

---

Built by [Canopy HQ LLC](https://canopyhq.io). Apache 2.0 License.
