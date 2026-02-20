package cmd

import (
	"context"
	"fmt"

	"github.com/CanopyHQ/phloem/internal/memory"
	"github.com/spf13/cobra"
)

var dreamsCmd = &cobra.Command{
	Use:   "dreams",
	Short: "Run one offline curation pass",
	Long: `Run one offline curation pass (confidence decay on stale citations).

Examples:
  phloem dreams`,
	RunE: func(cmd *cobra.Command, args []string) error { return runDreams() },
}

var decayCmd = &cobra.Command{
	Use:   "decay",
	Short: "Decay citation confidence scores",
	Long: `Apply time-based decay to citation confidence scores.

Citations that haven't been verified recently have their confidence
reduced by 10% per day since last verification, down to a minimum of 10%.

Examples:
  phloem decay`,
	RunE: func(cmd *cobra.Command, args []string) error { return runDecay() },
}

var verifyCmd = &cobra.Command{
	Use:   "verify <memory_id>",
	Short: "Verify all citations for a memory",
	Long: `Verify all citations for a given memory by checking if the
referenced files and code snippets still exist.

Examples:
  phloem verify abc123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error { return runVerify(args[0]) },
}

// runVerify verifies all citations for a memory
func runVerify(memoryID string) error {
	store, err := memory.NewStore()
	if err != nil {
		return fmt.Errorf("failed to open memory store: %w", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Get citations
	citations, err := store.GetCitations(ctx, memoryID)
	if err != nil {
		return fmt.Errorf("failed to get citations: %w", err)
	}

	if len(citations) == 0 {
		fmt.Printf("Memory %s has no citations to verify.\n", memoryID)
		return nil
	}

	fmt.Printf("Verifying %d citation(s) for memory %s...\n\n", len(citations), memoryID)

	verified := 0
	invalid := 0

	for _, citation := range citations {
		_, valid, err := store.VerifyCitation(ctx, citation.ID)
		if err != nil {
			fmt.Printf("⚠️  Error verifying citation %s: %v\n", citation.ID, err)
			continue
		}

		// Re-fetch to get updated confidence
		updatedCitations, _ := store.GetCitations(ctx, memoryID)
		for _, c := range updatedCitations {
			if c.ID == citation.ID {
				if valid {
					verified++
					fmt.Printf("✅ %s (confidence: %.0f%%)\n", c.FilePath, c.Confidence*100)
				} else {
					invalid++
					fmt.Printf("❌ %s (confidence: %.0f%%)\n", c.FilePath, c.Confidence*100)
				}
				break
			}
		}
	}

	// Get aggregate confidence
	confidence, _ := store.GetMemoryConfidence(ctx, memoryID)

	fmt.Printf("\nResults: %d verified, %d invalid\n", verified, invalid)
	fmt.Printf("Aggregate confidence: %.0f%%\n", confidence*100)

	return nil
}

// runDecay applies time-based decay to citation confidence scores
func runDecay() error {
	store, err := memory.NewStore()
	if err != nil {
		return fmt.Errorf("failed to open memory store: %w", err)
	}
	defer store.Close()

	ctx := context.Background()

	fmt.Println("Applying citation confidence decay...")
	updated, err := store.DecayCitations(ctx)
	if err != nil {
		return fmt.Errorf("failed to decay citations: %w", err)
	}

	if updated == 0 {
		fmt.Println("✅ No citations needed decay (all are recent)")
	} else {
		fmt.Printf("✅ Updated confidence for %d citation(s)\n", updated)
		fmt.Println("   (Decay: 10% per day since last verification, minimum 10%)")
	}

	return nil
}
