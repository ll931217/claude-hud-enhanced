package sections

import (
	"github.com/ll931217/claude-hud-enhanced/internal/config"
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
	name    string
	enabled bool
	order   int
	config  *config.Config
}

// NewBaseSection creates a new base section
func NewBaseSection(name string, cfg *config.Config) *BaseSection {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	return &BaseSection{
		name:    name,
		enabled: cfg.IsSectionEnabled(name),
		order:   cfg.GetSectionOrder(name),
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
