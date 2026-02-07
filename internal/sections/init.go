package sections

import (
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
)

func init() {
	// Register all built-in section types
	registry.Register("session", NewSessionSection)
	registry.Register("beads", NewBeadsSection)
	registry.Register("status", NewStatusSection)
	registry.Register("workspace", NewWorkspaceSection)
	registry.Register("tools", NewToolsSection)
	registry.Register("sysinfo", NewSysInfoSection)
}
