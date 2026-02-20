// Package memory: Stage 2 features (Cursor 2) — Compose, Prefetch, Dreams.

package memory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

// ComposeResult holds the result of a compositional recall.
type ComposeResult struct {
	Memories    []*Memory `json:"memories"`
	Explanation string    `json:"explanation"`
}

// Compose recalls memories for multiple queries and merges them (dedupe by ID, best score wins).
// Returns combined memories and a short explanation for how the result was derived.
func (s *Store) Compose(ctx context.Context, queries []string, limit int) (*ComposeResult, error) {
	if limit <= 0 {
		limit = 10
	}
	// Require at least one non-empty query
	hasQuery := false
	for _, q := range queries {
		if strings.TrimSpace(q) != "" {
			hasQuery = true
			break
		}
	}
	if !hasQuery {
		return nil, fmt.Errorf("at least one non-empty query is required")
	}
	seen := make(map[string]*Memory)
	var explanationParts []string

	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		mems, err := s.Recall(ctx, q, limit*2, nil)
		if err != nil {
			return nil, fmt.Errorf("recall query %q: %w", q, err)
		}
		for _, m := range mems {
			if existing, ok := seen[m.ID]; !ok || m.Similarity > existing.Similarity {
				seen[m.ID] = m
			}
		}
		explanationParts = append(explanationParts, fmt.Sprintf("%q (%d)", q, len(mems)))
	}

	// Sort by similarity and take top limit
	merged := make([]*Memory, 0, len(seen))
	for _, m := range seen {
		merged = append(merged, m)
	}
	sortBySimilarity(merged)
	if len(merged) > limit {
		merged = merged[:limit]
	}

	explanation := "Combined " + strings.Join(explanationParts, ", ") + " → " + fmt.Sprintf("%d memories", len(merged))
	return &ComposeResult{Memories: merged, Explanation: explanation}, nil
}

func sortBySimilarity(mems []*Memory) {
	sort.Slice(mems, func(i, j int) bool { return mems[i].Similarity > mems[j].Similarity })
}

// PrefetchSuggest returns memories likely needed next given current context (e.g. open file path or last query).
// Simple implementation: recall on currentContext and return as suggested preload.
func (s *Store) PrefetchSuggest(ctx context.Context, currentContext string, limit int) ([]*Memory, error) {
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}
	return s.Recall(ctx, currentContext, limit, nil)
}

// DreamStats holds stats from a dream run.
type DreamStats struct {
	DecayedCitations int   `json:"decayed_citations"`
	ReembeddedCount  int   `json:"reembedded_count,omitempty"`
	DurationMs       int64 `json:"duration_ms"`
}

// DreamRun runs a single offline curation pass: confidence decay on stale citations.
// Citations older than maxAge have confidence multiplied by decayFactor (e.g. 0.99).
func (s *Store) DreamRun(ctx context.Context, maxAge time.Duration, decayFactor float64) (DreamStats, error) {
	start := time.Now()
	cutoff := time.Now().Add(-maxAge)
	if decayFactor <= 0 || decayFactor > 1 {
		decayFactor = 0.99
	}

	res, err := s.db.ExecContext(ctx, `
		UPDATE citations SET confidence = confidence * ?
		WHERE (COALESCE(verified_at, created_at) < ?) AND confidence > 0.01
	`, decayFactor, cutoff)
	if err != nil {
		return DreamStats{}, fmt.Errorf("dream decay: %w", err)
	}
	affected, _ := res.RowsAffected()

	return DreamStats{
		DecayedCitations: int(affected),
		DurationMs:       time.Since(start).Milliseconds(),
	}, nil
}
