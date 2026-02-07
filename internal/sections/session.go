package sections

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/errors"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
	"github.com/ll931217/claude-hud-enhanced/internal/statusline"
	"github.com/ll931217/claude-hud-enhanced/internal/theme"
	"github.com/ll931217/claude-hud-enhanced/internal/transcript"
)

// SessionSection displays Claude Code session information
type SessionSection struct {
	*BaseSection
	parser *transcript.Parser
}

// NewSessionSection creates a new session section (factory function for registry)
func NewSessionSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	// Get transcript path from environment or use default
	transcriptPath := getTranscriptPath()

	base := NewBaseSection("session", appConfig)

	return &SessionSection{
		BaseSection: base,
		parser:      transcript.NewParser(transcriptPath),
	}, nil
}

// Render returns the session section output
func (s *SessionSection) Render() string {
	var parts []string

	// Try to get model name from statusline context first (doesn't require transcript)
	model := statusline.GetModelName()

	if model != "" {
		// Shorten model name
		model = strings.ReplaceAll(model, "Claude ", "")
		model = strings.ReplaceAll(model, "Sonnet", "SN")
		model = strings.ReplaceAll(model, "Haiku", "HK")
		model = strings.ReplaceAll(model, "Opus", "OP")
		parts = append(parts, model)
	}

	// Try to parse transcript for additional information
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = s.parser.Parse(ctx) // Try to parse, but don't fail if it doesn't work

	// Add context bar if available
	if contextBar := s.getContextBar(); contextBar != "" {
		parts = append(parts, contextBar)
	}

	// Add duration if available
	if duration := s.getDuration(); duration != "" {
		parts = append(parts, duration)
	}

	// Add tool activity if available
	if tools := s.getToolActivity(); tools != "" {
		parts = append(parts, tools)
	}

	// Add agent activity if available
	if agents := s.getAgentActivity(); agents != "" {
		parts = append(parts, agents)
	}

	// Add todo progress if available
	if todos := s.getTodoProgress(); todos != "" {
		parts = append(parts, todos)
	}

	// Add cost if available
	if cost := s.getCost(); cost != "" {
		parts = append(parts, cost)
	}

	if len(parts) == 0 {
		return "[Session: waiting for data]"
	}

	return strings.Join(parts, " ")
}

// getModelName returns the short model name
func (s *SessionSection) getModelName() string {
	event := s.parser.GetLatestEvent(transcript.EventTypeAssistantMessage)
	if event == nil || event.Message == nil {
		return ""
	}

	model := event.Message.Model
	if model == "" {
		return ""
	}

	// Shorten model name
	model = strings.ReplaceAll(model, "Claude ", "")
	model = strings.ReplaceAll(model, "Sonnet", "SN")
	model = strings.ReplaceAll(model, "Haiku", "HK")
	model = strings.ReplaceAll(model, "Opus", "OP")

	return model
}

// getContextBar returns the context window progress bar with color coding
func (s *SessionSection) getContextBar() string {
	// First, try to get context window data from Claude Code's JSON input (most reliable)
	windowSize := statusline.GetContextWindowSize()
	inputTokens := statusline.GetContextInputTokens()
	cacheTokens := statusline.GetContextCacheTokens()

	if windowSize > 0 {
		// Calculate percentage from JSON input data
		totalTokens := inputTokens + cacheTokens
		percentage := (totalTokens * 100) / windowSize
		if percentage > 100 {
			percentage = 100
		}
		if percentage < 0 {
			percentage = 0
		}

		bar := s.progressBar(percentage, 10) // 10-char width
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
	cw := s.parser.GetContextWindow()
	if cw == nil {
		// No context window data available
		return ""
	}
	if cw.ContextWindowSize == 0 {
		// Debug: log why context window size is 0
		errors.Debug("session", "context window size is 0 - trying to infer from model")
		// Try to infer context window size from model name
		if model := s.getModelName(); model != "" {
			inferredSize := inferContextWindowFromModel(model)
			if inferredSize > 0 {
				// Create a new context window with inferred size
				cw.ContextWindowSize = inferredSize
				errors.Debug("session", "inferred context window size %d from model %s", inferredSize, model)
			}
		}
		// If still 0, return empty
		if cw.ContextWindowSize == 0 {
			return ""
		}
	}

	percentage := s.parser.GetContextPercentage()
	bar := s.progressBar(percentage, 10) // 10-char width
	color := theme.ContextColor(percentage)

	// Show format: "72%" without brackets as user requested
	// At high usage, show token breakdown
	result := fmt.Sprintf("%s%s %d%%", color, bar, percentage)
	if color != "" {
		result += theme.Reset
	}

	// Add token breakdown at high context usage
	if percentage >= 85 {
		breakdown := s.getTokenBreakdown(cw)
		if breakdown != "" {
			result += fmt.Sprintf("%s %s%s", theme.Dim, breakdown, theme.Reset)
		}
	}

	return result
}

// inferContextWindowFromModel infers context window size from model name
func inferContextWindowFromModel(model string) int {
	// All current Claude models have 200k token context
	// This may change in the future, but it's a reasonable fallback
	return 200000
}

// progressBar creates a visual progress bar
func (s *SessionSection) progressBar(percentage, width int) string {
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
func (s *SessionSection) getTokenBreakdown(cw *transcript.ContextWindow) string {
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

// getDuration returns the session duration
func (s *SessionSection) getDuration() string {
	return s.parser.GetDuration()
}

// getToolActivity returns active and completed tool usage
func (s *SessionSection) getToolActivity() string {
	tools := s.parser.GetToolActivity()
	if len(tools) == 0 {
		return ""
	}

	// Aggregate tools by name
	toolCounts := make(map[string]int)
	for _, tool := range tools {
		toolCounts[tool.Name]++
	}

	// Format: ✓ Read ×2 | ✓ Edit ×1
	var parts []string
	for name, count := range toolCounts {
		parts = append(parts, fmt.Sprintf("✓ %s ×%d", name, count))
	}

	return strings.Join(parts, " | ")
}

// getAgentActivity returns active and completed agent runs
func (s *SessionSection) getAgentActivity() string {
	agents := s.parser.GetAgentActivity()
	if len(agents) == 0 {
		return ""
	}

	// Format active agents
	var parts []string
	for _, agent := range agents {
		name := agent.AgentName
		if name == "" {
			name = agent.Type
		}
		if name == "" {
			name = "Agent"
		}

		// Calculate elapsed time if we have timestamp
		elapsed := ""
		if start := s.parser.GetSessionStart(); !start.IsZero() {
			duration := time.Since(start)
			elapsed = fmt.Sprintf("(%d%s)", int(duration.Seconds()), "s")
		}

		parts = append(parts, fmt.Sprintf("↻ %s %s", name, elapsed))
	}

	return strings.Join(parts, " | ")
}

// getTranscriptPath returns the transcript path from context, environment, or default
func getTranscriptPath() string {
	// Check global context from Claude Code first
	if path := statusline.GetTranscriptPath(); path != "" {
		return path
	}

	// Fallback to environment variable (for standalone mode or wrapper script)
	if path := os.Getenv("CLAUDE_HUD_TRANSCRIPT_PATH"); path != "" {
		return path
	}

	// For standalone mode, try to find the latest transcript in cwd
	// Look for .claude/transcript.json or similar
	return ""
}

// getTodoProgress returns todo progress information
func (s *SessionSection) getTodoProgress() string {
	total, completed := s.parser.GetTodoCount()
	if total == 0 {
		return ""
	}

	// Check if there's a current in-progress todo
	current := s.parser.GetCurrentTodo()

	if current != nil {
		// Show current in-progress task
		content := current.Content
		if len(content) > 30 {
			content = content[:27] + "..."
		}
		return fmt.Sprintf("◐ %s", content)
	}

	// All todos completed
	if completed == total {
		return fmt.Sprintf("✓ All todos complete (%d/%d)", total, completed)
	}

	// Show progress
	return fmt.Sprintf("✓ %d/%d", completed, total)
}

// getCost returns the estimated session cost
func (s *SessionSection) getCost() string {
	cost := s.parser.CalculateCost()
	if cost == 0 {
		return ""
	}

	// Format cost
	if cost < 0.01 {
		return fmt.Sprintf("$%.4f", cost)
	} else if cost < 1 {
		return fmt.Sprintf("$%.2f", cost)
	}
	return fmt.Sprintf("$%.2f", cost)
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
