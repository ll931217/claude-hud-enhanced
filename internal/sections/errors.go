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

// ErrorsSection displays recent errors from transcript
type ErrorsSection struct {
	*BaseSection
	parser *transcript.Parser
}

// NewErrorsSection creates a new errors section (factory function for registry)
func NewErrorsSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	// Get transcript path from environment or use default
	transcriptPath := getTranscriptPath()

	base := NewBaseSection("errors", appConfig)
	base.SetPriority(registry.PriorityImportant) // Important to see errors
	base.SetMinWidth(15)                         // Minimum width for error display

	return &ErrorsSection{
		BaseSection: base,
		parser:      transcript.NewParser(transcriptPath),
	}, nil
}

// Render returns the errors section output
func (e *ErrorsSection) Render() string {
	// Get transcript path dynamically from global context
	transcriptPath := getTranscriptPath()
	if transcriptPath == "" {
		return "" // Hide section if no transcript path
	}

	// Create a parser for the current transcript path
	parser := transcript.NewParser(transcriptPath)

	// Parse transcript for error data
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := parser.Parse(ctx); err != nil {
		return "" // Hide section on parse error
	}

	// Get error counts (total and recent)
	total, recent := parser.GetErrorCount(5) // Errors in last 5 minutes
	if total == 0 {
		return "" // Hide section when no errors
	}

	var parts []string

	// Show total errors
	if total > 0 {
		parts = append(parts, fmt.Sprintf("⚠️  %d", total))
	}

	// Show recent errors if any
	if recent > 0 && recent != total {
		parts = append(parts, fmt.Sprintf("(%d recent)", recent))
	}

	// Get last error details
	recentErrors := parser.GetRecentErrors(1)
	if len(recentErrors) > 0 {
		lastError := recentErrors[0]
		if lastError.ToolName != "" {
			parts = append(parts, fmt.Sprintf("[%s]", shortenToolName(lastError.ToolName)))
		}
	}

	return strings.Join(parts, " ")
}

func init() {
	registry.Register("errors", NewErrorsSection)
}
