package sections

import (
	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
)

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
