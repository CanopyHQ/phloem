package memory

import (
	"context"
	"testing"
)

func TestAddEdge_EdgesFrom(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	m1, _ := store.Remember(ctx, "memory one", []string{"a"}, "")
	m2, _ := store.Remember(ctx, "memory two", []string{"b"}, "")

	if err := store.AddEdge(ctx, m1.ID, m2.ID, "causal", ""); err != nil {
		t.Fatalf("AddEdge: %v", err)
	}
	// Filter by "causal" so we ignore the automatic temporal edge from Remember()
	edges, err := store.GetEdgesFrom(ctx, m1.ID, "causal")
	if err != nil {
		t.Fatalf("GetEdgesFrom: %v", err)
	}
	if len(edges) != 1 {
		t.Fatalf("expected 1 causal edge, got %d", len(edges))
	}
	if edges[0].TargetID != m2.ID || edges[0].EdgeType != "causal" {
		t.Errorf("edge: TargetID=%q EdgeType=%q", edges[0].TargetID, edges[0].EdgeType)
	}
}

func TestEdgesTo(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	m1, _ := store.Remember(ctx, "one", nil, "")
	m2, _ := store.Remember(ctx, "two", nil, "")
	_ = store.AddEdge(ctx, m1.ID, m2.ID, "causal", "")

	// Filter by "causal" so we ignore the automatic temporal edge from Remember()
	edges, err := store.GetEdgesTo(ctx, m2.ID, "causal")
	if err != nil {
		t.Fatalf("GetEdgesTo: %v", err)
	}
	if len(edges) != 1 || edges[0].SourceID != m1.ID {
		t.Errorf("expected one causal edge from m1 to m2, got %+v", edges)
	}
}

// TestEdgesFrom_Store tests the graph.go convenience wrapper EdgesFrom (no type filter).
func TestEdgesFrom_Store(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	m1, _ := store.Remember(ctx, "one", nil, "")
	m2, _ := store.Remember(ctx, "two", nil, "")
	_ = store.AddEdge(ctx, m1.ID, m2.ID, "causal", "")

	edges, err := store.EdgesFrom(ctx, m1.ID)
	if err != nil {
		t.Fatalf("EdgesFrom: %v", err)
	}
	if len(edges) < 1 {
		t.Errorf("expected at least 1 edge from m1, got %d", len(edges))
	}
}

// TestEdgesTo_Store tests the graph.go convenience wrapper EdgesTo (no type filter).
func TestEdgesTo_Store(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	m1, _ := store.Remember(ctx, "one", nil, "")
	m2, _ := store.Remember(ctx, "two", nil, "")
	_ = store.AddEdge(ctx, m1.ID, m2.ID, "causal", "")

	edges, err := store.EdgesTo(ctx, m2.ID)
	if err != nil {
		t.Fatalf("EdgesTo: %v", err)
	}
	if len(edges) < 1 {
		t.Errorf("expected at least 1 edge to m2, got %d", len(edges))
	}
}
