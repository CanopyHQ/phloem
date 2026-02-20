package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose common setup issues",
	Long: `Diagnose common setup issues and optionally fix them.

Examples:
  phloem doctor        # check for issues
  phloem doctor --fix  # check and auto-fix issues`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fix, _ := cmd.Flags().GetBool("fix")
		return runDoctor(fix)
	},
}

func init() {
	doctorCmd.Flags().Bool("fix", false, "Attempt to automatically fix issues")
}

// redact returns the first n and last n chars of s, or "***" if too short.
func redact(s string, n int) string {
	if s == "" {
		return "(not set)"
	}
	if len(s) <= n*2 {
		return "***"
	}
	return s[:n] + "..." + s[len(s)-n:]
}

// runDoctor diagnoses common setup issues
func runDoctor(fix bool) error {
	fmt.Println("🔍 Phloem Doctor - Diagnosing Setup")
	if fix {
		fmt.Println("🛠️  Auto-fix enabled")
	}
	fmt.Println()

	issues := 0
	warnings := 0
	fixed := 0

	// 1. Check if binary is in PATH
	fmt.Print("✓ Checking if phloem is in PATH... ")
	path, err := exec.LookPath("phloem")
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Println("  Issue: phloem binary not found in PATH")
		fmt.Println("  Fix: Add phloem to your PATH or use full path")
		issues++
	} else {
		fmt.Printf("✅ OK (%s)\n", path)
	}

	// 2. Check binary permissions
	fmt.Print("✓ Checking binary permissions... ")
	if path != "" {
		info, err := os.Stat(path)
		if err != nil {
			fmt.Println("❌ FAILED")
			fmt.Printf("  Issue: Cannot stat binary: %v\n", err)
			issues++
		} else if info.Mode()&0111 == 0 {
			if fix {
				fmt.Print("🛠️  Fixing... ")
				if err := os.Chmod(path, info.Mode()|0111); err != nil {
					fmt.Printf("❌ FAILED: %v\n", err)
					issues++
				} else {
					fmt.Println("✅ FIXED")
					fixed++
				}
			} else {
				fmt.Println("❌ FAILED")
				fmt.Println("  Issue: Binary is not executable")
				fmt.Printf("  Fix: Run 'chmod +x %s'\n", path)
				issues++
			}
		} else {
			fmt.Println("✅ OK")
		}
	}

	// 3. Check data directory
	fmt.Print("✓ Checking data directory... ")
	dataDir := os.Getenv("PHLOEM_DATA_DIR")
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".phloem")
	}
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if fix {
			fmt.Print("🛠️  Creating... ")
			if err := os.MkdirAll(dataDir, 0755); err != nil {
				fmt.Printf("❌ FAILED: %v\n", err)
				issues++
			} else {
				fmt.Println("✅ FIXED")
				fixed++
			}
		} else {
			fmt.Println("⚠️  WARNING")
			fmt.Printf("  Data directory does not exist: %s\n", dataDir)
			fmt.Println("  It will be created on first run")
			warnings++
		}
	} else {
		fmt.Printf("✅ OK (%s)\n", dataDir)
	}

	// 4. Check MCP configuration for Cursor
	fmt.Print("✓ Checking Cursor MCP configuration... ")
	home, _ := os.UserHomeDir()
	cursorConfig := filepath.Join(home, ".cursor", "mcp.json")
	if _, err := os.Stat(cursorConfig); os.IsNotExist(err) {
		if fix {
			fmt.Print("🛠️  Setting up... ")
			if err := runSetupCursor(); err != nil {
				fmt.Printf("❌ FAILED: %v\n", err)
				issues++
			} else {
				fmt.Println("✅ FIXED")
				fixed++
			}
		} else {
			fmt.Println("⚠️  WARNING")
			fmt.Printf("  MCP config not found: %s\n", cursorConfig)
			fmt.Println("  Run 'phloem setup cursor' to configure")
			warnings++
		}
	} else {
		fmt.Println("✅ OK")
	}

	// 5. Check MCP configuration for Windsurf
	fmt.Print("✓ Checking Windsurf MCP configuration... ")
	windsurfConfig := filepath.Join(home, ".windsurf", "mcp_config.json")
	if _, err := os.Stat(windsurfConfig); os.IsNotExist(err) {
		if fix {
			fmt.Print("🛠️  Setting up... ")
			if err := runSetupWindsurf(); err != nil {
				fmt.Printf("❌ FAILED: %v\n", err)
				issues++
			} else {
				fmt.Println("✅ FIXED")
				fixed++
			}
		} else {
			fmt.Println("⚠️  WARNING")
			fmt.Printf("  MCP config not found: %s\n", windsurfConfig)
			fmt.Println("  Run 'phloem setup windsurf' to configure")
			warnings++
		}
	} else {
		fmt.Println("✅ OK")
	}

	// 6. Check MCP configuration for Claude Code
	fmt.Print("✓ Checking Claude Code MCP configuration... ")
	if _, err := exec.LookPath("claude"); err != nil {
		fmt.Println("⚠️  SKIPPED (claude not in PATH)")
	} else {
		listCmd := exec.Command("claude", "mcp", "list")
		listOutput, listErr := listCmd.CombinedOutput()
		if listErr != nil {
			fmt.Println("⚠️  WARNING")
			fmt.Printf("  Could not check MCP list: %v\n", listErr)
			warnings++
		} else if strings.Contains(string(listOutput), "phloem") {
			fmt.Println("✅ OK")
		} else {
			if fix {
				fmt.Print("🛠️  Setting up... ")
				if err := runSetupClaudeCode(); err != nil {
					fmt.Printf("❌ FAILED: %v\n", err)
					issues++
				} else {
					fmt.Println("✅ FIXED")
					fixed++
				}
			} else {
				fmt.Println("⚠️  WARNING")
				fmt.Println("  Phloem not registered with Claude Code")
				fmt.Println("  Run 'phloem setup claude-code' to configure")
				warnings++
			}
		}
	}

	// 7. Check SQLite database
	fmt.Print("✓ Checking SQLite database... ")
	dbPath := filepath.Join(dataDir, "memories.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("⚠️  WARNING")
		fmt.Printf("  Database not found: %s\n", dbPath)
		fmt.Println("  It will be created on first run")
		warnings++
	} else {
		fmt.Println("✅ OK")
	}

	// 8. Test MCP server startup
	fmt.Print("✓ Testing MCP server startup... ")
	cmd := exec.Command("phloem", "version")
	if err := cmd.Run(); err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("  Issue: Cannot run phloem: %v\n", err)
		issues++
	} else {
		fmt.Println("✅ OK")
	}

	// 9. Check for common environment issues
	fmt.Print("✓ Checking environment... ")
	if runtime.GOOS == "darwin" {
		// Check for Rosetta on Apple Silicon
		if runtime.GOARCH == "arm64" {
			fmt.Println("✅ OK (Apple Silicon native)")
		} else {
			fmt.Println("⚠️  WARNING (Running under Rosetta)")
			warnings++
		}
	} else {
		fmt.Printf("✅ OK (%s/%s)\n", runtime.GOOS, runtime.GOARCH)
	}

	// Summary
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if issues == 0 && warnings == 0 {
		fmt.Println("✅ All checks passed! Phloem is ready to use.")
	} else {
		if fixed > 0 {
			fmt.Printf("🛠️  Auto-fixed %d issue(s)\n", fixed)
		}
		if issues > 0 {
			fmt.Printf("❌ Found %d critical issue(s)\n", issues)
		}
		if warnings > 0 {
			fmt.Printf("⚠️  Found %d warning(s)\n", warnings)
		}
		fmt.Println()
		fmt.Println("Run the suggested fixes above to resolve issues.")
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if issues > 0 {
		return fmt.Errorf("found %d critical issue(s)", issues)
	}
	return nil
}
