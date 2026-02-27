package sections

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
	"github.com/ll931217/claude-hud-enhanced/internal/system"
)

// TestCoverageSection displays test coverage percentage
type TestCoverageSection struct {
	*BaseSection
	mu            sync.RWMutex
	lastCoverage  string
	lastCheck     time.Time
	cacheDuration time.Duration
}

// NewTestCoverageSection creates a new test coverage section (factory function for registry)
func NewTestCoverageSection(cfg interface{}) (registry.Section, error) {
	appConfig, ok := cfg.(*config.Config)
	if !ok {
		appConfig = config.DefaultConfig()
	}

	base := NewBaseSection("testcoverage", appConfig)
	base.SetPriority(registry.PriorityImportant) // Important for TDD workflow
	base.SetMinWidth(15)                         // Minimum width for coverage display

	return &TestCoverageSection{
		BaseSection:   base,
		cacheDuration: 30 * time.Second, // Cache coverage for 30 seconds
	}, nil
}

// Render returns the test coverage section output
func (t *TestCoverageSection) Render() string {
	// Check cache first
	t.mu.RLock()
	if time.Since(t.lastCheck) < t.cacheDuration && t.lastCoverage != "" {
		cached := t.lastCoverage
		t.mu.RUnlock()
		return cached
	}
	t.mu.RUnlock()

	// Detect language
	lang := system.DetectLanguage(".")
	if lang == "" {
		return "" // Hide section if language not detected
	}

	// Get coverage based on language
	coverage := t.getCoverage(lang)

	// Update cache
	t.mu.Lock()
	t.lastCoverage = coverage
	t.lastCheck = time.Now()
	t.mu.Unlock()

	return coverage
}

// getCoverage retrieves coverage for the detected language
func (t *TestCoverageSection) getCoverage(lang string) string {
	var coverage float64
	var err error

	switch lang {
	case "Go":
		coverage, err = t.getGoCoverage()
	case "JavaScript", "TypeScript":
		coverage, err = t.getJSCoverage()
	case "Python":
		coverage, err = t.getPythonCoverage()
	default:
		return "" // Unsupported language
	}

	if err != nil {
		return "" // Hide section on error
	}

	if coverage == 0 {
		return "" // Hide section if no coverage
	}

	// Color code based on coverage
	// Green: >= 80%, Yellow: 60-79%, Red: < 60%
	icon := "🧪"
	if coverage >= 80 {
		icon = "✓"
	} else if coverage < 60 {
		icon = "⚠️"
	}

	return fmt.Sprintf("%s %.1f%%", icon, coverage)
}

// getGoCoverage gets coverage for Go projects
func (t *TestCoverageSection) getGoCoverage() (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "test", "-cover", "./...")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}

	// Parse output for coverage
	// Expected format: "coverage: 85.2% of statements"
	re := regexp.MustCompile(`coverage:\s+([\d.]+)%`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return 0, fmt.Errorf("coverage not found in output")
	}

	coverage, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, err
	}

	return coverage, nil
}

// getJSCoverage gets coverage for JavaScript/TypeScript projects
func (t *TestCoverageSection) getJSCoverage() (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try Jest first (most common)
	cmd := exec.CommandContext(ctx, "npm", "run", "test:coverage", "--", "--silent")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try alternative command
		cmd = exec.CommandContext(ctx, "npx", "jest", "--coverage", "--silent")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return 0, err
		}
	}

	// Parse Jest coverage output
	// Expected format: "All files  | 85.2 | ..."
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "All files") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				coverageStr := strings.TrimSpace(fields[2])
				coverage, err := strconv.ParseFloat(coverageStr, 64)
				if err == nil {
					return coverage, nil
				}
			}
		}
	}

	return 0, fmt.Errorf("coverage not found in output")
}

// getPythonCoverage gets coverage for Python projects
func (t *TestCoverageSection) getPythonCoverage() (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try pytest with coverage
	cmd := exec.CommandContext(ctx, "pytest", "--cov", "--cov-report=term", "-q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}

	// Parse coverage output
	// Expected format: "TOTAL    100    20    80%"
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "TOTAL") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				coverageStr := strings.TrimSuffix(fields[len(fields)-1], "%")
				coverage, err := strconv.ParseFloat(coverageStr, 64)
				if err == nil {
					return coverage, nil
				}
			}
		}
	}

	return 0, fmt.Errorf("coverage not found in output")
}

func init() {
	registry.Register("testcoverage", NewTestCoverageSection)
}
