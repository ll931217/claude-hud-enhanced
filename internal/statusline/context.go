package statusline

import (
	"sync"
)

// ClaudeCodeContext holds data from Claude Code's statusline JSON input
type ClaudeCodeContext struct {
	mu             sync.RWMutex
	TranscriptPath string
	WorkspaceDir   string
	ModelName      string
	Available      bool // true if JSON was successfully parsed
}

// Global context instance
var globalContext = &ClaudeCodeContext{}

// SetContext updates the global context from parsed JSON
func SetContext(transcriptPath, workspaceDir, modelName string) {
	globalContext.mu.Lock()
	defer globalContext.mu.Unlock()
	globalContext.TranscriptPath = transcriptPath
	globalContext.WorkspaceDir = workspaceDir
	globalContext.ModelName = modelName
	globalContext.Available = true
}

// GetTranscriptPath returns the transcript path from context
func GetTranscriptPath() string {
	globalContext.mu.RLock()
	defer globalContext.mu.RUnlock()
	return globalContext.TranscriptPath
}

// GetWorkspaceDir returns the workspace directory from context
func GetWorkspaceDir() string {
	globalContext.mu.RLock()
	defer globalContext.mu.RUnlock()
	return globalContext.WorkspaceDir
}

// GetModelName returns the model name from context
func GetModelName() string {
	globalContext.mu.RLock()
	defer globalContext.mu.RUnlock()
	return globalContext.ModelName
}

// IsContextAvailable returns true if Claude Code context was set
func IsContextAvailable() bool {
	globalContext.mu.RLock()
	defer globalContext.mu.RUnlock()
	return globalContext.Available
}
