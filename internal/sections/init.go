package sections

import (
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
)

func init() {
	// Note: Most sections register themselves via their own init() functions.
	// This init() is kept for backward compatibility but is technically redundant.
	// The sections below are handled by their individual init() functions:
	// - model, contextbar, duration, beads, status, workspace, tools, sysinfo

	// This block ensures sections are imported and their init() functions run.
	// The import statement in main.go (_ "github.com/.../internal/sections")
	// triggers all init() functions in this package.
	_ = registry.DefaultRegistry() // Force import of registry package
}
