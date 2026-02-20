package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/CanopyHQ/phloem/internal/importer"
	"github.com/CanopyHQ/phloem/internal/memory"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import <source> <path>",
	Short: "Import AI history (chatgpt or claude)",
	Long: `Import AI conversation history from ChatGPT or Claude exports.

Supported sources:
  chatgpt  - Import from ChatGPT JSON export
  claude   - Import from Claude JSON export

The path can be a single JSON file or a directory containing JSON files.

Examples:
  phloem import chatgpt ~/Downloads/conversations.json
  phloem import claude ~/Downloads/claude-export/`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error { return runImport(args[0], args[1]) },
}

var exportCmd = &cobra.Command{
	Use:   "export [format] [output]",
	Short: "Export all memories",
	Long: `Export all memories to a file.

Supported formats:
  json      - JSON format (default)
  markdown  - Markdown format

If no output path is given, a default filename is generated.

Examples:
  phloem export
  phloem export json memories.json
  phloem export markdown memories.md`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		format, output := "json", ""
		if len(args) >= 1 {
			format = args[0]
		}
		if len(args) >= 2 {
			output = args[1]
		}
		return runExport(format, output)
	},
}

func runImport(source, path string) error {
	store, err := memory.NewStore()
	if err != nil {
		return fmt.Errorf("failed to open memory store: %w", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Check if path is file or directory
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access path: %w", err)
	}

	var result *importer.ImportResult

	switch source {
	case "chatgpt":
		imp := importer.NewChatGPTImporter(store)
		if info.IsDir() {
			fmt.Printf("Importing ChatGPT conversations from directory: %s\n", path)
			result, err = imp.ImportFromDirectory(ctx, path)
		} else {
			fmt.Printf("Importing ChatGPT conversations from file: %s\n", path)
			result, err = imp.ImportFromFile(ctx, path)
		}

	case "claude":
		imp := importer.NewClaudeImporter(store)
		if info.IsDir() {
			fmt.Printf("Importing Claude conversations from directory: %s\n", path)
			result, err = imp.ImportFromDirectory(ctx, path)
		} else {
			fmt.Printf("Importing Claude conversations from file: %s\n", path)
			result, err = imp.ImportFromFile(ctx, path)
		}

	default:
		return fmt.Errorf("unknown source: %s (supported: chatgpt, claude)", source)
	}

	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	fmt.Printf("\n✅ Import Complete!\n")
	fmt.Printf("   Conversations processed: %d\n", result.ConversationsProcessed)
	fmt.Printf("   Memories created: %d\n", result.MemoriesCreated)
	fmt.Printf("   Duration: %s\n", result.Duration.Round(time.Millisecond))

	if len(result.Errors) > 0 {
		fmt.Printf("\n⚠️  Errors (%d):\n", len(result.Errors))
		for i, e := range result.Errors {
			if i >= 5 {
				fmt.Printf("   ... and %d more\n", len(result.Errors)-5)
				break
			}
			fmt.Printf("   - %s\n", e)
		}
	}

	return nil
}

// runExport exports all memories to a file
func runExport(format, output string) error {
	store, err := memory.NewStore()
	if err != nil {
		return fmt.Errorf("failed to open memory store: %w", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Get all memories
	memories, err := store.List(ctx, 100000, nil) // Get all
	if err != nil {
		return fmt.Errorf("failed to list memories: %w", err)
	}

	if len(memories) == 0 {
		fmt.Println("No memories to export.")
		return nil
	}

	var data []byte

	switch format {
	case "json":
		// Export as JSON
		type ExportMemory struct {
			ID        string    `json:"id"`
			Content   string    `json:"content"`
			Context   string    `json:"context,omitempty"`
			Tags      []string  `json:"tags"`
			CreatedAt time.Time `json:"created_at"`
		}

		exportData := make([]ExportMemory, len(memories))
		for i, m := range memories {
			exportData[i] = ExportMemory{
				ID:        m.ID,
				Content:   m.Content,
				Context:   m.Context,
				Tags:      m.Tags,
				CreatedAt: m.CreatedAt,
			}
		}

		data, err = json.MarshalIndent(exportData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}

	case "markdown", "md":
		// Export as Markdown
		var sb strings.Builder
		sb.WriteString("# Phloem Memory Export\n\n")
		sb.WriteString(fmt.Sprintf("Exported: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
		sb.WriteString(fmt.Sprintf("Total memories: %d\n\n", len(memories)))
		sb.WriteString("---\n\n")

		for _, m := range memories {
			// Title from first line
			title := m.Content
			if idx := strings.Index(title, "\n"); idx > 0 {
				title = title[:idx]
			}
			if len(title) > 80 {
				title = title[:80] + "..."
			}

			sb.WriteString(fmt.Sprintf("## %s\n\n", title))
			sb.WriteString(fmt.Sprintf("*%s*", m.CreatedAt.Format("2006-01-02 15:04")))
			if len(m.Tags) > 0 {
				sb.WriteString(fmt.Sprintf(" | Tags: %s", strings.Join(m.Tags, ", ")))
			}
			sb.WriteString("\n\n")
			sb.WriteString(m.Content)
			sb.WriteString("\n\n---\n\n")
		}

		data = []byte(sb.String())

	default:
		return fmt.Errorf("unknown format: %s (supported: json, markdown)", format)
	}

	// Output
	if output == "" {
		// Generate default filename
		timestamp := time.Now().Format("2006-01-02")
		ext := format
		if format == "markdown" {
			ext = "md"
		}
		output = fmt.Sprintf("phloem-export-%s.%s", timestamp, ext)
	}

	if err := os.WriteFile(output, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✅ Exported %d memories to %s\n", len(memories), output)
	return nil
}
