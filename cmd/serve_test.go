package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestExecute_Version(t *testing.T) {
	defer setArgs("phloem", "version")()
	out, err := captureStdout(func() {
		if e := Execute(); e != nil {
			t.Fatalf("Execute(version): %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if out == "" {
		t.Error("version should print to stdout")
	}
	if !strings.Contains(out, "phloem") {
		t.Errorf("version output should contain 'phloem': %q", out)
	}
}

func TestExecute_Status(t *testing.T) {
	tmpDir := t.TempDir()
	orig := os.Getenv("PHLOEM_DATA_DIR")
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer func() {
		os.Setenv("PHLOEM_DATA_DIR", orig)
	}()

	defer setArgs("phloem", "status")()
	out, err := captureStdout(func() {
		if e := Execute(); e != nil {
			t.Fatalf("Execute(status): %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Phloem Memory Status") {
		t.Errorf("status output: %q", out)
	}
}
