#!/bin/bash
# Claude HUD Enhanced Statusline Wrapper for Claude Code
# This script bridges claude-hud with Claude Code's statusline protocol

set -euo pipefail

# Paths - prioritize project directory binary during development
CLAUDE_HUD_BIN="${CLAUDE_HUD_BIN:-}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Find working binary - check multiple locations
if [[ -x "$SCRIPT_DIR/../bin/claude-hud" ]]; then
    # Development: script is in project root/scripts/
    CLAUDE_HUD_BIN="$SCRIPT_DIR/../bin/claude-hud"
elif [[ -x "/home/ll931217/Projects/claude-hud-enhanced/bin/claude-hud" ]]; then
    # Hardcoded project path for development
    CLAUDE_HUD_BIN="/home/ll931217/Projects/claude-hud-enhanced/bin/claude-hud"
elif [[ -x "$HOME/.claude/claude-hud-new" ]]; then
    CLAUDE_HUD_BIN="$HOME/.claude/claude-hud-new"
elif [[ -x "$HOME/.claude/claude-hud" ]]; then
    CLAUDE_HUD_BIN="$HOME/.claude/claude-hud"
fi

# Default values
TRANSCRIPT_PATH=""
CWD="${PWD}"

# Parse JSON from Claude Code if available
if [[ -t 0 ]]; then
    # No stdin, use current directory
    :
else
    # Parse JSON from Claude Code
    INPUT=$(cat)
    CWD=$(echo "$INPUT" | jq -r '.workspace.current_dir // .cwd // "'"$PWD"'"')
    TRANSCRIPT_PATH=$(echo "$INPUT" | jq -r '.transcript_path // ""')
fi

# Change to working directory
cd "$CWD" 2>/dev/null || true

# Export environment variables for claude-hud
export CLAUDE_HUD_TRANSCRIPT_PATH="$TRANSCRIPT_PATH"

# Run claude-hud if available
if [[ -x "$CLAUDE_HUD_BIN" ]]; then
    # Run in single-shot mode with timeout
    timeout 3s "$CLAUDE_HUD_BIN" --statusline 2>/dev/null || echo "Claude HUD"
else
    # Fallback if claude-hud not found
    echo "Claude HUD"
fi
