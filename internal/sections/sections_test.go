package sections

import (
	"testing"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
)

func TestSectionRegistry(t *testing.T) {
	// Test that all built-in sections are registered
	t.Run("List returns all registered sections", func(t *testing.T) {
		sections := registry.List()

		expectedSections := []string{"model", "contextbar", "duration", "beads", "status", "workspace", "tools", "sysinfo"}

		for _, expected := range expectedSections {
			found := false
			for _, section := range sections {
				if section == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected section %q not found in registered sections: %v", expected, sections)
			}
		}
	})

	// Test creating sections
	t.Run("Create returns valid sections", func(t *testing.T) {
		testCases := []string{"model", "contextbar", "duration", "beads", "status", "workspace", "tools", "sysinfo"}

		for _, sectionType := range testCases {
			section, err := registry.Create(sectionType, nil)
			if err != nil {
				t.Errorf("Failed to create section %q: %v", sectionType, err)
				continue
			}

			// Verify Section interface is implemented
			if section.Name() != sectionType {
				t.Errorf("Expected section name %q, got %q", sectionType, section.Name())
			}

			if !section.Enabled() {
				t.Errorf("Expected section %q to be enabled by default", sectionType)
			}

			// Order should be set by default config (1-8)
			if section.Order() < 1 || section.Order() > 8 {
				t.Errorf("Expected section %q to have order between 1-8, got %d", sectionType, section.Order())
			}

			// Note: Some sections may return empty strings in test environment
			// - model: needs statusline context with model name
			// - contextbar, duration: needs transcript file
			// - tools: needs transcript file for tool activity
			// - sysinfo: monitor may fail to update in test
			rendered := section.Render()
			allowEmpty := (sectionType == "model" || sectionType == "tools" || sectionType == "sysinfo" || sectionType == "contextbar" || sectionType == "duration")
			if rendered == "" && !allowEmpty {
				t.Errorf("Expected section %q to render non-empty string", sectionType)
			}
		}
	})

	// Test configuration-based enable/disable
	t.Run("Create respects config.Enabled", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.Sections.Model.Enabled = false
		cfg.Sections.Model.Order = 5

		section, err := registry.Create("model", cfg)
		if err != nil {
			t.Fatalf("Failed to create section: %v", err)
		}

		if section.Enabled() {
			t.Error("Expected section to be disabled when config.Enabled is false")
		}

		if section.Order() != 5 {
			t.Errorf("Expected order 5, got %d", section.Order())
		}
	})

	// Test creating unregistered section type
	t.Run("Create fails for unregistered type", func(t *testing.T) {
		_, err := registry.Create("nonexistent", nil)
		if err == nil {
			t.Error("Expected error when creating unregistered section type")
		}
	})

	// Test Register function
	t.Run("Register adds new section type", func(t *testing.T) {
		customFactory := func(config interface{}) (registry.Section, error) {
			return &mockSection{name: "custom"}, nil
		}

		registry.Register("custom", customFactory)

		section, err := registry.Create("custom", nil)
		if err != nil {
			t.Fatalf("Failed to create custom section: %v", err)
		}

		if section.Name() != "custom" {
			t.Errorf("Expected section name 'custom', got %q", section.Name())
		}
	})
}

// mockSection is a test implementation of Section
type mockSection struct {
	name string
}

func (m *mockSection) Render() string {
	return "mock"
}

func (m *mockSection) Enabled() bool {
	return true
}

func (m *mockSection) Order() int {
	return 0
}

func (m *mockSection) Name() string {
	return m.name
}

func (m *mockSection) Priority() registry.Priority {
	return registry.PriorityImportant
}

func (m *mockSection) MinWidth() int {
	return 0
}
