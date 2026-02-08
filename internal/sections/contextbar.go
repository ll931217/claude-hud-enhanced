package sections

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/errors"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
	"github.com/ll931217/claude-hud-enhanced/internal/statusline"
	"github.com/ll931217/claude-hud-enhanced/internal/theme"
	"github.com/ll931217/claude-hud-enhanced/internal/transcript"
)

// ContextBarSection displays context window progress bar with color coding
type ContextBarSection struct {
	*BaseSection
	parser *transcript.Parser
}

// NewContextBarSection creates a new context bar section (factory function for registry)
func NewContextBarSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	transcriptPath := getTranscriptPath()
	base := NewBaseSection("contextbar", appConfig)
	base.SetPriority(registry.PriorityEssential) // Essential - always show context
	base.SetMinWidth(6)                          // "█ 0%" minimum

	return &ContextBarSection{
		BaseSection: base,
		parser:      transcript.NewParser(transcriptPath),
	}, nil
}

func init() {
	registry.Register("contextbar", NewContextBarSection)
}

// Render returns the context bar section output
func (c *ContextBarSection) Render() string {
	// First, try to get context window data from Claude Code's JSON input (most reliable)
	windowSize := statusline.GetContextWindowSize()
	inputTokens := statusline.GetContextInputTokens()
	cacheTokens := statusline.GetContextCacheTokens()

	// Only use stdin data if we have actual token counts (not just zeros)
	if windowSize > 0 && (inputTokens > 0 || cacheTokens > 0) {
		// Calculate percentage from JSON input data
		totalTokens := inputTokens + cacheTokens
		percentage := (totalTokens * 100) / windowSize
		if percentage > 100 {
			percentage = 100
		}
		if percentage < 0 {
			percentage = 0
		}

		bar := c.progressBar(percentage, 10) // 10-char width
		color := theme.ContextColor(percentage)

		// Show format: "72%" without brackets as user requested
		result := fmt.Sprintf("%s%s %d%%", color, bar, percentage)
		if color != "" {
			result += theme.Reset
		}

		// Add token breakdown at high context usage
		if percentage >= 85 {
			var parts []string
			if inputTokens > 0 {
				parts = append(parts, fmt.Sprintf("in: %s", formatTokens(inputTokens)))
			}
			if cacheTokens > 0 {
				parts = append(parts, fmt.Sprintf("cache: %s", formatTokens(cacheTokens)))
			}
			if len(parts) > 0 {
				result += fmt.Sprintf("%s (%s)%s", theme.Dim, strings.Join(parts, ", "), theme.Reset)
			}
		}

		return result
	}

	// Fallback: Try to get from transcript parser
	// (also used when stdin data exists but has zero tokens)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = c.parser.Parse(ctx) // Try to parse, but don't fail if it doesn't work

	cw := c.parser.GetContextWindow()
	if cw == nil {
		// No context window data available
		return ""
	}
	if cw.ContextWindowSize == 0 {
		// Debug: log why context window size is 0
		errors.Debug("contextbar", "context window size is 0 - trying to infer from model")
		// Try to infer context window size from model name
		// Note: We can't easily get model name here without duplicating logic
		// For now, return empty
		return ""
	}

	percentage := c.parser.GetContextPercentage()
	bar := c.progressBar(percentage, 10) // 10-char width
	color := theme.ContextColor(percentage)

	// Show format: "72%" without brackets as user requested
	// At high usage, show token breakdown
	result := fmt.Sprintf("%s%s %d%%", color, bar, percentage)
	if color != "" {
		result += theme.Reset
	}

	// Add token breakdown at high context usage
	if percentage >= 85 {
		breakdown := c.getTokenBreakdown(cw)
		if breakdown != "" {
			result += fmt.Sprintf("%s %s%s", theme.Dim, breakdown, theme.Reset)
		}
	}

	return result
}

// progressBar creates a visual progress bar
func (c *ContextBarSection) progressBar(percentage, width int) string {
	if width <= 0 {
		width = 10 // Default to 10 chars
	}

	filled := percentage * width / 100
	if filled > width {
		filled = width
	}

	empty := width - filled
	if empty < 0 {
		empty = 0
	}

	return strings.Repeat("█", filled) + strings.Repeat("░", empty)
}

// getTokenBreakdown returns token breakdown at high context usage
func (c *ContextBarSection) getTokenBreakdown(cw *transcript.ContextWindow) string {
	usage := cw.CurrentUsage

	inputTokens := usage.InputTokens
	cacheTokens := usage.CacheCreationInputTokens + usage.CacheReadInputTokens

	// Only show breakdown if there are actual tokens
	if inputTokens == 0 && cacheTokens == 0 {
		return ""
	}

	var parts []string
	if inputTokens > 0 {
		parts = append(parts, fmt.Sprintf("in: %s", formatTokens(inputTokens)))
	}
	if cacheTokens > 0 {
		parts = append(parts, fmt.Sprintf("cache: %s", formatTokens(cacheTokens)))
	}

	if len(parts) == 0 {
		return ""
	}

	return fmt.Sprintf("(%s)", strings.Join(parts, ", "))
}

// formatTokens formats a token count with suffix (k, M)
func formatTokens(tokens int) string {
	if tokens >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(tokens)/1_000_000)
	}
	if tokens >= 1_000 {
		return fmt.Sprintf("%dk", tokens/1_000)
	}
	return fmt.Sprintf("%d", tokens)
}
