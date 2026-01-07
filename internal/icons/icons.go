package icons

// Icon provides Nerd Font and ASCII fallback icons
type Icon struct {
	NerdFont string
	ASCII    string
}

// String returns the Nerd Font icon (or ASCII if unavailable)
func (i Icon) String() string {
	if i.NerdFont != "" {
		return i.NerdFont
	}
	return i.ASCII
}

// Fallback returns the ASCII fallback icon
func (i Icon) Fallback() string {
	return i.ASCII
}

// Predefined icons for various UI elements
var (
	// Git icons
	GitBranch    = Icon{NerdFont: "", ASCII: "branch"}
	GitWorktree  = Icon{NerdFont: "🌿", ASCII: "worktree"}
	GitDirty     = Icon{NerdFont: "", ASCII: "*"}
	GitAhead     = Icon{NerdFont: "↑", ASCII: "ahead"}
	GitBehind    = Icon{NerdFont: "↓", ASCII: "behind"}
	GitStash     = Icon{NerdFont: "≡", ASCII: "stash"}

	// Beads (issue tracker) status icons
	BeadsOpen       = Icon{NerdFont: "✗", ASCII: "[open]"}
	BeadsClosed     = Icon{NerdFont: "✓", ASCII: "[done]"}
	BeadsInProgress = Icon{NerdFont: "◐", ASCII: "[in progress]"}

	// Resource icons
	CPU  = Icon{NerdFont: "💻", ASCII: "CPU"}
	RAM  = Icon{NerdFont: "🎯", ASCII: "RAM"}
	Disk = Icon{NerdFont: "💾", ASCII: "Disk"}

	// Language icons
	Go         = Icon{NerdFont: "🐹", ASCII: "Go"}
	Python     = Icon{NerdFont: "🐍", ASCII: "Py"}
	Rust       = Icon{NerdFont: "🦀", ASCII: "Rs"}
	Ruby       = Icon{NerdFont: "💎", ASCII: "Rb"}
	JavaScript = Icon{NerdFont: "🟨", ASCII: "JS"}
	TypeScript = Icon{NerdFont: "💎", ASCII: "TS"}
	Java       = Icon{NerdFont: "☕", ASCII: "Java"}
	C          = Icon{NerdFont: "🔧", ASCII: "C"}
	CPP        = Icon{NerdFont: "⚙️", ASCII: "C++"}
	CSharp     = Icon{NerdFont: "🔷", ASCII: "C#"}
	Swift      = Icon{NerdFont: "🍎", ASCII: "Sw"}
	Shell      = Icon{NerdFont: "📜", ASCII: "Sh"}
	PHP        = Icon{NerdFont: "🐘", ASCII: "PHP"}
	Kotlin     = Icon{NerdFont: "🎯", ASCII: "Kt"}

	// Directory and file icons
	Directory = Icon{NerdFont: "📁", ASCII: "dir"}
	File      = Icon{NerdFont: "📄", ASCII: "file"}

	// Time and session icons
	Clock       = Icon{NerdFont: "⏱️", ASCII: "time"}
	Session     = Icon{NerdFont: "🤖", ASCII: "AI"}
	Context     = Icon{NerdFont: "📊", ASCII: "ctx"}
	Agent       = Icon{NerdFont: "↻", ASCII: "agent"}
	Tool        = Icon{NerdFont: "✓", ASCII: "ok"}

	// Status icons
	Loading = Icon{NerdFont: "◐", ASCII: "..."}
	Waiting = Icon{NerdFont: "◌", ASCII: "-"}
	Success = Icon{NerdFont: "✓", ASCII: "OK"}
	Error   = Icon{NerdFont: "✗", ASCII: "X"}
	Warning = Icon{NerdFont: "⚠", ASCII: "!"}
	Info    = Icon{NerdFont: "ℹ", ASCII: "i"}

	// Priority icons
	PriorityCritical = Icon{NerdFont: "🔴", ASCII: "P0"}
	PriorityHigh    = Icon{NerdFont: "🟠", ASCII: "P1"}
	PriorityMedium  = Icon{NerdFont: "🟡", ASCII: "P2"}
	PriorityLow     = Icon{NerdFont: "🟢", ASCII: "P3"}
	PriorityBacklog = Icon{NerdFont: "⚪", ASCII: "P4"}
)

// LanguageIcon returns the icon for a programming language
func LanguageIcon(lang string) Icon {
	switch lang {
	case "Go":
		return Go
	case "Python":
		return Python
	case "Rust":
		return Rust
	case "Ruby":
		return Ruby
	case "JavaScript":
		return JavaScript
	case "TypeScript":
		return TypeScript
	case "Java":
		return Java
	case "C":
		return C
	case "C++":
		return CPP
	case "C#":
		return CSharp
	case "Swift":
		return Swift
	case "Shell":
		return Shell
	case "PHP":
		return PHP
	case "Kotlin":
		return Kotlin
	default:
		return File
	}
}

// PriorityIcon returns the icon for a priority level
func PriorityIcon(priority string) Icon {
	switch priority {
	case "P0", "0", "critical":
		return PriorityCritical
	case "P1", "1", "high":
		return PriorityHigh
	case "P2", "2", "medium":
		return PriorityMedium
	case "P3", "3", "low":
		return PriorityLow
	case "P4", "4", "backlog":
		return PriorityBacklog
	default:
		return Info
	}
}

// UseASCIIFallback forces all icons to use ASCII fallback
// Set this to true if the terminal doesn't support Nerd Fonts
var UseASCIIFallback = false

// Get returns the appropriate icon based on terminal support
func Get(i Icon) string {
	if UseASCIIFallback {
		return i.Fallback()
	}
	return i.String()
}
