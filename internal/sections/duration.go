package sections

import (
	"context"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
	"github.com/ll931217/claude-hud-enhanced/internal/transcript"
)

// DurationSection displays session duration
type DurationSection struct {
	*BaseSection
	parser *transcript.Parser
}

// NewDurationSection creates a new duration section (factory function for registry)
func NewDurationSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	transcriptPath := getTranscriptPath()
	base := NewBaseSection("duration", appConfig)
	base.SetPriority(registry.PriorityImportant) // Important but not essential
	base.SetMinWidth(2)                          // "0s" minimum

	return &DurationSection{
		BaseSection: base,
		parser:      transcript.NewParser(transcriptPath),
	}, nil
}

func init() {
	registry.Register("duration", NewDurationSection)
}

// Render returns the duration section output
func (d *DurationSection) Render() string {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_ = d.parser.Parse(ctx)
	return d.parser.GetDuration()
}
