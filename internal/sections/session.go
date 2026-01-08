package sections

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
	"github.com/ll931217/claude-hud-enhanced/internal/statusline"
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
	// Parse transcript
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := s.parser.Parse(ctx); err != nil {
		// Graceful degradation - return placeholder
		return "[Session: no transcript]"
	}

	var parts []string

	// Add model name if available
	if model := s.getModelName(); model != "" {
		parts = append(parts, model)
	}

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

// getContextBar returns the context window progress bar
func (s *SessionSection) getContextBar() string {
	cw := s.parser.GetContextWindow()
	if cw == nil || cw.ContextWindowSize == 0 {
		return ""
	}

	percentage := s.parser.GetContextPercentage()

	// Create progress bar
	bar := s.progressBar(percentage, 15)

	return fmt.Sprintf("[%s]%d%%", bar, percentage)
}

// progressBar creates a visual progress bar
func (s *SessionSection) progressBar(percentage, width int) string {
	if width <= 0 {
		width = 15
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
