package config

import (
	"os"
	"testing"
)

// TestConfigMatchesSpecification verifies the config structure matches the exact YAML spec
func TestConfigMatchesSpecification(t *testing.T) {
	yamlContent := `sections:
  session:
    enabled: true
    order: 1
  beads:
    enabled: true
    order: 2
  status:
    enabled: true
    order: 3
  workspace:
    enabled: true
    order: 4
colors:
  primary: "blue"
  secondary: "green"
  error: "red"
  warning: "yellow"
  info: "cyan"
  success: "green"
  muted: "gray"
refresh_interval_ms: 300
debug: false
`

	// Write test config
	tmpFile := "/tmp/test_spec_config.yaml"
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	defer os.Remove(tmpFile)

	// Load and verify
	cfg := LoadFromPath(tmpFile)

	// Verify all fields match the specification
	if cfg.RefreshIntervalMs != 300 {
		t.Errorf("Expected refresh_interval_ms=300, got %d", cfg.RefreshIntervalMs)
	}

	if cfg.Debug != false {
		t.Errorf("Expected debug=false, got %v", cfg.Debug)
	}

	// Verify sections
	sections := cfg.GetEnabledSections()
	if len(sections) != 6 {
		t.Errorf("Expected 6 enabled sections, got %d", len(sections))
	}

	// Verify colors
	tests := []struct {
		field string
		value string
		got   string
	}{
		{"primary", "blue", cfg.Colors.Primary},
		{"secondary", "green", cfg.Colors.Secondary},
		{"error", "red", cfg.Colors.Error},
		{"warning", "yellow", cfg.Colors.Warning},
		{"info", "cyan", cfg.Colors.Info},
		{"success", "green", cfg.Colors.Success},
		{"muted", "gray", cfg.Colors.Muted},
	}

	for _, tt := range tests {
		if tt.got != tt.value {
			t.Errorf("Expected color %s=%s, got %s", tt.field, tt.value, tt.got)
		}
	}

	t.Log("✅ Config structure matches specification exactly!")
}
