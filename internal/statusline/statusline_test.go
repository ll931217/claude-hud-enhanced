package statusline

import (
	"context"
	"testing"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
)

// MockSection is a test implementation of registry.Section
type MockSection struct {
	name    string
	enabled bool
	order   int
	content string
	panicOn string // if set, will panic when this content is set
}

func (m *MockSection) Render() string {
	if m.panicOn != "" && m.content == m.panicOn {
		panic("intentional panic for testing")
	}
	return m.content
}

func (m *MockSection) Enabled() bool {
	return m.enabled
}

func (m *MockSection) Order() int {
	return m.order
}

func (m *MockSection) Name() string {
	return m.name
}

func (m *MockSection) Priority() registry.Priority {
	return registry.PriorityImportant
}

func (m *MockSection) MinWidth() int {
	return 0
}

func (m *MockSection) SetContent(content string) {
	m.content = content
}

func TestNewStatusline(t *testing.T) {
	cfg := config.DefaultConfig()
	statusline, err := New(cfg, nil)

	if err != nil {
		t.Fatalf("New() should not return error, got: %v", err)
	}

	if statusline == nil {
		t.Fatal("New() should return a Statusline instance")
	}

	if statusline.config != cfg {
		t.Error("Statusline should store the provided config")
	}

	if statusline.refreshInterval != 300*time.Millisecond {
		t.Errorf("Expected refresh interval 300ms, got %v", statusline.refreshInterval)
	}
}

func TestNewStatuslineWithNilConfig(t *testing.T) {
	_, err := New(nil, nil)

	if err == nil {
		t.Error("New() with nil config should return error")
	}
}

func TestAddSection(t *testing.T) {
	cfg := config.DefaultConfig()
	statusline, _ := New(cfg, nil)

	section := &MockSection{
		name:    "test",
		enabled: true,
		order:   1,
		content: "test content",
	}

	statusline.AddSection(section)

	sections := statusline.GetSections()
	if len(sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(sections))
	}

	if sections[0].Name() != "test" {
		t.Errorf("Expected section name 'test', got %s", sections[0].Name())
	}
}

func TestRemoveSection(t *testing.T) {
	cfg := config.DefaultConfig()
	statusline, _ := New(cfg, nil)

	section1 := &MockSection{name: "test1", enabled: true, order: 1, content: "content1"}
	section2 := &MockSection{name: "test2", enabled: true, order: 2, content: "content2"}

	statusline.AddSection(section1)
	statusline.AddSection(section2)
	statusline.RemoveSection("test1")

	sections := statusline.GetSections()
	if len(sections) != 1 {
		t.Fatalf("Expected 1 section after removal, got %d", len(sections))
	}

	if sections[0].Name() != "test2" {
		t.Errorf("Expected remaining section name 'test2', got %s", sections[0].Name())
	}
}

func TestSetSections(t *testing.T) {
	cfg := config.DefaultConfig()
	statusline, _ := New(cfg, nil)

	section1 := &MockSection{name: "test1", enabled: true, order: 1, content: "content1"}
	section2 := &MockSection{name: "test2", enabled: true, order: 2, content: "content2"}

	statusline.SetSections([]registry.Section{section1, section2})

	sections := statusline.GetSections()
	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}
}

func TestSectionSorting(t *testing.T) {
	cfg := config.DefaultConfig()
	statusline, _ := New(cfg, nil)

	// Add sections in random order
	statusline.AddSection(&MockSection{name: "c", enabled: true, order: 3, content: "c"})
	statusline.AddSection(&MockSection{name: "a", enabled: true, order: 1, content: "a"})
	statusline.AddSection(&MockSection{name: "b", enabled: true, order: 2, content: "b"})

	sections := statusline.GetSections()
	if len(sections) != 3 {
		t.Fatalf("Expected 3 sections, got %d", len(sections))
	}

	// Check they are sorted by order
	if sections[0].Name() != "a" || sections[1].Name() != "b" || sections[2].Name() != "c" {
		t.Error("Sections should be sorted by order")
	}
}

func TestRender(t *testing.T) {
	cfg := config.DefaultConfig()
	statusline, _ := New(cfg, nil)

	section1 := &MockSection{name: "test1", enabled: true, order: 1, content: "line 1"}
	section2 := &MockSection{name: "test2", enabled: true, order: 2, content: "line 2"}

	statusline.AddSection(section1)
	statusline.AddSection(section2)

	// Render should not error
	err := statusline.Render()
	if err != nil {
		t.Errorf("Render() should not return error, got: %v", err)
	}
}

func TestRenderSkipsDisabledSections(t *testing.T) {
	cfg := config.DefaultConfig()
	statusline, _ := New(cfg, nil)

	section1 := &MockSection{name: "enabled", enabled: true, order: 1, content: "enabled content"}
	section2 := &MockSection{name: "disabled", enabled: false, order: 2, content: "disabled content"}

	statusline.AddSection(section1)
	statusline.AddSection(section2)

	err := statusline.Render()
	if err != nil {
		t.Errorf("Render() should not return error, got: %v", err)
	}

	// The disabled section should not cause any output
	// (we can't easily test stdout, but we can ensure no error)
}

func TestRenderSkipsEmptySections(t *testing.T) {
	cfg := config.DefaultConfig()
	statusline, _ := New(cfg, nil)

	section1 := &MockSection{name: "content", enabled: true, order: 1, content: "has content"}
	section2 := &MockSection{name: "empty", enabled: true, order: 2, content: ""}

	statusline.AddSection(section1)
	statusline.AddSection(section2)

	err := statusline.Render()
	if err != nil {
		t.Errorf("Render() should not return error, got: %v", err)
	}
}

func TestRenderHandlesPanic(t *testing.T) {
	cfg := config.DefaultConfig()
	statusline, _ := New(cfg, nil)

	// Create a section that will panic
	panicSection := &MockSection{
		name:    "panic",
		enabled: true,
		order:   1,
		content: "panic content",
		panicOn: "panic content",
	}

	normalSection := &MockSection{
		name:    "normal",
		enabled: true,
		order:   2,
		content: "normal content",
	}

	statusline.AddSection(panicSection)
	statusline.AddSection(normalSection)

	// Render should not panic and should not return error
	err := statusline.Render()
	if err != nil {
		t.Errorf("Render() should not return error even with panicking section, got: %v", err)
	}
}

func TestSetRefreshInterval(t *testing.T) {
	cfg := config.DefaultConfig()
	statusline, _ := New(cfg, nil)

	newInterval := 500 * time.Millisecond
	statusline.SetRefreshInterval(newInterval)

	if statusline.refreshInterval != newInterval {
		t.Errorf("Expected refresh interval %v, got %v", newInterval, statusline.refreshInterval)
	}

	// Test that zero or negative values are handled
	statusline.SetRefreshInterval(0)
	if statusline.refreshInterval != 300*time.Millisecond {
		t.Errorf("Expected default interval 300ms for zero value, got %v", statusline.refreshInterval)
	}

	statusline.SetRefreshInterval(-100 * time.Millisecond)
	if statusline.refreshInterval != 300*time.Millisecond {
		t.Errorf("Expected default interval 300ms for negative value, got %v", statusline.refreshInterval)
	}
}

func TestRefresh(t *testing.T) {
	cfg := config.DefaultConfig()
	statusline, _ := New(cfg, nil)

	section := &MockSection{name: "test", enabled: true, order: 1, content: "test content"}
	statusline.AddSection(section)

	err := statusline.Refresh()
	if err != nil {
		t.Errorf("Refresh() should not return error, got: %v", err)
	}
}

func TestRun(t *testing.T) {
	cfg := config.DefaultConfig()
	// Set a short interval for testing
	cfg.RefreshIntervalMs = 10
	statusline, _ := New(cfg, nil)

	section := &MockSection{name: "test", enabled: true, order: 1, content: "test content"}
	statusline.AddSection(section)

	// Create a context that will be cancelled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Run should complete when context is done
	err := statusline.Run(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("Run() should return DeadlineExceeded, got: %v", err)
	}
}

func TestStop(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RefreshIntervalMs = 100
	statusline, _ := New(cfg, nil)

	section := &MockSection{name: "test", enabled: true, order: 1, content: "test content"}
	statusline.AddSection(section)

	// Start Run in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	go func() {
		done <- statusline.Run(ctx)
	}()

	// Let it run a bit, then stop
	time.Sleep(50 * time.Millisecond)
	statusline.Stop()
	cancel()

	// Run should return
	select {
	case err := <-done:
		if err != nil && err != context.Canceled {
			t.Errorf("Run() returned unexpected error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Run() did not return after Stop()")
	}
}

func TestRenderWithNoSections(t *testing.T) {
	cfg := config.DefaultConfig()
	statusline, _ := New(cfg, nil)

	// Render with no sections should not error
	err := statusline.Render()
	if err != nil {
		t.Errorf("Render() with no sections should not return error, got: %v", err)
	}
}
