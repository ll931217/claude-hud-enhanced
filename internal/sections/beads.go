package sections

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/beads"
	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/git"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
)

// BeadsSection displays beads issue tracking information
type BeadsSection struct {
	*BaseSection
	reader   *beads.Reader
	detector *git.Detector
}

// NewBeadsSection creates a new beads section (factory function for registry)
func NewBeadsSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	// Get current directory or use git repo root
	repoPath := getRepoPath()

	return &BeadsSection{
		BaseSection: NewBaseSection("beads", appConfig),
		reader:      beads.NewReader(repoPath),
		detector:    git.NewDetector(repoPath),
	}, nil
}

// Render returns the beads section output
func (b *BeadsSection) Render() string {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Load issues
	if err := b.reader.Load(ctx); err != nil {
		// Graceful degradation
		return "[Beads: not available]"
	}

	// Get current issue
	issue := b.reader.GetCurrentIssue()
	if issue == nil {
		// No active issue, show summary
		return b.formatSummary()
	}

	// Format issue display
	return b.formatIssue(issue)
}

// formatIssue formats an issue for display
func (b *BeadsSection) formatIssue(issue *beads.Issue) string {
	var parts []string

	// Status icon
	parts = append(parts, issue.Status.Icon())

	// Issue ID
	parts = append(parts, issue.ID)

	// Title (truncated if needed)
	title := issue.Title
	if len(title) > 40 {
		title = title[:37] + "..."
	}
	parts = append(parts, title)

	// Priority
	parts = append(parts, issue.GetPriorityLabel())

	// Todo progress (if available in description)
	if progress := b.extractTodoProgress(issue); progress != "" {
		parts = append(parts, progress)
	}

	return strings.Join(parts, " • ")
}

// extractTodoProgress extracts todo progress from issue description
func (b *BeadsSection) extractTodoProgress(issue *beads.Issue) string {
	// Look for todo patterns in description
	// Format: "- [x]" for completed, "- [ ]" for open
	desc := issue.Description

	var completed, total int
	lines := strings.Split(desc, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- [x]") || strings.HasPrefix(line, "* [x]") {
			completed++
			total++
		} else if strings.HasPrefix(line, "- [ ]") || strings.HasPrefix(line, "* [ ]") {
			total++
		}
	}

	if total > 0 {
		return fmt.Sprintf("%d/%d todos", completed, total)
	}

	return ""
}

// getRepoPath returns the git repository root path
func getRepoPath() string {
	// Try to get git root
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to current directory
		if cwd, err := os.Getwd(); err == nil {
			return cwd
		}
		return "."
	}

	return strings.TrimSpace(string(output))
}

// formatSummary formats a summary when no active issue
func (b *BeadsSection) formatSummary() string {
	// Get counts by status
	inProgressCount := b.reader.CountByStatus(beads.StatusInProgress)
	closedCount := b.reader.CountByStatus(beads.StatusClosed)

	total := b.reader.Count()

	// If no issues at all
	if total == 0 {
		return "[Beads: no issues]"
	}

	// Build summary
	var parts []string
	parts = append(parts, fmt.Sprintf("☍ %d total", total))

	if inProgressCount > 0 {
		parts = append(parts, fmt.Sprintf("↻ %d in progress", inProgressCount))
	}
	if openCount := b.reader.CountByStatus(beads.StatusOpen); openCount > 0 {
		parts = append(parts, fmt.Sprintf("○ %d open", openCount))
	}
	if closedCount > 0 {
		parts = append(parts, fmt.Sprintf("✓ %d closed", closedCount))
	}

	return strings.Join(parts, " • ")
}

// getStatusSection returns the git status section
func (b *BeadsSection) getStatusSection() *registry.Section {
	// This would be used to combine beads and status sections
	return nil
}
