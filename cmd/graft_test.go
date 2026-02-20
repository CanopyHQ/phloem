package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecute_Graft_Usage(t *testing.T) {
	defer setArgs("phloem", "graft")()
	out, err := captureStdout(func() {
		if e := Execute(); e != nil {
			t.Fatalf("Execute(graft): %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = out // graft may print usage or subcommand list
}

func TestExecute_Graft_Inspect(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	defer setArgs("phloem", "graft", "inspect", filepath.Join(tmpDir, "nonexistent.graft"))()
	out, _ := captureStdout(func() { Execute() })
	_ = out
}

func TestExecute_Graft_Export_NoOutput(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	defer setArgs("phloem", "graft", "export")()
	err := Execute()
	if err == nil {
		t.Fatal("expected error for graft export without --output")
	}
	if !strings.Contains(err.Error(), "output") {
		t.Errorf("expected error about --output: %v", err)
	}
}

func TestExecute_Graft_Export_WithOutput(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	outPath := filepath.Join(tmpDir, "export.graft")
	// Add a memory first so export has something to package
	restoreRemember := setArgs("phloem", "remember", "graft export test memory")
	_ = Execute()
	restoreRemember()
	restoreExport := setArgs("phloem", "graft", "export", "--output", outPath)
	err := Execute()
	restoreExport()
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Error("expected graft file to be created")
	}
}

func TestExecute_Graft_Import(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	exportPath := filepath.Join(tmpDir, "e.graft")
	restoreRemember := setArgs("phloem", "remember", "graft import test")
	_ = Execute()
	restoreRemember()
	restoreExport := setArgs("phloem", "graft", "export", "--output", exportPath)
	if err := Execute(); err != nil {
		restoreExport()
		t.Fatalf("export: %v", err)
	}
	restoreExport()
	restoreImport := setArgs("phloem", "graft", "import", exportPath)
	err := Execute()
	restoreImport()
	if err != nil {
		t.Fatalf("Execute(import): %v", err)
	}
}

func TestExecute_Graft_Help(t *testing.T) {
	defer setArgs("phloem", "graft", "help")()
	out, err := captureStdout(func() {
		if e := Execute(); e != nil {
			t.Fatalf("Execute: %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Graft") && !strings.Contains(out, "export") {
		t.Errorf("expected graft help: %q", out)
	}
}

func TestExecute_Graft_Unknown(t *testing.T) {
	defer setArgs("phloem", "graft", "unknown_cmd")()
	// Cobra prints help for unknown subcommands and returns nil
	_ = Execute()
}

func TestExecute_Graft_Inspect_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	graftPath := filepath.Join(tmpDir, "valid.graft")
	// Create graft file via export
	defer setArgs("phloem", "remember", "graft inspect valid test")()
	_ = Execute()
	defer setArgs("phloem", "graft", "export", "--output", graftPath)()
	if err := Execute(); err != nil {
		t.Fatalf("graft export: %v", err)
	}
	defer setArgs("phloem", "graft", "inspect", graftPath)()
	out, err := captureStdout(func() {
		if e := Execute(); e != nil {
			t.Fatalf("Execute(graft inspect): %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Graft Manifest") && !strings.Contains(out, "Memories") {
		t.Errorf("expected manifest output: %q", out)
	}
}
