package sections

import (
	"strings"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
	"github.com/ll931217/claude-hud-enhanced/internal/system"
)

// WorkspaceSection displays workspace information
type WorkspaceSection struct {
	*BaseSection
	monitor *system.Monitor
}

// NewWorkspaceSection creates a new workspace section (factory function for registry)
func NewWorkspaceSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	return &WorkspaceSection{
		BaseSection: NewBaseSection("workspace", appConfig),
		monitor:     system.NewMonitor(),
	}, nil
}

// Render returns the workspace section output
func (w *WorkspaceSection) Render() string {
	// Update monitor for language detection and directory
	if err := w.monitor.Update(); err != nil {
		return "[Workspace: unavailable]"
	}

	var parts []string

	// Language first (with icon)
	if lang := w.monitor.FormatLanguageDisplay(); lang != "" {
		parts = append(parts, lang)
	}

	// Then directory
	if dir := w.monitor.FormatDirDisplay(); dir != "" {
		parts = append(parts, dir)
	}

	// Note: System metrics (CPU, RAM, Disk) are now in sysinfo section

	if len(parts) == 0 {
		return "[Workspace: waiting for data]"
	}

	return strings.Join(parts, " | ")
}
