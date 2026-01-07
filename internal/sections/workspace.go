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
	// Update system metrics
	if err := w.monitor.Update(); err != nil {
		return "[Workspace: unavailable]"
	}

	var parts []string

	// Add directory path
	if dir := w.monitor.FormatDirDisplay(); dir != "" {
		parts = append(parts, "📁", dir)
	}

	// Add language detection
	if lang := w.monitor.FormatLanguageDisplay(); lang != "" {
		parts = append(parts, lang)
	}

	// Add system resources
	var resources []string

	if cpu := w.monitor.FormatCPUDisplay(); cpu != "" {
		resources = append(resources, cpu)
	}

	if mem := w.monitor.FormatMemoryDisplay(); mem != "" {
		resources = append(resources, mem)
	}

	if disk := w.monitor.FormatDiskDisplay(); disk != "" {
		resources = append(resources, disk)
	}

	if len(resources) > 0 {
		parts = append(parts, strings.Join(resources, " | "))
	}

	if len(parts) == 0 {
		return "[Workspace: waiting for data]"
	}

	return strings.Join(parts, " ")
}
