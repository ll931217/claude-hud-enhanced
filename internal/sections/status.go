package sections

import (
	"context"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/git"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
)

// StatusSection displays git status information
type StatusSection struct {
	*BaseSection
	detector *git.Detector
}

// NewStatusSection creates a new status section (factory function for registry)
func NewStatusSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	repoPath := getRepoPath()

	return &StatusSection{
		BaseSection: NewBaseSection("status", appConfig),
		detector:    git.NewDetector(repoPath),
	}, nil
}

// Render returns the status section output
func (s *StatusSection) Render() string {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	status, err := s.detector.Detect(ctx)
	if err != nil || status == nil {
		return "[Status: not a git repo]"
	}

	return status.FormatStatus()
}
