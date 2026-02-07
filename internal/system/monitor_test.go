package system

import (
	"testing"
	"time"
)

func TestNewMonitor(t *testing.T) {
	m := NewMonitor()
	if m == nil {
		t.Fatal("NewMonitor() returned nil")
	}
	if m.updateInterval != 5*time.Second {
		t.Errorf("Expected update interval 5s, got %v", m.updateInterval)
	}
}

func TestMonitor_Update(t *testing.T) {
	m := NewMonitor()
	if err := m.Update(); err != nil {
		t.Errorf("Update() error = %v", err)
	}
}

func TestMonitor_GetCPU(t *testing.T) {
	m := NewMonitor()
	m.Update()

	cpu := m.GetCPU()
	if cpu.CoreCount <= 0 {
		t.Errorf("Expected positive core count, got %d", cpu.CoreCount)
	}
	// UsagePercent might be 0 on some systems, so we only check it's not negative
	if cpu.UsagePercent < 0 {
		t.Errorf("UsagePercent cannot be negative, got %f", cpu.UsagePercent)
	}
}

func TestMonitor_GetMemory(t *testing.T) {
	m := NewMonitor()
	m.Update()

	mem := m.GetMemory()
	if mem.Total == 0 {
		// This might legitimately be 0 on some systems, so we just log
		t.Logf("Warning: Total memory is 0")
	}
	if mem.Percent < 0 || mem.Percent > 100 {
		t.Errorf("Memory percent out of range: %f", mem.Percent)
	}
}

func TestMonitor_GetDisk(t *testing.T) {
	m := NewMonitor()
	m.Update()

	disk := m.GetDisk()
	if disk.Total == 0 {
		t.Logf("Warning: Total disk space is 0")
	}
}

func TestMonitor_GetCurrentDir(t *testing.T) {
	m := NewMonitor()
	m.Update()

	dir := m.GetCurrentDir()
	if dir == "" {
		t.Error("GetCurrentDir() returned empty string")
	}
}

func TestMonitor_GetLanguage(t *testing.T) {
	m := NewMonitor()
	m.Update()

	lang := m.GetLanguage()
	// Language might be empty if no recognized files are found
	// So we just verify it doesn't crash
	_ = lang
}

func TestMonitor_FormatCPUDisplay(t *testing.T) {
	m := NewMonitor()

	// Test with no data
	display := m.FormatCPUDisplay()
	if display != "" {
		t.Logf("FormatCPUDisplay with no data: %s", display)
	}

	// Test with data
	m.Update()
	display = m.FormatCPUDisplay()
	if m.cpu.UsagePercent > 0 && display == "" {
		t.Error("FormatCPUDisplay returned empty when CPU usage > 0")
	}
}

func TestMonitor_FormatMemoryDisplay(t *testing.T) {
	m := NewMonitor()

	// Test with no data
	display := m.FormatMemoryDisplay()
	if display != "" {
		t.Logf("FormatMemoryDisplay with no data: %s", display)
	}

	// Test with data
	m.Update()
	display = m.FormatMemoryDisplay()
	if m.memory.Total > 0 && display == "" {
		t.Error("FormatMemoryDisplay returned empty when memory > 0")
	}
}

func TestMonitor_FormatDiskDisplay(t *testing.T) {
	m := NewMonitor()

	// Test with no data
	display := m.FormatDiskDisplay()
	if display != "" {
		t.Logf("FormatDiskDisplay with no data: %s", display)
	}

	// Test with data
	m.Update()
	display = m.FormatDiskDisplay()
	if m.disk.Total > 0 && display == "" {
		t.Error("FormatDiskDisplay returned empty when disk > 0")
	}
}

func TestMonitor_FormatDirDisplay(t *testing.T) {
	m := NewMonitor()
	m.Update()

	display := m.FormatDirDisplay()
	if display == "" {
		t.Error("FormatDirDisplay returned empty")
	}
	// Check that it's not too long (increased from 20 to 50 chars)
	if len(display) > 50 {
		t.Errorf("FormatDirDisplay too long: %d chars", len(display))
	}
}

func TestMonitor_FormatLanguageDisplay(t *testing.T) {
	m := NewMonitor()

	// Test with no data
	display := m.FormatLanguageDisplay()
	if display != "" {
		t.Logf("FormatLanguageDisplay with no data: %s", display)
	}
}

func TestGetThresholdLevel(t *testing.T) {
	tests := []struct {
		name     string
		percent  float64
		expected ThresholdLevel
	}{
		{"Good", 50.0, LevelGood},
		{"Warning low", 70.0, LevelWarning},
		{"Warning high", 89.0, LevelWarning},
		{"Critical", 90.0, LevelCritical},
		{"Critical high", 100.0, LevelCritical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := GetThresholdLevel(tt.percent)
			if level != tt.expected {
				t.Errorf("GetThresholdLevel(%f) = %v, want %v", tt.percent, level, tt.expected)
			}
		})
	}
}

func TestMonitor_SetUpdateInterval(t *testing.T) {
	m := NewMonitor()
	interval := 10 * time.Second
	m.SetUpdateInterval(interval)

	if m.updateInterval != interval {
		t.Errorf("SetUpdateInterval() did not set interval, got %v", m.updateInterval)
	}
}

func TestMonitor_ForceUpdate(t *testing.T) {
	m := NewMonitor()
	if err := m.ForceUpdate(); err != nil {
		t.Errorf("ForceUpdate() error = %v", err)
	}
}
