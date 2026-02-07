package theme

// ANSI escape codes
const (
	Reset = "\033[0m"
	Bold  = "\033[1m"
	Dim   = "\033[2m"
)

// ANSI color codes (256-color mode)
const (
	Green  = "\033[38;5;40m"
	Yellow = "\033[38;5;215m"
	Red    = "\033[38;5;203m"
)

// ContextColor returns the ANSI color code for a given context percentage
func ContextColor(percentage int) string {
	if percentage >= 85 {
		return Red
	}
	if percentage >= 70 {
		return Yellow
	}
	return "" // No color for low usage (user request)
}
