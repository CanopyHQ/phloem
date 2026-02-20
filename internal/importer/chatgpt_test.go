package importer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/CanopyHQ/phloem/internal/memory"
)

func TestNewChatGPTImporter(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", dir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	store, err := memory.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	imp := NewChatGPTImporter(store)
	if imp == nil || imp.store != store {
		t.Error("NewChatGPTImporter failed")
	}
}

func TestChatGPTImporter_ImportFromFile_invalidPath(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", dir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	store, err := memory.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	imp := NewChatGPTImporter(store)
	_, err = imp.ImportFromFile(context.Background(), "/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestChatGPTImporter_ImportFromFile_validMinimal(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", dir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	store, err := memory.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	// Minimal valid ChatGPT export: one conversation with one user message
	conv := ChatGPTConversation{
		Title:      "Test",
		CreateTime: 1,
		UpdateTime: 1,
		Mapping: map[string]ChatGPTNode{
			"node1": {
				ID: "node1",
				Message: &ChatGPTMessage{
					ID:      "msg1",
					Author:  ChatGPTAuthor{Role: "user"},
					Content: ChatGPTContent{ContentType: "text", Parts: []string{"Hello"}},
				},
			},
		},
	}
	data, _ := json.Marshal([]ChatGPTConversation{conv})
	fpath := filepath.Join(t.TempDir(), "chatgpt.json")
	os.WriteFile(fpath, data, 0644)

	imp := NewChatGPTImporter(store)
	result, err := imp.ImportFromFile(context.Background(), fpath)
	if err != nil {
		t.Fatalf("ImportFromFile: %v", err)
	}
	if result.ConversationsProcessed < 1 {
		t.Errorf("expected at least 1 conversation processed, got %d", result.ConversationsProcessed)
	}
}

// TestChatGPTImporter_ImportFromFile_createsMemories uses a Q&A pair that passes isWorthRemembering (long enough, not greeting).
func TestChatGPTImporter_ImportFromFile_createsMemories(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", dir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	store, err := memory.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer store.Close()

	userQ := "How do I implement a binary search in Python with recursion and what's the time complexity?"
	assistantA := "You can implement binary search recursively by splitting the array and recurring on the half that might contain the target. Time complexity is O(log n) because we halve the search space each step. Here's a minimal implementation: def bin_search(arr, t, lo, hi): ..."
	conv := ChatGPTConversation{
		Title: "Python binary search",
		Mapping: map[string]ChatGPTNode{
			"root": {
				ID: "root",
				Message: &ChatGPTMessage{
					ID:      "m1",
					Author:  ChatGPTAuthor{Role: "user"},
					Content: ChatGPTContent{ContentType: "text", Parts: []string{userQ}},
				},
				Children: []string{"child"},
			},
			"child": {
				ID: "child",
				Message: &ChatGPTMessage{
					ID:      "m2",
					Author:  ChatGPTAuthor{Role: "assistant"},
					Content: ChatGPTContent{ContentType: "text", Parts: []string{assistantA}},
				},
			},
		},
	}
	data, _ := json.Marshal([]ChatGPTConversation{conv})
	fpath := filepath.Join(dir, "chatgpt_mem.json")
	os.WriteFile(fpath, data, 0644)

	imp := NewChatGPTImporter(store)
	result, err := imp.ImportFromFile(context.Background(), fpath)
	if err != nil {
		t.Fatalf("ImportFromFile: %v", err)
	}
	if result.MemoriesCreated < 1 {
		t.Errorf("expected at least 1 memory created (Q&A worth remembering), got %d", result.MemoriesCreated)
	}
}

// TestChatGPTImporter_ImportFromDirectory walks a directory and imports JSON files.
func TestChatGPTImporter_ImportFromDirectory(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", dir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	store, err := memory.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer store.Close()

	sub := filepath.Join(dir, "sub")
	os.MkdirAll(sub, 0755)
	conv := ChatGPTConversation{
		Title: "Dir import test",
		Mapping: map[string]ChatGPTNode{
			"n1": {ID: "n1", Message: &ChatGPTMessage{
				ID: "m1", Author: ChatGPTAuthor{Role: "user"}, Content: ChatGPTContent{ContentType: "text", Parts: []string{"Hi"}},
			}},
		},
	}
	data, _ := json.Marshal([]ChatGPTConversation{conv})
	os.WriteFile(filepath.Join(sub, "a.json"), data, 0644)

	imp := NewChatGPTImporter(store)
	result, err := imp.ImportFromDirectory(context.Background(), sub)
	if err != nil {
		t.Fatalf("ImportFromDirectory: %v", err)
	}
	if result.ConversationsProcessed < 1 {
		t.Errorf("expected at least 1 conversation, got %d", result.ConversationsProcessed)
	}
}
