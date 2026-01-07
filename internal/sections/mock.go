package sections

import (
	"fmt"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
)

// MockSection is a simple mock section for testing
type MockSection struct {
	*BaseSection
	content string
}

// NewMockSection creates a new mock section
func NewMockSection(name string, content string, cfg *config.Config) *MockSection {
	return &MockSection{
		BaseSection: NewBaseSection(name, cfg),
		content:     content,
	}
}

// Render returns the mock content
func (m *MockSection) Render() string {
	return m.content
}

// SetContent updates the mock content
func (m *MockSection) SetContent(content string) {
	m.content = content
}

// ExampleSessionSection creates a mock session section
func ExampleSessionSection(cfg *config.Config) Section {
	return NewMockSection("session", "[Opus 4.5] ████░░░░░░ 19% | 2 CLAUDE.md | 8 rules | ⏱️ 1m", cfg)
}

// ExampleBeadsSection creates a mock beads section
func ExampleBeadsSection(cfg *config.Config) Section {
	return NewMockSection("beads", "✗ bd-auth: Fix JWT token validation | P1 | 3/5 todos", cfg)
}

// ExampleGitSection creates a mock git section
func ExampleGitSection(cfg *config.Config) Section {
	return NewMockSection("git", "🌿 feature-auth | 📁 src/auth | 🦀 Rust", cfg)
}

// ExampleResourcesSection creates a mock resources section
func ExampleResourcesSection(cfg *config.Config) Section {
	return NewMockSection("resources", "💻 CPU: 45% | RAM: 4.2GB/16GB | 💾 120GB free", cfg)
}

// ExampleEmptySection creates a section that returns empty content
func ExampleEmptySection(cfg *config.Config) Section {
	return NewMockSection("empty", "", cfg)
}

// ExampleErrorSection creates a section that simulates an error
type ErrorSection struct {
	*BaseSection
}

// NewErrorSection creates a new error section
func NewErrorSection(name string, cfg *config.Config) *ErrorSection {
	return &ErrorSection{
		BaseSection: NewBaseSection(name, cfg),
	}
}

// Render simulates an error by returning an error indicator
func (e *ErrorSection) Render() string {
	return fmt.Sprintf("⚠️ Error in %s section", e.Name())
}
