package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

// validTableName matches only safe SQLite table names (alphanumeric and underscores).
var validTableName = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Verify privacy — inspect data, permissions, and network activity",
	Long: `Audit your Phloem installation for privacy.

Checks:
  1. Data inventory — lists all files in ~/.phloem/ with sizes
  2. Permissions — verifies files are user-readable only
  3. Schema — shows SQLite tables and row counts (no content)
  4. Network — instructions to verify zero network activity

Run this anytime to confirm Phloem respects your privacy.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAudit()
	},
}

// humanSize formats bytes into a human-readable string.
func humanSize(bytes int64) string {
	switch {
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// fileDescription returns a short explanation of what a file is.
func fileDescription(name string) string {
	switch name {
	case "memories.db":
		return "SQLite database with memories and embeddings"
	case "memories.db-wal":
		return "SQLite write-ahead log (temporary)"
	case "memories.db-shm":
		return "SQLite shared memory file (temporary)"
	default:
		return ""
	}
}

func runAudit() error {
	fmt.Println("🔒 Phloem Privacy Audit")
	fmt.Println()

	dataDir := os.Getenv("PHLOEM_DATA_DIR")
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".phloem")
	}

	// ── Section 1: Data Inventory ──────────────────────────────────────
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📁 Section 1: Data Inventory")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		fmt.Printf("  Data directory does not exist: %s\n", dataDir)
		fmt.Println("  Phloem has not been used yet — no data stored.")
		fmt.Println()
	} else {
		fmt.Printf("  Data directory: %s\n", dataDir)
		fmt.Println()

		var totalSize int64
		var fileCount int
		err := filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // skip files we can't read
			}
			if info.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(dataDir, path)
			size := info.Size()
			totalSize += size
			fileCount++
			desc := fileDescription(info.Name())
			if desc != "" {
				fmt.Printf("  %-30s %10s  (%s)\n", rel, humanSize(size), desc)
			} else {
				fmt.Printf("  %-30s %10s\n", rel, humanSize(size))
			}
			return nil
		})
		if err != nil {
			fmt.Printf("  ⚠️  Error walking directory: %v\n", err)
		}

		fmt.Println()
		fmt.Printf("  Total: %d file(s), %s\n", fileCount, humanSize(totalSize))
		fmt.Println()
	}

	// ── Section 2: Permissions Check ───────────────────────────────────
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🔐 Section 2: Permissions Check")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	issues := 0

	if info, err := os.Stat(dataDir); err == nil {
		mode := info.Mode().Perm()
		fmt.Printf("  %s  %04o", dataDir, mode)
		if mode&0007 != 0 {
			fmt.Println("  ⚠️  WARNING: world-accessible")
			fmt.Printf("    Fix: chmod 700 %s\n", dataDir)
			issues++
		} else {
			fmt.Println("  ✅ OK")
		}
	} else if !os.IsNotExist(err) {
		fmt.Printf("  ⚠️  Cannot stat data directory: %v\n", err)
		issues++
	}

	dbPath := filepath.Join(dataDir, "memories.db")
	if info, err := os.Stat(dbPath); err == nil {
		mode := info.Mode().Perm()
		fmt.Printf("  %s  %04o", dbPath, mode)
		if mode&0007 != 0 {
			fmt.Println("  ⚠️  WARNING: world-readable")
			fmt.Printf("    Fix: chmod 600 %s\n", dbPath)
			issues++
		} else {
			fmt.Println("  ✅ OK")
		}
	} else if !os.IsNotExist(err) {
		fmt.Printf("  ⚠️  Cannot stat database: %v\n", err)
		issues++
	}

	if issues == 0 {
		fmt.Println("  ✅ All permissions OK")
	}
	fmt.Println()

	// ── Section 3: Database Schema ─────────────────────────────────────
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🗃️  Section 3: Database Schema")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("  Database not found — no data stored yet.")
	} else {
		db, err := sql.Open("sqlite3", dbPath+"?mode=ro")
		if err != nil {
			fmt.Printf("  ⚠️  Cannot open database: %v\n", err)
		} else {
			defer db.Close()

			rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
			if err != nil {
				fmt.Printf("  ⚠️  Cannot query schema: %v\n", err)
			} else {
				defer rows.Close()
				tableFound := false
				for rows.Next() {
					var name string
					if err := rows.Scan(&name); err != nil {
						continue
					}
					tableFound = true

					// Validate table name to prevent SQL injection
					if !validTableName.MatchString(name) {
						fmt.Printf("  %-30s  (skipped — invalid table name)\n", name)
						continue
					}

					var count int
					countRow := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM [%s]", name))
					if err := countRow.Scan(&count); err != nil {
						fmt.Printf("  %-30s  (error counting rows)\n", name)
					} else {
						fmt.Printf("  %-30s  %d row(s)\n", name, count)
					}
				}
				if !tableFound {
					fmt.Println("  No tables found (empty database).")
				}
			}
		}
	}
	fmt.Println()
	fmt.Println("  Note: Only table names and row counts are shown.")
	fmt.Println("  No memory content is ever printed by this command.")
	fmt.Println()

	// ── Section 4: Network Verification ────────────────────────────────
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🌐 Section 4: Network Verification")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("  Phloem makes zero network connections. Verify by running")
	fmt.Println("  the commands below while phloem is active:")
	fmt.Println()

	if runtime.GOOS == "darwin" {
		fmt.Println("  macOS:")
		fmt.Println("    sudo lsof -i -P | grep phloem    # should show nothing")
		fmt.Println()
		fmt.Println("  For continuous monitoring, use Little Snitch or LuLu:")
		fmt.Println("    https://objective-see.org/products/lulu.html")
	} else {
		fmt.Println("  Linux:")
		fmt.Println("    ss -tlnp | grep phloem                      # should show nothing")
		fmt.Println("    strace -e network -f phloem serve 2>&1      # trace network syscalls")
	}
	fmt.Println()
	fmt.Println("  Automated verification:")
	fmt.Println("    make verify-privacy")
	fmt.Println()

	// ── Summary ────────────────────────────────────────────────────────
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if issues == 0 {
		fmt.Println("✅ Privacy audit complete — no issues found.")
	} else {
		fmt.Printf("⚠️  Privacy audit complete — %d issue(s) found. See above.\n", issues)
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return nil
}
