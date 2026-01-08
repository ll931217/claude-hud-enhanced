package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// ClaudeCodeInput represents the JSON input from Claude Code
type ClaudeCodeInput struct {
	Workspace      WorkspaceInfo `json:"workspace"`
	TranscriptPath string        `json:"transcript_path"`
	Model          ModelInfo     `json:"model"`
}

type WorkspaceInfo struct {
	CurrentDir string `json:"current_dir"`
}

type ModelInfo struct {
	DisplayName string `json:"display_name"`
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
