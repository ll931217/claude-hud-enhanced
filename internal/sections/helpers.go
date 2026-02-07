package sections

import (
	"os"

	"github.com/ll931217/claude-hud-enhanced/internal/statusline"
)

// getTranscriptPath returns the transcript path from context, environment, or default
func getTranscriptPath() string {
	// Check global context from Claude Code first
	if path := statusline.GetTranscriptPath(); path != "" {
		return path
	}

	// Fallback to environment variable (for standalone mode or wrapper script)
	if path := os.Getenv("CLAUDE_HUD_TRANSCRIPT_PATH"); path != "" {
		return path
	}

	// For standalone mode, try to find the latest transcript in cwd
	// Look for .claude/transcript.json or similar
	return ""
}
