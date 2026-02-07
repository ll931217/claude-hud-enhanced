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
	ContextWindowSize int
	ContextInputTokens int
	ContextCacheTokens int
	Available      bool // true if JSON was successfully parsed
}

// Global context instance
var globalContext = &ClaudeCodeContext{}

// SetContext updates the global context from parsed JSON
func SetContext(transcriptPath, workspaceDir, modelName string) {
	SetContextWithWindow(transcriptPath, workspaceDir, modelName, 0, 0, 0)
}

// SetContextWithWindow updates the global context including context window data
func SetContextWithWindow(transcriptPath, workspaceDir, modelName string, contextWindowSize, contextInputTokens, contextCacheTokens int) {
	globalContext.mu.Lock()
	defer globalContext.mu.Unlock()
	globalContext.TranscriptPath = transcriptPath
	globalContext.WorkspaceDir = workspaceDir
	globalContext.ModelName = modelName
	globalContext.ContextWindowSize = contextWindowSize
	globalContext.ContextInputTokens = contextInputTokens
	globalContext.ContextCacheTokens = contextCacheTokens
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

// GetContextWindowSize returns the context window size from JSON input
func GetContextWindowSize() int {
	globalContext.mu.RLock()
	defer globalContext.mu.RUnlock()
	return globalContext.ContextWindowSize
}

// GetContextInputTokens returns the input token count from JSON input
func GetContextInputTokens() int {
	globalContext.mu.RLock()
	defer globalContext.mu.RUnlock()
	return globalContext.ContextInputTokens
}

// GetContextCacheTokens returns the cache token count from JSON input
func GetContextCacheTokens() int {
	globalContext.mu.RLock()
	defer globalContext.mu.RUnlock()
	return globalContext.ContextCacheTokens
}
