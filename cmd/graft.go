package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CanopyHQ/phloem/internal/graft"
	"github.com/CanopyHQ/phloem/internal/memory"
	"github.com/spf13/cobra"
)

var graftCmd = &cobra.Command{
	Use:   "graft",
	Short: "Shareable memory bundles",
	Long: `Create, import, and inspect shareable memory bundles (grafts).

Examples:
  phloem graft export --tags "architecture,patterns" --output arch.graft
  phloem graft import arch.graft
  phloem graft inspect arch.graft`,
}

func init() {
	// graft export
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Create a graft file from memories",
		RunE: func(cmd *cobra.Command, args []string) error {
			tags, _ := cmd.Flags().GetString("tags")
			since, _ := cmd.Flags().GetString("since")
			output, _ := cmd.Flags().GetString("output")
			name, _ := cmd.Flags().GetString("name")
			desc, _ := cmd.Flags().GetString("desc")
			author, _ := cmd.Flags().GetString("author")
			return runGraftExport(tags, since, output, name, desc, author)
		},
	}
	exportCmd.Flags().String("tags", "", "Comma-separated tags to include")
	exportCmd.Flags().String("since", "", "Include memories since duration (e.g. 24h)")
	exportCmd.Flags().String("output", "", "Output filename (required)")
	exportCmd.Flags().String("name", "", "Graft name")
	exportCmd.Flags().String("desc", "", "Graft description")
	exportCmd.Flags().String("author", "", "Author name")
	graftCmd.AddCommand(exportCmd)

	// graft import
	importCmd := &cobra.Command{
		Use:   "import [file.graft]",
		Short: "Import a graft file (local or from registry)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			from, _ := cmd.Flags().GetString("from")
			var filePath string
			if len(args) >= 1 {
				filePath = args[0]
			}
			return runGraftImport(from, filePath)
		},
	}
	importCmd.Flags().String("from", "", "Import from registry URL")
	graftCmd.AddCommand(importCmd)

	// graft inspect
	graftCmd.AddCommand(&cobra.Command{
		Use:   "inspect <file.graft>",
		Short: "View graft manifest without importing",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGraftInspect(args[0])
		},
	})
}

func runGraftExport(tagsStr, sinceStr, output, name, desc, author string) error {
	if output == "" {
		return fmt.Errorf("--output is required")
	}

	store, err := memory.NewStore()
	if err != nil {
		return fmt.Errorf("failed to open store: %w", err)
	}
	defer store.Close()

	var tagList []string
	if tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			tagList = append(tagList, strings.TrimSpace(t))
		}
	}

	allMemories, err := store.List(context.Background(), 10000, nil)
	if err != nil {
		return fmt.Errorf("failed to list memories: %w", err)
	}

	var filteredMemories []memory.Memory
	var sinceTime time.Time
	if sinceStr != "" {
		dur, err := time.ParseDuration(sinceStr)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
		sinceTime = time.Now().Add(-dur)
	}

	for _, m := range allMemories {
		if !sinceTime.IsZero() && m.CreatedAt.Before(sinceTime) {
			continue
		}

		if len(tagList) > 0 {
			matched := false
			for _, tag := range tagList {
				for _, mTag := range m.Tags {
					if strings.EqualFold(tag, mTag) {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
			if !matched {
				continue
			}
		}

		filteredMemories = append(filteredMemories, *m)
	}

	if len(filteredMemories) == 0 {
		fmt.Println("No memories matched criteria.")
		return nil
	}

	manifest := graft.Manifest{
		Name:        name,
		Description: desc,
		Author:      author,
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		MemoryCount: len(filteredMemories),
		Tags:        tagList,
	}

	if manifest.Name == "" {
		manifest.Name = fmt.Sprintf("Export %s", time.Now().Format("2006-01-02"))
	}
	if manifest.Author == "" {
		user, _ := os.UserHomeDir()
		manifest.Author = filepath.Base(user)
	}

	err = graft.Package(manifest, filteredMemories, nil, output)
	if err != nil {
		return fmt.Errorf("failed to package graft: %w", err)
	}

	fmt.Printf("📦 Created %s (%d memories)\n", output, len(filteredMemories))
	return nil
}

func runGraftImport(fromURL, filePath string) error {
	var inputPath string
	if fromURL != "" {
		inputPath = downloadGraftFromRegistry(fromURL)
		if inputPath == "" {
			return fmt.Errorf("failed to download graft from registry")
		}
		defer os.Remove(inputPath)
	} else if filePath != "" {
		inputPath = filePath
	} else {
		return fmt.Errorf("provide a file path or --from URL")
	}

	fmt.Printf("📦 Reading %s...\n", inputPath)
	payload, err := graft.Unpack(inputPath)
	if err != nil {
		return fmt.Errorf("failed to unpack graft: %w", err)
	}

	fmt.Printf("🔓 Verifying format... OK\n")
	fmt.Printf("📄 Manifest: %s by %s\n", payload.Manifest.Name, payload.Manifest.Author)
	if payload.Manifest.Description != "" {
		fmt.Printf("   %s\n", payload.Manifest.Description)
	}

	store, err := memory.NewStore()
	if err != nil {
		return fmt.Errorf("failed to open store: %w", err)
	}
	defer store.Close()

	source := fmt.Sprintf("graft:%s:%s", payload.Manifest.Name, payload.Manifest.Author)

	count := 0
	ctx := context.Background()
	for _, m := range payload.Memories {
		m.Source = source

		hasGraftTag := false
		for _, t := range m.Tags {
			if t == "graft" {
				hasGraftTag = true
				break
			}
		}
		if !hasGraftTag {
			m.Tags = append(m.Tags, "graft")
		}

		if err := store.Add(ctx, m); err != nil {
			fmt.Printf("⚠️ Failed to import memory %s: %v\n", m.ID, err)
			continue
		}
		count++
	}

	fmt.Printf("✨ Imported %d memories (source: %s)\n", count, source)
	return nil
}

// downloadGraftFromRegistry downloads a .graft file from a URL to a temp file
func downloadGraftFromRegistry(rawURL string) string {
	fmt.Printf("📥 Downloading graft from %s...\n", rawURL)

	if !strings.HasPrefix(rawURL, "https://") {
		fmt.Println("❌ Only HTTPS URLs are supported for graft downloads")
		return ""
	}

	resp, err := http.Get(rawURL)
	if err != nil {
		fmt.Printf("❌ Failed to download: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Server returned %d\n", resp.StatusCode)
		return ""
	}

	tmpFile, err := os.CreateTemp("", "graft-*.graft")
	if err != nil {
		fmt.Printf("❌ Failed to create temp file: %v\n", err)
		return ""
	}

	// Limit download to 50MB to prevent disk exhaustion
	limited := io.LimitReader(resp.Body, 50*1024*1024)
	if _, err := io.Copy(tmpFile, limited); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		fmt.Printf("❌ Failed to save graft: %v\n", err)
		return ""
	}
	tmpFile.Close()

	return tmpFile.Name()
}

func runGraftInspect(inputPath string) error {
	manifest, err := graft.Inspect(inputPath)
	if err != nil {
		return fmt.Errorf("failed to inspect graft: %w", err)
	}

	fmt.Println("📦 Graft Manifest")
	fmt.Println("================")
	fmt.Printf("Name:        %s\n", manifest.Name)
	fmt.Printf("Description: %s\n", manifest.Description)
	fmt.Printf("Author:      %s\n", manifest.Author)
	fmt.Printf("Version:     %s\n", manifest.Version)
	fmt.Printf("Created:     %s\n", manifest.CreatedAt)
	fmt.Printf("Memories:    %d\n", manifest.MemoryCount)
	fmt.Printf("Tags:        %v\n", manifest.Tags)

	return nil
}
