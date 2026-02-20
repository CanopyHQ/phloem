package cmd

import (
	"github.com/spf13/cobra"
)

// Build-time variables
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// SetVersion sets the version info from main
func SetVersion(v, c, d string) {
	Version = v
	Commit = c
	Date = d
}

var rootCmd = &cobra.Command{
	Use:   "phloem",
	Short: "Phloem MCP - AI Memory Layer",
	Long:  "Local-first memory for AI tools via Model Context Protocol.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServe()
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the phloem command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// serve, version, status (defined in serve.go)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(statusCmd)

	// import, export (defined in import_export.go)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(exportCmd)

	// dreams, decay, verify (defined in misc.go)
	rootCmd.AddCommand(dreamsCmd)
	rootCmd.AddCommand(decayCmd)
	rootCmd.AddCommand(verifyCmd)

	// doctor (defined in doctor.go)
	rootCmd.AddCommand(doctorCmd)

	// remember (defined in remember.go)
	rootCmd.AddCommand(rememberCmd)

	// setup (defined in setup.go)
	rootCmd.AddCommand(setupCmd)

	// graft (defined in graft.go)
	rootCmd.AddCommand(graftCmd)
}
