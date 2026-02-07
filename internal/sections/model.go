package sections

import (
	"strings"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
	"github.com/ll931217/claude-hud-enhanced/internal/statusline"
)

// ModelSection displays the Claude model name
type ModelSection struct {
	*BaseSection
}

// NewModelSection creates a new model section (factory function for registry)
func NewModelSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	base := NewBaseSection("model", appConfig)
	base.SetPriority(registry.PriorityEssential) // Essential - always show model
	base.SetMinWidth(6)                          // Shortest model name like "gpt-4"

	return &ModelSection{
		BaseSection: base,
	}, nil
}

func init() {
	registry.Register("model", NewModelSection)
}

// Render returns the model section output
func (m *ModelSection) Render() string {
	model := statusline.GetModelName()
	if model == "" {
		return ""
	}

	// Shorten model name
	model = strings.ReplaceAll(model, "Claude ", "")
	model = strings.ReplaceAll(model, "Sonnet", "SN")
	model = strings.ReplaceAll(model, "Haiku", "HK")
	model = strings.ReplaceAll(model, "Opus", "OP")

	return model
}
