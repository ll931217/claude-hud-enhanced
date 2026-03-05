package claudestats

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCollector_Collect(t *testing.T) {
	// Create a temporary directory for test settings
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	// Create test settings file
	settings := map[string]interface{}{
		"enabledPlugins": map[string]bool{
			"skill1": true,
			"skill2": true,
			"skill3": true,
		},
		"hooks": map[string]interface{}{
			"PreToolUse": []map[string]interface{}{
				{
					"matcher": "*",
					"hooks":   []interface{}{map[string]string{"type": "prompt"}},
				},
			},
			"SessionStart": []map[string]interface{}{
				{
					"matcher": "*",
					"hooks":   []interface{}{map[string]string{"type": "prompt"}, map[string]string{"type": "command"}},
				},
			},
		},
	}

	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("Failed to marshal test settings: %v", err)
	}
	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test settings: %v", err)
	}

	// Create collector with test settings path
	collector := &Collector{
		settingsPath: settingsPath,
		mcpClient:    nil, // Will return 0 for MCP count
		cacheTTL:     5 * time.Second,
	}

	// Collect stats
	ctx := context.Background()
	stats := collector.Collect(ctx)

	// Verify results
	if stats.CoreCount != len(coreTools) {
		t.Errorf("Expected CoreCount %d, got %d", len(coreTools), stats.CoreCount)
	}

	// MCP count will be 0 since we don't have a real client
	if stats.MCPCount != 0 {
		t.Errorf("Expected MCPCount 0, got %d", stats.MCPCount)
	}

	// Plugins count should match enabledPlugins
	if stats.PluginsCount != 3 {
		t.Errorf("Expected PluginsCount 3, got %d", stats.PluginsCount)
	}

	// Hooks count should be 3 (1 from PreToolUse + 2 from SessionStart)
	if stats.HooksCount != 3 {
		t.Errorf("Expected HooksCount 3, got %d", stats.HooksCount)
	}
}

func TestCollector_CacheTTL(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	// Create minimal settings
	settings := map[string]interface{}{
		"enabledPlugins": map[string]bool{
			"skill1": true,
		},
		"hooks": map[string]interface{}{
			"SessionStart": []map[string]interface{}{
				{
					"matcher": "*",
					"hooks":   []interface{}{map[string]string{"type": "prompt"}},
				},
			},
		},
	}

	data, _ := json.Marshal(settings)
	os.WriteFile(settingsPath, data, 0644)

	collector := &Collector{
		settingsPath: settingsPath,
		cacheTTL:     100 * time.Millisecond,
	}

	ctx := context.Background()

	// First collection
	stats1 := collector.Collect(ctx)
	firstTimestamp := stats1.Timestamp

	// Immediate second collection should use cache
	stats2 := collector.Collect(ctx)
	if stats2.Timestamp != firstTimestamp {
		t.Error("Cache was not used for immediate second collection")
	}

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Third collection should refresh cache
	stats3 := collector.Collect(ctx)
	if stats3.Timestamp == firstTimestamp {
		t.Error("Cache was not refreshed after TTL expired")
	}
}

func TestCollector_MissingSettingsFile(t *testing.T) {
	// Use a non-existent path
	collector := &Collector{
		settingsPath: "/nonexistent/path/settings.json",
		cacheTTL:     5 * time.Second,
	}

	ctx := context.Background()
	stats := collector.Collect(ctx)

	// Should return 0 for skills and hooks, but still have core count
	if stats.CoreCount == 0 {
		t.Error("CoreCount should not be 0 even with missing settings")
	}
	if stats.PluginsCount != 0 {
		t.Errorf("Expected PluginsCount 0 with missing settings, got %d", stats.PluginsCount)
	}
	if stats.HooksCount != 0 {
		t.Errorf("Expected HooksCount 0 with missing settings, got %d", stats.HooksCount)
	}
}

func TestCollector_InvalidSettingsJSON(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	// Write invalid JSON
	if err := os.WriteFile(settingsPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("Failed to write invalid settings: %v", err)
	}

	collector := &Collector{
		settingsPath: settingsPath,
		cacheTTL:     5 * time.Second,
	}

	ctx := context.Background()
	stats := collector.Collect(ctx)

	// Should return 0 for skills and hooks due to parse error
	if stats.PluginsCount != 0 {
		t.Errorf("Expected PluginsCount 0 with invalid JSON, got %d", stats.PluginsCount)
	}
	if stats.HooksCount != 0 {
		t.Errorf("Expected HooksCount 0 with invalid JSON, got %d", stats.HooksCount)
	}
}

func TestCollector_EmptySettings(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	// Write empty settings
	settings := map[string]interface{}{}
	data, _ := json.Marshal(settings)
	os.WriteFile(settingsPath, data, 0644)

	collector := &Collector{
		settingsPath: settingsPath,
		cacheTTL:     5 * time.Second,
	}

	ctx := context.Background()
	stats := collector.Collect(ctx)

	// Core count should still be available
	if stats.CoreCount == 0 {
		t.Error("CoreCount should not be 0 even with empty settings")
	}
	// Skills and hooks should be 0
	if stats.PluginsCount != 0 {
		t.Errorf("Expected PluginsCount 0 with empty settings, got %d", stats.PluginsCount)
	}
	if stats.HooksCount != 0 {
		t.Errorf("Expected HooksCount 0 with empty settings, got %d", stats.HooksCount)
	}
}

func TestCollector_PluginHooks(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")
	pluginsDir := filepath.Join(tmpDir, "plugins")

	// Create settings with no hooks
	settings := map[string]interface{}{
		"enabledPlugins": map[string]bool{},
		"hooks":          map[string]interface{}{},
	}
	data, _ := json.Marshal(settings)
	os.WriteFile(settingsPath, data, 0644)

	// Create plugin directory structure
	pluginInstallPath := filepath.Join(pluginsDir, "cache", "test-marketplace", "test-plugin", "1.0.0")
	pluginDir := filepath.Join(pluginInstallPath, ".claude-plugin")
	os.MkdirAll(pluginDir, 0755)

	// Create plugin.json with inline hooks
	pluginJSON := map[string]interface{}{
		"name": "test-plugin",
		"hooks": map[string]interface{}{
			"SessionStart": []map[string]interface{}{
				{
					"matcher": "",
					"hooks": []map[string]string{
						{"type": "command", "command": "echo hello"},
					},
				},
			},
			"PreCompact": []map[string]interface{}{
				{
					"matcher": "",
					"hooks": []map[string]string{
						{"type": "command", "command": "echo compact"},
						{"type": "prompt", "prompt": "remember"},
					},
				},
			},
		},
	}
	pluginData, _ := json.Marshal(pluginJSON)
	os.WriteFile(filepath.Join(pluginDir, "plugin.json"), pluginData, 0644)

	// Create installed_plugins.json
	installed := map[string]interface{}{
		"version": 2,
		"plugins": map[string]interface{}{
			"test-plugin@test-marketplace": []map[string]interface{}{
				{"installPath": pluginInstallPath},
			},
		},
	}
	installedData, _ := json.Marshal(installed)
	os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), installedData, 0644)

	collector := &Collector{
		settingsPath: settingsPath,
		pluginsDir:   pluginsDir,
		cacheTTL:     5 * time.Second,
	}

	ctx := context.Background()
	stats := collector.Collect(ctx)

	// Should have 3 plugin hooks (1 from SessionStart + 2 from PreCompact)
	if stats.HooksCount != 3 {
		t.Errorf("Expected HooksCount 3, got %d", stats.HooksCount)
	}
}

func TestCollector_MissingPluginsDir(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	settings := map[string]interface{}{
		"enabledPlugins": map[string]bool{},
		"hooks":          map[string]interface{}{},
	}
	data, _ := json.Marshal(settings)
	os.WriteFile(settingsPath, data, 0644)

	collector := &Collector{
		settingsPath: settingsPath,
		pluginsDir:   filepath.Join(tmpDir, "nonexistent-plugins"),
		cacheTTL:     5 * time.Second,
	}

	ctx := context.Background()
	stats := collector.Collect(ctx)

	// Should gracefully return 0 hooks
	if stats.HooksCount != 0 {
		t.Errorf("Expected HooksCount 0 with missing plugins dir, got %d", stats.HooksCount)
	}
}

func TestCollector_CoreToolsSet(t *testing.T) {
	// Verify that coreTools is properly defined
	if len(coreTools) == 0 {
		t.Error("coreTools set should not be empty")
	}

	// Check for some expected tools
	expectedTools := []string{"read", "edit", "write", "bash", "grep", "glob"}
	for _, tool := range expectedTools {
		if !coreTools[tool] {
			t.Errorf("Expected tool %s to be in coreTools", tool)
		}
	}
}
