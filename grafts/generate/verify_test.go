package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/CanopyHQ/phloem/internal/graft"
)

func TestGraftFiles(t *testing.T) {
	// Find the grafts directory relative to the test file
	graftsDir := filepath.Join("..")

	files := []struct {
		filename    string
		wantName    string
		wantMinMem  int
		wantMaxMem  int
		wantVersion string
	}{
		{"go-best-practices.graft", "Go Best Practices", 10, 15, "1.0.0"},
		{"react-patterns.graft", "React Patterns", 10, 15, "1.0.0"},
		{"security-essentials.graft", "Security Essentials", 10, 15, "1.0.0"},
	}

	for _, f := range files {
		t.Run(f.filename, func(t *testing.T) {
			path := filepath.Join(graftsDir, f.filename)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Fatalf("graft file does not exist: %s", path)
			}

			// Inspect manifest
			manifest, err := graft.Inspect(path)
			if err != nil {
				t.Fatalf("failed to inspect %s: %v", f.filename, err)
			}

			if manifest.Name != f.wantName {
				t.Errorf("name = %q, want %q", manifest.Name, f.wantName)
			}
			if manifest.Version != f.wantVersion {
				t.Errorf("version = %q, want %q", manifest.Version, f.wantVersion)
			}
			if manifest.MemoryCount < f.wantMinMem || manifest.MemoryCount > f.wantMaxMem {
				t.Errorf("memory count = %d, want between %d and %d", manifest.MemoryCount, f.wantMinMem, f.wantMaxMem)
			}
			if manifest.Author != "Canopy Team" {
				t.Errorf("author = %q, want %q", manifest.Author, "Canopy Team")
			}

			t.Logf("Name:        %s", manifest.Name)
			t.Logf("Description: %s", manifest.Description)
			t.Logf("Author:      %s", manifest.Author)
			t.Logf("Version:     %s", manifest.Version)
			t.Logf("Memories:    %d", manifest.MemoryCount)
			t.Logf("Tags:        %v", manifest.Tags)

			// Full unpack to verify memories
			payload, err := graft.Unpack(path)
			if err != nil {
				t.Fatalf("failed to unpack %s: %v", f.filename, err)
			}

			if len(payload.Memories) != manifest.MemoryCount {
				t.Errorf("actual memories = %d, manifest says %d", len(payload.Memories), manifest.MemoryCount)
			}

			for i, m := range payload.Memories {
				if m.ID == "" {
					t.Errorf("memory[%d] has empty ID", i)
				}
				if m.Content == "" {
					t.Errorf("memory[%d] (%s) has empty content", i, m.ID)
				}
				if len(m.Tags) == 0 {
					t.Errorf("memory[%d] (%s) has no tags", i, m.ID)
				}
				t.Logf("  [%d] %s (tags: %v)", i, m.ID, m.Tags)
			}
		})
	}
}
