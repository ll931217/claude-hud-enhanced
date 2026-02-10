package sections

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/claudestats"
	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
)

// ClaudeStatsSection displays Claude capability statistics
type ClaudeStatsSection struct {
	*BaseSection
	collector *claudestats.Collector
}

// NewClaudeStatsSection creates a new claudestats section (factory function)
func NewClaudeStatsSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	base := NewBaseSection("claudestats", appConfig)
	base.SetPriority(registry.PriorityImportant) // Show on medium+ terminals
	base.SetMinWidth(30)                         // Minimum width for "Core:8 | MCP:5"

	return &ClaudeStatsSection{
		BaseSection: base,
		collector:   claudestats.NewCollector(),
	}, nil
}

// Render returns the claudestats section output
func (s *ClaudeStatsSection) Render() string {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	stats := s.collector.Collect(ctx)

	var parts []string

	// Only show categories with non-zero counts
	if stats.CoreCount > 0 {
		parts = append(parts, fmt.Sprintf("Core:%d", stats.CoreCount))
	}
	if stats.MCPCount > 0 {
		parts = append(parts, fmt.Sprintf("MCP:%d", stats.MCPCount))
	}
	if stats.SkillsCount > 0 {
		parts = append(parts, fmt.Sprintf("Skills:%d", stats.SkillsCount))
	}
	if stats.HooksCount > 0 {
		parts = append(parts, fmt.Sprintf("Hooks:%d", stats.HooksCount))
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " | ")
}

func init() {
	registry.Register("claudestats", NewClaudeStatsSection)
}
