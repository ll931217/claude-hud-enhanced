package sections

import (
	"strings"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
	"github.com/ll931217/claude-hud-enhanced/internal/system"
)

// SysInfoSection displays system resource usage (CPU, RAM, Disk)
type SysInfoSection struct {
	*BaseSection
	monitor *system.Monitor
}

// NewSysInfoSection creates a new sysinfo section (factory function for registry)
func NewSysInfoSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	base := NewBaseSection("sysinfo", appConfig)
	base.SetPriority(registry.PriorityImportant) // Show on medium+ terminals (80+ cols)

	return &SysInfoSection{
		BaseSection: base,
		monitor:     system.NewMonitor(),
	}, nil
}

// Render returns the sysinfo section output
func (s *SysInfoSection) Render() string {
	// Update system metrics
	if err := s.monitor.Update(); err != nil {
		return ""
	}

	var parts []string

	// Add CPU usage
	if cpu := s.monitor.FormatCPUDisplay(); cpu != "" {
		parts = append(parts, cpu)
	}

	// Add Memory usage
	if mem := s.monitor.FormatMemoryDisplay(); mem != "" {
		parts = append(parts, mem)
	}

	// Add Disk usage
	if disk := s.monitor.FormatDiskDisplay(); disk != "" {
		parts = append(parts, disk)
	}

	// Add File Descriptor count
	if fd := s.monitor.FormatFDDisplay(); fd != "" {
		parts = append(parts, fd)
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " | ")
}

func init() {
	registry.Register("sysinfo", NewSysInfoSection)
}
