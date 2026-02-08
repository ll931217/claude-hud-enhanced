package statusline

import (
	"strings"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
	"github.com/ll931217/claude-hud-enhanced/internal/terminal"
)

// BreakpointLevel represents terminal size category
type BreakpointLevel int

const (
	BreakpointSmall BreakpointLevel = iota
	BreakpointMedium
	BreakpointLarge
)

// ResponsiveRenderer handles adaptive layout based on terminal size
type ResponsiveRenderer struct {
	config   *config.Config
	sections map[string]registry.Section
}

// NewResponsiveRenderer creates a new responsive renderer
func NewResponsiveRenderer(cfg *config.Config, sections map[string]registry.Section) *ResponsiveRenderer {
	return &ResponsiveRenderer{
		config:   cfg,
		sections: sections,
	}
}

// RenderLayout renders sections according to available space
func (r *ResponsiveRenderer) RenderLayout() []string {
	termWidth := terminal.AvailableWidth()

	// If terminal width is 0 (non-TTY/statusline mode), assume large terminal
	// Claude Code will handle the actual layout
	if termWidth == 0 {
		sections := r.getAllSections()
		return r.layoutSections(sections, 0)
	}

	// Determine breakpoint level
	breakpoint := r.getBreakpoint(termWidth)

	// Filter sections by priority based on breakpoint
	filteredSections := r.filterSectionsByPriority(breakpoint)

	// Layout sections into lines
	lines := r.layoutSections(filteredSections, termWidth)

	return lines
}

func (r *ResponsiveRenderer) getAllSections() []registry.Section {
	var result []registry.Section
	for _, section := range r.sections {
		if section.Enabled() {
			result = append(result, section)
		}
	}
	return result
}

func (r *ResponsiveRenderer) getBreakpoint(width int) BreakpointLevel {
	cfg := r.config.Layout.Responsive

	if !cfg.Enabled {
		return BreakpointLarge // No constraints
	}

	if width < cfg.Small {
		return BreakpointSmall // < 80 cols
	}
	if width < cfg.Medium {
		return BreakpointMedium // 80-119 cols
	}
	return BreakpointLarge // 120+ cols
}

func (r *ResponsiveRenderer) filterSectionsByPriority(level BreakpointLevel) []registry.Section {
	var result []registry.Section

	for _, section := range r.sections {
		if !section.Enabled() {
			continue
		}

		// Filter based on breakpoint
		switch level {
		case BreakpointSmall:
			// Only essential sections
			if section.Priority() == registry.PriorityEssential {
				result = append(result, section)
			}
		case BreakpointMedium:
			// Essential + Important
			if section.Priority() <= registry.PriorityImportant {
				result = append(result, section)
			}
		case BreakpointLarge:
			// All sections
			result = append(result, section)
		}
	}

	return result
}

func (r *ResponsiveRenderer) layoutSections(sections []registry.Section, maxWidth int) []string {
	// Group sections by their configured line
	lineGroups := r.groupSectionsByLine(sections)

	var lines []string

	for _, group := range lineGroups {
		line := r.buildLine(group, maxWidth)
		if line != "" {
			lines = append(lines, line)
		}
	}

	return lines
}

func (r *ResponsiveRenderer) groupSectionsByLine(sections []registry.Section) [][]registry.Section {
	if len(r.config.Layout.Lines) == 0 {
		// No layout configured, put all sections on one line
		return [][]registry.Section{sections}
	}

	// Create a map of section name to section
	sectionMap := make(map[string]registry.Section)
	for _, section := range sections {
		sectionMap[section.Name()] = section
	}

	// Group sections by their configured line
	var lineGroups [][]registry.Section

	for _, lineConfig := range r.config.Layout.Lines {
		var group []registry.Section
		for _, sectionName := range lineConfig.Sections {
			if section, ok := sectionMap[sectionName]; ok {
				group = append(group, section)
			}
		}
		if len(group) > 0 {
			lineGroups = append(lineGroups, group)
		}
	}

	return lineGroups
}

func (r *ResponsiveRenderer) buildLine(sections []registry.Section, maxWidth int) string {
	var parts []string
	currentWidth := 0

	for _, section := range sections {
		content := section.Render()
		if content == "" {
			continue
		}

		contentWidth := len(content) + len(" | ") // Include separator

		// Check if we have space (maxWidth of 0 means no limit)
		if maxWidth > 0 && currentWidth+contentWidth > maxWidth {
			// Try to fit by truncating or skipping
			if currentWidth == 0 {
				// First item, force fit with truncation
				parts = append(parts, truncate(content, maxWidth))
			}
			break // Skip this item
		}

		parts = append(parts, content)
		currentWidth += contentWidth
	}

	return strings.Join(parts, " | ")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}
