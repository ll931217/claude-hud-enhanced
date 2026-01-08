#!/bin/bash
# Claude HUD Enhanced Statusline Wrapper for Claude Code
# This script bridges claude-hud with Claude Code's statusline protocol

set -euo pipefail

# Paths - prioritize project directory binary during development
CLAUDE_HUD_BIN="${CLAUDE_HUD_BIN:-}"
PROJECT_BIN="$(dirname "${BASH_SOURCE[0]}")/../bin/claude-hud"

# Find working binary
if [[ -x "$PROJECT_BIN" ]]; then
    CLAUDE_HUD_BIN="$PROJECT_BIN"
elif [[ -x "$HOME/.claude/claude-hud-new" ]]; then
    CLAUDE_HUD_BIN="$HOME/.claude/claude-hud-new"
elif [[ -x "$HOME/.claude/claude-hud" ]]; then
    CLAUDE_HUD_BIN="$HOME/.claude/claude-hud"
fi

# Change to working directory first
if [[ -t 0 ]]; then
    # No stdin, use current directory
    cd "${PWD}" 2>/dev/null || true
else
    # Parse JSON from Claude Code
    INPUT=$(cat)
    CWD=$(echo "$INPUT" | jq -r '.workspace.current_dir // .cwd // "'"$PWD"'"')
    cd "$CWD" 2>/dev/null || true
fi

# Run claude-hud if available
if [[ -x "$CLAUDE_HUD_BIN" ]]; then
    # Run in single-shot mode with timeout
    timeout 3s "$CLAUDE_HUD_BIN" --statusline 2>/dev/null || echo "Claude HUD"
else
    # Fallback if claude-hud not found
    echo "Claude HUD"
fi
