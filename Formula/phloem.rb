# Homebrew Formula for Phloem
# This is a template - actual formula lives in CanopyHQ/homebrew-tap repository
# This file shows the structure and is used by GoReleaser to generate the real formula

class Phloem < Formula
  desc "Local-first AI memory with causal graphs"
  homepage "https://github.com/CanopyHQ/phloem"
  version "1.0.0"
  license "Apache-2.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/CanopyHQ/phloem/releases/download/v#{version}/phloem-darwin-arm64.tar.gz"
      sha256 "PLACEHOLDER_ARM64_SHA256"  # Replaced by GoReleaser
    else
      url "https://github.com/CanopyHQ/phloem/releases/download/v#{version}/phloem-darwin-amd64.tar.gz"
      sha256 "PLACEHOLDER_AMD64_SHA256"  # Replaced by GoReleaser
    end
  end

  def install
    bin.install "phloem"
  end

  test do
    # Test MCP protocol initialization
    assert_match "protocolVersion", shell_output("echo '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\",\"params\":{}}' | #{bin}/phloem serve 2>/dev/null | head -1")
  end

  def caveats
    <<~EOS
      Phloem is installed!

      Your AI memories travel with you across tools.

      Phloem is a local-first AI memory layer with causal graphs, keeping your memory usable across sessions, tools, and models.

      Quick setup for agentic IDE tools (beta: Cursor, Windsurf):
        phloem setup cursor
        phloem setup windsurf

      Or manually add to ~/.cursor/mcp.json:
        {
          "mcpServers": {
            "phloem": {
              "command": "#{opt_bin}/phloem",
              "args": ["serve"]
            }
          }
        }

      Commands:
        phloem serve    - Start MCP server
        phloem status   - View memory stats
        phloem graft    - Export/import memory grafts
        phloem help     - Full command list

      Documentation: https://github.com/CanopyHQ/phloem
    EOS
  end
end
