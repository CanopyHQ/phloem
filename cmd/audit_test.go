package cmd

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/CanopyHQ/phloem/internal/memory"
)

func TestRunAudit_EmptyDataDir(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")

	out, err := captureStdout(func() {
		if e := runAudit(); e != nil {
			t.Fatalf("runAudit: %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Privacy Audit") {
		t.Errorf("expected audit header in output: %q", out)
	}
	if !strings.Contains(out, "Data Inventory") {
		t.Errorf("expected Data Inventory section: %q", out)
	}
}

func TestRunAudit_WithMemories(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")

	// Create a store and add a memory to populate the DB
	store, err := memory.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	ctx := context.Background()
	_, err = store.Remember(ctx, "audit test memory", []string{"audit"}, "")
	if err != nil {
		t.Fatalf("Remember: %v", err)
	}
	store.Close()

	out, capErr := captureStdout(func() {
		if e := runAudit(); e != nil {
			t.Fatalf("runAudit: %v", e)
		}
	})
	if capErr != nil {
		t.Fatal(capErr)
	}

	// Should show table row counts
	if !strings.Contains(out, "row(s)") {
		t.Errorf("expected row counts in output: %q", out)
	}
	// Should show the database file in the inventory
	if !strings.Contains(out, "memories.db") {
		t.Errorf("expected memories.db in data inventory: %q", out)
	}
}

func TestExecute_Audit(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")

	defer setArgs("phloem", "audit")()
	out, err := captureStdout(func() {
		if e := Execute(); e != nil {
			t.Fatalf("Execute(audit): %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Privacy Audit") {
		t.Errorf("expected audit output: %q", out)
	}
}
