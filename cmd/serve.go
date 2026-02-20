package cmd

import (
	"fmt"
	"os"

	"github.com/CanopyHQ/phloem/internal/mcp"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:     "serve",
	Aliases: []string{"mcp"},
	Short:   "Start MCP server (default)",
	Long: `Start the MCP server using stdio transport.

The server communicates via JSON-RPC over stdin/stdout and is designed
to be connected to by an MCP client such as Claude Code, Cursor, etc.

Examples:
  phloem serve
  phloem mcp`,
	RunE: func(cmd *cobra.Command, args []string) error { return runServe() },
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("phloem %s (commit: %s, built: %s)\n", Version, Commit, Date)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show memory statistics",
	Long: `Show current memory statistics including total memories,
database size, and last activity.

Examples:
  phloem status`,
	RunE: func(cmd *cobra.Command, args []string) error { return runStatus() },
}

func runServe() error {
	fmt.Fprintln(os.Stderr, "🧠 Phloem MCP - AI Memory Layer")
	fmt.Fprintln(os.Stderr, "Starting MCP server (stdio transport)...")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "This server communicates via JSON-RPC over stdin/stdout.")
	fmt.Fprintln(os.Stderr, "It is not an interactive CLI — connect an MCP client (Claude Code, Cursor, etc.).")
	fmt.Fprintln(os.Stderr, "Press Ctrl+C to stop. Run 'phloem help' for available commands.")
	fmt.Fprintln(os.Stderr, "")

	mcp.Version = Version

	server, err := mcp.NewServer()
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	return server.Start()
}

func runStatus() error {
	server, err := mcp.NewServer()
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	stats := server.GetMemoryStats()
	fmt.Printf("Phloem Memory Status:\n")
	fmt.Printf("  Total Memories: %d\n", stats.TotalMemories)
	fmt.Printf("  Database Size: %s\n", stats.DatabaseSize)
	fmt.Printf("  Last Activity: %s\n", stats.LastActivity)
	return nil
}
