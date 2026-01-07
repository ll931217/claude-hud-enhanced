package registry

// Section represents a renderable component in the HUD
type Section interface {
	// Render returns the formatted string representation of this section
	Render() string

	// Enabled returns true if this section should be displayed
	Enabled() bool

	// Order returns the display order for this section (lower values appear first)
	Order() int

	// Name returns the unique identifier for this section type
	Name() string
}
