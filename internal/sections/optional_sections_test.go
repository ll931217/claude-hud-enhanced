package sections

import (
	"testing"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
)

// TestAgentsSectionCreation tests that the agents section can be created
func TestAgentsSectionCreation(t *testing.T) {
	cfg := config.DefaultConfig()
	section, err := NewAgentsSection(cfg)
	if err != nil {
		t.Fatalf("Failed to create agents section: %v", err)
	}

	if section == nil {
		t.Fatal("Expected section to be non-nil")
	}

	if section.Name() != "agents" {
		t.Errorf("Expected name 'agents', got '%s'", section.Name())
	}

	if section.Priority() != registry.PriorityEssential {
		t.Errorf("Expected priority Essential, got %v", section.Priority())
	}

	if section.MinWidth() != 30 {
		t.Errorf("Expected min width 30, got %d", section.MinWidth())
	}
}

// TestAgentsSectionRender tests agents section rendering
func TestAgentsSectionRender(t *testing.T) {
	cfg := config.DefaultConfig()
	section, err := NewAgentsSection(cfg)
	if err != nil {
		t.Fatalf("Failed to create agents section: %v", err)
	}

	// Render should not panic
	output := section.Render()

	// Output can be empty if no transcript or no agents
	// Just verify it doesn't panic
	_ = output
}

// TestCostSectionCreation tests that the cost section can be created
func TestCostSectionCreation(t *testing.T) {
	cfg := config.DefaultConfig()
	section, err := NewCostSection(cfg)
	if err != nil {
		t.Fatalf("Failed to create cost section: %v", err)
	}

	if section == nil {
		t.Fatal("Expected section to be non-nil")
	}

	if section.Name() != "cost" {
		t.Errorf("Expected name 'cost', got '%s'", section.Name())
	}

	if section.Priority() != registry.PriorityImportant {
		t.Errorf("Expected priority Important, got %v", section.Priority())
	}

	if section.MinWidth() != 10 {
		t.Errorf("Expected min width 10, got %d", section.MinWidth())
	}
}

// TestCostSectionRender tests cost section rendering
func TestCostSectionRender(t *testing.T) {
	cfg := config.DefaultConfig()
	section, err := NewCostSection(cfg)
	if err != nil {
		t.Fatalf("Failed to create cost section: %v", err)
	}

	// Render should not panic
	output := section.Render()

	// Output can be empty if no transcript or no costs
	_ = output
}

// TestTodoProgressSectionCreation tests that the todo progress section can be created
func TestTodoProgressSectionCreation(t *testing.T) {
	cfg := config.DefaultConfig()
	section, err := NewTodoProgressSection(cfg)
	if err != nil {
		t.Fatalf("Failed to create todo progress section: %v", err)
	}

	if section == nil {
		t.Fatal("Expected section to be non-nil")
	}

	if section.Name() != "todoprogress" {
		t.Errorf("Expected name 'todoprogress', got '%s'", section.Name())
	}

	if section.Priority() != registry.PriorityEssential {
		t.Errorf("Expected priority Essential, got %v", section.Priority())
	}

	if section.MinWidth() != 20 {
		t.Errorf("Expected min width 20, got %d", section.MinWidth())
	}
}

// TestTodoProgressSectionRender tests todo progress section rendering
func TestTodoProgressSectionRender(t *testing.T) {
	cfg := config.DefaultConfig()
	section, err := NewTodoProgressSection(cfg)
	if err != nil {
		t.Fatalf("Failed to create todo progress section: %v", err)
	}

	// Render should not panic
	output := section.Render()

	// Output can be empty if no transcript or no todos
	_ = output
}

// TestErrorsSectionCreation tests that the errors section can be created
func TestErrorsSectionCreation(t *testing.T) {
	cfg := config.DefaultConfig()
	section, err := NewErrorsSection(cfg)
	if err != nil {
		t.Fatalf("Failed to create errors section: %v", err)
	}

	if section == nil {
		t.Fatal("Expected section to be non-nil")
	}

	if section.Name() != "errors" {
		t.Errorf("Expected name 'errors', got '%s'", section.Name())
	}

	if section.Priority() != registry.PriorityImportant {
		t.Errorf("Expected priority Important, got %v", section.Priority())
	}

	if section.MinWidth() != 15 {
		t.Errorf("Expected min width 15, got %d", section.MinWidth())
	}
}

// TestErrorsSectionRender tests errors section rendering
func TestErrorsSectionRender(t *testing.T) {
	cfg := config.DefaultConfig()
	section, err := NewErrorsSection(cfg)
	if err != nil {
		t.Fatalf("Failed to create errors section: %v", err)
	}

	// Render should not panic
	output := section.Render()

	// Output can be empty if no transcript or no errors
	_ = output
}

// TestTestCoverageSectionCreation tests that the test coverage section can be created
func TestTestCoverageSectionCreation(t *testing.T) {
	cfg := config.DefaultConfig()
	section, err := NewTestCoverageSection(cfg)
	if err != nil {
		t.Fatalf("Failed to create test coverage section: %v", err)
	}

	if section == nil {
		t.Fatal("Expected section to be non-nil")
	}

	if section.Name() != "testcoverage" {
		t.Errorf("Expected name 'testcoverage', got '%s'", section.Name())
	}

	if section.Priority() != registry.PriorityImportant {
		t.Errorf("Expected priority Important, got %v", section.Priority())
	}

	if section.MinWidth() != 15 {
		t.Errorf("Expected min width 15, got %d", section.MinWidth())
	}
}

// TestTestCoverageSectionRender tests test coverage section rendering
func TestTestCoverageSectionRender(t *testing.T) {
	cfg := config.DefaultConfig()
	section, err := NewTestCoverageSection(cfg)
	if err != nil {
		t.Fatalf("Failed to create test coverage section: %v", err)
	}

	// Render should not panic
	output := section.Render()

	// Output can be empty if language not detected or coverage not available
	_ = output
}

// TestBuildStatusSectionCreation tests that the build status section can be created
func TestBuildStatusSectionCreation(t *testing.T) {
	cfg := config.DefaultConfig()
	section, err := NewBuildStatusSection(cfg)
	if err != nil {
		t.Fatalf("Failed to create build status section: %v", err)
	}

	if section == nil {
		t.Fatal("Expected section to be non-nil")
	}

	if section.Name() != "buildstatus" {
		t.Errorf("Expected name 'buildstatus', got '%s'", section.Name())
	}

	if section.Priority() != registry.PriorityImportant {
		t.Errorf("Expected priority Important, got %v", section.Priority())
	}

	if section.MinWidth() != 12 {
		t.Errorf("Expected min width 12, got %d", section.MinWidth())
	}
}

// TestBuildStatusSectionRender tests build status section rendering
func TestBuildStatusSectionRender(t *testing.T) {
	cfg := config.DefaultConfig()
	section, err := NewBuildStatusSection(cfg)
	if err != nil {
		t.Fatalf("Failed to create build status section: %v", err)
	}

	// Render should not panic
	output := section.Render()

	// Output can be empty if language not detected or build not available
	_ = output
}

// TestShortenAgentName tests the agent name shortening function
func TestShortenAgentName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"planner", "Plan"},
		{"code-reviewer", "Review"},
		{"architect", "Arch"},
		{"tdd-guide", "TDD"},
		{"security-reviewer", "Sec"},
		{"build-error-resolver", "Build"},
		{"e2e-runner", "E2E"},
		{"refactor-cleaner", "Refactor"},
		{"doc-updater", "Docs"},
		{"debugger", "Debug"},
		{"general-purpose", "GP"},
		{"Explore", "Explore"},
		{"unknown-agent", "unknown-"},
		{"short", "short"},
		{"verylongagentname", "verylong"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := shortenAgentName(tt.input)
			if result != tt.expected {
				t.Errorf("shortenAgentName(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestTruncateTaskName tests the task name truncation function
func TestTruncateTaskName(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"Short task", 30, "Short task"},
		{"This is a very long task name that needs truncation", 30, "This is a very long task na..."},
		{"activeForm: Doing something", 30, "Doing something"},
		{"activeForm:   Spaced prefix", 30, "Spaced prefix"},
		{"", 30, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncateTaskName(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateTaskName(%q, %d) = %q; want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestOptionalSectionsRegistered tests that all optional sections are registered
func TestOptionalSectionsRegistered(t *testing.T) {
	requiredSections := []string{
		"agents",
		"cost",
		"todoprogress",
		"errors",
		"testcoverage",
		"buildstatus",
	}

	cfg := config.DefaultConfig()

	for _, name := range requiredSections {
		t.Run(name, func(t *testing.T) {
			section, err := registry.Create(name, cfg)
			if err != nil {
				t.Fatalf("Section %q not registered: %v", name, err)
			}
			if section == nil {
				t.Fatalf("Section %q created but is nil", name)
			}
			if section.Name() != name {
				t.Errorf("Section name mismatch: got %q, want %q", section.Name(), name)
			}
		})
	}
}
