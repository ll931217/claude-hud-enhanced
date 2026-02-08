package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ClaudeCodeInput represents the JSON input from Claude Code
type ClaudeCodeInput struct {
	Workspace      WorkspaceInfo       `json:"workspace"`
	TranscriptPath string              `json:"transcript_path"`
	Model          ModelInfo           `json:"model"`
	ContextWindow  *ContextWindowInput `json:"context_window,omitempty"`
}

type WorkspaceInfo struct {
	CurrentDir string `json:"current_dir"`
}

type ModelInfo struct {
	DisplayName string `json:"display_name"`
}

// ContextWindowInput contains context usage information from Claude Code
type ContextWindowInput struct {
	CurrentUsage      UsageInfoInput `json:"current_usage"`
	ContextWindowSize int            `json:"context_window_size"`
}

// UsageInfoInput contains token usage breakdown
type UsageInfoInput struct {
	InputTokens              int `json:"input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	OutputTokens             int `json:"output_tokens"`
}

// readStdinJSON reads and parses JSON from stdin
func readStdinJSON() (*ClaudeCodeInput, error) {
	// Check if stdin is a terminal (no input)
	fileInfo, _ := os.Stdin.Stat()
	if fileInfo.Mode()&os.ModeCharDevice != 0 {
		return nil, nil // No stdin data
	}

	// Read all input from stdin
	var input ClaudeCodeInput
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &input, nil
}

// logStdinDebug logs the stdin JSON input to a file for debugging
func logStdinDebug(input *ClaudeCodeInput) {
	debugLogMutex.Lock()
	defer debugLogMutex.Unlock()

	const debugLogPath = "/tmp/claude-hud-debug.log"

	// Marshal input to formatted JSON
	data, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		data = []byte(fmt.Sprintf("Error marshaling JSON: %v", err))
	}

	// Create log entry with timestamp
	logEntry := fmt.Sprintf("=== %s ===\n%s\n\n", time.Now().Format(time.RFC3339), data)

	// Append to log file
	f, err := os.OpenFile(debugLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(logEntry)
}
