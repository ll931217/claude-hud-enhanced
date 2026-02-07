package registry

// Priority represents the display priority of a section for responsive layout
type Priority int

const (
	PriorityEssential Priority = iota // Always show (model, context)
	PriorityImportant                 // Show when space permits (todos, git, language)
	PriorityOptional                  // Hide first when space constrained (tools, cost, system info)
)

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

	// Priority returns the display priority for responsive layouts
	Priority() Priority

	// MinWidth returns the minimum columns needed to display this section
	MinWidth() int
}
