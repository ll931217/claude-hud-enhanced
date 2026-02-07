package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Check refresh interval
	if config.RefreshIntervalMs != 300 {
		t.Errorf("Expected refresh interval 300ms, got %d", config.RefreshIntervalMs)
	}

	// Check debug mode
	if config.Debug {
		t.Error("Expected debug mode to be false by default")
	}

	tests := []struct {
		name     string
		enabled  bool
		order    int
		checkFn  func() (bool, int)
	}{
		{"model", true, 1, func() (bool, int) { return config.Sections.Model.Enabled, config.Sections.Model.Order }},
		{"contextbar", true, 2, func() (bool, int) { return config.Sections.ContextBar.Enabled, config.Sections.ContextBar.Order }},
		{"duration", true, 3, func() (bool, int) { return config.Sections.Duration.Enabled, config.Sections.Duration.Order }},
		{"beads", true, 4, func() (bool, int) { return config.Sections.Beads.Enabled, config.Sections.Beads.Order }},
		{"status", true, 5, func() (bool, int) { return config.Sections.Status.Enabled, config.Sections.Status.Order }},
		{"workspace", true, 6, func() (bool, int) { return config.Sections.Workspace.Enabled, config.Sections.Workspace.Order }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enabled, order := tt.checkFn()
			if enabled != tt.enabled {
				t.Errorf("Expected %s enabled=%v, got %v", tt.name, tt.enabled, enabled)
			}
			if order != tt.order {
				t.Errorf("Expected %s order=%d, got %d", tt.name, tt.order, order)
			}
		})
	}

	// Check colors - Catppuccin Mocha theme
	colorTests := []struct {
		field   string
		value   string
		checkFn func() string
	}{
		{"Primary", "#89dceb", func() string { return config.Colors.Primary }},
		{"Secondary", "#cba6f7", func() string { return config.Colors.Secondary }},
		{"Error", "#f38ba8", func() string { return config.Colors.Error }},
		{"Warning", "#fab387", func() string { return config.Colors.Warning }},
		{"Info", "#b4befe", func() string { return config.Colors.Info }},
		{"Success", "#a6e3a1", func() string { return config.Colors.Success }},
		{"Muted", "#6c7086", func() string { return config.Colors.Muted }},
	}

	for _, tt := range colorTests {
		t.Run(tt.field, func(t *testing.T) {
			value := tt.checkFn()
			if value != tt.value {
				t.Errorf("Expected Colors.%s=%s, got %s", tt.field, tt.value, value)
			}
		})
	}
}

func TestLoadFromPath_NonExistent(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	config := LoadFromPath(configPath)

	if config == nil {
		t.Fatal("LoadFromPath returned nil for non-existent file")
	}

	// Should return default config
	if config.RefreshIntervalMs != 300 {
		t.Errorf("Expected default refresh interval, got %d", config.RefreshIntervalMs)
	}
}

func TestLoadFromPath_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	config := LoadFromPath(configPath)

	if config == nil {
		t.Fatal("LoadFromPath returned nil for invalid YAML")
	}

	// Should return default config on parse error
	if config.RefreshIntervalMs != 300 {
		t.Errorf("Expected default refresh interval on parse error, got %d", config.RefreshIntervalMs)
	}
}

func TestLoadFromPath_ValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "valid.yaml")

	yamlContent := `
sections:
  model:
    enabled: true
    order: 1
  beads:
    enabled: false
    order: 2
  status:
    enabled: true
    order: 3
  workspace:
    enabled: false
    order: 4
colors:
  primary: "cyan"
  secondary: "magenta"
  error: "red"
  warning: "yellow"
  info: "blue"
  success: "green"
  muted: "gray"
refresh_interval_ms: 500
debug: true
`

	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	config := LoadFromPath(configPath)

	if config == nil {
		t.Fatal("LoadFromPath returned nil")
	}

	// Check custom values
	if config.RefreshIntervalMs != 500 {
		t.Errorf("Expected refresh interval 500, got %d", config.RefreshIntervalMs)
	}

	if !config.Debug {
		t.Error("Expected debug mode to be true")
	}

	if config.Sections.Beads.Enabled {
		t.Error("Expected beads section to be disabled")
	}

	if config.Sections.Model.Enabled {
		// This is expected
	} else {
		t.Error("Expected model section to be enabled")
	}

	if config.Colors.Primary != "cyan" {
		t.Errorf("Expected primary color 'cyan', got '%s'", config.Colors.Primary)
	}
}

func TestLoadFromPath_PartialConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "partial.yaml")

	// Write partial YAML - only override some values
	yamlContent := `
sections:
  beads:
    enabled: false
    order: 2
refresh_interval_ms: 1000
`

	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	config := LoadFromPath(configPath)

	if config == nil {
		t.Fatal("LoadFromPath returned nil")
	}

	// Check overridden values
	if config.RefreshIntervalMs != 1000 {
		t.Errorf("Expected refresh interval 1000, got %d", config.RefreshIntervalMs)
	}

	if config.Sections.Beads.Enabled {
		t.Error("Expected beads section to be disabled")
	}

	// Check default values are still present
	if config.Sections.Model.Enabled != true {
		t.Error("Expected model section to use default enabled=true")
	}

	if config.Colors.Primary != "#89dceb" {
		t.Errorf("Expected default primary color '#89dceb', got '%s'", config.Colors.Primary)
	}
}

func TestValidate_RefreshIntervalClamping(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"Too low", 50, 100},
		{"Too high", 10000, 5000},
		{"Valid low", 100, 100},
		{"Valid high", 5000, 5000},
		{"Valid middle", 500, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.RefreshIntervalMs = tt.input
			config.validate()

			if config.RefreshIntervalMs != tt.expected {
				t.Errorf("Expected refresh interval %d after validation, got %d",
					tt.expected, config.RefreshIntervalMs)
			}
		})
	}
}

func TestValidate_ColorDefaults(t *testing.T) {
	config := DefaultConfig()

	// Set some colors to empty
	config.Colors.Primary = ""
	config.Colors.Error = ""

	config.validate()

	if config.Colors.Primary != "#89dceb" {
		t.Errorf("Expected default primary color '#89dceb', got '%s'", config.Colors.Primary)
	}

	if config.Colors.Error != "#f38ba8" {
		t.Errorf("Expected default error color '#f38ba8', got '%s'", config.Colors.Error)
	}
}

func TestNormalizeSectionOrders(t *testing.T) {
	config := DefaultConfig()

	// Disable new sections for this test to maintain expected behavior
	config.Sections.ContextBar.Enabled = false
	config.Sections.Duration.Enabled = false
	config.Sections.Tools.Enabled = false
	config.Sections.SysInfo.Enabled = false

	// Set custom orders with gaps and duplicates
	config.Sections.Model.Order = 10
	config.Sections.Beads.Order = 5
	config.Sections.Status.Order = 10
	config.Sections.Workspace.Order = 1

	config.normalizeSectionOrders()

	// Check orders are normalized starting from 1
	orders := []struct {
		name   string
		order  int
		checkFn func() int
	}{
		{"model", 3, func() int { return config.Sections.Model.Order }},
		{"beads", 2, func() int { return config.Sections.Beads.Order }},
		{"status", 4, func() int { return config.Sections.Status.Order }},
		{"workspace", 1, func() int { return config.Sections.Workspace.Order }},
	}

	for _, tt := range orders {
		t.Run(tt.name, func(t *testing.T) {
			order := tt.checkFn()
			if order != tt.order {
				t.Errorf("Expected %s order=%d after normalization, got %d",
					tt.name, tt.order, order)
			}
		})
	}
}

func TestGetEnabledSections(t *testing.T) {
	config := DefaultConfig()

	// Disable some sections (including new ones)
	config.Sections.ContextBar.Enabled = false
	config.Sections.Duration.Enabled = false
	config.Sections.Beads.Enabled = false
	config.Sections.Workspace.Enabled = false
	config.Sections.Tools.Enabled = false
	config.Sections.SysInfo.Enabled = false

	enabled := config.GetEnabledSections()

	expected := []string{"model", "status"}
	if len(enabled) != len(expected) {
		t.Fatalf("Expected %d enabled sections, got %d", len(expected), len(enabled))
	}

	for i, section := range enabled {
		if section != expected[i] {
			t.Errorf("Expected section %d to be '%s', got '%s'", i, expected[i], section)
		}
	}
}

func TestGetEnabledSections_CustomOrder(t *testing.T) {
	config := DefaultConfig()

	// Change order (include new sections)
	config.Sections.Model.Order = 3
	config.Sections.ContextBar.Order = 2
	config.Sections.Duration.Order = 8
	config.Sections.Beads.Order = 1
	config.Sections.Status.Order = 5
	config.Sections.Workspace.Order = 4
	config.Sections.Tools.Order = 6
	config.Sections.SysInfo.Order = 7

	enabled := config.GetEnabledSections()

	expected := []string{"beads", "contextbar", "model", "workspace", "status", "tools", "sysinfo", "duration"}
	if len(enabled) != len(expected) {
		t.Fatalf("Expected %d enabled sections, got %d", len(expected), len(enabled))
	}

	for i, section := range enabled {
		if section != expected[i] {
			t.Errorf("Expected section %d to be '%s', got '%s'", i, expected[i], section)
		}
	}
}

func TestIsSectionEnabled(t *testing.T) {
	config := DefaultConfig()

	tests := []struct {
		section string
		enabled bool
	}{
		{"model", true},
		{"beads", true},
		{"status", true},
		{"workspace", true},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.section, func(t *testing.T) {
			result := config.IsSectionEnabled(tt.section)
			if result != tt.enabled {
				t.Errorf("Expected IsSectionEnabled(%s)=%v, got %v",
					tt.section, tt.enabled, result)
			}
		})
	}
}

func TestGetRefreshInterval(t *testing.T) {
	config := DefaultConfig()
	config.RefreshIntervalMs = 500

	interval := config.GetRefreshInterval()

	expected := 500 * time.Millisecond
	if interval != expected {
		t.Errorf("Expected refresh interval %v, got %v", expected, interval)
	}
}

func TestGetSectionOrder(t *testing.T) {
	config := DefaultConfig()

	config.Sections.Model.Order = 10
	config.Sections.ContextBar.Order = 20
	config.Sections.Duration.Order = 30
	config.Sections.Beads.Order = 40
	config.Sections.Status.Order = 50
	config.Sections.Workspace.Order = 60

	tests := []struct {
		section string
		order   int
	}{
		{"model", 10},
		{"contextbar", 20},
		{"duration", 30},
		{"beads", 40},
		{"status", 50},
		{"workspace", 60},
		{"nonexistent", 999},
	}

	for _, tt := range tests {
		t.Run(tt.section, func(t *testing.T) {
			result := config.GetSectionOrder(tt.section)
			if result != tt.order {
				t.Errorf("Expected GetSectionOrder(%s)=%d, got %d",
					tt.section, tt.order, result)
			}
		})
	}
}

func TestToYAML(t *testing.T) {
	config := DefaultConfig()

	yaml, err := config.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML returned error: %v", err)
	}

	if yaml == "" {
		t.Error("ToYAML returned empty string")
	}

	// Check that it contains expected keys
	expectedKeys := []string{
		"sections:",
		"model:",
		"beads:",
		"colors:",
		"primary:",
		"refresh_interval_ms:",
	}

	for _, key := range expectedKeys {
		if !contains(yaml, key) {
			t.Errorf("Expected YAML to contain '%s'", key)
		}
	}
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()

	// Override the config path for testing
	homeDir := os.Getenv("HOME")
	defer func() {
		_ = os.Setenv("HOME", homeDir)
	}()
	_ = os.Setenv("HOME", tmpDir)

	config := DefaultConfig()
	config.RefreshIntervalMs = 750
	config.Debug = true

	err := config.Save()
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(tmpDir, ".config", "claude-hud", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load and verify
	loadedConfig := LoadFromPath(configPath)
	if loadedConfig.RefreshIntervalMs != 750 {
		t.Errorf("Expected saved refresh interval 750, got %d", loadedConfig.RefreshIntervalMs)
	}

	if !loadedConfig.Debug {
		t.Error("Expected saved debug mode to be true")
	}
}

func TestLoad_GracefulDegradation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (string, func())
		wantErr bool
	}{
		{
			name: "File not found",
			setup: func() (string, func()) {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "nonexistent.yaml")
				return path, func() {}
			},
			wantErr: false,
		},
		{
			name: "Invalid YAML syntax",
			setup: func() (string, func()) {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "invalid.yaml")
				_ = os.WriteFile(path, []byte("invalid: yaml: ["), 0644)
				return path, func() {}
			},
			wantErr: false,
		},
		{
			name: "Valid YAML with invalid types",
			setup: func() (string, func()) {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "badtypes.yaml")
				_ = os.WriteFile(path, []byte("refresh_interval_ms: \"not a number\""), 0644)
				return path, func() {}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := tt.setup()
			defer cleanup()

			config := LoadFromPath(path)

			// Should never return nil - always returns a valid config
			if config == nil {
				t.Error("LoadFromPath returned nil, should always return valid config")
			}

			// Should have default values as fallback
			if config.RefreshIntervalMs < 100 || config.RefreshIntervalMs > 5000 {
				t.Errorf("Config should have valid refresh interval, got %d", config.RefreshIntervalMs)
			}
		})
	}
}

func TestLoadWithMissingOptionalFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "minimal.yaml")

	// Minimal valid YAML with only some fields
	yamlContent := `
sections:
  model:
    enabled: true
    order: 1
refresh_interval_ms: 250
`

	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	config := LoadFromPath(configPath)

	if config == nil {
		t.Fatal("LoadFromPath returned nil")
	}

	// Check that default values filled in missing fields
	if config.Colors.Primary == "" {
		t.Error("Expected default primary color to be filled in")
	}

	if config.Colors.Error == "" {
		t.Error("Expected default error color to be filled in")
	}

	// Check that specified values were used
	if config.RefreshIntervalMs != 250 {
		t.Errorf("Expected refresh interval 250, got %d", config.RefreshIntervalMs)
	}

	if !config.Sections.Model.Enabled {
		t.Error("Expected model section to be enabled")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
