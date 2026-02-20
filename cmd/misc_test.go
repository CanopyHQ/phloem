package cmd

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/CanopyHQ/phloem/internal/memory"
)

func TestExecute_Dreams(t *testing.T) {
	tmpDir := t.TempDir()
	orig := os.Getenv("PHLOEM_DATA_DIR")
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer func() { os.Setenv("PHLOEM_DATA_DIR", orig) }()

	defer setArgs("phloem", "dreams")()
	err := Execute()
	if err != nil {
		t.Fatalf("Execute(dreams): %v", err)
	}
}

func TestExecute_Verify(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	defer setArgs("phloem", "verify", "nonexistent-id")()
	out, err := captureStdout(func() {
		if e := Execute(); e != nil {
			t.Fatalf("Execute(verify): %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = out
}

func TestExecute_Decay(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	defer setArgs("phloem", "decay")()
	err := Execute()
	if err != nil {
		t.Fatalf("Execute(decay): %v", err)
	}
}

// runVerify and runDecay are not in Execute() switch; test them directly for coverage
func TestRunVerify_NoCitations(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	out, err := captureStdout(func() {
		if e := runVerify("nonexistent-id"); e != nil {
			t.Fatalf("runVerify: %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "no citations") && !strings.Contains(out, "Memory") {
		t.Errorf("expected no citations message: %q", out)
	}
}

func TestRunVerify_WithCitations(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	store, err := memory.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer store.Close()
	ctx := context.Background()
	mem, err := store.Remember(ctx, "verify test memory", []string{"verify"}, "")
	if err != nil {
		t.Fatalf("Remember: %v", err)
	}
	_, err = store.AddCitation(ctx, mem.ID, "/path/to/file.go", 10, 20, "", "snippet")
	if err != nil {
		t.Fatalf("AddCitation: %v", err)
	}
	out, err := captureStdout(func() {
		if e := runVerify(mem.ID); e != nil {
			t.Fatalf("runVerify: %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Verifying") && !strings.Contains(out, "citation") {
		t.Errorf("expected verify output: %q", out)
	}
}

func TestRunDecay(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	out, err := captureStdout(func() {
		if e := runDecay(); e != nil {
			t.Fatalf("runDecay: %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "decay") && !strings.Contains(out, "citation") && !strings.Contains(out, "Applying") {
		t.Errorf("expected decay message: %q", out)
	}
}

func TestRunDecay_WithCitations(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	store, err := memory.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer store.Close()
	ctx := context.Background()
	mem, err := store.Remember(ctx, "decay test memory", nil, "")
	if err != nil {
		t.Fatalf("Remember: %v", err)
	}
	_, err = store.AddCitation(ctx, mem.ID, "/path/to/file.go", 1, 2, "", "snippet")
	if err != nil {
		t.Fatalf("AddCitation: %v", err)
	}
	out, err := captureStdout(func() {
		if e := runDecay(); e != nil {
			t.Fatalf("runDecay: %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "decay") && !strings.Contains(out, "citation") {
		t.Errorf("expected decay output: %q", out)
	}
}
