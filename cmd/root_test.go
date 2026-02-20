package cmd

import (
	"io"
	"os"
	"strings"
	"testing"
)

func setArgs(args ...string) func() {
	orig := os.Args
	os.Args = args
	return func() { os.Args = orig }
}

func captureStdout(f func()) (string, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	old := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = old; w.Close() }()
	f()
	w.Close()
	data, _ := io.ReadAll(r)
	return string(data), nil
}

func TestExecute_Help(t *testing.T) {
	defer setArgs("phloem", "help")()
	out, err := captureStdout(func() {
		if e := Execute(); e != nil {
			t.Fatalf("Execute(help): %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Phloem") {
		t.Errorf("help output should contain 'Phloem': %q", out)
	}
}

func TestExecute_HelpShortFlag(t *testing.T) {
	defer setArgs("phloem", "-h")()
	out, err := captureStdout(func() {
		if e := Execute(); e != nil {
			t.Fatalf("Execute(-h): %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Error("help -h should print")
	}
}

func TestSetVersion(t *testing.T) {
	SetVersion("1.2.3", "abc123", "2026-01-01")
	if Version != "1.2.3" || Commit != "abc123" || Date != "2026-01-01" {
		t.Errorf("SetVersion: got Version=%q Commit=%q Date=%q", Version, Commit, Date)
	}
	// Restore for other tests
	SetVersion("dev", "none", "unknown")
}
