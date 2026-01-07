package icons

import "testing"

func TestIcon_String(t *testing.T) {
	icon := Icon{NerdFont: "🦀", ASCII: "Rs"}
	if icon.String() != "🦀" {
		t.Errorf("Expected NerdFont icon, got %s", icon.String())
	}
}

func TestIcon_Fallback(t *testing.T) {
	icon := Icon{NerdFont: "🦀", ASCII: "Rs"}
	if icon.Fallback() != "Rs" {
		t.Errorf("Expected ASCII fallback, got %s", icon.Fallback())
	}
}

func TestIcon_String_EmptyNerdFont(t *testing.T) {
	icon := Icon{NerdFont: "", ASCII: "Rs"}
	if icon.String() != "Rs" {
		t.Errorf("Expected ASCII when NerdFont empty, got %s", icon.String())
	}
}

func TestGitIcons(t *testing.T) {
	tests := []struct {
		name string
		icon Icon
	}{
		{"GitBranch", GitBranch},
		{"GitWorktree", GitWorktree},
		{"GitDirty", GitDirty},
		{"GitAhead", GitAhead},
		{"GitBehind", GitBehind},
		{"GitStash", GitStash},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.icon.String() == "" && tt.icon.ASCII == "" {
				t.Errorf("Icon %s has both NerdFont and ASCII empty", tt.name)
			}
		})
	}
}

func TestBeadsIcons(t *testing.T) {
	tests := []struct {
		name string
		icon Icon
	}{
		{"BeadsOpen", BeadsOpen},
		{"BeadsClosed", BeadsClosed},
		{"BeadsInProgress", BeadsInProgress},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.icon.String() == "" && tt.icon.ASCII == "" {
				t.Errorf("Icon %s has both NerdFont and ASCII empty", tt.name)
			}
		})
	}
}

func TestResourceIcons(t *testing.T) {
	if CPU.String() == "" {
		t.Error("CPU icon is empty")
	}
	if RAM.String() == "" {
		t.Error("RAM icon is empty")
	}
	if Disk.String() == "" {
		t.Error("Disk icon is empty")
	}
}

func TestLanguageIcon(t *testing.T) {
	tests := []struct {
		lang     string
		contains string
	}{
		{"Go", "🐹"},
		{"Python", "🐍"},
		{"Rust", "🦀"},
		{"Ruby", "💎"},
		{"JavaScript", "🟨"},
		{"TypeScript", "💎"},
		{"Java", "☕"},
		{"Shell", "📜"},
		{"Unknown", "file"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			icon := LanguageIcon(tt.lang)
			if icon.String() == "" {
				t.Errorf("LanguageIcon(%s) returned empty", tt.lang)
			}
		})
	}
}

func TestPriorityIcon(t *testing.T) {
	tests := []struct {
		priority string
		expected Icon
	}{
		{"P0", PriorityCritical},
		{"P1", PriorityHigh},
		{"P2", PriorityMedium},
		{"P3", PriorityLow},
		{"P4", PriorityBacklog},
		{"critical", PriorityCritical},
		{"high", PriorityHigh},
		{"medium", PriorityMedium},
		{"low", PriorityLow},
		{"backlog", PriorityBacklog},
	}

	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			icon := PriorityIcon(tt.priority)
			if icon != tt.expected {
				t.Errorf("PriorityIcon(%s) = %v, want %v", tt.priority, icon, tt.expected)
			}
		})
	}
}

func TestGet(t *testing.T) {
	icon := Icon{NerdFont: "🦀", ASCII: "Rs"}

	// Test default (UseASCIIFallback = false)
	UseASCIIFallback = false
	if Get(icon) != "🦀" {
		t.Errorf("Expected NerdFont when UseASCIIFallback=false, got %s", Get(icon))
	}

	// Test ASCII fallback
	UseASCIIFallback = true
	if Get(icon) != "Rs" {
		t.Errorf("Expected ASCII when UseASCIIFallback=true, got %s", Get(icon))
	}

	// Reset for other tests
	UseASCIIFallback = false
}
