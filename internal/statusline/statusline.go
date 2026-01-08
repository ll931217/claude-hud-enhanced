package statusline

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
)

// Statusline manages the rendering of the statusline display
type Statusline struct {
	// config holds the application configuration
	config *config.Config

	// registry manages section creation
	registry *registry.SectionRegistry

	// sections holds the active sections to render
	sections []registry.Section

	// mu protects concurrent access to sections
	mu sync.RWMutex

	// done is used to signal shutdown
	done chan struct{}

	// refreshInterval is how often to refresh the display
	refreshInterval time.Duration
}

// New creates a new Statusline instance
func New(cfg *config.Config, reg *registry.SectionRegistry) (*Statusline, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if reg == nil {
		reg = registry.DefaultRegistry()
	}

	interval := cfg.GetRefreshInterval()
	if interval <= 0 {
		interval = 300 * time.Millisecond
	}

	return &Statusline{
		config:         cfg,
		registry:       reg,
		sections:       make([]registry.Section, 0),
		done:           make(chan struct{}),
		refreshInterval: interval,
	}, nil
}

// AddSection adds a section to the statusline
func (s *Statusline) AddSection(section registry.Section) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sections = append(s.sections, section)
	s.sortSections()
}

// RemoveSection removes a section by name
func (s *Statusline) RemoveSection(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newSections := make([]registry.Section, 0, len(s.sections))
	for _, section := range s.sections {
		if section.Name() != name {
			newSections = append(newSections, section)
		}
	}
	s.sections = newSections
}

// SetSections replaces all sections with the provided list
func (s *Statusline) SetSections(sections []registry.Section) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sections = make([]registry.Section, len(sections))
	copy(s.sections, sections)
	s.sortSections()
}

// sortSections sorts sections by their order
func (s *Statusline) sortSections() {
	// Simple bubble sort for small lists
	n := len(s.sections)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if s.sections[j].Order() > s.sections[j+1].Order() {
				s.sections[j], s.sections[j+1] = s.sections[j+1], s.sections[j]
			}
		}
	}
}

// Render renders all enabled sections and outputs to stdout
func (s *Statusline) Render() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var lines []string

	// Render each section
	for _, section := range s.sections {
		// Skip disabled sections
		if !section.Enabled() {
			continue
		}

		// Render the section with error handling
		content := s.renderSection(section)

		// Skip empty sections
		if content == "" {
			continue
		}

		lines = append(lines, content)
	}

	// Output to stdout (for Claude Code statusline API)
	s.output(lines)

	return nil
}

// renderSection renders a single section with error handling
func (s *Statusline) renderSection(section registry.Section) string {
	// Recover from panics during rendering
	defer func() {
		if r := recover(); r != nil {
			if s.config.Debug {
				log.Printf("Panic rendering section %s: %v", section.Name(), r)
			}
		}
	}()

	// Render the section
	content := section.Render()

	// Handle render errors or empty results
	if content == "" {
		return ""
	}

	return content
}

// output writes the rendered lines to stdout
func (s *Statusline) output(lines []string) {
	// Clear previous output using ANSI escape code
	// Move cursor to beginning of line and clear
	fmt.Print("\r\033[K")

	// Output each section on its own line
	for i, line := range lines {
		if i > 0 {
			fmt.Println()
		}
		fmt.Print(line)
	}

	// Ensure the output is displayed immediately
	os.Stdout.Sync()
}

// Run starts the refresh loop
func (s *Statusline) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.refreshInterval)
	defer ticker.Stop()

	// Initial render
	if err := s.Render(); err != nil {
		if s.config.Debug {
			log.Printf("Initial render error: %v", err)
		}
	}

	// Refresh loop
	for {
		select {
		case <-ticker.C:
			if err := s.Render(); err != nil {
				if s.config.Debug {
					log.Printf("Render error: %v", err)
				}
				// Continue running despite render errors
			}

		case <-ctx.Done():
			// Shutdown requested
			return ctx.Err()

		case <-s.done:
			// Internal shutdown signal
			return nil
		}
	}
}

// Stop gracefully stops the statusline refresh loop
func (s *Statusline) Stop() {
	close(s.done)
}

// SetRefreshInterval updates the refresh interval
func (s *Statusline) SetRefreshInterval(interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if interval <= 0 {
		interval = 300 * time.Millisecond
	}

	s.refreshInterval = interval
}

// GetSections returns a copy of the current sections list
func (s *Statusline) GetSections() []registry.Section {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sections := make([]registry.Section, len(s.sections))
	copy(sections, s.sections)
	return sections
}

// Refresh triggers an immediate render
func (s *Statusline) Refresh() error {
	return s.Render()
}

// RenderStatuslineMode renders for Claude Code statusline (multiline, no ANSI clear codes)
func (s *Statusline) RenderStatuslineMode() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if compact mode is enabled
	if s.config.CompactMode {
		return s.renderCompactMode()
	}

	var lines []string

	// Render each section
	for _, section := range s.sections {
		// Skip disabled sections
		if !section.Enabled() {
			continue
		}

		// Render the section with error handling
		content := s.renderSection(section)

		// Skip empty sections
		if content == "" {
			continue
		}

		lines = append(lines, content)
	}

	// Output each line on its own line (no ANSI codes for Claude Code)
	for i, line := range lines {
		if i > 0 {
			fmt.Println()
		}
		fmt.Print(line)
	}

	return nil
}

// renderCompactMode renders sections in compact 2-line mode
func (s *Statusline) renderCompactMode() error {
	var line1, line2 []string

	// Line 1: Session + Beads + Git (project state)
	for _, section := range s.sections {
		if !section.Enabled() {
			continue
		}
		switch section.Name() {
		case "session", "beads", "status":
			content := s.renderSection(section)
			if content != "" {
				line1 = append(line1, content)
			}
		}
	}

	// Line 2: Workspace (environment)
	for _, section := range s.sections {
		if !section.Enabled() {
			continue
		}
		if section.Name() == "workspace" {
			content := s.renderSection(section)
			if content != "" {
				line2 = append(line2, content)
			}
		}
	}

	// Output with consistent separator
	if len(line1) > 0 {
		fmt.Print(strings.Join(line1, " | "))
	}
	if len(line2) > 0 {
		if len(line1) > 0 {
			fmt.Println()
		}
		fmt.Print(strings.Join(line2, " | "))
	}

	return nil
}
