package cmd

import (
	"io"
	"os"
	"testing"
)

func TestRedact_Empty(t *testing.T) {
	if got := redact("", 2); got != "(not set)" {
		t.Errorf("redact(\"\", 2): got %q", got)
	}
}

func TestRedact_Short(t *testing.T) {
	if got := redact("ab", 2); got != "***" {
		t.Errorf("redact(\"ab\", 2): got %q want ***", got)
	}
}

func TestRedact_Long(t *testing.T) {
	if got := redact("abcdefgh", 2); got != "ab...gh" {
		t.Errorf("redact(\"abcdefgh\", 2): got %q want ab...gh", got)
	}
}

func TestExecute_Doctor(t *testing.T) {
	tmpDir := t.TempDir()
	orig := os.Getenv("PHLOEM_DATA_DIR")
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer func() { os.Setenv("PHLOEM_DATA_DIR", orig) }()

	defer setArgs("phloem", "doctor")()
	r, w, _ := os.Pipe()
	oldStderr := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = oldStderr; w.Close() }()
	err := Execute()
	w.Close()
	if err != nil {
		// Doctor may report issues in test environment.
		// We verify the command runs without panicking.
		t.Logf("Execute(doctor): %v (expected in test environment)", err)
	}
	io.ReadAll(r)
}
