package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if !client.IsEnabled() {
		t.Error("NewClient() should be enabled by default")
	}
}

func TestClient_DetectServers_NoConfig(t *testing.T) {
	client := NewClient()
	client.configPath = "/nonexistent/path/settings.json"

	ctx := context.Background()
	err := client.DetectServers(ctx)
	if err != nil {
		t.Errorf("DetectServers() should not error with missing config, got: %v", err)
	}

	if client.ServerCount() != 0 {
		t.Errorf("Expected 0 servers with missing config, got %d", client.ServerCount())
	}
}

func TestClient_DetectServers_InvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "settings.json")

	// Write invalid JSON
	if err := os.WriteFile(configPath, []byte("invalid json"), 0644); err != nil {
		t.Fatal(err)
	}

	client := NewClient()
	client.configPath = configPath

	ctx := context.Background()
	err := client.DetectServers(ctx)
	if err == nil {
		t.Error("DetectServers() should error with invalid JSON")
	}
}

func TestClient_DetectServers_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "settings.json")

	// Write valid config with MCP servers
	configJSON := `{
		"mcpServers": {
			"test-server": {
				"command": "node",
				"args": ["test.js"]
			},
			"disabled-server": {
				"command": "python",
				"args": ["test.py"],
				"disabled": true
			}
		}
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatal(err)
	}

	client := NewClient()
	client.configPath = configPath

	ctx := context.Background()
	if err := client.DetectServers(ctx); err != nil {
		t.Fatalf("DetectServers() error = %v", err)
	}

	if client.ServerCount() != 1 {
		t.Errorf("Expected 1 server (disabled should be filtered), got %d", client.ServerCount())
	}

	servers := client.GetServers()
	if len(servers) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(servers))
	}

	if servers[0].Name != "test-server" {
		t.Errorf("Expected server name 'test-server', got %s", servers[0].Name)
	}

	if servers[0].Command != "node" {
		t.Errorf("Expected command 'node', got %s", servers[0].Command)
	}
}

func TestClient_GetServers(t *testing.T) {
	client := NewClient()
	servers := client.GetServers()

	if servers == nil {
		t.Error("GetServers() should never return nil")
	}
}

func TestClient_QueryAll(t *testing.T) {
	client := NewClient()

	ctx := context.Background()
	results := client.QueryAll(ctx)

	if results == nil {
		t.Error("QueryAll() should never return nil")
	}
}

func TestClient_IsEnabled(t *testing.T) {
	client := NewClient()

	if !client.IsEnabled() {
		t.Error("New client should be enabled")
	}

	client.SetEnabled(false)
	if client.IsEnabled() {
		t.Error("Client should be disabled after SetEnabled(false)")
	}

	client.SetEnabled(true)
	if !client.IsEnabled() {
		t.Error("Client should be enabled after SetEnabled(true)")
	}
}

func TestClient_SetTimeout(t *testing.T) {
	client := NewClient()
	timeout := 5 * time.Second

	client.SetTimeout(timeout)
	// Can't directly check timeout as it's private, but we can verify it doesn't crash
	client.SetTimeout(1 * time.Second)
	client.SetTimeout(100 * time.Millisecond)
}

func TestClient_SetCacheTTL(t *testing.T) {
	client := NewClient()
	ttl := 10 * time.Second

	client.SetCacheTTL(ttl)
	// Can't directly check TTL as it's private, but we can verify it doesn't crash
	client.SetCacheTTL(1 * time.Second)
	client.SetCacheTTL(100 * time.Millisecond)
}

func TestClient_GetStatus(t *testing.T) {
	client := NewClient()

	status := client.GetStatus()
	if status == "" {
		t.Error("GetStatus() should not return empty string")
	}

	// Test disabled status
	client.SetEnabled(false)
	status = client.GetStatus()
	if status != "disabled" {
		t.Errorf("Expected 'disabled' status, got %s", status)
	}

	// Re-enable for other tests
	client.SetEnabled(true)
}

func TestClient_FormatStatus(t *testing.T) {
	client := NewClient()

	// With no servers, should return empty string
	formatted := client.FormatStatus()
	if formatted != "" {
		t.Logf("FormatStatus with no servers returned: %s", formatted)
	}

	// Disabled should return empty
	client.SetEnabled(false)
	formatted = client.FormatStatus()
	if formatted != "" {
		t.Errorf("FormatStatus() when disabled should return empty, got: %s", formatted)
	}
}

func TestClient_GetServerNames(t *testing.T) {
	client := NewClient()
	names := client.GetServerNames()

	if names == nil {
		t.Error("GetServerNames() should never return nil")
	}

	// Should be empty with no servers
	if len(names) != 0 {
		t.Errorf("Expected 0 server names, got %d", len(names))
	}
}

func TestClient_Refresh(t *testing.T) {
	client := NewClient()

	ctx := context.Background()
	err := client.Refresh(ctx)
	if err != nil {
		// Refresh might fail if config doesn't exist, which is ok
		t.Logf("Refresh() returned error (may be expected): %v", err)
	}
}

func TestClient_QueryServer(t *testing.T) {
	client := NewClient()

	ctx := context.Background()

	// Query non-existent server
	_, err := client.QueryServer(ctx, "nonexistent")
	if err == nil {
		t.Error("QueryServer() should error for non-existent server")
	}

	// Query existing server (after adding one)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "settings.json")

	configJSON := `{
		"mcpServers": {
			"test-server": {
				"command": "node",
				"args": ["test.js"]
			}
		}
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatal(err)
	}

	client.configPath = configPath
	if err := client.DetectServers(ctx); err != nil {
		t.Fatal(err)
	}

	data, err := client.QueryServer(ctx, "test-server")
	if err != nil {
		t.Fatalf("QueryServer() error = %v", err)
	}

	if data == nil {
		t.Error("QueryServer() should return data, got nil")
	}

	if data.ServerName != "test-server" {
		t.Errorf("Expected server name 'test-server', got %s", data.ServerName)
	}
}

func TestMCPServer_Disabled(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "settings.json")

	configJSON := `{
		"mcpServers": {
			"enabled-server": {
				"command": "node",
				"args": ["test.js"],
				"disabled": false
			},
			"disabled-server": {
				"command": "python",
				"args": ["test.py"],
				"disabled": true
			}
		}
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatal(err)
	}

	client := NewClient()
	client.configPath = configPath

	ctx := context.Background()
	if err := client.DetectServers(ctx); err != nil {
		t.Fatalf("DetectServers() error = %v", err)
	}

	// Only non-disabled servers should be detected
	if client.ServerCount() != 1 {
		t.Errorf("Expected 1 server (disabled filtered), got %d", client.ServerCount())
	}

	servers := client.GetServers()
	if len(servers) != 1 || servers[0].Name != "enabled-server" {
		t.Error("Disabled server should be filtered out")
	}
}

func TestClient_QueryCaching(t *testing.T) {
	client := NewClient()
	client.SetCacheTTL(1 * time.Second)

	ctx := context.Background()

	// First query
	start1 := time.Now()
	results1 := client.QueryAll(ctx)
	duration1 := time.Since(start1)

	// Second query should be cached (much faster)
	start2 := time.Now()
	results2 := client.QueryAll(ctx)
	duration2 := time.Since(start2)

	// Results should be the same
	if len(results1) != len(results2) {
		t.Errorf("Cached results should match: %d vs %d", len(results1), len(results2))
	}

	// Cached query should be faster (or at least not significantly slower)
	// Note: This is a soft check since timing can vary
	if duration2 > duration1*10 && duration1 > 0 {
		t.Logf("Warning: Cached query took longer than expected: %v vs %v", duration2, duration1)
	}

	// Wait for cache to expire
	time.Sleep(client.cacheTTL + 100*time.Millisecond)

	// Third query should not be cached
	results3 := client.QueryAll(ctx)
	if len(results3) != len(results2) {
		t.Logf("Results after cache expiry: %d vs %d", len(results3), len(results2))
	}
}

func TestMCPData_Timestamp(t *testing.T) {
	client := NewClient()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "settings.json")

	configJSON := `{
		"mcpServers": {
			"test-server": {
				"command": "node",
				"args": ["test.js"]
			}
		}
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatal(err)
	}

	client.configPath = configPath

	ctx := context.Background()
	if err := client.DetectServers(ctx); err != nil {
		t.Fatal(err)
	}

	before := time.Now()
	results := client.QueryAll(ctx)
	after := time.Now()

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	data := results[0]
	if data.Timestamp.Before(before) || data.Timestamp.After(after) {
		t.Error("Timestamp should be between query start and end")
	}

	if data.ServerName != "test-server" {
		t.Errorf("Expected server name 'test-server', got %s", data.ServerName)
	}
}
