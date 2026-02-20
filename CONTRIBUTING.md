# Contributing to Phloem

Thanks for your interest in contributing to Phloem. This guide covers everything you need to get started.

## Code of Conduct

Be respectful. Be constructive. We're building something together.

## How to Contribute

### Reporting Bugs

1. Check [existing issues](https://github.com/CanopyHQ/phloem/issues) first
2. Run `phloem doctor` and include the output
3. Include your OS, Go version, and IDE
4. Provide steps to reproduce

### Suggesting Features

Open a [discussion](https://github.com/CanopyHQ/phloem/discussions) first. We'll promote it to an issue if it fits the roadmap.

### Pull Requests

All changes require a pull request. Direct pushes to `main` are blocked.

1. Fork the repo and create a branch from `main`
2. Make your changes
3. Add or update tests
4. Run the quality checks (see below)
5. Open a PR with a clear description

#### PR Guidelines

- **One concern per PR.** Don't mix a bug fix with a refactor.
- **Write tests.** We maintain 70%+ coverage.
- **Follow existing patterns.** Read surrounding code before adding new patterns.
- **Keep commits clean.** Squash fixups before requesting review.

## Development Setup

### Prerequisites

- Go 1.24+
- C compiler (GCC, MinGW, or Xcode CLT) — required for sqlite-vec/CGO
- `golangci-lint` (optional, for linting)

### Build and Test

```bash
# Build
CGO_ENABLED=1 go build -o phloem .

# Run all tests
go test ./... -v -race

# Run linter
golangci-lint run

# Quick preflight (build + tests + vet)
make preflight

# Full preflight (includes acceptance tests + privacy verification)
make preflight-release
```

### MCP Protocol Testing

```bash
# Verify MCP server starts and responds
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./phloem serve 2>/dev/null | head -1 | jq .

# Verify all 14 tools are registered
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | ./phloem serve 2>/dev/null | tail -1 | jq '.result.tools | length'
```

## Architecture

Phloem is a single Go binary with no runtime dependencies:

```
cmd/           CLI commands (cobra)
internal/
  mcp/         MCP JSON-RPC server (stdio transport)
  memory/      Memory store (SQLite + sqlite-vec), embeddings, causal graph
grafts/        Shareable memory packs
docs/          Documentation
test/          Acceptance tests (godog/Gherkin)
```

### Key Design Principles

- **Zero network.** No HTTP clients, no outbound connections, no telemetry. Ever.
- **Single file storage.** Everything in `~/.phloem/memories.db` (SQLite).
- **MCP standard.** Phloem is a tool-agnostic MCP server. No IDE-specific code in the core.
- **CGO required.** sqlite-vec needs CGO. This is a deliberate tradeoff for local vector search.

## What We're Looking For

High-impact areas where contributions are especially welcome:

- **New IDE setup support** — Add auto-configuration for more MCP-compatible tools
- **Import formats** — More conversation import sources beyond ChatGPT and Claude
- **Embedding improvements** — Better local embedding models or strategies
- **Graft packs** — Curated memory packs for popular frameworks and languages
- **Documentation** — Tutorials, guides, translations

## License

By contributing, you agree that your contributions will be licensed under the Apache 2.0 License.
