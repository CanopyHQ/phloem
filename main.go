// Phloem MCP - AI Memory Layer
// Local-first memory for AI tools via Model Context Protocol
package main

import (
	"fmt"
	"os"

	"github.com/CanopyHQ/phloem/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersion(version, commit, date)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
