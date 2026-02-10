package claudestats

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/errors"
	"github.com/ll931217/claude-hud-enhanced/internal/mcp"
)

// coreTools is the set of built-in Claude Code tools
var coreTools = map[string]bool{
	"read":            true,
	"edit":            true,
	"write":           true,
	"bash":            true,
	"grep":            true,
	"glob":            true,
	"askuserquestion": true,
	"todowrite":       true,
	"taskupdate":      true,
	"taskget":         true,
	"tasklist":        true,
	"taskoutput":      true,
	"skill":           true,
	"websearch":       true,
	"webfetch":        true,
	"mcp__search":     true,
	"notebookedit":    true,
	"killshell":       true,
	"exitplanmode":    true,
	"enterplanmode":   true,
	"readfile":        true,
	"repository":      true,
}

// StatsCache holds cached statistics
type StatsCache struct {
	CoreCount   int
	MCPCount    int
	SkillsCount int
	HooksCount  int
	Timestamp   time.Time
}

// Collector gathers Claude capability statistics
type Collector struct {
	mu           sync.RWMutex
	mcpClient    *mcp.Client
	settingsPath string
	cache        *StatsCache
	lastUpdate   time.Time
	cacheTTL     time.Duration
}

// NewCollector creates a new statistics collector
func NewCollector() *Collector {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		errors.Warn("claudestats", "failed to get home directory: %v", err)
		return &Collector{
			mcpClient: mcp.NewClient(),
			cacheTTL:  5 * time.Second,
		}
	}

	return &Collector{
		mcpClient:    mcp.NewClient(),
		settingsPath: filepath.Join(homeDir, ".claude", "settings.json"),
		cacheTTL:     5 * time.Second,
	}
}

// Collect gathers all statistics with caching
func (c *Collector) Collect(ctx context.Context) *StatsCache {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check cache
	if c.cache != nil && time.Since(c.lastUpdate) < c.cacheTTL {
		return c.cache
	}

	// Collect fresh data
	stats := &StatsCache{
		CoreCount:   len(coreTools),
		MCPCount:    c.collectMCPCount(ctx),
		SkillsCount: c.collectSkillsCount(ctx),
		HooksCount:  c.collectHooksCount(ctx),
		Timestamp:   time.Now(),
	}

	c.cache = stats
	c.lastUpdate = time.Now()
	return stats
}

// collectMCPCount returns MCP server count
func (c *Collector) collectMCPCount(ctx context.Context) int {
	if c.mcpClient == nil {
		return 0
	}
	if err := c.mcpClient.DetectServers(ctx); err != nil {
		errors.Debug("claudestats", "failed to detect MCP servers: %v", err)
		return 0
	}
	return c.mcpClient.ServerCount()
}

// collectSkillsCount returns enabled skills count
func (c *Collector) collectSkillsCount(ctx context.Context) int {
	data, err := os.ReadFile(c.settingsPath)
	if err != nil {
		errors.Debug("claudestats", "failed to read settings file: %v", err)
		return 0
	}

	var settings struct {
		EnabledPlugins map[string]bool `json:"enabledPlugins"`
	}

	if err := json.Unmarshal(data, &settings); err != nil {
		errors.Debug("claudestats", "failed to parse enabledPlugins: %v", err)
		return 0
	}

	return len(settings.EnabledPlugins)
}

// collectHooksCount returns configured hooks count
func (c *Collector) collectHooksCount(ctx context.Context) int {
	data, err := os.ReadFile(c.settingsPath)
	if err != nil {
		errors.Debug("claudestats", "failed to read settings file for hooks: %v", err)
		return 0
	}

	var settings struct {
		Hooks map[string][]struct {
			Matcher string            `json:"matcher"`
			Hooks   []json.RawMessage `json:"hooks"`
		} `json:"hooks"`
	}

	if err := json.Unmarshal(data, &settings); err != nil {
		errors.Debug("claudestats", "failed to parse hooks: %v", err)
		return 0
	}

	count := 0
	for _, hookGroup := range settings.Hooks {
		for _, group := range hookGroup {
			count += len(group.Hooks)
		}
	}
	return count
}
