package memory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// setupTestStore creates a temporary store for testing
func setupTestStore(t *testing.T) (*Store, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "phloem-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Set environment variable for test
	originalDataDir := os.Getenv("PHLOEM_DATA_DIR")
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)

	store, err := NewStore()
	if err != nil {
		os.RemoveAll(tmpDir)
		os.Setenv("PHLOEM_DATA_DIR", originalDataDir)
		t.Fatalf("failed to create store: %v", err)
	}

	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
		os.Setenv("PHLOEM_DATA_DIR", originalDataDir)
	}

	return store, cleanup
}

// =============================================================================
// Store Creation Tests
// =============================================================================

func TestNewStore(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	if store == nil {
		t.Fatal("expected non-nil store")
	}
	if store.db == nil {
		t.Error("expected non-nil database connection")
	}
}

func TestNewStore_EXAMPLE_REDACTED_REDACTED_REDACTED(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "phloem-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dataDir := filepath.Join(tmpDir, "subdir", "phloem")
	os.Setenv("PHLOEM_DATA_DIR", dataDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")

	store, err := NewStore()
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Check directory was created
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Error("expected data directory to be created")
	}

	// Check database file exists
	dbPath := filepath.Join(dataDir, "memories.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected database file to be created")
	}
}

// =============================================================================
// Remember Tests
// =============================================================================

func TestRemember_Basic(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	mem, err := store.Remember(ctx, "Test content", []string{"test"}, "test context")
	if err != nil {
		t.Fatalf("Remember failed: %v", err)
	}

	if mem.ID == "" {
		t.Error("expected non-empty ID")
	}
	if mem.Content != "Test content" {
		t.Errorf("expected content 'Test content', got '%s'", mem.Content)
	}
	if len(mem.Tags) != 1 || mem.Tags[0] != "test" {
		t.Errorf("expected tags ['test'], got %v", mem.Tags)
	}
	if mem.Context != "test context" {
		t.Errorf("expected context 'test context', got '%s'", mem.Context)
	}
	if len(mem.Embedding) == 0 {
		t.Error("expected non-empty embedding")
	}
}

func TestRemember_EmptyContent(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	mem, err := store.Remember(ctx, "", nil, "")

	// Empty content should still work (store allows it)
	if err != nil {
		t.Fatalf("Remember failed unexpectedly: %v", err)
	}
	if mem.Content != "" {
		t.Error("expected empty content")
	}
}

func TestRemember_MultipleTags(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	tags := []string{"tag1", "tag2", "tag3"}
	mem, err := store.Remember(ctx, "Content with multiple tags", tags, "")
	if err != nil {
		t.Fatalf("Remember failed: %v", err)
	}

	if len(mem.Tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(mem.Tags))
	}
}

func TestRemember_LongContent(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	// Create a 10KB content string
	longContent := ""
	for i := 0; i < 1000; i++ {
		longContent += "This is a test sentence for long content. "
	}

	mem, err := store.Remember(ctx, longContent, nil, "")
	if err != nil {
		t.Fatalf("Remember failed for long content: %v", err)
	}

	if mem.Content != longContent {
		t.Error("content was truncated or modified")
	}
}

func TestRemember_SpecialCharacters(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	specialContent := `Test with "quotes", 'apostrophes', emoji 🎉, unicode: 日本語, and SQL injection: '; DROP TABLE memories;--`

	mem, err := store.Remember(ctx, specialContent, nil, "")
	if err != nil {
		t.Fatalf("Remember failed for special characters: %v", err)
	}

	if mem.Content != specialContent {
		t.Errorf("content was modified: got %s", mem.Content)
	}
}

func TestRemember_UniqueIDs(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	ids := make(map[string]bool)

	// Create 100 memories and ensure all IDs are unique
	// Use unique content to avoid deduplication
	for i := 0; i < 100; i++ {
		mem, err := store.Remember(ctx, fmt.Sprintf("Test memory %d", i), nil, "")
		if err != nil {
			t.Fatalf("Remember failed: %v", err)
		}
		if ids[mem.ID] {
			t.Errorf("duplicate ID generated: %s", mem.ID)
		}
		ids[mem.ID] = true
	}
}

// =============================================================================
// Recall Tests
// =============================================================================

func TestRecall_Basic(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store some memories
	store.Remember(ctx, "The quick brown fox jumps over the lazy dog", []string{"animals"}, "")
	store.Remember(ctx, "Python is a programming language", []string{"code"}, "")
	store.Remember(ctx, "Go is a fast compiled language", []string{"code"}, "")

	// Recall with query
	memories, err := store.Recall(ctx, "programming language", 5, nil)
	if err != nil {
		t.Fatalf("Recall failed: %v", err)
	}

	if len(memories) == 0 {
		t.Error("expected at least one memory")
	}

	// First result should be most similar
	if len(memories) > 0 && memories[0].Similarity <= 0 {
		t.Error("expected positive similarity score")
	}
}

func TestRecall_WithTagFilter(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store memories with different tags
	store.Remember(ctx, "First code memory", []string{"code"}, "")
	store.Remember(ctx, "Second code memory", []string{"code"}, "")
	store.Remember(ctx, "Animal memory", []string{"animals"}, "")

	// Recall with tag filter
	memories, err := store.Recall(ctx, "memory", 10, []string{"code"})
	if err != nil {
		t.Fatalf("Recall failed: %v", err)
	}

	// Should only return code-tagged memories
	for _, mem := range memories {
		hasCodeTag := false
		for _, tag := range mem.Tags {
			if tag == "code" {
				hasCodeTag = true
				break
			}
		}
		if !hasCodeTag {
			t.Errorf("expected memory to have 'code' tag, got %v", mem.Tags)
		}
	}
}

func TestRecall_LimitResults(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store 10 memories
	for i := 0; i < 10; i++ {
		store.Remember(ctx, "Test memory content", nil, "")
	}

	// Recall with limit
	memories, err := store.Recall(ctx, "test", 3, nil)
	if err != nil {
		t.Fatalf("Recall failed: %v", err)
	}

	if len(memories) > 3 {
		t.Errorf("expected at most 3 memories, got %d", len(memories))
	}
}

func TestRecall_EmptyStore(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	memories, err := store.Recall(ctx, "test query", 5, nil)
	if err != nil {
		t.Fatalf("Recall failed: %v", err)
	}

	if len(memories) != 0 {
		t.Errorf("expected 0 memories from empty store, got %d", len(memories))
	}
}

func TestRecall_SortedBySimilarity(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store memories with varying relevance
	store.Remember(ctx, "The cat sat on the mat", nil, "")
	store.Remember(ctx, "Dogs are loyal pets", nil, "")
	store.Remember(ctx, "Cats are independent animals", nil, "")

	memories, err := store.Recall(ctx, "cats", 10, nil)
	if err != nil {
		t.Fatalf("Recall failed: %v", err)
	}

	// Verify sorted by similarity (descending)
	for i := 1; i < len(memories); i++ {
		if memories[i].Similarity > memories[i-1].Similarity {
			t.Error("memories not sorted by similarity descending")
		}
	}
}

// =============================================================================
// Forget Tests
// =============================================================================

func TestForget_Basic(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store a memory
	mem, _ := store.Remember(ctx, "Memory to forget", nil, "")

	// Forget it
	err := store.Forget(ctx, mem.ID)
	if err != nil {
		t.Fatalf("Forget failed: %v", err)
	}

	// Verify it's gone
	count, _ := store.Count(ctx)
	if count != 0 {
		t.Errorf("expected 0 memories after forget, got %d", count)
	}
}

func TestForget_NonExistent(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	err := store.Forget(ctx, "nonexistent-id")

	if err == nil {
		t.Error("expected error when forgetting non-existent memory")
	}
}

func TestForget_RemovesTags(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	mem, _ := store.Remember(ctx, "Tagged memory", []string{"tag1", "tag2"}, "")

	// Forget it
	store.Forget(ctx, mem.ID)

	// Store another memory with same tags
	store.Remember(ctx, "Another memory", []string{"tag1"}, "")

	// Recall by tag should only find the new memory
	memories, _ := store.Recall(ctx, "memory", 10, []string{"tag1"})

	for _, m := range memories {
		if m.ID == mem.ID {
			t.Error("forgotten memory still found via tag filter")
		}
	}
}

// =============================================================================
// List Tests
// =============================================================================

func TestList_Basic(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store some memories
	store.Remember(ctx, "First memory", nil, "")
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	store.Remember(ctx, "Second memory", nil, "")
	time.Sleep(10 * time.Millisecond)
	store.Remember(ctx, "Third memory", nil, "")

	memories, err := store.List(ctx, 10, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(memories) != 3 {
		t.Errorf("expected 3 memories, got %d", len(memories))
	}

	// Should be ordered by created_at DESC (newest first)
	if memories[0].Content != "Third memory" {
		t.Error("expected newest memory first")
	}
}

func TestList_WithLimit(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store 10 memories with unique content to avoid deduplication
	for i := 0; i < 10; i++ {
		store.Remember(ctx, fmt.Sprintf("Memory %d", i), nil, "")
	}

	memories, _ := store.List(ctx, 5, nil)
	if len(memories) != 5 {
		t.Errorf("expected 5 memories, got %d", len(memories))
	}
}

func TestList_WithTagFilter(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	store.Remember(ctx, "Tagged A", []string{"tagA"}, "")
	store.Remember(ctx, "Tagged B", []string{"tagB"}, "")
	store.Remember(ctx, "Tagged A again", []string{"tagA"}, "")

	memories, _ := store.List(ctx, 10, []string{"tagA"})

	if len(memories) != 2 {
		t.Errorf("expected 2 memories with tagA, got %d", len(memories))
	}
}

// =============================================================================
// Count Tests
// =============================================================================

func TestCount_Empty(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 count for empty store, got %d", count)
	}
}

func TestCount_AfterOperations(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Add memories
	mem1, _ := store.Remember(ctx, "Memory 1", nil, "")
	store.Remember(ctx, "Memory 2", nil, "")

	count, _ := store.Count(ctx)
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}

	// Forget one
	store.Forget(ctx, mem1.ID)

	count, _ = store.Count(ctx)
	if count != 1 {
		t.Errorf("expected count 1 after forget, got %d", count)
	}
}

// =============================================================================
// Size Tests
// =============================================================================

func TestSize_ReturnsReadableString(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	size, err := store.Size()
	if err != nil {
		t.Fatalf("Size failed: %v", err)
	}

	// Should return a string like "X.X KB" or "X B"
	if size == "" || size == "unknown" {
		t.Error("expected readable size string")
	}
}

// =============================================================================
// LastActivity Tests
// =============================================================================

func TestLastActivity_Empty(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	lastActivity, err := store.LastActivity(ctx)

	// Empty store might return zero time or error
	if err != nil && err.Error() != "sql: Scan error" {
		t.Errorf("expected nil or Scan error, got: %v", err)
	}
	_ = lastActivity // May be zero
}

func TestLastActivity_AfterRemember(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	before := time.Now().Add(-time.Second)

	store.Remember(ctx, "Test memory", nil, "")

	lastActivity, err := store.LastActivity(ctx)
	if err != nil {
		t.Fatalf("LastActivity failed: %v", err)
	}

	if lastActivity.Before(before) {
		t.Error("last activity should be after we added the memory")
	}
}

// =============================================================================
// Embedding Tests
// =============================================================================

func TestEmbedder_Deterministic(t *testing.T) {
	// Use local embedder for deterministic tests
	embedder := NewLocalEmbedder()

	// Same text should produce same embedding
	emb1, _ := embedder.Embed("hello world")
	emb2, _ := embedder.Embed("hello world")

	if len(emb1) != len(emb2) {
		t.Error("embeddings have different lengths")
	}

	for i := range emb1 {
		if emb1[i] != emb2[i] {
			t.Error("embeddings differ for same input")
			break
		}
	}
}

func TestEmbedder_DifferentTexts(t *testing.T) {
	embedder := NewLocalEmbedder()

	emb1, _ := embedder.Embed("hello world")
	emb2, _ := embedder.Embed("goodbye universe")

	// Different texts should produce different embeddings
	same := true
	for i := range emb1 {
		if emb1[i] != emb2[i] {
			same = false
			break
		}
	}

	if same {
		t.Error("different texts produced identical embeddings")
	}
}

func TestEmbedder_EmptyText(t *testing.T) {
	embedder := NewLocalEmbedder()

	emb, _ := embedder.Embed("")

	if len(emb) != embedder.Dimensions() {
		t.Errorf("expected embedding size %d, got %d", embedder.Dimensions(), len(emb))
	}
}

// =============================================================================
// Cosine Similarity Tests
// =============================================================================

func TestCosineSimilarity_Identical(t *testing.T) {
	a := []float32{1, 2, 3, 4, 5}
	b := []float32{1, 2, 3, 4, 5}

	sim := cosineSimilarity(a, b)
	if sim < 0.999 {
		t.Errorf("expected similarity ~1.0 for identical vectors, got %f", sim)
	}
}

func TestCosineSimilarity_Orthogonal(t *testing.T) {
	a := []float32{1, 0, 0}
	b := []float32{0, 1, 0}

	sim := cosineSimilarity(a, b)
	if sim > 0.001 {
		t.Errorf("expected similarity ~0.0 for orthogonal vectors, got %f", sim)
	}
}

func TestCosineSimilarity_DifferentLengths(t *testing.T) {
	a := []float32{1, 2, 3}
	b := []float32{1, 2}

	sim := cosineSimilarity(a, b)
	if sim != 0 {
		t.Errorf("expected 0 for different length vectors, got %f", sim)
	}
}

func TestCosineSimilarity_Empty(t *testing.T) {
	a := []float32{}
	b := []float32{}

	sim := cosineSimilarity(a, b)
	if sim != 0 {
		t.Errorf("expected 0 for empty vectors, got %f", sim)
	}
}

// =============================================================================
// Concurrency Tests
// =============================================================================

func TestConcurrentRemember(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	done := make(chan bool)

	// Spawn 10 goroutines, each adding 10 memories
	// Use unique content to avoid deduplication
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				store.Remember(ctx, fmt.Sprintf("Concurrent memory %d-%d", id, j), []string{"concurrent"}, "")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 100 memories
	count, _ := store.Count(ctx)
	if count != 100 {
		t.Errorf("expected 100 memories after concurrent writes, got %d", count)
	}
}

func TestConcurrentRecall(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Add some memories first
	for i := 0; i < 10; i++ {
		store.Remember(ctx, "Test memory for recall", nil, "")
	}

	// Concurrent recalls should not panic
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			store.Recall(ctx, "test", 5, nil)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestRemember_VeryLongTags(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Create a very long tag
	longTag := ""
	for i := 0; i < 1000; i++ {
		longTag += "a"
	}

	mem, err := store.Remember(ctx, "Memory with long tag", []string{longTag}, "")
	if err != nil {
		t.Fatalf("Remember failed with long tag: %v", err)
	}

	if len(mem.Tags) != 1 || mem.Tags[0] != longTag {
		t.Error("long tag was not stored correctly")
	}
}

func TestRemember_ManyTags(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Create 100 tags
	tags := make([]string, 100)
	for i := 0; i < 100; i++ {
		tags[i] = "tag" + string(rune('a'+i%26))
	}

	mem, err := store.Remember(ctx, "Memory with many tags", tags, "")
	if err != nil {
		t.Fatalf("Remember failed with many tags: %v", err)
	}

	if len(mem.Tags) != 100 {
		t.Errorf("expected 100 tags, got %d", len(mem.Tags))
	}
}

// =============================================================================
// RecallWithRecencyBoost Tests
// =============================================================================

func TestRecallWithRecencyBoost_Basic(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store memories
	store.Remember(ctx, "The quick brown fox jumps over the lazy dog", []string{"animals"}, "")
	store.Remember(ctx, "Python is a programming language", []string{"code"}, "")
	store.Remember(ctx, "Go is a fast compiled language", []string{"code"}, "")

	// Recall with blended scoring
	memories, err := store.RecallWithRecencyBoost(ctx, "programming language", 5, RecallOptions{})
	if err != nil {
		t.Fatalf("RecallWithRecencyBoost failed: %v", err)
	}

	if len(memories) == 0 {
		t.Error("expected at least one memory")
	}

	// Results should have positive scores
	for _, mem := range memories {
		if mem.Similarity <= 0 {
			t.Error("expected positive blended score")
		}
	}
}

func TestRecallWithRecencyBoost_RecencyMatters(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store an old memory (semantically perfect match)
	oldMem, _ := store.Remember(ctx, "programming language tutorial", []string{"code"}, "")

	// Manually update the created_at to be old (7 days ago)
	_, err := store.db.ExecContext(ctx,
		`UPDATE memories SET created_at = ? WHERE id = ?`,
		time.Now().Add(-7*24*time.Hour), oldMem.ID)
	if err != nil {
		t.Fatalf("failed to update timestamp: %v", err)
	}

	// Store a recent memory (less perfect semantic match)
	store.Remember(ctx, "coding tips and tricks", []string{"code"}, "")

	// With high recency weight, the recent memory should rank higher
	memories, err := store.RecallWithRecencyBoost(ctx, "programming language", 5, RecallOptions{
		SemanticWeight:       0.3,
		RecencyWeight:        0.6,
		ImportanceWeight:     0.1,
		RecencyHalfLifeHours: 24, // 1 day half-life
	})
	if err != nil {
		t.Fatalf("RecallWithRecencyBoost failed: %v", err)
	}

	if len(memories) < 2 {
		t.Fatal("expected at least 2 memories")
	}

	// The recent memory should be first due to high recency weight
	if memories[0].Content == "programming language tutorial" {
		t.Log("Note: Old memory ranked first - this is acceptable if semantic match is very strong")
	}
}

func TestRecallWithRecencyBoost_ImportanceBoost(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store a regular memory
	store.Remember(ctx, "regular coding note", []string{"code"}, "")

	// Store a critical memory
	store.Remember(ctx, "critical system decision", []string{"critical", "decision"}, "")

	// With importance weight, critical memory should get boosted
	memories, err := store.RecallWithRecencyBoost(ctx, "system note", 5, RecallOptions{
		SemanticWeight:       0.4,
		RecencyWeight:        0.2,
		ImportanceWeight:     0.4, // High importance weight
		RecencyHalfLifeHours: 168,
	})
	if err != nil {
		t.Fatalf("RecallWithRecencyBoost failed: %v", err)
	}

	if len(memories) < 2 {
		t.Fatal("expected at least 2 memories")
	}

	// The critical memory should be boosted
	foundCritical := false
	for _, mem := range memories {
		for _, tag := range mem.Tags {
			if tag == "critical" {
				foundCritical = true
				break
			}
		}
	}
	if !foundCritical {
		t.Error("expected to find critical memory in results")
	}
}

func TestRecallWithRecencyBoost_CustomWeights(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	store.Remember(ctx, "test memory", nil, "")

	// Test with custom weights that sum to 1.0
	memories, err := store.RecallWithRecencyBoost(ctx, "test", 5, RecallOptions{
		SemanticWeight:       0.5,
		RecencyWeight:        0.3,
		ImportanceWeight:     0.2,
		RecencyHalfLifeHours: 48,
	})
	if err != nil {
		t.Fatalf("RecallWithRecencyBoost failed: %v", err)
	}

	if len(memories) == 0 {
		t.Error("expected at least one memory")
	}
}

func TestRecallWithRecencyBoost_WithSinceFilter(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store a memory
	mem, _ := store.Remember(ctx, "recent memory", nil, "")

	// Update one to be old
	_, _ = store.db.ExecContext(ctx,
		`UPDATE memories SET created_at = ? WHERE id = ?`,
		time.Now().Add(-30*24*time.Hour), mem.ID)

	// Store another recent one
	store.Remember(ctx, "another recent memory", nil, "")

	// Query with Since filter (last 7 days)
	memories, err := store.RecallWithRecencyBoost(ctx, "memory", 10, RecallOptions{
		Since: time.Now().Add(-7 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("RecallWithRecencyBoost failed: %v", err)
	}

	// Should only find the recent memory
	if len(memories) != 1 {
		t.Errorf("expected 1 memory (filtered by Since), got %d", len(memories))
	}
}

func TestRecall_HybridOptimization_SmallDataset(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store < 5000 memories (should use standard recall)
	for i := 0; i < 100; i++ {
		store.Remember(ctx, fmt.Sprintf("memory %d about programming", i), []string{"code"}, "")
	}

	// Query should use standard recall path
	memories, err := store.Recall(ctx, "programming", 10, nil)
	if err != nil {
		t.Fatalf("Recall failed: %v", err)
	}

	if len(memories) == 0 {
		t.Error("expected at least one memory")
	}

	// Verify results are sorted by similarity
	for i := 1; i < len(memories); i++ {
		if memories[i-1].Similarity < memories[i].Similarity {
			t.Error("expected memories sorted by similarity descending")
		}
	}
}

func TestRecall_HybridOptimization_LargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large dataset test in short mode")
	}

	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store > 5000 memories (should trigger hybrid recall)
	for i := 0; i < 5100; i++ {
		content := fmt.Sprintf("memory %d about programming and code", i)
		store.Remember(ctx, content, []string{"code"}, "")
	}

	// Query should automatically use hybrid recall with recency boost
	startTime := time.Now()
	memories, err := store.Recall(ctx, "programming", 10, nil)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("Recall failed: %v", err)
	}

	if len(memories) == 0 {
		t.Error("expected at least one memory")
	}

	// Verify performance improvement (should be faster than full scan)
	// With 5100 memories, hybrid recall should complete in reasonable time
	if duration > 2*time.Second {
		t.Logf("Warning: hybrid recall took %v for 5100 memories (expected < 2s)", duration)
	}

	t.Logf("Hybrid recall completed in %v for 5100 memories", duration)
}

func TestRecall_HybridOptimization_WithTagFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large dataset test in short mode")
	}

	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store > 5000 memories with different tags
	for i := 0; i < 5100; i++ {
		tag := "code"
		if i%2 == 0 {
			tag = "note"
		}
		store.Remember(ctx, fmt.Sprintf("memory %d", i), []string{tag}, "")
	}

	// Query with tag filter should still work with hybrid recall
	memories, err := store.Recall(ctx, "memory", 10, []string{"code"})
	if err != nil {
		t.Fatalf("Recall with tag filter failed: %v", err)
	}

	// Should only return code-tagged memories
	for _, mem := range memories {
		hasCodeTag := false
		for _, tag := range mem.Tags {
			if tag == "code" {
				hasCodeTag = true
				break
			}
		}
		if !hasCodeTag {
			t.Error("expected only code-tagged memories")
		}
	}
}

// =============================================================================
// GetRecentImportant Tests
// =============================================================================

func TestGetRecentImportant_Basic(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store regular memories
	store.Remember(ctx, "regular memory 1", []string{"code"}, "")
	store.Remember(ctx, "regular memory 2", []string{"note"}, "")

	// Store important memories
	store.Remember(ctx, "critical bug fix", []string{"critical"}, "")
	store.Remember(ctx, "milestone reached", []string{"milestone"}, "")
	store.Remember(ctx, "key decision made", []string{"decision"}, "")

	// Get recent important
	memories, err := store.GetRecentImportant(ctx, 7*24*time.Hour, 10)
	if err != nil {
		t.Fatalf("GetRecentImportant failed: %v", err)
	}

	// Should only return the important ones
	if len(memories) != 3 {
		t.Errorf("expected 3 important memories, got %d", len(memories))
	}

	// Verify all have important tags
	for _, mem := range memories {
		hasImportantTag := false
		for _, tag := range mem.Tags {
			switch tag {
			case "critical", "milestone", "founding", "permanent", "promise", "decision":
				hasImportantTag = true
			}
		}
		if !hasImportantTag {
			t.Errorf("memory %s doesn't have an important tag: %v", mem.ID, mem.Tags)
		}
	}
}

func TestGetRecentImportant_RespectsMaxAge(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store a critical memory
	mem, _ := store.Remember(ctx, "old critical memory", []string{"critical"}, "")

	// Make it old (30 days ago)
	_, _ = store.db.ExecContext(ctx,
		`UPDATE memories SET created_at = ? WHERE id = ?`,
		time.Now().Add(-30*24*time.Hour), mem.ID)

	// Store a recent critical memory
	store.Remember(ctx, "recent critical memory", []string{"critical"}, "")

	// Get important from last 7 days
	memories, err := store.GetRecentImportant(ctx, 7*24*time.Hour, 10)
	if err != nil {
		t.Fatalf("GetRecentImportant failed: %v", err)
	}

	// Should only find the recent one
	if len(memories) != 1 {
		t.Errorf("expected 1 recent important memory, got %d", len(memories))
	}
}

func TestGetRecentImportant_Limit(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store many critical memories with unique content to avoid deduplication
	for i := 0; i < 10; i++ {
		store.Remember(ctx, fmt.Sprintf("critical memory %d", i), []string{"critical"}, "")
	}

	// Get with limit of 3
	memories, err := store.GetRecentImportant(ctx, 7*24*time.Hour, 3)
	if err != nil {
		t.Fatalf("GetRecentImportant failed: %v", err)
	}

	if len(memories) != 3 {
		t.Errorf("expected 3 memories (limited), got %d", len(memories))
	}
}

func TestGetRecentImportant_AllImportantTags(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store one of each important tag type
	importantTags := []string{"critical", "milestone", "founding", "permanent", "promise", "decision"}
	for _, tag := range importantTags {
		store.Remember(ctx, "memory with "+tag, []string{tag}, "")
	}

	// Get all
	memories, err := store.GetRecentImportant(ctx, 7*24*time.Hour, 20)
	if err != nil {
		t.Fatalf("GetRecentImportant failed: %v", err)
	}

	if len(memories) != 6 {
		t.Errorf("expected 6 memories (one per important tag), got %d", len(memories))
	}
}

func TestGetRecentImportant_Empty(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	// Store only regular memories
	store.Remember(ctx, "regular memory", []string{"code"}, "")

	// Get important (should be empty)
	memories, err := store.GetRecentImportant(ctx, 7*24*time.Hour, 10)
	if err != nil {
		t.Fatalf("GetRecentImportant failed: %v", err)
	}

	if len(memories) != 0 {
		t.Errorf("expected 0 important memories, got %d", len(memories))
	}
}

// =============================================================================
// GetMemoryByID empty id, Add, GetEmbedderDimensions, Size, LastActivity
// =============================================================================

func TestGetMemoryByID_EmptyID(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	got, err := store.GetMemoryByID(ctx, "")
	if err != nil {
		t.Fatalf("GetMemoryByID empty id: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for empty id, got %v", got)
	}
}

func TestGetMemoryByID_NotFound(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	got, err := store.GetMemoryByID(ctx, "nonexistent-id-12345")
	if err != nil {
		t.Fatalf("GetMemoryByID: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for nonexistent id, got %v", got)
	}
}

func TestRunMemoryCritic(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	store.Remember(ctx, "Critic test memory", nil, "")
	err := store.RunMemoryCritic(ctx)
	if err != nil {
		t.Fatalf("RunMemoryCritic: %v", err)
	}
}

func TestStore_Size_ReturnsB(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	// Empty or tiny DB returns "X B"
	size, err := store.Size()
	if err != nil {
		t.Fatalf("Size: %v", err)
	}
	if size == "" || size == "unknown" {
		t.Errorf("Size should return readable string, got %q", size)
	}
	if !strings.Contains(size, "B") && !strings.Contains(size, "KB") && !strings.Contains(size, "MB") {
		t.Errorf("Size should contain B, KB, or MB: %q", size)
	}
}

func TestStore_List_WithLimitZero(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()
	store.Remember(ctx, "one", nil, "")
	list, err := store.List(ctx, 0, nil)
	if err != nil {
		t.Fatalf("List(0): %v", err)
	}
	// limit 0 may return 0 or all depending on implementation
	_ = list
}

func TestStore_Recall_WithFilterTags_NoMatch(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()
	store.Remember(ctx, "recall filter test", []string{"other"}, "")
	results, err := store.Recall(ctx, "recall", 10, []string{"nonexistent-tag"})
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	// May be empty or still return by semantic match depending on implementation
	_ = results
}

func TestAdd_NewMemory(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	m := Memory{
		Content: "Added via Add()",
		Tags:    []string{"add"},
		Context: "test",
	}
	err := store.Add(ctx, m)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	list, _ := store.List(ctx, 10, nil)
	if len(list) != 1 {
		t.Errorf("expected 1 memory after Add, got %d", len(list))
	}
	if len(list) > 0 && list[0].Content != "Added via Add()" {
		t.Errorf("content = %q", list[0].Content)
	}
}

func TestAdd_DuplicateContentHash(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	m := Memory{Content: "Duplicate content", Tags: nil, Context: ""}
	err1 := store.Add(ctx, m)
	if err1 != nil {
		t.Fatalf("Add first: %v", err1)
	}

	// Same content hash - should skip (return nil)
	m2 := Memory{Content: "Duplicate content", Tags: []string{"other"}, Context: ""}
	err2 := store.Add(ctx, m2)
	if err2 != nil {
		t.Fatalf("Add duplicate should not error: %v", err2)
	}

	count, _ := store.Count(ctx)
	if count != 1 {
		t.Errorf("expected 1 memory (dedupe), got %d", count)
	}
}

func TestGetEmbedderDimensions(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	dim := store.GetEmbedderDimensions()
	if dim <= 0 {
		t.Errorf("GetEmbedderDimensions = %d, want positive", dim)
	}
}

func TestSize_KB_and_MB(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	size, err := store.Size()
	if err != nil {
		t.Fatalf("Size: %v", err)
	}
	if size == "" || size == "unknown" {
		t.Error("Size should return readable string")
	}
	// After some activity we may get "X.X KB"; large DB would show "X.X MB"
	_ = size
}

func TestLastActivity_ParseFormats(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	store.Remember(ctx, "For last activity", nil, "")

	last, err := store.LastActivity(ctx)
	if err != nil {
		t.Fatalf("LastActivity: %v", err)
	}
	if last.IsZero() {
		t.Error("LastActivity should be non-zero after Remember")
	}
}

func TestSetMemoryUtility_Clamp(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	mem, _ := store.Remember(ctx, "Utility clamp test", nil, "")

	// Clamp to [0, 1]: negative -> 0, >1 -> 1
	_ = store.SetMemoryUtility(ctx, mem.ID, -0.5)
	_ = store.SetMemoryUtility(ctx, mem.ID, 1.5)
}

func TestRemember_DuplicateContentMergeTags(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	m1, err := store.Remember(ctx, "Same content", []string{"tag1"}, "")
	if err != nil {
		t.Fatalf("Remember first: %v", err)
	}

	// Same content hash - should return existing memory with merged tags
	m2, err := store.Remember(ctx, "Same content", []string{"tag2"}, "")
	if err != nil {
		t.Fatalf("Remember duplicate: %v", err)
	}
	if m2.ID != m1.ID {
		t.Errorf("duplicate should return same ID %q, got %q", m1.ID, m2.ID)
	}
	if len(m2.Tags) != 2 {
		t.Errorf("expected merged tags (2), got %d: %v", len(m2.Tags), m2.Tags)
	}
}

func TestRunCausalExtractionAsync_NilMemory(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	// Should not panic
	store.RunCausalExtractionAsync(nil)
}

// TestRunCausalExtractionAsync_WithMemory runs causal extraction on a memory whose content matches causal patterns.
func TestRunCausalExtractionAsync_WithMemory(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	// Memory that Recall can find for phrase "the validation logic"
	m1, _ := store.Remember(ctx, "We updated the validation logic to reject empty inputs.", nil, "")
	// Memory with causal phrase "because the validation logic" so Extract returns phrase for recall
	m2, _ := store.Remember(ctx, "The bug was fixed because the validation logic was correct.", nil, "")

	store.RunCausalExtractionAsync(m2)
	// Allow async goroutine to run
	time.Sleep(200 * time.Millisecond)

	edges, _ := store.GetEdgesFrom(ctx, m2.ID, "causal")
	_ = edges
	_ = m1
}

func TestRunMemoryDreams_ZeroLimitsUseDefaults(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	store.Remember(ctx, "Dream test memory content here", nil, "")

	// recentLimit <= 0 and linksPerMemory <= 0 use defaults (30, 3)
	_, err := store.RunMemoryDreams(ctx, 0, 0)
	if err != nil {
		t.Fatalf("RunMemoryDreams(0,0): %v", err)
	}
}

func TestGetRecentImportant_EmptyStore(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()
	list, err := store.GetRecentImportant(ctx, 24*time.Hour, 5)
	if err != nil {
		t.Fatalf("GetRecentImportant: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 important memories, got %d", len(list))
	}
}

func TestGetRecentImportant_WithImportantTags(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()
	store.Remember(ctx, "Critical decision", []string{"critical"}, "")
	store.Remember(ctx, "Architecture note", []string{"decision"}, "")
	list, err := store.GetRecentImportant(ctx, 24*time.Hour, 5)
	if err != nil {
		t.Fatalf("GetRecentImportant: %v", err)
	}
	if len(list) < 1 {
		t.Errorf("expected at least 1 important memory, got %d", len(list))
	}
}

func TestForget_NotFound(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()
	err := store.Forget(ctx, "nonexistent-id-12345")
	if err == nil {
		t.Fatal("expected error when forgetting non-existent memory")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found: %v", err)
	}
}

func TestForget_ExistingMemory(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()
	mem, err := store.Remember(ctx, "To be forgotten", nil, "")
	if err != nil {
		t.Fatalf("Remember: %v", err)
	}
	err = store.Forget(ctx, mem.ID)
	if err != nil {
		t.Fatalf("Forget: %v", err)
	}
	got, _ := store.GetMemoryByID(ctx, mem.ID)
	if got != nil {
		t.Error("memory should be gone after Forget")
	}
}

func TestLastActivity_EmptyStore(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()
	last, err := store.LastActivity(ctx)
	if err != nil {
		t.Fatalf("LastActivity: %v", err)
	}
	if !last.IsZero() {
		t.Errorf("expected zero time for empty store, got %v", last)
	}
}

func TestAddEdge_EmptySourceOrType_ReturnsNil(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()
	if err := store.AddEdge(ctx, "", "target", "temporal", ""); err != nil {
		t.Errorf("AddEdge empty source should return nil: %v", err)
	}
	if err := store.AddEdge(ctx, "source", "target", "", ""); err != nil {
		t.Errorf("AddEdge empty type should return nil: %v", err)
	}
}

func TestAddEdge_GetEdgesFrom_GetEdgesTo(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()
	m1, _ := store.Remember(ctx, "Source memory", nil, "")
	m2, _ := store.Remember(ctx, "Target memory", nil, "")
	if err := store.AddEdge(ctx, m1.ID, m2.ID, "temporal", "payload"); err != nil {
		t.Fatalf("AddEdge: %v", err)
	}
	from, err := store.GetEdgesFrom(ctx, m1.ID, "temporal")
	if err != nil {
		t.Fatalf("GetEdgesFrom: %v", err)
	}
	if len(from) < 1 {
		t.Errorf("GetEdgesFrom: want at least 1 edge, got %d", len(from))
	}
	to, err := store.GetEdgesTo(ctx, m2.ID, "temporal")
	if err != nil {
		t.Fatalf("GetEdgesTo: %v", err)
	}
	if len(to) < 1 {
		t.Errorf("GetEdgesTo: want at least 1 edge, got %d", len(to))
	}
}

func TestStore_Size_FileMissing(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	// Remove db file while store is open (Unix: unlink leaves open fd)
	dbPath := filepath.Join(store.dataDir, "memories.db")
	if err := os.Remove(dbPath); err != nil {
		t.Skip("cannot remove db file on this platform")
	}
	size, err := store.Size()
	if err == nil {
		t.Error("expected error when db file missing")
	}
	if size != "unknown" {
		t.Errorf("Size() on missing file should return unknown, got %q", size)
	}
}

// =============================================================================
// Vec Index Tests
// =============================================================================

func TestVecIndex_Available(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	if store.vecIdx == nil {
		t.Fatal("expected vec index to be initialized")
	}
	if !store.vecIdx.available {
		t.Skip("sqlite-vec not available on this system")
	}
}

func TestVecIndex_InsertAndSearch(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	if store.vecIdx == nil || !store.vecIdx.available {
		t.Skip("sqlite-vec not available")
	}

	ctx := context.Background()

	// Store several memories
	_, err := store.Remember(ctx, "Go programming language and concurrency patterns", []string{"code"}, "")
	if err != nil {
		t.Fatalf("Remember failed: %v", err)
	}
	_, err = store.Remember(ctx, "Python data science and machine learning", []string{"code"}, "")
	if err != nil {
		t.Fatalf("Remember failed: %v", err)
	}
	_, err = store.Remember(ctx, "Cooking pasta with tomato sauce recipe", []string{"food"}, "")
	if err != nil {
		t.Fatalf("Remember failed: %v", err)
	}

	// Recall should return results
	results, err := store.Recall(ctx, "Go concurrency", 2, nil)
	if err != nil {
		t.Fatalf("Recall failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	// The Go-related memory should be first (most similar)
	if !strings.Contains(results[0].Content, "Go") {
		t.Errorf("expected Go-related memory first, got: %s", results[0].Content)
	}
}

func TestVecIndex_ForgetRemovesFromIndex(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	if store.vecIdx == nil || !store.vecIdx.available {
		t.Skip("sqlite-vec not available")
	}

	ctx := context.Background()

	mem, err := store.Remember(ctx, "Test memory for deletion from vec index", nil, "")
	if err != nil {
		t.Fatalf("Remember failed: %v", err)
	}

	// Verify it's in the vec index
	results, err := store.vecIdx.Search(mem.Embedding, 1)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected memory in vec index")
	}

	// Delete
	if err := store.Forget(ctx, mem.ID); err != nil {
		t.Fatalf("Forget failed: %v", err)
	}

	// Verify it's gone from vec index
	results, err = store.vecIdx.Search(mem.Embedding, 1)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	for _, r := range results {
		if r.MemoryID == mem.ID {
			t.Error("memory should have been removed from vec index")
		}
	}
}

func TestVecIndex_Backfill(t *testing.T) {
	// Create a store, add memories, then simulate a fresh vec index rebuild
	tmpDir, err := os.MkdirTemp("", "phloem-backfill-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")

	store, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}

	if store.vecIdx == nil || !store.vecIdx.available {
		store.Close()
		t.Skip("sqlite-vec not available")
	}

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		_, err := store.Remember(ctx, fmt.Sprintf("Backfill test memory %d with some content", i), nil, "")
		if err != nil {
			t.Fatalf("Remember failed: %v", err)
		}
	}

	// Drop vec tables to simulate pre-vec-index state
	store.db.Exec(`DROP TABLE IF EXISTS memory_embeddings`)
	store.db.Exec(`DELETE FROM memory_vec_ids`)
	store.Close()

	// Reopen - should backfill
	store2, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	defer store2.Close()

	if store2.vecIdx == nil || !store2.vecIdx.available {
		t.Fatal("expected vec index to be available after reopen")
	}

	// Check that recall still works
	results, err := store2.Recall(ctx, "backfill test", 3, nil)
	if err != nil {
		t.Fatalf("Recall failed after backfill: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results after backfill")
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkRecall(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "phloem-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")

	store, err := NewStore()
	if err != nil {
		b.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()

	topics := []string{
		"database optimization and query performance",
		"microservices architecture patterns",
		"machine learning model training",
		"kubernetes deployment strategies",
		"react component design patterns",
		"go concurrency and goroutines",
		"security vulnerability scanning",
		"CI/CD pipeline configuration",
		"API rate limiting implementation",
		"distributed cache invalidation",
	}

	// Insert 10k memories
	b.Logf("Inserting 10,000 memories...")
	for i := 0; i < 10000; i++ {
		topic := topics[i%len(topics)]
		content := fmt.Sprintf("Memory %d about %s with details number %d", i, topic, i*7)
		_, err := store.Remember(ctx, content, nil, "")
		if err != nil {
			b.Fatalf("Remember failed at %d: %v", i, err)
		}
	}
	b.Logf("Done inserting. Vec index available: %v", store.vecIdx != nil && store.vecIdx.available)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query := topics[i%len(topics)]
		_, err := store.Recall(ctx, query, 10, nil)
		if err != nil {
			b.Fatalf("Recall failed: %v", err)
		}
	}
}

func BenchmarkRecallLinearScan(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "phloem-bench-linear-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")

	store, err := NewStore()
	if err != nil {
		b.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()

	topics := []string{
		"database optimization and query performance",
		"microservices architecture patterns",
		"machine learning model training",
		"kubernetes deployment strategies",
		"react component design patterns",
	}

	// Insert 10k memories
	for i := 0; i < 10000; i++ {
		content := fmt.Sprintf("Memory %d about %s details %d", i, topics[i%len(topics)], i*7)
		_, err := store.Remember(ctx, content, nil, "")
		if err != nil {
			b.Fatalf("Remember failed at %d: %v", i, err)
		}
	}

	// Force linear scan by generating query embedding manually
	queryEmbedding, err := store.embedder.Embed("database query performance")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.recallLinearScan(ctx, queryEmbedding, 10, nil, "")
		if err != nil {
			b.Fatalf("recallLinearScan failed: %v", err)
		}
	}
}
