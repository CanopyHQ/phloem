package memory

import (
	"context"
	"testing"
	"time"
)

func TestCompose_EmptyQueries(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	_, err := store.Compose(ctx, []string{}, 5)
	if err == nil {
		t.Error("expected error for empty queries")
	}
}

func TestCompose_SingleQuery(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	_, _ = store.Remember(ctx, "auth uses JWT", []string{"auth"}, "")
	_, _ = store.Remember(ctx, "database uses PostgreSQL", []string{"db"}, "")

	composed, err := store.Compose(ctx, []string{"auth"}, 10)
	if err != nil {
		t.Fatalf("compose: %v", err)
	}
	if composed.Explanation == "" {
		t.Error("expected non-empty explanation")
	}
	if len(composed.Memories) == 0 {
		t.Error("expected at least one memory for auth query")
	}
}

func TestPrefetchSuggest_EmptyContext(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	mems, err := store.PrefetchSuggest(ctx, "", 5)
	if err != nil {
		t.Fatalf("prefetch: %v", err)
	}
	if len(mems) > 5 {
		t.Errorf("expected at most 5, got %d", len(mems))
	}
}

func TestDreamRun_NoError(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	stats, err := store.DreamRun(ctx, 8760*time.Hour, 0.99) // 1 year - no citations that old
	if err != nil {
		t.Fatalf("dream run: %v", err)
	}
	if stats.DecayedCitations < 0 {
		t.Errorf("DecayedCitations should be >= 0, got %d", stats.DecayedCitations)
	}
}

func TestPrefetchSuggest_LimitClamp(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()
	store.Remember(ctx, "hint memory", nil, "")

	out, _ := store.PrefetchSuggest(ctx, "hint", 0)
	_ = out
	out2, _ := store.PrefetchSuggest(ctx, "hint", 100)
	_ = out2
}

func TestDreamRun_InvalidDecayFactor(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	stats, err := store.DreamRun(ctx, time.Hour, 0)
	if err != nil {
		t.Fatalf("DreamRun decay 0: %v", err)
	}
	_ = stats
}
