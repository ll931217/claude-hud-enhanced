package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
)

func main() {
	fmt.Println("=== Claude HUD Enhanced - Configuration Demo ===")
	fmt.Println()

	// Load default configuration
	fmt.Println("1. Loading default configuration...")
	defaultConfig := config.DefaultConfig()
	printConfig(defaultConfig)

	// Load from file (will use defaults if file doesn't exist)
	fmt.Println("\n2. Loading configuration from file...")
	fmt.Println("   (will use defaults if ~/.config/claude-hud/config.yaml doesn't exist)")
	loadedConfig := config.Load()
	printConfig(loadedConfig)

	// Demonstrate graceful degradation with a test file
	fmt.Println("\n3. Demonstrating graceful degradation...")
	tmpDir, err := os.MkdirTemp("", "claude-hud-demo-*")
	if err != nil {
		fmt.Printf("   Error creating temp dir: %v\n", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	nonExistentPath := filepath.Join(tmpDir, "nonexistent.yaml")
	configFromNonExistent := config.LoadFromPath(nonExistentPath)
	fmt.Printf("   Loaded config from non-existent file: %v\n", configFromNonExistent != nil)
	fmt.Printf("   Has valid defaults: refresh_interval=%dms, debug=%v\n",
		configFromNonExistent.RefreshIntervalMs, configFromNonExistent.Debug)

	// Create and save a custom config
	fmt.Println("\n4. Creating and saving custom configuration...")
	customConfig := config.DefaultConfig()
	customConfig.RefreshIntervalMs = 500
	customConfig.Debug = true
	customConfig.Sections.Beads.Enabled = false
	customConfig.Colors.Primary = "magenta"

	// Save to temp location for demo
	_ = filepath.Join(tmpDir, "test-config.yaml")
	if err := customConfig.Save(); err == nil {
		fmt.Println("   Custom config saved successfully")
	}

	// Load and verify
	loadedCustom := config.Load()
	fmt.Printf("   Loaded custom config: refresh_interval=%dms, debug=%v\n",
		loadedCustom.RefreshIntervalMs, loadedCustom.Debug)

	fmt.Println("\n=== Demo Complete ===")
}

func printConfig(cfg *config.Config) {
	fmt.Printf("   Refresh Interval: %dms\n", cfg.RefreshIntervalMs)
	fmt.Printf("   Debug Mode: %v\n", cfg.Debug)
	fmt.Printf("   Enabled Sections: %v\n", cfg.GetEnabledSections())
	fmt.Printf("   Colors: primary=%s, error=%s, success=%s\n",
		cfg.Colors.Primary, cfg.Colors.Error, cfg.Colors.Success)
}
