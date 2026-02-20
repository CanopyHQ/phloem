package cmd

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ensurePhloemInPath builds the phloem binary and adds it to PATH for tests
// that call exec.LookPath("phloem"). Returns a cleanup function.
func ensurePhloemInPath(t *testing.T) func() {
	t.Helper()

	// Check if phloem is already in PATH
	if _, err := exec.LookPath("phloem"); err == nil {
		return func() {}
	}

	// Build phloem into a temp bin directory
	binDir := t.TempDir()
	binary := filepath.Join(binDir, "phloem")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}

	// Build from the module root (one level up from cmd/)
	cmd := exec.Command("go", "build", "-o", binary, ".")
	cmd.Dir = filepath.Join("..")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("cannot build phloem binary for test: %v\n%s", err, out)
	}

	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+origPath)
	return func() {
		os.Setenv("PATH", origPath)
	}
}

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
	if !strings.Contains(out, "Usage") && !strings.Contains(out, "Setting up") && !strings.Contains(out, "configured") && !strings.Contains(out, "Phloem") {
		t.Errorf("setup without args should print usage or setup output: %q", out[:min(200, len(out))])
	}
}

func TestExecute_Setup_UnknownIDE(t *testing.T) {
	defer setArgs("phloem", "setup", "unknown")()
	// Cobra prints help for unknown subcommands and returns nil
	_ = Execute()
}

func TestExecute_Setup_Cursor(t *testing.T) {
	defer ensurePhloemInPath(t)()
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
	defer ensurePhloemInPath(t)()
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

// ---------------------------------------------------------------------------
// Installation verification tests
// ---------------------------------------------------------------------------

// setupTestHome creates a temp dir, sets HOME to it, and returns the path
// plus a cleanup function that restores the original environment.
func setupTestHome(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	origData := os.Getenv("PHLOEM_DATA_DIR")
	os.Setenv("PHLOEM_DATA_DIR", filepath.Join(tmpDir, ".phloem"))
	return tmpDir, func() {
		os.Setenv("HOME", origHome)
		if origData == "" {
			os.Unsetenv("PHLOEM_DATA_DIR")
		} else {
			os.Setenv("PHLOEM_DATA_DIR", origData)
		}
	}
}

// readJSONConfig reads a JSON file into a generic map.
func readJSONConfig(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to parse %s: %v", path, err)
	}
	return config
}

// getMCPServers extracts the mcpServers map from a parsed config.
func getMCPServers(t *testing.T, config map[string]interface{}) map[string]interface{} {
	t.Helper()
	servers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatal("mcpServers not found or not a map")
	}
	return servers
}

func TestSetupCursor_CreatesConfig(t *testing.T) {
	defer ensurePhloemInPath(t)()
	home, cleanup := setupTestHome(t)
	defer cleanup()

	cursorDir := filepath.Join(home, ".cursor")
	if err := os.MkdirAll(cursorDir, 0755); err != nil {
		t.Fatal(err)
	}

	_, err := captureStdout(func() {
		if e := runSetupCursor(); e != nil {
			t.Fatalf("runSetupCursor: %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(cursorDir, "mcp.json")
	config := readJSONConfig(t, configPath)
	servers := getMCPServers(t, config)

	if _, ok := servers["phloem"]; !ok {
		t.Error("expected phloem server in mcpServers")
	}

	phloemServer, ok := servers["phloem"].(map[string]interface{})
	if !ok {
		t.Fatal("phloem server is not a map")
	}
	if _, ok := phloemServer["command"]; !ok {
		t.Error("phloem server missing 'command' field")
	}
	if _, ok := phloemServer["args"]; !ok {
		t.Error("phloem server missing 'args' field")
	}
}

func TestSetupCursor_PreservesExistingServers(t *testing.T) {
	defer ensurePhloemInPath(t)()
	home, cleanup := setupTestHome(t)
	defer cleanup()

	cursorDir := filepath.Join(home, ".cursor")
	if err := os.MkdirAll(cursorDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Seed with an existing server
	existingConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"other-server": map[string]interface{}{
				"command": "/usr/bin/other",
				"args":    []string{"run"},
			},
		},
	}
	data, _ := json.MarshalIndent(existingConfig, "", "  ")
	configPath := filepath.Join(cursorDir, "mcp.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := captureStdout(func() {
		if e := runSetupCursor(); e != nil {
			t.Fatalf("runSetupCursor: %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	config := readJSONConfig(t, configPath)
	servers := getMCPServers(t, config)

	if _, ok := servers["other-server"]; !ok {
		t.Error("existing 'other-server' was not preserved")
	}
	if _, ok := servers["phloem"]; !ok {
		t.Error("phloem server was not added")
	}
}

func TestSetupCursor_Idempotent(t *testing.T) {
	defer ensurePhloemInPath(t)()
	home, cleanup := setupTestHome(t)
	defer cleanup()

	cursorDir := filepath.Join(home, ".cursor")
	if err := os.MkdirAll(cursorDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Run setup twice
	for i := 0; i < 2; i++ {
		_, err := captureStdout(func() {
			if e := runSetupCursor(); e != nil {
				t.Fatalf("runSetupCursor (run %d): %v", i+1, e)
			}
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	configPath := filepath.Join(cursorDir, "mcp.json")
	config := readJSONConfig(t, configPath)
	servers := getMCPServers(t, config)

	if _, ok := servers["phloem"]; !ok {
		t.Error("phloem server missing after idempotent run")
	}
}

func TestSetupWindsurf_CreatesConfig(t *testing.T) {
	defer ensurePhloemInPath(t)()
	home, cleanup := setupTestHome(t)
	defer cleanup()

	windsurfDir := filepath.Join(home, ".windsurf")
	if err := os.MkdirAll(windsurfDir, 0755); err != nil {
		t.Fatal(err)
	}

	_, err := captureStdout(func() {
		if e := runSetupWindsurf(); e != nil {
			t.Fatalf("runSetupWindsurf: %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(windsurfDir, "mcp_config.json")
	config := readJSONConfig(t, configPath)
	servers := getMCPServers(t, config)

	if _, ok := servers["phloem"]; !ok {
		t.Error("expected phloem server in mcpServers")
	}

	phloemServer, ok := servers["phloem"].(map[string]interface{})
	if !ok {
		t.Fatal("phloem server is not a map")
	}
	if _, ok := phloemServer["command"]; !ok {
		t.Error("phloem server missing 'command' field")
	}
	if _, ok := phloemServer["args"]; !ok {
		t.Error("phloem server missing 'args' field")
	}
}

func TestSetupWindsurf_PreservesExistingServers(t *testing.T) {
	defer ensurePhloemInPath(t)()
	home, cleanup := setupTestHome(t)
	defer cleanup()

	windsurfDir := filepath.Join(home, ".windsurf")
	if err := os.MkdirAll(windsurfDir, 0755); err != nil {
		t.Fatal(err)
	}

	existingConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"other-server": map[string]interface{}{
				"command": "/usr/bin/other",
				"args":    []string{"run"},
			},
		},
	}
	data, _ := json.MarshalIndent(existingConfig, "", "  ")
	configPath := filepath.Join(windsurfDir, "mcp_config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := captureStdout(func() {
		if e := runSetupWindsurf(); e != nil {
			t.Fatalf("runSetupWindsurf: %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	config := readJSONConfig(t, configPath)
	servers := getMCPServers(t, config)

	if _, ok := servers["other-server"]; !ok {
		t.Error("existing 'other-server' was not preserved")
	}
	if _, ok := servers["phloem"]; !ok {
		t.Error("phloem server was not added")
	}
}

func TestSetupAutoDetect(t *testing.T) {
	defer ensurePhloemInPath(t)()
	home, cleanup := setupTestHome(t)
	defer cleanup()

	// Create both IDE directories
	os.MkdirAll(filepath.Join(home, ".cursor"), 0755)
	os.MkdirAll(filepath.Join(home, ".windsurf"), 0755)

	out, err := captureStdout(func() {
		if e := runSetup(); e != nil {
			t.Fatalf("runSetup: %v", e)
		}
	})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "Detected Cursor") {
		t.Error("auto-detect did not find Cursor")
	}
	if !strings.Contains(out, "Detected Windsurf") {
		t.Error("auto-detect did not find Windsurf")
	}

	// Verify both configs were created
	cursorConfig := filepath.Join(home, ".cursor", "mcp.json")
	if _, err := os.Stat(cursorConfig); os.IsNotExist(err) {
		t.Error("Cursor mcp.json was not created")
	}
	windsurfConfig := filepath.Join(home, ".windsurf", "mcp_config.json")
	if _, err := os.Stat(windsurfConfig); os.IsNotExist(err) {
		t.Error("Windsurf mcp_config.json was not created")
	}
}
