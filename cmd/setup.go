package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Auto-configure IDE",
	Long: `Auto-detect and configure IDEs for Phloem.

Without arguments, auto-detects installed IDEs and configures them.
Specify an IDE to configure only that one.

Examples:
  phloem setup              # auto-detect and configure all IDEs
  phloem setup cursor       # configure Cursor only
  phloem setup windsurf     # configure Windsurf only
  phloem setup claude-code  # configure Claude Code only`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetup()
	},
}

func init() {
	setupCmd.AddCommand(&cobra.Command{
		Use:   "cursor",
		Short: "Configure Phloem for Cursor",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupCursor()
		},
	})

	setupCmd.AddCommand(&cobra.Command{
		Use:   "windsurf",
		Short: "Configure Phloem for Windsurf",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupWindsurf()
		},
	})

	setupCmd.AddCommand(&cobra.Command{
		Use:   "claude-code",
		Short: "Configure Phloem for Claude Code",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupClaudeCode()
		},
	})

	setupCmd.AddCommand(&cobra.Command{
		Use:   "vscode",
		Short: "Configure Phloem for VS Code (GitHub Copilot)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupVSCode()
		},
	})

	setupCmd.AddCommand(&cobra.Command{
		Use:   "zed",
		Short: "Configure Phloem for Zed",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupZed()
		},
	})

	setupCmd.AddCommand(&cobra.Command{
		Use:   "cline",
		Short: "Configure Phloem for Cline (VS Code extension)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupCline()
		},
	})

	setupCmd.AddCommand(&cobra.Command{
		Use:   "neovim",
		Short: "Configure Phloem for Neovim (mcphub.nvim)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupNeovim()
		},
	})

	setupCmd.AddCommand(&cobra.Command{
		Use:   "warp",
		Short: "Show Warp MCP setup instructions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupWarp()
		},
	})
}

// runSetup auto-detects and configures IDEs
func runSetup() error {
	fmt.Println("🔍 Auto-detecting IDEs for Phloem setup...")
	fmt.Println()

	home, _ := os.UserHomeDir()
	detected := 0

	// Check for Cursor
	cursorDir := filepath.Join(home, ".cursor")
	if _, err := os.Stat(cursorDir); err == nil {
		fmt.Println("👉 Detected Cursor")
		if err := runSetupCursor(); err != nil {
			fmt.Printf("   ❌ Cursor setup failed: %v\n", err)
		} else {
			detected++
		}
	}

	// Check for Windsurf
	windsurfDir := filepath.Join(home, ".windsurf")
	if _, err := os.Stat(windsurfDir); err == nil {
		fmt.Println("👉 Detected Windsurf")
		if err := runSetupWindsurf(); err != nil {
			fmt.Printf("   ❌ Windsurf setup failed: %v\n", err)
		} else {
			detected++
		}
	}

	// Check for Claude Code
	if _, err := exec.LookPath("claude"); err == nil {
		fmt.Println("👉 Detected Claude Code")
		if err := runSetupClaudeCode(); err != nil {
			fmt.Printf("   ❌ Claude Code setup failed: %v\n", err)
		} else {
			detected++
		}
	}

	// Check for VS Code
	vscodeMCPPath := vscodeMCPConfigPath()
	if vscodeMCPPath != "" {
		// Check if VS Code is installed by looking for its config parent dir
		if _, err := os.Stat(filepath.Dir(vscodeMCPPath)); err == nil {
			fmt.Println("👉 Detected VS Code")
			if err := runSetupVSCode(); err != nil {
				fmt.Printf("   ❌ VS Code setup failed: %v\n", err)
			} else {
				detected++
			}
		}
	}

	// Check for Zed
	zedSettingsPath := zedSettingsFilePath()
	if _, err := os.Stat(filepath.Dir(zedSettingsPath)); err == nil {
		fmt.Println("👉 Detected Zed")
		if err := runSetupZed(); err != nil {
			fmt.Printf("   ❌ Zed setup failed: %v\n", err)
		} else {
			detected++
		}
	}

	// Check for Cline
	clinePath := clineMCPConfigPath()
	if clinePath != "" {
		if _, err := os.Stat(filepath.Dir(clinePath)); err == nil {
			fmt.Println("👉 Detected Cline")
			if err := runSetupCline(); err != nil {
				fmt.Printf("   ❌ Cline setup failed: %v\n", err)
			} else {
				detected++
			}
		}
	}

	// Check for Neovim (mcphub.nvim)
	neovimPath := neovimMCPConfigPath()
	if _, err := os.Stat(filepath.Dir(neovimPath)); err == nil {
		fmt.Println("👉 Detected Neovim (mcphub)")
		if err := runSetupNeovim(); err != nil {
			fmt.Printf("   ❌ Neovim setup failed: %v\n", err)
		} else {
			detected++
		}
	}

	if detected == 0 {
		fmt.Println("⚠️  No IDEs automatically detected.")
		fmt.Println("   You can still manually setup using:")
		fmt.Println("   phloem setup cursor")
		fmt.Println("   phloem setup windsurf")
		fmt.Println("   phloem setup claude-code")
	} else {
		fmt.Printf("\n✅ Successfully configured %d IDE(s)!\n", detected)
	}

	return nil
}

// runSetupCursor auto-configures Cursor MCP settings
func runSetupCursor() error {
	fmt.Println("🔧 Setting up Phloem for Cursor...")
	fmt.Println()

	// 1. Find phloem binary path
	phloemPath, err := exec.LookPath("phloem")
	if err != nil {
		return fmt.Errorf("phloem binary not found in PATH. Please install via Homebrew or add to PATH")
	}
	fmt.Printf("✓ Found phloem at: %s\n", phloemPath)

	// 2. Locate Cursor config directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	cursorDir := filepath.Join(home, ".cursor")
	configPath := filepath.Join(cursorDir, "mcp.json")

	// 3. Create .cursor directory if it doesn't exist
	if _, err := os.Stat(cursorDir); os.IsNotExist(err) {
		fmt.Printf("✓ Creating Cursor config directory: %s\n", cursorDir)
		if err := os.MkdirAll(cursorDir, 0755); err != nil {
			return fmt.Errorf("failed to create .cursor directory: %w", err)
		}
	}

	// 4. Read existing config or create new one
	var config map[string]interface{}
	if data, err := os.ReadFile(configPath); err == nil {
		// Config exists, parse it
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse existing mcp.json: %w", err)
		}
		fmt.Println("✓ Found existing mcp.json")
	} else {
		// Create new config
		config = make(map[string]interface{})
		fmt.Println("✓ Creating new mcp.json")
	}

	// 5. Add or update phloem server
	if config["mcpServers"] == nil {
		config["mcpServers"] = make(map[string]interface{})
	}

	mcpServers := config["mcpServers"].(map[string]interface{})
	mcpServers["phloem"] = map[string]interface{}{
		"command": phloemPath,
		"args":    []string{"serve"},
	}

	// 6. Write config back
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write mcp.json: %w", err)
	}

	fmt.Printf("✓ Updated mcp.json: %s\n", configPath)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Phloem is now configured for Cursor!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Restart Cursor")
	fmt.Println("  2. Open a file and start coding")
	fmt.Println("  3. Phloem will automatically remember your conversations")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  phloem status   - View memory statistics")
	fmt.Println("  phloem help     - See all commands")
	fmt.Println()
	fmt.Println("Tip: You can also set up from within your IDE.")
	fmt.Println("  Ask your AI assistant: \"please run phloem setup cursor in a terminal\"")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return nil
}

// runSetupClaudeCode registers Phloem as an MCP server in Claude Code
func runSetupClaudeCode() error {
	fmt.Println("🔧 Setting up Phloem for Claude Code...")
	fmt.Println()

	// 1. Find claude binary
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude binary not found in PATH. Install Claude Code first")
	}
	fmt.Printf("✓ Found claude at: %s\n", claudePath)

	// 2. Find phloem binary path
	phloemPath, err := exec.LookPath("phloem")
	if err != nil {
		return fmt.Errorf("phloem binary not found in PATH. Please install via Homebrew or add to PATH")
	}
	fmt.Printf("✓ Found phloem at: %s\n", phloemPath)

	// 3. Check if phloem is already registered
	fmt.Print("✓ Checking existing MCP registrations... ")
	listCmd := exec.Command(claudePath, "mcp", "list")
	listOutput, err := listCmd.CombinedOutput()
	if err != nil {
		fmt.Println("⚠️  Could not list MCP servers (continuing)")
	} else if strings.Contains(string(listOutput), "phloem") {
		fmt.Println("already registered")
		fmt.Println()
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("✅ Phloem is already configured for Claude Code!")
		fmt.Println()
		fmt.Println("To re-register, first remove:")
		fmt.Println("  claude mcp remove phloem")
		fmt.Println("Then run:")
		fmt.Println("  phloem setup claude-code")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		return nil
	} else {
		fmt.Println("not yet registered")
	}

	// 4. Determine PHLOEM_DATA_DIR
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	dataDir := filepath.Join(home, ".phloem")

	// 5. Register phloem MCP server with Claude Code
	fmt.Print("✓ Registering phloem MCP server... ")
	addCmd := exec.Command(claudePath, "mcp", "add",
		"-e", "PHLOEM_DATA_DIR="+dataDir,
		"--scope", "user",
		"phloem",
		"--",
		phloemPath, "serve",
	)
	addOutput, err := addCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to register MCP server: %w\nOutput: %s", err, string(addOutput))
	}
	fmt.Println("done")

	// 6. Verify registration
	fmt.Print("✓ Verifying registration... ")
	verifyCmd := exec.Command(claudePath, "mcp", "list")
	verifyOutput, err := verifyCmd.CombinedOutput()
	if err != nil {
		fmt.Println("⚠️  Could not verify (may still be registered)")
	} else if strings.Contains(string(verifyOutput), "phloem") {
		fmt.Println("confirmed")
	} else {
		return fmt.Errorf("phloem not found in MCP list after registration")
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Phloem is now configured for Claude Code!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Start a new Claude Code session")
	fmt.Println("  2. MCP server will auto-start on first tool use")
	fmt.Println("  3. Use 'remember' tool to store memories")
	fmt.Println("  4. Use 'recall' tool to search memories")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  phloem status   - View memory statistics")
	fmt.Println("  phloem help     - See all commands")
	fmt.Println()
	fmt.Println("Tip: You can also set up from within your IDE.")
	fmt.Println("  Ask your AI assistant: \"please run phloem setup claude-code in a terminal\"")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return nil
}

// runSetupWindsurf auto-configures Windsurf MCP settings
func runSetupWindsurf() error {
	fmt.Println("🔧 Setting up Phloem for Windsurf...")
	fmt.Println()

	// 1. Find phloem binary path
	phloemPath, err := exec.LookPath("phloem")
	if err != nil {
		return fmt.Errorf("phloem binary not found in PATH. Please install via Homebrew or add to PATH")
	}
	fmt.Printf("✓ Found phloem at: %s\n", phloemPath)

	// 2. Locate Windsurf config directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	windsurfDir := filepath.Join(home, ".windsurf")
	configPath := filepath.Join(windsurfDir, "mcp_config.json")

	// 3. Create .windsurf directory if it doesn't exist
	if _, err := os.Stat(windsurfDir); os.IsNotExist(err) {
		fmt.Printf("✓ Creating Windsurf config directory: %s\n", windsurfDir)
		if err := os.MkdirAll(windsurfDir, 0755); err != nil {
			return fmt.Errorf("failed to create .windsurf directory: %w", err)
		}
	}

	// 4. Read existing config or create new one
	var config map[string]interface{}
	if data, err := os.ReadFile(configPath); err == nil {
		// Config exists, parse it
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse existing mcp_config.json: %w", err)
		}
		fmt.Println("✓ Found existing mcp_config.json")
	} else {
		// Create new config
		config = make(map[string]interface{})
		fmt.Println("✓ Creating new mcp_config.json")
	}

	// 5. Add or update phloem server
	if config["mcpServers"] == nil {
		config["mcpServers"] = make(map[string]interface{})
	}

	mcpServers := config["mcpServers"].(map[string]interface{})
	mcpServers["phloem"] = map[string]interface{}{
		"command": phloemPath,
		"args":    []string{"serve"},
	}

	// 6. Write config back
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write mcp_config.json: %w", err)
	}

	fmt.Printf("✓ Updated mcp_config.json: %s\n", configPath)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Phloem is now configured for Windsurf!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Restart Windsurf")
	fmt.Println("  2. Open a file and start coding")
	fmt.Println("  3. Phloem will automatically remember your conversations")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  phloem status   - View memory statistics")
	fmt.Println("  phloem help     - See all commands")
	fmt.Println()
	fmt.Println("Tip: You can also set up from within your IDE.")
	fmt.Println("  Ask your AI assistant: \"please run phloem setup windsurf in a terminal\"")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return nil
}

// vscodeMCPConfigPath returns the user-level VS Code MCP config path
func vscodeMCPConfigPath() string {
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "Code", "User", "mcp.json")
	case "linux":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "Code", "User", "mcp.json")
	case "windows":
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			return ""
		}
		return filepath.Join(appdata, "Code", "User", "mcp.json")
	}
	return ""
}

// runSetupVSCode configures VS Code (GitHub Copilot) MCP settings
func runSetupVSCode() error {
	fmt.Println("🔧 Setting up Phloem for VS Code...")
	fmt.Println()

	phloemPath, err := exec.LookPath("phloem")
	if err != nil {
		return fmt.Errorf("phloem binary not found in PATH")
	}
	fmt.Printf("✓ Found phloem at: %s\n", phloemPath)

	configPath := vscodeMCPConfigPath()
	if configPath == "" {
		return fmt.Errorf("could not determine VS Code config path for this platform")
	}

	// Create parent directory if needed
	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Read or create config
	var config map[string]interface{}
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse existing mcp.json: %w", err)
		}
		fmt.Println("✓ Found existing mcp.json")
	} else {
		config = make(map[string]interface{})
		fmt.Println("✓ Creating new mcp.json")
	}

	// VS Code uses "servers" key (not "mcpServers")
	if config["servers"] == nil {
		config["servers"] = make(map[string]interface{})
	}
	servers := config["servers"].(map[string]interface{})
	servers["phloem"] = map[string]interface{}{
		"type":    "stdio",
		"command": phloemPath,
		"args":    []string{"serve"},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write mcp.json: %w", err)
	}

	fmt.Printf("✓ Updated: %s\n", configPath)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Phloem is now configured for VS Code!")
	fmt.Println()
	fmt.Println("Requires VS Code 1.99+ with GitHub Copilot (Agent Mode).")
	fmt.Println("No restart needed — VS Code detects config changes automatically.")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return nil
}

// zedSettingsFilePath returns the Zed settings.json path
func zedSettingsFilePath() string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, ".zed", "settings.json")
	case "linux":
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, "zed", "settings.json")
		}
		return filepath.Join(home, ".config", "zed", "settings.json")
	}
	return filepath.Join(home, ".zed", "settings.json")
}

// runSetupZed configures Zed MCP settings
func runSetupZed() error {
	fmt.Println("🔧 Setting up Phloem for Zed...")
	fmt.Println()

	phloemPath, err := exec.LookPath("phloem")
	if err != nil {
		return fmt.Errorf("phloem binary not found in PATH")
	}
	fmt.Printf("✓ Found phloem at: %s\n", phloemPath)

	settingsPath := zedSettingsFilePath()
	settingsDir := filepath.Dir(settingsPath)

	if _, err := os.Stat(settingsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(settingsDir, 0755); err != nil {
			return fmt.Errorf("failed to create Zed config directory: %w", err)
		}
	}

	// Read or create settings
	var settings map[string]interface{}
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("failed to parse existing settings.json: %w", err)
		}
		fmt.Println("✓ Found existing settings.json")
	} else {
		settings = make(map[string]interface{})
		fmt.Println("✓ Creating new settings.json")
	}

	// Zed uses "context_servers" key
	if settings["context_servers"] == nil {
		settings["context_servers"] = make(map[string]interface{})
	}
	servers := settings["context_servers"].(map[string]interface{})
	servers["phloem"] = map[string]interface{}{
		"command": phloemPath,
		"args":    []string{"serve"},
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	if err := os.WriteFile(settingsPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write settings.json: %w", err)
	}

	fmt.Printf("✓ Updated: %s\n", settingsPath)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Phloem is now configured for Zed!")
	fmt.Println()
	fmt.Println("No restart needed — Zed hot-reloads settings.")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return nil
}

// clineMCPConfigPath returns the Cline MCP settings path
func clineMCPConfigPath() string {
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "Code", "User",
			"globalStorage", "saoudrizwan.claude-dev", "settings", "cline_mcp_settings.json")
	case "linux":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "Code", "User",
			"globalStorage", "saoudrizwan.claude-dev", "settings", "cline_mcp_settings.json")
	case "windows":
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			return ""
		}
		return filepath.Join(appdata, "Code", "User",
			"globalStorage", "saoudrizwan.claude-dev", "settings", "cline_mcp_settings.json")
	}
	return ""
}

// runSetupCline configures Cline (VS Code extension) MCP settings
func runSetupCline() error {
	fmt.Println("🔧 Setting up Phloem for Cline...")
	fmt.Println()

	phloemPath, err := exec.LookPath("phloem")
	if err != nil {
		return fmt.Errorf("phloem binary not found in PATH")
	}
	fmt.Printf("✓ Found phloem at: %s\n", phloemPath)

	configPath := clineMCPConfigPath()
	if configPath == "" {
		return fmt.Errorf("could not determine Cline config path for this platform")
	}

	// Create parent directory if needed
	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create Cline config directory: %w", err)
		}
	}

	var config map[string]interface{}
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse existing cline_mcp_settings.json: %w", err)
		}
		fmt.Println("✓ Found existing cline_mcp_settings.json")
	} else {
		config = make(map[string]interface{})
		fmt.Println("✓ Creating new cline_mcp_settings.json")
	}

	if config["mcpServers"] == nil {
		config["mcpServers"] = make(map[string]interface{})
	}
	mcpServers := config["mcpServers"].(map[string]interface{})
	mcpServers["phloem"] = map[string]interface{}{
		"command":  phloemPath,
		"args":     []string{"serve"},
		"disabled": false,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write cline_mcp_settings.json: %w", err)
	}

	fmt.Printf("✓ Updated: %s\n", configPath)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Phloem is now configured for Cline!")
	fmt.Println()
	fmt.Println("No restart needed — Cline detects config changes automatically.")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return nil
}

// neovimMCPConfigPath returns the mcphub.nvim servers.json path
func neovimMCPConfigPath() string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "linux":
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, "mcphub", "servers.json")
		}
		return filepath.Join(home, ".config", "mcphub", "servers.json")
	default: // darwin, windows
		return filepath.Join(home, ".config", "mcphub", "servers.json")
	}
}

// runSetupNeovim configures mcphub.nvim servers.json
func runSetupNeovim() error {
	fmt.Println("🔧 Setting up Phloem for Neovim (mcphub.nvim)...")
	fmt.Println()

	phloemPath, err := exec.LookPath("phloem")
	if err != nil {
		return fmt.Errorf("phloem binary not found in PATH")
	}
	fmt.Printf("✓ Found phloem at: %s\n", phloemPath)

	configPath := neovimMCPConfigPath()
	configDir := filepath.Dir(configPath)

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create mcphub config directory: %w", err)
		}
	}

	// Read or create config
	var config map[string]interface{}
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse existing servers.json: %w", err)
		}
		fmt.Println("✓ Found existing servers.json")
	} else {
		config = make(map[string]interface{})
		fmt.Println("✓ Creating new servers.json")
	}

	// mcphub uses "mcpServers" key
	if config["mcpServers"] == nil {
		config["mcpServers"] = make(map[string]interface{})
	}
	mcpServers := config["mcpServers"].(map[string]interface{})
	mcpServers["phloem"] = map[string]interface{}{
		"command": phloemPath,
		"args":    []string{"serve"},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write servers.json: %w", err)
	}

	fmt.Printf("✓ Updated: %s\n", configPath)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Phloem is now configured for Neovim!")
	fmt.Println()
	fmt.Println("Requires mcphub.nvim plugin. Install via lazy.nvim:")
	fmt.Println("  { \"ravitemer/mcphub.nvim\", build = \"npm install -g mcp-hub@latest\" }")
	fmt.Println()
	fmt.Println("No restart needed — run :MCPHub in Neovim to reload.")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return nil
}

// runSetupWarp prints Warp setup instructions (no local config file available)
func runSetupWarp() error {
	phloemPath, err := exec.LookPath("phloem")
	if err != nil {
		phloemPath = "/usr/local/bin/phloem"
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Warp MCP Setup (manual)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Warp stores MCP config in the cloud, not on disk.")
	fmt.Println("To add Phloem:")
	fmt.Println()
	fmt.Println("  1. Open Warp Settings > MCP Servers")
	fmt.Println("  2. Click + Add")
	fmt.Println("  3. Paste this JSON:")
	fmt.Println()
	fmt.Printf("  {\n    \"mcpServers\": {\n      \"phloem\": {\n        \"command\": \"%s\",\n        \"args\": [\"serve\"]\n      }\n    }\n  }\n", phloemPath)
	fmt.Println()
	fmt.Println("  4. Click Save")
	fmt.Println()
	fmt.Println("No restart needed — available on next message.")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return nil
}
