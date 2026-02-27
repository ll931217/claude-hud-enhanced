package sections

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
	"github.com/ll931217/claude-hud-enhanced/internal/transcript"
)

// AgentsSection displays agent activity (running and recently completed)
type AgentsSection struct {
	*BaseSection
	parser *transcript.Parser
}

// NewAgentsSection creates a new agents section (factory function for registry)
func NewAgentsSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	// Get transcript path from environment or use default
	transcriptPath := getTranscriptPath()

	base := NewBaseSection("agents", appConfig)
	base.SetPriority(registry.PriorityEssential) // Show on all terminals
	base.SetMinWidth(30)                         // Minimum width for agent names

	return &AgentsSection{
		BaseSection: base,
		parser:      transcript.NewParser(transcriptPath),
	}, nil
}

// Render returns the agents section output
func (a *AgentsSection) Render() string {
	// Get transcript path dynamically from global context
	transcriptPath := getTranscriptPath()
	if transcriptPath == "" {
		return "" // Hide section if no transcript path
	}

	// Create a parser for the current transcript path
	parser := transcript.NewParser(transcriptPath)

	// Parse transcript for agent data
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := parser.Parse(ctx); err != nil {
		return "" // Hide section on parse error
	}

	// Get agent activity
	agents := parser.GetAgentActivity()
	if len(agents) == 0 {
		return "" // Hide section when no agents active
	}

	// Separate running and completed agents
	var running, completed []agentDisplay
	for _, agent := range agents {
		display := agentDisplay{
			name:      shortenAgentName(agent.AgentName),
			status:    agent.Status,
			agentType: agent.Type,
		}

		if agent.Status == "running" {
			running = append(running, display)
		} else if agent.Status == "completed" || agent.Status == "success" {
			completed = append(completed, display)
		}
	}

	// If no running or completed agents, hide section
	if len(running) == 0 && len(completed) == 0 {
		return ""
	}

	var parts []string

	// Display running agents (max 2) with ◐ spinner
	for i, agent := range running {
		if i >= 2 {
			break
		}
		parts = append(parts, fmt.Sprintf("◐ %s", agent.name))
	}

	// Display recently completed agents (max 3) with ✓
	sort.Slice(completed, func(i, j int) bool {
		// Sort by completion time (most recent first)
		return true // Simplified - in real scenario, track timestamp
	})

	for i, agent := range completed {
		if i >= 3 {
			break
		}
		parts = append(parts, fmt.Sprintf("✓ %s", agent.name))
	}

	// Show count if there are more agents
	remaining := len(running) + len(completed) - 5
	if remaining > 0 {
		parts = append(parts, fmt.Sprintf("+%d", remaining))
	}

	return strings.Join(parts, " | ")
}

// agentDisplay holds formatted agent information for display
type agentDisplay struct {
	name      string
	status    string
	agentType string
}

// shortenAgentName shortens agent names for display
func shortenAgentName(name string) string {
	// Map of agent names to shortened versions
	shortNames := map[string]string{
		"planner":              "Plan",
		"code-reviewer":        "Review",
		"architect":            "Arch",
		"tdd-guide":            "TDD",
		"security-reviewer":    "Sec",
		"build-error-resolver": "Build",
		"e2e-runner":           "E2E",
		"refactor-cleaner":     "Refactor",
		"doc-updater":          "Docs",
		"debugger":             "Debug",
		"general-purpose":      "GP",
		"Explore":              "Explore",
	}

	// Check if we have a shortened version
	if short, ok := shortNames[name]; ok {
		return short
	}

	// Fallback: use first 8 characters
	if len(name) > 8 {
		return name[:8]
	}
	return name
}

func init() {
	registry.Register("agents", NewAgentsSection)
}
