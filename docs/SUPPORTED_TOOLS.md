# Supported Tools

Phloem works with any tool that supports the [Model Context Protocol (MCP)](https://modelcontextprotocol.io). Below are setup instructions for every tool we've verified.

**Don't see your tool?** If it supports MCP stdio servers, it works with Phloem. [Open an issue](https://github.com/CanopyHQ/phloem/issues) and we'll add setup instructions.

---

## Auto-Detect (Recommended)

```bash
phloem setup
```

This scans your machine for supported tools and configures each one automatically. Currently auto-detects: Claude Code, VS Code, Cursor, Windsurf, Zed, Cline, and Neovim.

---

## Claude Code

```bash
phloem setup claude-code
```

Or from within Claude Code: *"please run `phloem setup claude-code` in a terminal"*

No restart needed. The MCP server auto-starts on first tool use.

---

## VS Code (GitHub Copilot)

Requires VS Code 1.99+ with GitHub Copilot and Agent Mode enabled.

```bash
phloem setup vscode
```

This creates a user-level MCP config. Your config will look like:

**macOS:** `~/Library/Application Support/Code/User/mcp.json`
**Linux:** `~/.config/Code/User/mcp.json`
**Windows:** `%APPDATA%\Code\User\mcp.json`

```json
{
  "servers": {
    "phloem": {
      "type": "stdio",
      "command": "/usr/local/bin/phloem",
      "args": ["serve"]
    }
  }
}
```

No restart needed. VS Code detects config changes automatically.

---

## Cursor

```bash
phloem setup cursor
```

Restart Cursor after setup. Config at `~/.cursor/mcp.json`:

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

## Windsurf

```bash
phloem setup windsurf
```

Restart Windsurf after setup. Config at `~/.windsurf/mcp_config.json`:

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

## Zed

```bash
phloem setup zed
```

No restart needed — Zed hot-reloads settings. Config added to your Zed settings file:

**macOS:** `~/.zed/settings.json`
**Linux:** `~/.config/zed/settings.json`

```json
{
  "context_servers": {
    "phloem": {
      "command": "/usr/local/bin/phloem",
      "args": ["serve"]
    }
  }
}
```

---

## Cline (VS Code Extension)

```bash
phloem setup cline
```

No restart needed. Config at:

**macOS:** `~/Library/Application Support/Code/User/globalStorage/saoudrizwan.claude-dev/settings/cline_mcp_settings.json`
**Linux:** `~/.config/Code/User/globalStorage/saoudrizwan.claude-dev/settings/cline_mcp_settings.json`
**Windows:** `%APPDATA%\Code\User\globalStorage\saoudrizwan.claude-dev\settings\cline_mcp_settings.json`

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

## Neovim (via mcphub.nvim)

```bash
phloem setup neovim
```

Requires the [mcphub.nvim](https://github.com/ravitemer/mcphub.nvim) plugin. Config at `~/.config/mcphub/servers.json`:

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

No restart needed — run `:MCPHub` in Neovim to reload. Works with avante.nvim, codecompanion.nvim, and any plugin that integrates with mcphub.

---

## Warp

```bash
phloem setup warp
```

Warp stores MCP config in the cloud, not on disk. The command above prints the JSON to paste into Warp's UI:

1. Open **Settings > MCP Servers**
2. Click **+ Add**
3. Paste the JSON shown by the command
4. Click **Save**

No restart needed — available on next message.

---

## JetBrains IDEs (IntelliJ, WebStorm, PyCharm, etc.)

JetBrains IDEs with AI Assistant support MCP clients starting from 2025.1.

**Automated setup is not available** — JetBrains stores MCP config inside its settings UI, not in a standalone file.

To configure manually:

1. Open **Settings > Tools > AI Assistant > Model Context Protocol (MCP)**
2. Click **Add** and paste:

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

3. Click **OK**. No restart needed.

Replace `/usr/local/bin/phloem` with the output of `which phloem`.

---

## Continue (VS Code / JetBrains Extension)

Continue uses YAML configuration. Add to `~/.continue/config.yaml`:

```yaml
mcpServers:
  - name: phloem
    type: stdio
    command: /usr/local/bin/phloem
    args:
      - serve
```

Reload the Continue extension after saving.

Replace `/usr/local/bin/phloem` with the output of `which phloem`.

---

## Any Other MCP Client

Phloem is a standard MCP stdio server. To connect from any MCP client:

- **Command:** `phloem serve`
- **Transport:** stdio (JSON-RPC over stdin/stdout)
- **No environment variables required** (data defaults to `~/.phloem/`)
- **Optional:** Set `PHLOEM_DATA_DIR` to change the storage location

If your tool supports MCP and isn't listed here, it should work out of the box. Please [let us know](https://github.com/CanopyHQ/phloem/issues) so we can add instructions.
