package cmd

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestExecute_Setup_Usage(t *testing.T) {
	defer setArgs("phloem", "setup")()
	out, err := captureStdout(func() {
		if e := Execute(); e != nil {
			t.Fatalf("Execute(setup): %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	// setup without subcommand runs runSetup() which auto-detects IDEs; accept usage or setup output
	if !strings.Contains(out, "Usage") && !strings.Contains(out, "Setting up") && !strings.Contains(out, "configured") && !strings.Contains(out, "Canopy") {
		t.Errorf("setup without args should print usage or setup output: %q", out[:min(200, len(out))])
	}
}

func TestExecute_Setup_UnknownIDE(t *testing.T) {
	defer setArgs("phloem", "setup", "unknown")()
	// Cobra prints help for unknown subcommands and returns nil
	_ = Execute()
}

func TestExecute_Setup_Cursor(t *testing.T) {
	tmpDir := t.TempDir()
	orig := os.Getenv("PHLOEM_DATA_DIR")
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer func() { os.Setenv("PHLOEM_DATA_DIR", orig) }()

	defer setArgs("phloem", "setup", "cursor")()
	r, w, _ := os.Pipe()
	oldStderr := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = oldStderr; w.Close() }()
	err := Execute()
	w.Close()
	if err != nil {
		t.Fatalf("Execute(setup cursor): %v", err)
	}
	io.ReadAll(r)
}

func TestExecute_Setup_Windsurf(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	defer setArgs("phloem", "setup", "windsurf")()
	r, w, _ := os.Pipe()
	oldStderr := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = oldStderr; w.Close() }()
	err := Execute()
	w.Close()
	if err != nil {
		t.Fatalf("Execute(setup windsurf): %v", err)
	}
	io.ReadAll(r)
}

func TestExecute_Setup_UnknownIDE_Vscode(t *testing.T) {
	defer setArgs("phloem", "setup", "vscode")()
	// Cobra prints help for unknown subcommands and returns nil
	_ = Execute()
}
