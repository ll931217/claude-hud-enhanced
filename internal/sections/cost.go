package sections

import (
	"context"
	"fmt"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
	"github.com/ll931217/claude-hud-enhanced/internal/transcript"
)

// CostSection displays accumulated API costs for the session
type CostSection struct {
	*BaseSection
	parser *transcript.Parser
}

// NewCostSection creates a new cost section (factory function for registry)
func NewCostSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	// Get transcript path from environment or use default
	transcriptPath := getTranscriptPath()

	base := NewBaseSection("cost", appConfig)
	base.SetPriority(registry.PriorityImportant) // Important but not essential
	base.SetMinWidth(10)                          // Minimum width for cost display

	return &CostSection{
		BaseSection: base,
		parser:      transcript.NewParser(transcriptPath),
	}, nil
}

// Render returns the cost section output
func (c *CostSection) Render() string {
	// Get transcript path dynamically from global context
	transcriptPath := getTranscriptPath()
	if transcriptPath == "" {
		return "" // Hide section if no transcript path
	}

	// Create a parser for the current transcript path
	parser := transcript.NewParser(transcriptPath)

	// Parse transcript for token data
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := parser.Parse(ctx); err != nil {
		return "" // Hide section on parse error
	}

	// Calculate cost
	cost := parser.CalculateCost()
	if cost == 0 {
		return "" // Hide section if no cost yet
	}

	// Get session duration for rate calculation
	duration := time.Since(parser.GetSessionStart())
	if duration < time.Minute {
		// Too early to show cost
		return ""
	}

	// Format cost based on magnitude
	var costStr string
	if cost < 0.01 {
		costStr = fmt.Sprintf("$%.4f", cost)
	} else if cost < 1.0 {
		costStr = fmt.Sprintf("$%.3f", cost)
	} else {
		costStr = fmt.Sprintf("$%.2f", cost)
	}

	// Calculate rate per hour
	hoursElapsed := duration.Hours()
	if hoursElapsed > 0.1 { // Only show rate after 6 minutes
		ratePerHour := cost / hoursElapsed
		var rateStr string
		if ratePerHour < 0.01 {
			rateStr = fmt.Sprintf("$%.4f/h", ratePerHour)
		} else if ratePerHour < 1.0 {
			rateStr = fmt.Sprintf("$%.3f/h", ratePerHour)
		} else {
			rateStr = fmt.Sprintf("$%.2f/h", ratePerHour)
		}
		return fmt.Sprintf("💰 %s (%s)", costStr, rateStr)
	}

	return fmt.Sprintf("💰 %s", costStr)
}

func init() {
	registry.Register("cost", NewCostSection)
}
