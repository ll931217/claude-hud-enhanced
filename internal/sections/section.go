package sections

import (
	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
)

// Section represents a renderable section of the statusline
type Section interface {
	// Render returns the rendered content for this section
	// Returns an empty string if the section has nothing to display
	Render() string

	// Enabled returns true if this section should be displayed
	Enabled() bool

	// Name returns the section identifier
	Name() string

	// Order returns the display order (lower values first)
	Order() int
}

// BaseSection provides common functionality for all sections
type BaseSection struct {
	name     string
	enabled  bool
	order    int
	config   *config.Config
	priority registry.Priority
	minWidth int
}

// NewBaseSection creates a new base section
// Order is set to a default value (999) since actual ordering is determined by layout
func NewBaseSection(name string, cfg *config.Config) *BaseSection {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	return &BaseSection{
		name:    name,
		enabled: cfg.IsSectionEnabled(name),
		order:   999, // Default order - actual ordering determined by layout
		config:  cfg,
	}
}

// Name returns the section identifier
func (b *BaseSection) Name() string {
	return b.name
}

// Enabled returns true if this section should be displayed
func (b *BaseSection) Enabled() bool {
	return b.enabled
}

// Order returns the display order
func (b *BaseSection) Order() int {
	return b.order
}

// GetConfig returns the full config
func (b *BaseSection) GetConfig() *config.Config {
	return b.config
}

// Priority returns the display priority for responsive layouts
func (b *BaseSection) Priority() registry.Priority {
	if b.priority == 0 {
		return registry.PriorityImportant // Default
	}
	return b.priority
}

// MinWidth returns the minimum columns needed to display this section
func (b *BaseSection) MinWidth() int {
	return b.minWidth
}

// SetPriority sets the priority for this section
func (b *BaseSection) SetPriority(p registry.Priority) {
	b.priority = p
}

// SetMinWidth sets the minimum width for this section
func (b *BaseSection) SetMinWidth(w int) {
	b.minWidth = w
}
