package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/CanopyHQ/phloem/internal/memory"
)

// runDreams runs one offline curation pass (confidence decay on stale citations).
func runDreams() error {
	store, err := memory.NewStore()
	if err != nil {
		return fmt.Errorf("failed to open store: %w", err)
	}
	ctx := context.Background()
	// Decay citations older than 24h by factor 0.99
	stats, err := store.DreamRun(ctx, 24*time.Hour, 0.99)
	if err != nil {
		return fmt.Errorf("dream run failed: %w", err)
	}
	fmt.Printf("✨ Dreams run complete: decayed %d citations in %d ms\n",
		stats.DecayedCitations, stats.DurationMs)
	return nil
}
