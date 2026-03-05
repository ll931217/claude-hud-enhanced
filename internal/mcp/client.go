package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/errors"
)

const (
	// ClaudeConfigFile is the Claude global config file (~/.claude.json)
	ClaudeConfigFile = ".claude.json"

	// ClaudePluginsDir is the plugins directory under ~/.claude
	ClaudePluginsDir = ".claude/plugins"

	// DefaultTimeout is the default timeout for MCP queries
	DefaultTimeout = 2 * time.Second
)

// MCPServer represents an MCP server configuration
type MCPServer struct {
	Name     string                 `json:"name"`
	Command  string                 `json:"command"`
	Args     []string               `json:"args"`
	Env      map[string]string      `json:"env,omitempty"`
	Disabled bool                   `json:"disabled,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// MCPData represents data returned from MCP servers
type MCPData struct {
	ServerName string                 `json:"server_name"`
	Data       map[string]interface{} `json:"data"`
	Error      string                 `json:"error,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// Client represents an MCP client for querying Claude Code's MCP servers
type Client struct {
	mu            sync.RWMutex
	configPath    string
	pluginsDir    string
	servers       map[string]*MCPServer
	enabled       bool
	timeout       time.Duration
	lastQueryTime time.Time
	queryCache    map[string]*MCPData
	cacheTTL      time.Duration
}

// NewClient creates a new MCP client
func NewClient() *Client {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		errors.Warn("mcp", "failed to get home directory: %v", err)
		return &Client{
			enabled: false,
			timeout: DefaultTimeout,
		}
	}

	return &Client{
		configPath: filepath.Join(homeDir, ClaudeConfigFile),
		pluginsDir: filepath.Join(homeDir, ClaudePluginsDir),
		servers:    make(map[string]*MCPServer),
		enabled:    true,
		timeout:    DefaultTimeout,
		queryCache: make(map[string]*MCPData),
		cacheTTL:   5 * time.Second,
	}
}

// DetectServers detects MCP servers from Claude config and plugins
func (c *Client) DetectServers(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.enabled {
		return fmt.Errorf("MCP client is disabled")
	}

	c.servers = make(map[string]*MCPServer)

	// Load global MCP servers from ~/.claude.json
	c.loadGlobalServers()

	// Load plugin MCP servers from installed plugin .mcp.json files
	c.loadPluginServers()

	errors.Info("mcp", "detected %d MCP servers", len(c.servers))
	return nil
}

// loadGlobalServers loads MCP servers from ~/.claude.json
func (c *Client) loadGlobalServers() {
	if _, err := os.Stat(c.configPath); os.IsNotExist(err) {
		errors.Debug("mcp", "Claude config not found at %s", c.configPath)
		return
	}

	data, err := os.ReadFile(c.configPath)
	if err != nil {
		errors.Warn("mcp", "failed to read config file: %v", err)
		return
	}

	var config struct {
		MCPServers map[string]json.RawMessage `json:"mcpServers"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		errors.Warn("mcp", "failed to parse config file: %v", err)
		return
	}

	for name, serverData := range config.MCPServers {
		var server MCPServer
		if err := json.Unmarshal(serverData, &server); err != nil {
			errors.Warn("mcp", "failed to parse server %s: %v", name, err)
			continue
		}
		server.Name = name
		if !server.Disabled {
			c.servers[name] = &server
		}
	}
}

// loadPluginServers loads MCP servers from installed plugin .mcp.json files
func (c *Client) loadPluginServers() {
	if c.pluginsDir == "" {
		return
	}

	installedPath := filepath.Join(c.pluginsDir, "installed_plugins.json")
	data, err := os.ReadFile(installedPath)
	if err != nil {
		errors.Debug("mcp", "failed to read installed_plugins.json: %v", err)
		return
	}

	var installed struct {
		Plugins map[string][]struct {
			InstallPath string `json:"installPath"`
		} `json:"plugins"`
	}

	if err := json.Unmarshal(data, &installed); err != nil {
		errors.Debug("mcp", "failed to parse installed_plugins.json: %v", err)
		return
	}

	seen := make(map[string]bool)
	for _, installs := range installed.Plugins {
		for _, inst := range installs {
			if inst.InstallPath == "" || seen[inst.InstallPath] {
				continue
			}
			seen[inst.InstallPath] = true
			c.loadMCPFromPlugin(inst.InstallPath)
		}
	}
}

// loadMCPFromPlugin loads MCP servers from a plugin's .mcp.json file
func (c *Client) loadMCPFromPlugin(installPath string) {
	mcpPath := filepath.Join(installPath, ".mcp.json")
	data, err := os.ReadFile(mcpPath)
	if err != nil {
		return
	}

	var mcpConfig struct {
		MCPServers map[string]json.RawMessage `json:"mcpServers"`
	}

	if err := json.Unmarshal(data, &mcpConfig); err != nil {
		return
	}

	for name, serverData := range mcpConfig.MCPServers {
		if _, exists := c.servers[name]; exists {
			continue
		}
		var server MCPServer
		if err := json.Unmarshal(serverData, &server); err != nil {
			continue
		}
		server.Name = name
		if !server.Disabled {
			c.servers[name] = &server
		}
	}
}

// GetServers returns the list of detected MCP servers
func (c *Client) GetServers() []*MCPServer {
	c.mu.RLock()
	defer c.mu.RUnlock()

	servers := make([]*MCPServer, 0, len(c.servers))
	for _, server := range c.servers {
		servers = append(servers, server)
	}
	return servers
}

// QueryAll queries all detected MCP servers for data
// This is non-blocking and returns cached data if available
func (c *Client) QueryAll(ctx context.Context) []*MCPData {
	c.mu.Lock()
	defer c.mu.Unlock()

	results := make([]*MCPData, 0, len(c.servers))

	// Check if we should use cache
	if time.Since(c.lastQueryTime) < c.cacheTTL {
		for _, data := range c.queryCache {
			results = append(results, data)
		}
		return results
	}

	// Query each server
	for _, server := range c.servers {
		data := c.queryServer(ctx, server)
		results = append(results, data)
		c.queryCache[server.Name] = data
	}

	c.lastQueryTime = time.Now()
	return results
}

// queryServer queries a single MCP server
func (c *Client) queryServer(ctx context.Context, server *MCPServer) *MCPData {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// For now, return a placeholder since we can't actually run MCP commands
	// In a real implementation, this would execute the server command and parse output
	data := &MCPData{
		ServerName: server.Name,
		Data: map[string]interface{}{
			"status":  "detected",
			"command": server.Command,
			"args":    server.Args,
		},
		Timestamp: time.Now(),
	}

	return data
}

// QueryServer queries a specific MCP server by name
func (c *Client) QueryServer(ctx context.Context, serverName string) (*MCPData, error) {
	c.mu.RLock()
	server, exists := c.servers[serverName]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("server %s not found", serverName)
	}

	data := c.queryServer(ctx, server)
	return data, nil
}

// IsEnabled returns whether MCP client is enabled
func (c *Client) IsEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.enabled
}

// SetEnabled sets whether MCP client is enabled
func (c *Client) SetEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enabled = enabled
}

// SetTimeout sets the query timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timeout = timeout
}

// SetCacheTTL sets the cache TTL
func (c *Client) SetCacheTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cacheTTL = ttl
}

// ServerCount returns the number of detected servers
func (c *Client) ServerCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.servers)
}

// Refresh re-detects MCP servers
func (c *Client) Refresh(ctx context.Context) error {
	return c.DetectServers(ctx)
}

// GetStatus returns the current status of the MCP client
func (c *Client) GetStatus() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.enabled {
		return "disabled"
	}

	if len(c.servers) == 0 {
		return "no servers"
	}

	return fmt.Sprintf("%d servers", len(c.servers))
}

// FormatStatus formats the MCP status for display
func (c *Client) FormatStatus() string {
	status := c.GetStatus()
	if status == "disabled" || status == "no servers" {
		return ""
	}

	return fmt.Sprintf("MCP: %s", status)
}

// GetServerNames returns the names of all detected servers
func (c *Client) GetServerNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.servers))
	for name := range c.servers {
		names = append(names, name)
	}
	return names
}
