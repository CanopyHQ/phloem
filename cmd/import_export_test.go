package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecute_Import_Usage(t *testing.T) {
	defer setArgs("phloem", "import")()
	err := Execute()
	if err == nil {
		t.Fatal("import without args should return error")
	}
}

func TestExecute_Import_ChatGPT_File(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")

	// Minimal ChatGPT export: one conversation
	jsonContent := `[{"title":"Test","create_time":1,"update_time":1,"mapping":{"n1":{"id":"n1","message":{"id":"m1","author":{"role":"user"},"content":{"content_type":"text","parts":["Hi"]}}}}}]`
	importPath := filepath.Join(tmpDir, "chatgpt.json")
	os.WriteFile(importPath, []byte(jsonContent), 0644)

	defer setArgs("phloem", "import", "chatgpt", importPath)()
	err := Execute()
	if err != nil {
		t.Fatalf("Execute(import chatgpt): %v", err)
	}
}

func TestExecute_Import_Claude_File(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")

	jsonContent := `[{"uuid":"u1","name":"Test","chat_messages":[{"text":"Hi","sender":"human"},{"text":"Hello","sender":"assistant"}]}]`
	importPath := filepath.Join(tmpDir, "claude.json")
	os.WriteFile(importPath, []byte(jsonContent), 0644)

	defer setArgs("phloem", "import", "claude", importPath)()
	err := Execute()
	if err != nil {
		t.Fatalf("Execute(import claude): %v", err)
	}
}

func TestExecute_Export(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")

	defer setArgs("phloem", "export", "json", filepath.Join(tmpDir, "out.json"))()
	err := Execute()
	if err != nil {
		t.Fatalf("Execute(export): %v", err)
	}
}

func TestExecute_Export_WithMemories_Json(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")

	restoreRemember := setArgs("phloem", "remember", "export test memory content")
	_ = Execute()
	restoreRemember()

	outPath := filepath.Join(tmpDir, "export.json")
	restoreExport := setArgs("phloem", "export", "json", outPath)
	err := Execute()
	restoreExport()
	if err != nil {
		t.Fatalf("Execute(export json): %v", err)
	}
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Error("expected export file to be created")
	}
}

func TestExecute_Export_WithMemories_Markdown(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")

	restoreRemember := setArgs("phloem", "remember", "export markdown test")
	_ = Execute()
	restoreRemember()

	outPath := filepath.Join(tmpDir, "export.md")
	restoreExport := setArgs("phloem", "export", "markdown", outPath)
	err := Execute()
	restoreExport()
	if err != nil {
		t.Fatalf("Execute(export markdown): %v", err)
	}
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Error("expected export file to be created")
	}
}

func TestExecute_Export_Markdown_WithMemories(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	defer setArgs("phloem", "remember", "export markdown test content")()
	_ = Execute()
	outPath := filepath.Join(tmpDir, "out.md")
	defer setArgs("phloem", "export", "markdown", outPath)()
	err := Execute()
	if err != nil {
		t.Fatalf("Execute(export markdown): %v", err)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	if !strings.Contains(string(data), "Phloem Memory Export") || !strings.Contains(string(data), "export markdown test") {
		t.Errorf("expected markdown content: %q", string(data))
	}
}

func TestExecute_Export_UnknownFormat(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	// Add a memory so runExport hits the format switch (not "no memories")
	defer setArgs("phloem", "remember", "export unknown format test")()
	_ = Execute()
	defer setArgs("phloem", "export", "csv")()
	err := Execute()
	if err == nil {
		t.Error("Execute(export csv) should fail with unknown format")
	}
	if err != nil && !strings.Contains(err.Error(), "unknown format") {
		t.Errorf("expected unknown format error, got: %v", err)
	}
}

func TestExecute_Export_DefaultOutput(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")

	// Change to a temp dir so the default output file doesn't leak into the source tree
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	outputDir := t.TempDir()
	if err := os.Chdir(outputDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer os.Chdir(origDir)

	restoreRemember := setArgs("phloem", "remember", "export default output test")
	_ = Execute()
	restoreRemember()
	// export json with no output path -> default filename phloem-export-YYYY-MM-DD.json
	defer setArgs("phloem", "export", "json")()
	err = Execute()
	if err != nil {
		t.Fatalf("Execute(export json): %v", err)
	}
}
