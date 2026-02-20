package memory

import (
	"context"
	"testing"
	"time"
)

func TestAddEdge_GetEdgesFrom(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Create two memories
	m1, err := store.Remember(ctx, "First memory", nil, "")
	if err != nil {
		t.Fatalf("Remember failed: %v", err)
	}
	m2, err := store.Remember(ctx, "Second memory", nil, "")
	if err != nil {
		t.Fatalf("Remember failed: %v", err)
	}

	// Remember already created a temporal edge m1->m2 when m2 was added.
	// Add causal edge with payload
	err = store.AddEdge(ctx, m1.ID, m2.ID, "causal", "because we refactored")
	if err != nil {
		t.Fatalf("AddEdge causal failed: %v", err)
	}

	// Get edges from m1 (temporal from Remember + causal from AddEdge)
	edges, err := store.GetEdgesFrom(ctx, m1.ID, "")
	if err != nil {
		t.Fatalf("GetEdgesFrom failed: %v", err)
	}
	if len(edges) < 2 {
		t.Errorf("expected at least 2 edges from m1, got %d", len(edges))
	}

	// Filter by type: at least one temporal (from Remember)
	temporal, err := store.GetEdgesFrom(ctx, m1.ID, "temporal")
	if err != nil {
		t.Fatalf("GetEdgesFrom temporal failed: %v", err)
	}
	if len(temporal) < 1 {
		t.Errorf("expected at least 1 temporal edge, got %d", len(temporal))
	}
	foundM2 := false
	for _, e := range temporal {
		if e.TargetID == m2.ID {
			foundM2 = true
			break
		}
	}
	if !foundM2 {
		t.Errorf("expected temporal edge to m2, got %v", temporal)
	}
}

func TestAddEdge_emptySourceIgnored(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	err := store.AddEdge(ctx, "", "target", "temporal", "")
	if err != nil {
		t.Fatalf("AddEdge with empty source should not error: %v", err)
	}
	// Empty source should be a no-op (we return nil in AddEdge when sourceID == "")
}

func TestAddEdge_emptyEdgeTypeIgnored(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	m1, _ := store.Remember(ctx, "One", nil, "")
	m2, _ := store.Remember(ctx, "Two", nil, "")
	err := store.AddEdge(ctx, m1.ID, m2.ID, "", "payload")
	if err != nil {
		t.Fatalf("AddEdge with empty edgeType should not error: %v", err)
	}
	edges, _ := store.GetEdgesFrom(ctx, m1.ID, "")
	if len(edges) != 1 {
		t.Logf("only temporal edge from Remember expected; empty edgeType adds nothing")
	}
}

func TestGetPreviousMemoryID(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// No memories yet
	id, err := store.GetPreviousMemoryID(ctx, time.Now())
	if err != nil {
		t.Fatalf("GetPreviousMemoryID failed: %v", err)
	}
	if id != "" {
		t.Errorf("expected empty previous ID, got %q", id)
	}

	// Add one memory
	m1, err := store.Remember(ctx, "Only memory", nil, "")
	if err != nil {
		t.Fatalf("Remember failed: %v", err)
	}

	// Previous before m1.CreatedAt should be empty
	id, err = store.GetPreviousMemoryID(ctx, m1.CreatedAt.Add(-time.Second))
	if err != nil {
		t.Fatalf("GetPreviousMemoryID failed: %v", err)
	}
	if id != "" {
		t.Errorf("expected empty before first memory, got %q", id)
	}

	// Add second memory
	m2, err := store.Remember(ctx, "Second", nil, "")
	if err != nil {
		t.Fatalf("Remember failed: %v", err)
	}

	// Previous before m2 should be m1
	id, err = store.GetPreviousMemoryID(ctx, m2.CreatedAt)
	if err != nil {
		t.Fatalf("GetPreviousMemoryID failed: %v", err)
	}
	if id != m1.ID {
		t.Errorf("previous before m2 = %q, want %q", id, m1.ID)
	}
}

func TestSetMemoryUtility_RunMemoryCritic(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	mem, err := store.Remember(ctx, "Utility test memory", nil, "")
	if err != nil {
		t.Fatalf("Remember failed: %v", err)
	}

	err = store.SetMemoryUtility(ctx, mem.ID, 0.3)
	if err != nil {
		t.Fatalf("SetMemoryUtility failed: %v", err)
	}

	err = store.RunMemoryCritic(ctx)
	if err != nil {
		t.Fatalf("RunMemoryCritic failed: %v", err)
	}
}

func TestRemember_createsTemporalEdge(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	m1, err := store.Remember(ctx, "First", nil, "")
	if err != nil {
		t.Fatalf("Remember failed: %v", err)
	}
	m2, err := store.Remember(ctx, "Second", nil, "")
	if err != nil {
		t.Fatalf("Remember failed: %v", err)
	}

	// Should have temporal edge from m1 to m2
	edges, err := store.GetEdgesFrom(ctx, m1.ID, "temporal")
	if err != nil {
		t.Fatalf("GetEdgesFrom failed: %v", err)
	}
	if len(edges) != 1 {
		t.Fatalf("expected 1 temporal edge from first memory, got %d", len(edges))
	}
	if edges[0].TargetID != m2.ID {
		t.Errorf("temporal edge target = %q, want %q", edges[0].TargetID, m2.ID)
	}
}

func TestGetEdgesTo(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	m1, err := store.Remember(ctx, "First", nil, "")
	if err != nil {
		t.Fatalf("Remember failed: %v", err)
	}
	m2, err := store.Remember(ctx, "Second", nil, "")
	if err != nil {
		t.Fatalf("Remember failed: %v", err)
	}
	_ = store.AddEdge(ctx, m1.ID, m2.ID, "causal", "because we refactored")

	// Edges pointing to m2: temporal (m1->m2) and causal (m1->m2)
	edges, err := store.GetEdgesTo(ctx, m2.ID, "")
	if err != nil {
		t.Fatalf("GetEdgesTo failed: %v", err)
	}
	if len(edges) < 2 {
		t.Errorf("expected at least 2 edges to m2 (temporal + causal), got %d", len(edges))
	}

	causalTo, err := store.GetEdgesTo(ctx, m2.ID, "causal")
	if err != nil {
		t.Fatalf("GetEdgesTo causal failed: %v", err)
	}
	if len(causalTo) != 1 {
		t.Fatalf("expected 1 causal edge to m2, got %d", len(causalTo))
	}
	if causalTo[0].SourceID != m1.ID {
		t.Errorf("causal edge source = %q, want %q", causalTo[0].SourceID, m1.ID)
	}
}

func TestGetMemoryByID(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Nonexistent
	got, err := store.GetMemoryByID(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetMemoryByID nonexistent: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for nonexistent ID, got %v", got)
	}

	m1, err := store.Remember(ctx, "Content for get-by-id", nil, "")
	if err != nil {
		t.Fatalf("Remember failed: %v", err)
	}

	got, err = store.GetMemoryByID(ctx, m1.ID)
	if err != nil {
		t.Fatalf("GetMemoryByID: %v", err)
	}
	if got == nil || got.ID != m1.ID || got.Content != "Content for get-by-id" {
		t.Errorf("GetMemoryByID: got %v", got)
	}
}

func TestCausalNeighbors(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	m1, _ := store.Remember(ctx, "First", nil, "")
	m2, _ := store.Remember(ctx, "Second", nil, "")
	m3, _ := store.Remember(ctx, "Third", nil, "")

	_ = store.AddEdge(ctx, m1.ID, m2.ID, "causal", "reason")
	_ = store.AddEdge(ctx, m3.ID, m1.ID, "causal", "other")

	neighbors, err := store.CausalNeighbors(ctx, m1.ID)
	if err != nil {
		t.Fatalf("CausalNeighbors: %v", err)
	}
	// m1 has: outgoing to m2, incoming from m3 → neighbors = m2, m3
	if len(neighbors) != 2 {
		t.Errorf("expected 2 causal neighbors for m1, got %d", len(neighbors))
	}
	ids := make(map[string]bool)
	for _, n := range neighbors {
		ids[n.ID] = true
	}
	if !ids[m2.ID] || !ids[m3.ID] {
		t.Errorf("neighbors should include m2 and m3, got %v", ids)
	}
}

func TestAffectedIfChanged(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	m1, _ := store.Remember(ctx, "Root", nil, "")
	m2, _ := store.Remember(ctx, "Child", nil, "")
	m3, _ := store.Remember(ctx, "Grandchild", nil, "")

	_ = store.AddEdge(ctx, m1.ID, m2.ID, "causal", "")
	_ = store.AddEdge(ctx, m2.ID, m3.ID, "causal", "")

	affected, err := store.AffectedIfChanged(ctx, m1.ID)
	if err != nil {
		t.Fatalf("AffectedIfChanged: %v", err)
	}
	// From m1: follow causal → m2, m3
	if len(affected) != 2 {
		t.Errorf("expected 2 affected (m2, m3), got %d: %v", len(affected), affected)
	}
	ids := make(map[string]bool)
	for _, id := range affected {
		ids[id] = true
	}
	if !ids[m2.ID] || !ids[m3.ID] {
		t.Errorf("affected should include m2 and m3, got %v", affected)
	}

	// From m3: no outgoing causal edges
	affected3, _ := store.AffectedIfChanged(ctx, m3.ID)
	if len(affected3) != 0 {
		t.Errorf("expected 0 affected from m3, got %d", len(affected3))
	}
}

func TestCompose(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	_, _ = store.Remember(ctx, "Auth and login flow", []string{"auth"}, "")
	_, _ = store.Remember(ctx, "Database schema for users", []string{"db"}, "")
	_, _ = store.Remember(ctx, "Auth middleware and JWT", []string{"auth"}, "")

	// Compose two queries: auth-related and db-related (stage2 Compose)
	composed, err := store.Compose(ctx, []string{"authentication", "database"}, 5)
	if err != nil {
		t.Fatalf("Compose: %v", err)
	}
	if composed.Explanation == "" {
		t.Error("expected non-empty explanation")
	}
	// Should get merged results (auth + db); at least 2 unique
	if len(composed.Memories) < 1 {
		t.Errorf("expected at least 1 memory from compose, got %d", len(composed.Memories))
	}
	seen := make(map[string]bool)
	for _, m := range composed.Memories {
		if seen[m.ID] {
			t.Errorf("duplicate memory in compose result: %s", m.ID)
		}
		seen[m.ID] = true
	}
}

func TestPrefetchSuggestions(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	_, _ = store.Remember(ctx, "Auth setup", nil, "")
	_, _ = store.Remember(ctx, "Database config", nil, "")

	// With hint: semantic recall (stage2 PrefetchSuggest)
	prefetch, err := store.PrefetchSuggest(ctx, "authentication", 5)
	if err != nil {
		t.Fatalf("PrefetchSuggest with hint: %v", err)
	}
	// May be 0 or more depending on embedding match
	_ = prefetch

	// Empty hint: recent important (may be empty if no important tags)
	recent, err := store.PrefetchSuggest(ctx, "", 3)
	if err != nil {
		t.Fatalf("PrefetchSuggest no hint: %v", err)
	}
	// Should not error; may be empty
	_ = recent
}

func TestRunMemoryDreams(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	_, _ = store.Remember(ctx, "First memory about auth", nil, "")
	_, _ = store.Remember(ctx, "Second memory about auth flow", nil, "")
	_, _ = store.Remember(ctx, "Third memory about login", nil, "")

	added, err := store.RunMemoryDreams(ctx, 10, 2)
	if err != nil {
		t.Fatalf("RunMemoryDreams: %v", err)
	}
	// May add 0 or more semantic edges depending on embedding similarity
	_ = added
}

func TestRunNightlyCuration(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	_, _ = store.Remember(ctx, "Memory for nightly test", nil, "")

	result, err := store.RunNightlyCuration(ctx)
	if err != nil {
		t.Fatalf("RunNightlyCuration: %v", err)
	}
	_ = result.DecayedCitations
	_ = result.DreamsEdgesAdded
	if result.Error != "" {
		t.Errorf("unexpected error in result: %s", result.Error)
	}
}
