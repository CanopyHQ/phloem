// Package memory provides tests for the Citation system
package memory

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// --- Citation Struct Tests ---

func TestCitation_Structure(t *testing.T) {
	now := time.Now()
	c := &Citation{
		ID:         "cit-123",
		MemoryID:   "mem-456",
		FilePath:   "/path/to/file.go",
		StartLine:  10,
		EndLine:    20,
		CommitSHA:  "abc123def",
		Content:    "func main() {}",
		Confidence: 0.95,
		VerifiedAt: now,
		CreatedAt:  now,
	}

	if c.ID != "cit-123" {
		t.Errorf("ID mismatch: got %q", c.ID)
	}
	if c.MemoryID != "mem-456" {
		t.Errorf("MemoryID mismatch: got %q", c.MemoryID)
	}
	if c.FilePath != "/path/to/file.go" {
		t.Errorf("FilePath mismatch: got %q", c.FilePath)
	}
	if c.StartLine != 10 {
		t.Errorf("StartLine mismatch: got %d", c.StartLine)
	}
	if c.EndLine != 20 {
		t.Errorf("EndLine mismatch: got %d", c.EndLine)
	}
	if c.CommitSHA != "abc123def" {
		t.Errorf("CommitSHA mismatch: got %q", c.CommitSHA)
	}
	if c.Content != "func main() {}" {
		t.Errorf("Content mismatch: got %q", c.Content)
	}
	if c.Confidence != 0.95 {
		t.Errorf("Confidence mismatch: got %f", c.Confidence)
	}
}

func TestCitation_OptionalFields(t *testing.T) {
	c := &Citation{
		ID:        "cit-123",
		MemoryID:  "mem-456",
		FilePath:  "/path/to/file.go",
		StartLine: 1,
		EndLine:   5,
		// CommitSHA and Content are optional
	}

	if c.CommitSHA != "" {
		t.Error("CommitSHA should be empty by default")
	}
	if c.Content != "" {
		t.Error("Content should be empty by default")
	}
}

// --- String Similarity Tests ---

func TestStringSimilarity_Identical(t *testing.T) {
	sim := stringSimilarity("hello world", "hello world")
	if sim != 1.0 {
		t.Errorf("Expected 1.0 for identical strings, got %f", sim)
	}
}

func TestStringSimilarity_Empty(t *testing.T) {
	sim := stringSimilarity("", "hello")
	if sim != 0.0 {
		t.Errorf("Expected 0.0 for empty string, got %f", sim)
	}

	sim = stringSimilarity("hello", "")
	if sim != 0.0 {
		t.Errorf("Expected 0.0 for empty string, got %f", sim)
	}
}

func TestStringSimilarity_BothEmpty(t *testing.T) {
	sim := stringSimilarity("", "")
	// Empty strings are equal
	if sim != 1.0 {
		t.Errorf("Expected 1.0 for both empty strings, got %f", sim)
	}
}

func TestStringSimilarity_NoOverlap(t *testing.T) {
	sim := stringSimilarity("hello world", "foo bar baz")
	if sim != 0.0 {
		t.Errorf("Expected 0.0 for no overlap, got %f", sim)
	}
}

func TestStringSimilarity_PartialOverlap(t *testing.T) {
	sim := stringSimilarity("hello world foo", "hello world bar")
	// 2 out of 3 words match
	if sim < 0.5 || sim > 0.8 {
		t.Errorf("Expected ~0.67 for partial overlap, got %f", sim)
	}
}

func TestStringSimilarity_CaseInsensitive(t *testing.T) {
	sim := stringSimilarity("Hello World", "hello world")
	if sim != 1.0 {
		t.Errorf("Expected 1.0 for case-insensitive match, got %f", sim)
	}
}

func TestStringSimilarity_DifferentLengths(t *testing.T) {
	sim := stringSimilarity("hello", "hello world foo bar baz")
	// 1 word matches out of 5
	if sim < 0.1 || sim > 0.3 {
		t.Errorf("Expected low similarity for different lengths, got %f", sim)
	}
}

// --- Citation Store Integration Tests ---

func TestStore_AddCitation(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	// First create a memory to cite
	mem, err := store.Remember(ctx, "Test memory content", []string{"test"}, "")
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	// Add a citation
	citation, err := store.AddCitation(ctx, mem.ID, "/path/to/file.go", 10, 20, "abc123", "func main() {}")
	if err != nil {
		t.Fatalf("Failed to add citation: %v", err)
	}

	if citation.ID == "" {
		t.Error("Citation ID should not be empty")
	}
	if citation.MemoryID != mem.ID {
		t.Errorf("MemoryID mismatch: got %q, want %q", citation.MemoryID, mem.ID)
	}
	if citation.FilePath != "/path/to/file.go" {
		t.Errorf("FilePath mismatch: got %q", citation.FilePath)
	}
	if citation.StartLine != 10 {
		t.Errorf("StartLine mismatch: got %d", citation.StartLine)
	}
	if citation.EndLine != 20 {
		t.Errorf("EndLine mismatch: got %d", citation.EndLine)
	}
	if citation.CommitSHA != "abc123" {
		t.Errorf("CommitSHA mismatch: got %q", citation.CommitSHA)
	}
	if citation.Content != "func main() {}" {
		t.Errorf("Content mismatch: got %q", citation.Content)
	}
	if citation.Confidence != 1.0 {
		t.Errorf("Initial confidence should be 1.0, got %f", citation.Confidence)
	}
}

func TestStore_AddCitation_NoContent(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	mem, err := store.Remember(ctx, "Test memory", []string{}, "")
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	// Add citation without content
	citation, err := store.AddCitation(ctx, mem.ID, "/path/to/file.go", 1, 5, "", "")
	if err != nil {
		t.Fatalf("Failed to add citation: %v", err)
	}

	if citation.Content != "" {
		t.Error("Content should be empty")
	}
	if citation.CommitSHA != "" {
		t.Error("CommitSHA should be empty")
	}
}

func TestStore_GetCitations(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	mem, err := store.Remember(ctx, "Test memory", []string{}, "")
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	// Add multiple citations
	_, err = store.AddCitation(ctx, mem.ID, "/path/to/file1.go", 1, 10, "", "content1")
	if err != nil {
		t.Fatalf("Failed to add citation 1: %v", err)
	}

	_, err = store.AddCitation(ctx, mem.ID, "/path/to/file2.go", 20, 30, "", "content2")
	if err != nil {
		t.Fatalf("Failed to add citation 2: %v", err)
	}

	// Get citations
	citations, err := store.GetCitations(ctx, mem.ID)
	if err != nil {
		t.Fatalf("Failed to get citations: %v", err)
	}

	if len(citations) != 2 {
		t.Errorf("Expected 2 citations, got %d", len(citations))
	}
}

func TestStore_GetCitations_Empty(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	mem, err := store.Remember(ctx, "Test memory", []string{}, "")
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	// Get citations for memory with none
	citations, err := store.GetCitations(ctx, mem.ID)
	if err != nil {
		t.Fatalf("Failed to get citations: %v", err)
	}

	if len(citations) != 0 {
		t.Errorf("Expected 0 citations, got %d", len(citations))
	}
}

func TestStore_GetCitations_NonexistentMemory(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	citations, err := store.GetCitations(ctx, "nonexistent-memory-id")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(citations) != 0 {
		t.Errorf("Expected 0 citations for nonexistent memory, got %d", len(citations))
	}
}

func TestStore_VerifyCitation_FileExists(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	// Create a temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	content := "line1\nline2\nline3\nline4\nline5"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	mem, err := store.Remember(ctx, "Test memory", []string{}, "")
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	// Add citation with content matching the file
	citation, err := store.AddCitation(ctx, mem.ID, tmpFile, 2, 4, "", "line2\nline3\nline4")
	if err != nil {
		t.Fatalf("Failed to add citation: %v", err)
	}

	// Verify citation
	verified, valid, err := store.VerifyCitation(ctx, citation.ID)
	if err != nil {
		t.Fatalf("Failed to verify citation: %v", err)
	}

	if !valid {
		t.Error("Citation should be valid")
	}
	if verified.Confidence != 1.0 {
		t.Errorf("Confidence should be 1.0, got %f", verified.Confidence)
	}
}

func TestStore_VerifyCitation_FileNotFound(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	mem, err := store.Remember(ctx, "Test memory", []string{}, "")
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	// Add citation to nonexistent file
	citation, err := store.AddCitation(ctx, mem.ID, "/nonexistent/file.go", 1, 5, "", "content")
	if err != nil {
		t.Fatalf("Failed to add citation: %v", err)
	}

	// Verify citation
	verified, valid, err := store.VerifyCitation(ctx, citation.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if valid {
		t.Error("Citation should be invalid for nonexistent file")
	}
	if verified.Confidence != 0.0 {
		t.Errorf("Confidence should be 0.0, got %f", verified.Confidence)
	}
}

func TestStore_VerifyCitation_ContentChanged(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	// Create a temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	content := "line1\nline2\nline3\nline4\nline5"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	mem, err := store.Remember(ctx, "Test memory", []string{}, "")
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	// Add citation with different content
	citation, err := store.AddCitation(ctx, mem.ID, tmpFile, 2, 4, "", "different\ncontent\nhere")
	if err != nil {
		t.Fatalf("Failed to add citation: %v", err)
	}

	// Verify citation
	verified, _, err := store.VerifyCitation(ctx, citation.ID)
	if err != nil {
		t.Fatalf("Failed to verify citation: %v", err)
	}

	// Confidence should be reduced but not zero (some similarity)
	if verified.Confidence >= 1.0 {
		t.Errorf("Confidence should be reduced for changed content, got %f", verified.Confidence)
	}
}

func TestStore_VerifyCitation_NonexistentCitation(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	_, _, err := store.VerifyCitation(ctx, "nonexistent-citation-id")
	if err == nil {
		t.Error("Expected error for nonexistent citation")
	}
}

func TestStore_VerifyCitation_NoContent(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	// Create a temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	content := "line1\nline2\nline3\nline4\nline5"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	mem, err := store.Remember(ctx, "Test memory", []string{}, "")
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	// Add citation without content
	citation, err := store.AddCitation(ctx, mem.ID, tmpFile, 2, 4, "", "")
	if err != nil {
		t.Fatalf("Failed to add citation: %v", err)
	}

	// Verify citation
	verified, valid, err := store.VerifyCitation(ctx, citation.ID)
	if err != nil {
		t.Fatalf("Failed to verify citation: %v", err)
	}

	// Should be valid with 0.9 confidence (file exists, lines in range)
	if !valid {
		t.Error("Citation should be valid when file exists")
	}
	if verified.Confidence != 0.9 {
		t.Errorf("Confidence should be 0.9 for no-content citation, got %f", verified.Confidence)
	}
}

func TestStore_GetMemoryConfidence(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	mem, err := store.Remember(ctx, "Test memory", []string{}, "")
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	// No citations - should return 1.0
	conf, err := store.GetMemoryConfidence(ctx, mem.ID)
	if err != nil {
		t.Fatalf("Failed to get confidence: %v", err)
	}
	if conf != 1.0 {
		t.Errorf("Expected 1.0 for no citations, got %f", conf)
	}
}

func TestStore_GetMemoryConfidence_WithCitations(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	mem, err := store.Remember(ctx, "Test memory", []string{}, "")
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	// Add citations (they start with confidence 1.0)
	_, err = store.AddCitation(ctx, mem.ID, "/file1.go", 1, 5, "", "")
	if err != nil {
		t.Fatalf("Failed to add citation: %v", err)
	}

	_, err = store.AddCitation(ctx, mem.ID, "/file2.go", 1, 5, "", "")
	if err != nil {
		t.Fatalf("Failed to add citation: %v", err)
	}

	// Both citations have 1.0 confidence
	conf, err := store.GetMemoryConfidence(ctx, mem.ID)
	if err != nil {
		t.Fatalf("Failed to get confidence: %v", err)
	}
	if conf != 1.0 {
		t.Errorf("Expected 1.0 for new citations, got %f", conf)
	}
}

func TestStore_DecayCitations(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	mem, err := store.Remember(ctx, "Test memory", []string{}, "")
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	// Add a citation
	_, err = store.AddCitation(ctx, mem.ID, "/file.go", 1, 5, "", "")
	if err != nil {
		t.Fatalf("Failed to add citation: %v", err)
	}

	// Decay citations (should not affect new citations)
	updated, err := store.DecayCitations(ctx)
	if err != nil {
		t.Fatalf("Failed to decay citations: %v", err)
	}

	// New citations (verified today) should not decay
	if updated != 0 {
		t.Errorf("Expected 0 updated citations for new citations, got %d", updated)
	}
}

func TestStore_VerifyCitation_LinesOutOfRange(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "short.txt")
	os.WriteFile(tmpFile, []byte("line1\nline2\n"), 0644)

	mem, _ := store.Remember(ctx, "Test memory", []string{}, "")
	citation, _ := store.AddCitation(ctx, mem.ID, tmpFile, 10, 20, "", "content") // lines 10-20 but file has 2 lines

	verified, valid, err := store.VerifyCitation(ctx, citation.ID)
	if err != nil {
		t.Fatalf("VerifyCitation: %v", err)
	}
	if valid {
		t.Error("citation should be invalid when lines out of range")
	}
	if verified.Confidence != 0.0 {
		t.Errorf("confidence should be 0, got %f", verified.Confidence)
	}
}

// Note: Uses setupTestStore from store_test.go
