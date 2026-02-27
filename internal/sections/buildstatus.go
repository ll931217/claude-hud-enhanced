package sections

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
	"github.com/ll931217/claude-hud-enhanced/internal/system"
)

// BuildStatusSection displays build status (pass/fail)
type BuildStatusSection struct {
	*BaseSection
	mu            sync.RWMutex
	lastStatus    string
	lastCheck     time.Time
	cacheDuration time.Duration
}

// NewBuildStatusSection creates a new build status section (factory function for registry)
func NewBuildStatusSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	base := NewBaseSection("buildstatus", appConfig)
	base.SetPriority(registry.PriorityImportant) // Important for code health
	base.SetMinWidth(12)                          // Minimum width for build status

	return &BuildStatusSection{
		BaseSection:   base,
		cacheDuration: 30 * time.Second, // Cache build status for 30 seconds
	}, nil
}

// Render returns the build status section output
func (b *BuildStatusSection) Render() string {
	// Check cache first
	b.mu.RLock()
	if time.Since(b.lastCheck) < b.cacheDuration && b.lastStatus != "" {
		cached := b.lastStatus
		b.mu.RUnlock()
		return cached
	}
	b.mu.RUnlock()

	// Detect language
	lang := system.DetectLanguage(".")
	if lang == "" {
		return "" // Hide section if language not detected
	}

	// Get build status based on language
	status := b.getBuildStatus(lang)

	// Update cache
	b.mu.Lock()
	b.lastStatus = status
	b.lastCheck = time.Now()
	b.mu.Unlock()

	return status
}

// getBuildStatus retrieves build status for the detected language
func (b *BuildStatusSection) getBuildStatus(lang string) string {
	var success bool
	var errorCount int
	var err error

	switch lang {
	case "Go":
		success, errorCount, err = b.getGoBuildStatus()
	case "JavaScript", "TypeScript":
		success, errorCount, err = b.getTSBuildStatus()
	case "Python":
		success, errorCount, err = b.getPythonBuildStatus()
	default:
		return "" // Unsupported language
	}

	if err != nil {
		return "" // Hide section on error
	}

	// Format status
	if success {
		return "🔨 ✓ Build"
	}

	if errorCount > 0 {
		return fmt.Sprintf("🔨 ✗ %d errors", errorCount)
	}

	return "🔨 ✗ Failed"
}

// getGoBuildStatus checks Go build status
func (b *BuildStatusSection) getGoBuildStatus() (bool, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "./...")
	output, err := cmd.CombinedOutput()

	if err == nil {
		return true, 0, nil
	}

	// Count errors in output
	errorCount := strings.Count(string(output), "error:")
	if errorCount == 0 {
		errorCount = 1 // At least 1 error if build failed
	}

	return false, errorCount, nil
}

// getTSBuildStatus checks TypeScript build status
func (b *BuildStatusSection) getTSBuildStatus() (bool, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Try tsc first
	cmd := exec.CommandContext(ctx, "npx", "tsc", "--noEmit")
	output, err := cmd.CombinedOutput()

	if err == nil {
		return true, 0, nil
	}

	// Count errors in output
	// TypeScript errors look like: "src/file.ts(10,5): error TS2322:"
	errorCount := strings.Count(string(output), ": error TS")
	if errorCount == 0 {
		// Try counting lines with "error" in them
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "error") {
				errorCount++
			}
		}
	}

	if errorCount == 0 {
		errorCount = 1 // At least 1 error if build failed
	}

	return false, errorCount, nil
}

// getPythonBuildStatus checks Python "build" status (syntax check)
func (b *BuildStatusSection) getPythonBuildStatus() (bool, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use mypy for type checking (if available)
	cmd := exec.CommandContext(ctx, "mypy", ".")
	output, err := cmd.CombinedOutput()

	if err == nil {
		return true, 0, nil
	}

	// Count errors in output
	errorCount := strings.Count(string(output), "error:")
	if errorCount == 0 {
		errorCount = 1
	}

	return false, errorCount, nil
}

func init() {
	registry.Register("buildstatus", NewBuildStatusSection)
}
