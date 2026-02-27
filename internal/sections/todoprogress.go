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

// TodoProgressSection displays todo list progress from TodoWrite
type TodoProgressSection struct {
	*BaseSection
	parser *transcript.Parser
}

// NewTodoProgressSection creates a new todo progress section (factory function for registry)
func NewTodoProgressSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	// Get transcript path from environment or use default
	transcriptPath := getTranscriptPath()

	base := NewBaseSection("todoprogress", appConfig)
	base.SetPriority(registry.PriorityEssential) // Show current task progress
	base.SetMinWidth(20)                         // Minimum width for progress display

	return &TodoProgressSection{
		BaseSection: base,
		parser:      transcript.NewParser(transcriptPath),
	}, nil
}

// Render returns the todo progress section output
func (t *TodoProgressSection) Render() string {
	// Get transcript path dynamically from global context
	transcriptPath := getTranscriptPath()
	if transcriptPath == "" {
		return "" // Hide section if no transcript path
	}

	// Create a parser for the current transcript path
	parser := transcript.NewParser(transcriptPath)

	// Parse transcript for todo data
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := parser.Parse(ctx); err != nil {
		return "" // Hide section on parse error
	}

	// Get todo counts
	total, completed := parser.GetTodoCount()
	if total == 0 {
		return "" // Hide section when no todos
	}

	// Get current in-progress todo
	currentTodo := parser.GetCurrentTodo()

	var parts []string

	// Show progress fraction
	parts = append(parts, fmt.Sprintf("📋 %d/%d", completed, total))

	// Show current task if available
	if currentTodo != nil {
		taskName := truncateTaskName(currentTodo.Content, 30)
		parts = append(parts, fmt.Sprintf("◐ %s", taskName))
	}

	return strings.Join(parts, " | ")
}

// truncateTaskName truncates a task name to max length
func truncateTaskName(task string, maxLen int) string {
	// Remove "activeForm" prefix if present
	task = strings.TrimPrefix(task, "activeForm:")
	task = strings.TrimSpace(task)

	if len(task) <= maxLen {
		return task
	}
	return task[:maxLen-3] + "..."
}

func init() {
	registry.Register("todoprogress", NewTodoProgressSection)
}
