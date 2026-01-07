package theme

// Theme defines color constants for the statusline
type Theme struct {
	// Background colors
	Background string

	// Foreground colors
	Primary   string
	Secondary string
	Muted     string

	// Semantic colors
	Success string
	Warning string
	Error   string
	Info    string
}

// CatppuccinMocha returns the Catppuccin Mocha theme colors
// Reference: https://catppuccin.com/
func CatppuccinMocha() *Theme {
	return &Theme{
		Background: "#1E1E2E",

		Primary:   "#89dceb", // Sky
		Secondary: "#cba6f7", // Mauve
		Muted:     "#6c7086", // Overlay 0

		Success: "#a6e3a1", // Green
		Warning: "#fab387", // Peach
		Error:   "#f38ba8", // Red
		Info:    "#b4befe", // Lavender
	}
}

// Default returns the default theme (Catppuccin Mocha)
func Default() *Theme {
	return CatppuccinMocha()
}

// ColorNames returns a map of color names to hex values
func ColorNames() map[string]string {
	t := CatppuccinMocha()
	return map[string]string{
		"background": t.Background,
		"primary":    t.Primary,
		"secondary":  t.Secondary,
		"muted":      t.Muted,
		"success":    t.Success,
		"warning":    t.Warning,
		"error":      t.Error,
		"info":       t.Info,
	}
}

// ANSIColors returns a map of semantic names to ANSI color codes
// These can be used for terminal output with color support
func ANSIColors() map[string]int {
	return map[string]int{
		"primary":   38, // Blue/Cyan
		"secondary": 141, // Purple
		"muted":     245, // Gray
		"success":   40, // Green
		"warning":   215, // Orange
		"error":     203, // Red
		"info":      146, // Lavender
	}
}
