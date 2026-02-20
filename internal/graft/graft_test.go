package graft

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/CanopyHQ/phloem/internal/memory"
)

func TestPackage_ParentDirMissing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graft-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// outputPath has a parent that does not exist, so os.Create fails
	outputPath := filepath.Join(tmpDir, "missing", "sub", "test.graft")
	manifest := Manifest{ID: "x", Name: "x", CreatedAt: time.Now(), MemoryCount: 0}
	err = Package(manifest, nil, nil, outputPath)
	if err == nil {
		t.Fatal("expected error when parent dir does not exist")
	}
	if !strings.Contains(err.Error(), "create file") && !strings.Contains(err.Error(), "failed to create") {
		t.Errorf("error should mention create/file: %v", err)
	}
}

func TestPackage_ValidGraft(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graft-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputPath := filepath.Join(tmpDir, "test.graft")

	manifest := Manifest{
		ID:          "test-graft-1",
		Name:        "Test Graft",
		Description: "A test graft for unit testing",
		Author:      "Test Author",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		MemoryCount: 2,
		Tags:        []string{"test", "unit"},
	}

	memories := []memory.Memory{
		{
			ID:        "mem-1",
			Content:   "Test memory content 1",
			Tags:      []string{"test"},
			CreatedAt: time.Now(),
		},
		{
			ID:        "mem-2",
			Content:   "Test memory content 2",
			Tags:      []string{"test"},
			CreatedAt: time.Now(),
		},
	}

	err = Package(manifest, memories, nil, outputPath)
	if err != nil {
		t.Fatalf("Package failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Graft file was not created")
	}

	// Verify file is not empty
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("Graft file is empty")
	}
}

func TestUnpack_ValidGraft(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graft-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// First create a graft
	outputPath := filepath.Join(tmpDir, "test.graft")
	manifest := Manifest{
		ID:          "test-graft-1",
		Name:        "Test Graft",
		Description: "A test graft",
		Author:      "Test Author",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		MemoryCount: 1,
		Tags:        []string{"test"},
	}

	memories := []memory.Memory{
		{
			ID:        "mem-1",
			Content:   "Test content",
			Tags:      []string{"test"},
			CreatedAt: time.Now(),
		},
	}

	if err := Package(manifest, memories, nil, outputPath); err != nil {
		t.Fatalf("failed to create test graft: %v", err)
	}

	// Now unpack it
	payload, err := Unpack(outputPath)
	if err != nil {
		t.Fatalf("Unpack failed: %v", err)
	}

	// Verify manifest
	if payload.Manifest.ID != manifest.ID {
		t.Errorf("expected manifest ID %s, got %s", manifest.ID, payload.Manifest.ID)
	}
	if payload.Manifest.Name != manifest.Name {
		t.Errorf("expected manifest name %s, got %s", manifest.Name, payload.Manifest.Name)
	}

	// Verify memories
	if len(payload.Memories) != 1 {
		t.Fatalf("expected 1 memory, got %d", len(payload.Memories))
	}
	if payload.Memories[0].Content != "Test content" {
		t.Errorf("expected memory content 'Test content', got '%s'", payload.Memories[0].Content)
	}
}

func TestUnpack_InvalidMagicBytes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graft-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	invalidFile := filepath.Join(tmpDir, "invalid.graft")
	if err := os.WriteFile(invalidFile, []byte("INVALID"), 0644); err != nil {
		t.Fatalf("failed to create invalid file: %v", err)
	}

	_, err = Unpack(invalidFile)
	if err == nil {
		t.Fatal("expected error for invalid magic bytes, got nil")
	}
	if err.Error() != "invalid file format: not a .graft file" {
		t.Errorf("expected 'invalid file format' error, got: %v", err)
	}
}

func TestInspect_ValidGraft(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graft-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a graft
	outputPath := filepath.Join(tmpDir, "test.graft")
	manifest := Manifest{
		ID:          "inspect-test",
		Name:        "Inspect Test Graft",
		Description: "For testing inspect",
		Author:      "Test Author",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		MemoryCount: 5,
		Tags:        []string{"inspect", "test"},
	}

	memories := make([]memory.Memory, 5)
	for i := 0; i < 5; i++ {
		memories[i] = memory.Memory{
			ID:        "mem-" + string(rune(i)),
			Content:   "Memory content",
			CreatedAt: time.Now(),
		}
	}

	if err := Package(manifest, memories, nil, outputPath); err != nil {
		t.Fatalf("failed to create test graft: %v", err)
	}

	// Inspect it
	inspected, err := Inspect(outputPath)
	if err != nil {
		t.Fatalf("Inspect failed: %v", err)
	}

	// Verify manifest
	if inspected.ID != manifest.ID {
		t.Errorf("expected ID %s, got %s", manifest.ID, inspected.ID)
	}
	if inspected.Name != manifest.Name {
		t.Errorf("expected name %s, got %s", manifest.Name, inspected.Name)
	}
	if inspected.MemoryCount != 5 {
		t.Errorf("expected memory count 5, got %d", inspected.MemoryCount)
	}
}

func TestPackage_EmptyMemories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graft-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputPath := filepath.Join(tmpDir, "empty.graft")

	manifest := Manifest{
		ID:          "empty-graft",
		Name:        "Empty Graft",
		Description: "A graft with no memories",
		Author:      "Test",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		MemoryCount: 0,
		Tags:        []string{},
	}

	err = Package(manifest, []memory.Memory{}, nil, outputPath)
	if err != nil {
		t.Fatalf("Package should succeed with empty memories: %v", err)
	}

	// Verify it can be unpacked
	payload, err := Unpack(outputPath)
	if err != nil {
		t.Fatalf("Failed to unpack empty graft: %v", err)
	}

	if len(payload.Memories) != 0 {
		t.Errorf("expected 0 memories, got %d", len(payload.Memories))
	}
}

func TestPackage_ManyMemories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graft-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputPath := filepath.Join(tmpDir, "large.graft")

	manifest := Manifest{
		ID:          "large-graft",
		Name:        "Large Graft",
		Description: "A graft with many memories",
		Author:      "Test",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		MemoryCount: 100,
		Tags:        []string{"large"},
	}

	memories := make([]memory.Memory, 100)
	for i := 0; i < 100; i++ {
		memories[i] = memory.Memory{
			ID:        "mem-" + string(rune(i)),
			Content:   "Memory content " + string(rune(i)),
			Tags:      []string{"test"},
			CreatedAt: time.Now(),
		}
	}

	err = Package(manifest, memories, nil, outputPath)
	if err != nil {
		t.Fatalf("Package failed with many memories: %v", err)
	}

	// Verify it can be unpacked
	payload, err := Unpack(outputPath)
	if err != nil {
		t.Fatalf("Failed to unpack large graft: %v", err)
	}

	if len(payload.Memories) != 100 {
		t.Errorf("expected 100 memories, got %d", len(payload.Memories))
	}
}

func TestUnpack_NonExistentFile(t *testing.T) {
	_, err := Unpack("/nonexistent/file.graft")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}

func TestUnpack_WrongVersion(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graft-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	badFile := filepath.Join(tmpDir, "bad.graft")
	f, err := os.Create(badFile)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	f.Write(MagicBytes)
	var v uint8 = 99
	binary.Write(f, binary.LittleEndian, v)
	f.Close()

	_, err = Unpack(badFile)
	if err == nil {
		t.Fatal("expected error for wrong version, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported version") {
		t.Errorf("expected unsupported version error, got: %v", err)
	}
}

func TestPackage_WithCitations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graft-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputPath := filepath.Join(tmpDir, "with-citations.graft")

	manifest := Manifest{
		ID:          "citations-test",
		Name:        "Graft with Citations",
		Description: "Testing citation support",
		Author:      "Test",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		MemoryCount: 1,
		Tags:        []string{"citations"},
	}

	memories := []memory.Memory{
		{
			ID:        "mem-1",
			Content:   "Memory with citation",
			CreatedAt: time.Now(),
		},
	}

	citations := []memory.Citation{
		{
			ID:         "cite-1",
			MemoryID:   "mem-1",
			FilePath:   "/path/to/file.go",
			StartLine:  42,
			EndLine:    45,
			Content:    "func example() {}",
			Confidence: 0.95,
			CreatedAt:  time.Now(),
		},
	}

	err = Package(manifest, memories, citations, outputPath)
	if err != nil {
		t.Fatalf("Package failed with citations: %v", err)
	}

	// Verify citations are included
	payload, err := Unpack(outputPath)
	if err != nil {
		t.Fatalf("Failed to unpack graft with citations: %v", err)
	}

	if len(payload.Citations) != 1 {
		t.Errorf("expected 1 citation, got %d", len(payload.Citations))
	}
	if payload.Citations[0].FilePath != "/path/to/file.go" {
		t.Errorf("expected file path '/path/to/file.go', got '%s'", payload.Citations[0].FilePath)
	}
}
