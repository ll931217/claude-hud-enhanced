package sections

import (
	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
)

// SessionSection displays session/worktrunk information
type SessionSection struct {
	*BaseSection
}

// NewSessionSection creates a new session section (factory function for registry)
func NewSessionSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	return &SessionSection{
		BaseSection: NewBaseSection("session", appConfig),
	}, nil
}

// Render returns the session section output
func (s *SessionSection) Render() string {
	return "[Session: worktrunk info]"
}

// BeadsSection displays beads issue tracking information
type BeadsSection struct {
	*BaseSection
}

// NewBeadsSection creates a new beads section (factory function for registry)
func NewBeadsSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	return &BeadsSection{
		BaseSection: NewBaseSection("beads", appConfig),
	}, nil
}

// Render returns the beads section output
func (b *BeadsSection) Render() string {
	return "[Beads: issue tracking]"
}

// StatusSection displays git status information
type StatusSection struct {
	*BaseSection
}

// NewStatusSection creates a new status section (factory function for registry)
func NewStatusSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	return &StatusSection{
		BaseSection: NewBaseSection("status", appConfig),
	}, nil
}

// Render returns the status section output
func (g *StatusSection) Render() string {
	return "[Status: git info]"
}

// WorkspaceSection displays workspace information
type WorkspaceSection struct {
	*BaseSection
}

// NewWorkspaceSection creates a new workspace section (factory function for registry)
func NewWorkspaceSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	return &WorkspaceSection{
		BaseSection: NewBaseSection("workspace", appConfig),
	}, nil
}

// Render returns the workspace section output
func (r *WorkspaceSection) Render() string {
	return "[Workspace: resources/directory]"
}
