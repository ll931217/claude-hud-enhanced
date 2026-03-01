package sections

import (
	"fmt"
	"strings"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
	"github.com/ll931217/claude-hud-enhanced/internal/theme"
	"github.com/ll931217/claude-hud-enhanced/internal/zai"
)

// ZaiUsageSection displays Z.ai coding plan usage metrics
type ZaiUsageSection struct {
	*BaseSection
	client *zai.Client
}

// NewZaiUsageSection creates a new Z.ai usage section (factory function for registry)
func NewZaiUsageSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	base := NewBaseSection("zaiusage", appConfig)
	base.SetPriority(registry.PriorityImportant) // Show on medium+ terminals (80+ cols)
	base.SetMinWidth(20)                         // Minimum width for display

	return &ZaiUsageSection{
		BaseSection: base,
		client:      zai.NewClient(),
	}, nil
}

// Render returns the Z.ai usage section output
func (s *ZaiUsageSection) Render() string {
	info := s.client.Fetch()
	if info == nil || info.IsEmpty() {
		return ""
	}

	var parts []string

	// Session usage (5-hour rolling window)
	if info.SessionPercent > 0 {
		sessionDisplay := fmt.Sprintf("%d%%", info.SessionPercent)
		color := s.getUsageColor(info.SessionPercent)
		if color != "" {
			sessionDisplay = fmt.Sprintf("%s%s%s", color, sessionDisplay, theme.Reset)
		}
		parts = append(parts, "🔋 "+sessionDisplay)
	}

	// Weekly usage
	if info.WeeklyPercent > 0 {
		weeklyDisplay := fmt.Sprintf("%d%%", info.WeeklyPercent)
		color := s.getUsageColor(info.WeeklyPercent)
		if color != "" {
			weeklyDisplay = fmt.Sprintf("%s%s%s", color, weeklyDisplay, theme.Reset)
		}
		parts = append(parts, "📊 "+weeklyDisplay)
	}

	// Search usage (monthly)
	if info.SearchPercent > 0 {
		searchDisplay := fmt.Sprintf("%d%%", info.SearchPercent)
		color := s.getUsageColor(info.SearchPercent)
		if color != "" {
			searchDisplay = fmt.Sprintf("%s%s%s", color, searchDisplay, theme.Reset)
		}
		parts = append(parts, "🔍 "+searchDisplay)
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " | ")
}

// getUsageColor returns the color based on usage percentage
func (s *ZaiUsageSection) getUsageColor(percent int) string {
	switch {
	case percent >= 90:
		return theme.Red // Red for critical
	case percent >= 70:
		return theme.Yellow // Yellow for warning
	default:
		return "" // Default terminal color (green implied by low usage)
	}
}

func init() {
	registry.Register("zaiusage", NewZaiUsageSection)
}
