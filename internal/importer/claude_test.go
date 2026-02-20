package importer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/CanopyHQ/phloem/internal/memory"
)

func TestNewClaudeImporter(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", dir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	store, err := memory.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	imp := NewClaudeImporter(store)
	if imp == nil || imp.store != store {
		t.Error("NewClaudeImporter failed")
	}
}

func TestClaudeImporter_ImportFromFile_invalidPath(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", dir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	store, err := memory.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	imp := NewClaudeImporter(store)
	_, err = imp.ImportFromFile(context.Background(), "/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestClaudeImporter_ImportFromFile_validJSON(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", dir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	store, err := memory.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	conv := ClaudeConversation{
		UUID: "u1",
		Name: "Test",
		ChatMessages: []ClaudeMessage{
			{Text: "user message", Sender: "human"},
			{Text: "assistant reply", Sender: "assistant"},
		},
	}
	data, _ := json.Marshal([]ClaudeConversation{conv})
	fpath := filepath.Join(t.TempDir(), "claude.json")
	os.WriteFile(fpath, data, 0644)

	imp := NewClaudeImporter(store)
	result, err := imp.ImportFromFile(context.Background(), fpath)
	if err != nil {
		t.Fatalf("ImportFromFile: %v", err)
	}
	if result.ConversationsProcessed < 1 {
		t.Errorf("expected at least 1 conversation processed, got %d", result.ConversationsProcessed)
	}
}

func TestClaudeImporter_ImportFromFile_createsMemories(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", dir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	store, err := memory.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer store.Close()

	userQ := "How do I implement a binary search in Go with recursion and what's the time complexity?"
	assistantA := "You can implement binary search recursively by splitting the slice and recurring on the half that might contain the target. Time complexity is O(log n) because we halve the search space each step. Here's a minimal implementation in Go: func binSearch(arr []int, t, lo, hi int) int { ... }"
	conv := ClaudeConversation{
		UUID: "u2",
		Name: "Go binary search",
		ChatMessages: []ClaudeMessage{
			{Text: userQ, Sender: "human"},
			{Text: assistantA, Sender: "assistant"},
		},
	}
	data, _ := json.Marshal([]ClaudeConversation{conv})
	fpath := filepath.Join(dir, "claude_mem.json")
	os.WriteFile(fpath, data, 0644)

	imp := NewClaudeImporter(store)
	result, err := imp.ImportFromFile(context.Background(), fpath)
	if err != nil {
		t.Fatalf("ImportFromFile: %v", err)
	}
	if result.MemoriesCreated < 1 {
		t.Errorf("expected at least 1 memory created, got %d", result.MemoriesCreated)
	}
}

func TestClaudeImporter_ImportFromDirectory(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", dir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	store, err := memory.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer store.Close()

	sub := filepath.Join(dir, "claude_sub")
	os.MkdirAll(sub, 0755)
	conv := ClaudeConversation{
		UUID: "u3",
		Name: "Dir test",
		ChatMessages: []ClaudeMessage{
			{Text: "Hi", Sender: "human"},
			{Text: "Hello", Sender: "assistant"},
		},
	}
	data, _ := json.Marshal([]ClaudeConversation{conv})
	os.WriteFile(filepath.Join(sub, "a.json"), data, 0644)

	imp := NewClaudeImporter(store)
	result, err := imp.ImportFromDirectory(context.Background(), sub)
	if err != nil {
		t.Fatalf("ImportFromDirectory: %v", err)
	}
	if result.ConversationsProcessed < 1 {
		t.Errorf("expected at least 1 conversation, got %d", result.ConversationsProcessed)
	}
}
