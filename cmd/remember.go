package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/CanopyHQ/phloem/internal/memory"
	"github.com/spf13/cobra"
)

var rememberCmd = &cobra.Command{
	Use:   "remember <content>",
	Short: "Store a memory in Phloem",
	Long: `Store a memory in Phloem with optional tags.

Examples:
  canopy remember "always use snake_case for Go test names"
  canopy remember "prefer composition over inheritance" --tags "architecture,patterns"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tagsStr, _ := cmd.Flags().GetString("tags")
		return runRemember(args[0], tagsStr)
	},
}

func init() {
	rememberCmd.Flags().String("tags", "", "Comma-separated tags")
}

func runRemember(content, tagsStr string) error {
	if content == "" {
		fmt.Println("Usage: canopy remember \"<content>\" [--tags \"tag1,tag2,...\"]")
		return nil
	}
	store, err := memory.NewStore()
	if err != nil {
		return fmt.Errorf("failed to open memory store: %w", err)
	}
	defer store.Close()
	var tags []string
	if tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			if s := strings.TrimSpace(t); s != "" {
				tags = append(tags, s)
			}
		}
	}
	ctx := context.Background()
	if _, err := store.Remember(ctx, content, tags, ""); err != nil {
		return fmt.Errorf("remember failed: %w", err)
	}
	fmt.Println("✅ Remembered.")
	return nil
}
