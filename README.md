# Phloem

**Your AI forgets everything when you close the tab. Phloem fixes that.**

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/CanopyHQ/phloem)](https://github.com/CanopyHQ/phloem/releases)
[![MCP](https://img.shields.io/badge/MCP-compatible-green)](https://modelcontextprotocol.io)

Phloem is a local [MCP](https://modelcontextprotocol.io) server that gives your AI persistent memory across sessions. It works with **any tool that supports the Model Context Protocol** — an open standard. Today that includes Claude Code, VS Code, Cursor, Windsurf, Zed, Neovim, Cline, Warp, Continue, JetBrains, and more arriving every week.

You install it once. Every MCP-compatible AI tool on your machine gets memory.

```bash
brew install phloemhq/tap/phloem   # or download binary from Releases
phloem setup                        # auto-detects your tools
```

---

## Why Not Just Use `.claude` Files or Markdown?

Every AI coding tool has some form of project context — `.claude/` directories, `CLAUDE.md`, `.cursorrules`, system prompts. They work. So why do you need Phloem?

**They're siloed.** Your `.claude` files don't help Cursor. Your `.cursorrules` don't help Claude Code. Switch tools and you start from zero. Phloem is one memory that works across all of them because MCP is a shared standard.

**They're static text.** You write rules by hand and hope the AI reads them. Phloem memories are created naturally during conversation, found via semantic search (not keyword matching), and linked to the code they reference.

**They can't track what changed.** You wrote "we use JWT with 15-minute expiry" in a markdown file. Then someone changed it to 30 minutes. The file still says 15. Phloem **citations** link memories to specific lines of code and **automatically decay confidence** when the code drifts. Your AI never confidently cites stale information.

**They don't capture *why*.** Markdown files tell the AI *what* to do. Phloem's **causal graph** captures the chain of decisions — *why* you chose JWT over sessions, what that caused you to change, and what would break if you reverted it.

---

## What Makes Phloem Different

### Semantic memory, not keyword matching

```
You: "Add the logout endpoint"
```

Phloem finds "JWT tokens expire in 15 minutes, refresh tokens in 7 days" — even though nothing in your query mentions JWT or tokens. It understands *meaning*, not just words.

### Citations that know when code changes

```
Memory: "Rate limiter uses sliding window, 100 req/min per API key"
  → src/middleware/rate_limit.go:42-67 (confidence: 0.95)
```

Refactor that file? Confidence decays. Delete it? Citation marked invalid. Your AI adapts.

### Causal graphs, not flat lists

```
"Switched from REST to gRPC"
  → caused: "Updated service clients to protobuf"
  → caused: "Added buf.gen.yaml to build pipeline"
  → caused: "Rewrote integration tests for streaming"
```

Ask "what would break if we reverted gRPC?" and get a real answer.

### Completely offline

Zero network requests. Not as a policy — there is no networking code in the binary. Your memories are a SQLite file on your disk. `phloem audit` proves it.

### Tool-agnostic by design

Phloem speaks MCP — the same open protocol that Claude Code, VS Code, Cursor, Zed, and others already support. Add a new AI tool? If it speaks MCP, it already has access to your full memory.

---

## Quick Start with Claude Code

```bash
brew install phloemhq/tap/phloem
phloem setup claude-code
```

No restart needed. Start a session and your AI has memory.

For **all other supported tools** (VS Code, Cursor, Windsurf, Zed, Neovim, Cline, Warp, Continue, JetBrains), see the **[Supported Tools Guide](docs/SUPPORTED_TOOLS.md)**.

Or just run `phloem setup` — it auto-detects everything on your machine.

---

## Install

**macOS (Homebrew):**

```bash
brew install phloemhq/tap/phloem
phloem setup
```

**Windows / Linux:**

Download the binary for your platform from [GitHub Releases](https://github.com/CanopyHQ/phloem/releases), extract it, add to PATH, then run `phloem setup`.

**From source:**

```bash
git clone https://github.com/CanopyHQ/phloem.git && cd phloem
CGO_ENABLED=1 go build -o phloem .
./phloem setup
```

Requires Go 1.24+ and a C compiler (for sqlite-vec/CGO).

---

## How It Works

**SQLite + sqlite-vec** — Everything in `~/.phloem/memories.db`. Vector embeddings power semantic search. No external services.

**Causal DAG** — Memories linked by cause and effect. Your AI traverses the graph to understand full chains of reasoning.

**Citation verification** — Memories attach to `file:line` ranges. When code drifts, confidence decays automatically.

**MCP Protocol** — JSON-RPC over stdio. No HTTP server, no ports, no network surface. Any MCP client connects instantly.

---

## Privacy

There is no networking code in the binary. Verify it:

```bash
phloem audit
sudo lsof -i -P | grep phloem    # should show nothing
```

No accounts. No telemetry. No analytics. Delete anytime: `rm -rf ~/.phloem`. [Full details →](docs/PRIVACY.md)

---

## Contributing

We welcome contributions. See **[CONTRIBUTING.md](CONTRIBUTING.md)** for guidelines.

- [Report issues](https://github.com/CanopyHQ/phloem/issues)
- [Discussions](https://github.com/CanopyHQ/phloem/discussions)

---

Apache 2.0 License. Built by [Canopy HQ LLC](https://canopyhq.io).
