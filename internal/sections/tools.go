package sections

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
	"github.com/ll931217/claude-hud-enhanced/internal/transcript"
)

// ToolsSection displays tool activity with recency tracking
type ToolsSection struct {
	*BaseSection
	parser *transcript.Parser
}

// NewToolsSection creates a new tools section (factory function for registry)
func NewToolsSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	// Get transcript path from environment or use default
	transcriptPath := getTranscriptPath()

	base := NewBaseSection("tools", appConfig)
	base.SetPriority(registry.PriorityImportant) // Show on medium+ terminals (80+ cols)

	return &ToolsSection{
		BaseSection: base,
		parser:      transcript.NewParser(transcriptPath),
	}, nil
}

// Render returns the tools section output
func (t *ToolsSection) Render() string {
	// Parse transcript for tool data
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = t.parser.Parse(ctx)

	// Get tools by recency (max 5)
	tools := t.parser.GetToolsByRecency(5)
	if len(tools) == 0 {
		return ""
	}

	var parts []string
	for _, tool := range tools {
		name := shortenToolName(tool.Name)
		parts = append(parts, fmt.Sprintf("%s×%d", name, tool.Count))
	}

	return strings.Join(parts, " | ")
}

// shortenToolName shortens verbose tool names for display
func shortenToolName(name string) string {
	// Shorten MCP plugin prefixes
	// mcp__plugin_playwright_playwright__browser_click → browser_click
	// mcp__morph__edit_file → edit_file
	// mcp__plugin_playwright_playwright__ → [browser]

	// Remove common MCP prefixes
	prefixes := []string{
		"mcp__plugin_playwright_playwright__",
		"mcp__plugin_playwright__",
		"mcp__morph__",
		"mcp__zai-mcp-server__",
		"mcp__4_5v_mcp__",
		"mcp__web-reader__",
		"mcp__web_search_prime__",
		"mcp__plugin_context7_context7__",
		"mcp__plugin_greptile_greptile__",
		"mcp__plugin_playwright__",
		"mcp__zread__",
	}

	for _, prefix := range prefixes {
		name = strings.TrimPrefix(name, prefix)
	}

	// If name is still long, try to extract the meaningful part
	if strings.Contains(name, "__") {
		parts := strings.Split(name, "__")
		if len(parts) > 0 {
			name = parts[len(parts)-1]
		}
	}

	// Shorten common Claude Code tools
	shortNames := map[string]string{
		"computer_20250124": "computer",
		"cli":               "cli",
		"text_editor_20250124": "editor",
	}

	if short, ok := shortNames[name]; ok {
		name = short
	}

	return name
}

func init() {
	registry.Register("tools", NewToolsSection)
}
