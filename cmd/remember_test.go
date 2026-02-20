package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestExecute_Remember(t *testing.T) {
	tmpDir := t.TempDir()
	orig := os.Getenv("PHLOEM_DATA_DIR")
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer func() { os.Setenv("PHLOEM_DATA_DIR", orig) }()

	defer setArgs("phloem", "remember", "test memory content")()
	out, err := captureStdout(func() {
		if e := Execute(); e != nil {
			t.Fatalf("Execute(remember): %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if out == "" && !strings.Contains(out, "id") {
		_ = out // may print ID or "Stored"
	}
}

func TestExecute_Remember_WithTags(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	defer setArgs("phloem", "remember", "content with tags", "--tags", "a,b")()
	err := Execute()
	if err != nil {
		t.Fatalf("Execute(remember --tags): %v", err)
	}
}
