# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Claude HUD Enhanced** is a sophisticated statusline plugin for Claude Code sessions, written in Go. It provides real-time visibility into your development environment through a multi-line statusline display that integrates deeply with Claude Code's session data.

### What This Project Does

- **Real-time Statusline**: Displays contextual information about your Claude Code session including model, context usage, tool activity, agent status, and todo progress
- **Context Health Monitoring**: Shows color-coded context progress bar (green/yellow/red) with token breakdown at high usage (>=85%)
- **Multi-Source Integration**: Combines data from Claude Code stdin, transcript files, Git, Beads issue tracker, and system monitoring
- **Flexible Layout**: Supports both 2-line compact mode and 4-line full mode display
- **Cross-Platform**: Builds for Linux and macOS with single binary distribution

## Development Commands

### Build Commands

```bash
# Build for current platform (development)
make build

# Build release (optimized, trimmed paths)
make release

# Build for all platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64)
make release-all

# Create release archives (tar.gz)
make archives

# Install statusline for Claude Code (copies to ~/.claude/claude-hud)
make install-statusline
```

### Testing

```bash
# Run all tests with race detection and coverage
make test

# Run fast CI tests (short mode, 5min timeout)
make test-ci

# Run benchmarks
make benchmark

# Development workflow (format + lint + test)
make dev
```

### Code Quality

```bash
# Format Go code
make fmt

# Run linter (golangci-lint)
make lint
```

### Running the Application

```bash
# Run standalone mode (continuous refresh)
make run

# Or directly:
./bin/claude-hud

# Show version information
./bin/claude-hud --version
./bin/claude-hud --build-info
```

## Technology Stack

- **Language**: Go 1.25.5+
- **Build System**: Makefile with cross-platform compilation support
- **Configuration**: YAML config files with Catppuccin Mocha color scheme by default
- **Concurrency**: Goroutines with proper synchronization using mutexes
- **Error Handling**: Comprehensive error recovery via `internal/errors` package with graceful degradation
- **Dependencies**:
  - `fsnotify` for file watching
  - `yaml.v3` for configuration parsing

## Architecture Overview

### Core Design Patterns

1. **Section Registry Factory Pattern** (`internal/registry/`)
   - Dynamic section registration and creation
   - Thread-safe with `sync.RWMutex`
   - Sections register themselves via `init()` functions
   - Factory type: `SectionFactory func(config interface{}) (Section, error)`

2. **Statusline Orchestrator** (`internal/statusline/`)
   - Manages section rendering with configurable refresh intervals
   - Handles concurrent access with mutex protection
   - Supports graceful shutdown via context cancellation
   - Two operation modes: standalone (continuous) and statusline (single-shot)

3. **Configuration System** (`internal/config/`)
   - YAML-based configuration with defaults
   - Graceful fallback to defaults on errors
   - Section-level enable/disable and ordering

4. **Transcript Parser** (`internal/transcript/`)
   - Streaming JSONL parser for efficient transcript processing
   - Extracts tool usage, agent activity, and todos
   - Context-aware timeouts for safety

### Data Flow

```
Claude Code stdin → JSON context → statusline context → section rendering → stdout
                          ↓                    ↓
                   transcript_path → JSONL parsing → section data
```

### Two Operation Modes

**Standalone Mode** (default when run from terminal):
- Continuous refresh loop at configured interval (default 500ms)
- Displays real-time status updates
- Handles SIGINT/SIGTERM for graceful shutdown

**Statusline Mode** (auto-detected when stdin has data):
- Single-shot render for Claude Code integration
- Reads JSON context from stdin
- Extracts workspace directory and transcript path
- Changes to workspace directory if specified
- Renders once and exits

### Built-in Sections

All sections implement the `Section` interface from `internal/registry/section.go`:

```go
type Section interface {
    Render() string    // Returns the rendered content
    Enabled() bool     // Returns whether the section is enabled
    Order() int        // Returns the display order
    Name() string      // Returns the section name
    Priority() Priority  // Returns display priority (Essential/Important/Optional)
    MinWidth() int      // Returns minimum width required
}
```

**Model Section** (`internal/sections/model.go`):
- Model name only (e.g., "glm-4.7", "claude-opus-4-5")
- Shortens model names (Sonnet→SN, Haiku→HK, Opus→OP)
- Priority: Essential

**ContextBar Section** (`internal/sections/contextbar.go`):
- Context progress bar with color coding (green/yellow/red)
- Token breakdown at high context usage (>=85%)
- Priority: Essential

**Duration Section** (`internal/sections/duration.go`):
- Session duration in human-readable format
- Priority: Essential

**Tools Section** (`internal/sections/tools.go`):
- Recently used tools with call counts (max 5)
- Sorted by most recently used
- MCP plugin names are shortened
- Integrates with `internal/transcript/` for tool tracking
- Priority: Essential

**Beads Section** (`internal/sections/beads.go`):
- Issue tracker status (total/closed/in-progress)
- Current active task display
- Integrates with `internal/beads/` for issue parsing
- Priority: Important

**Status Section** (`internal/sections/status.go`):
- Git branch and dirty state
- Ahead/behind status and worktree info
- Integrates with `internal/git/` for git operations
- Priority: Important

**Workspace Section** (`internal/sections/workspace.go`):
- Language detection (with icon) before directory
- Current directory (truncated)
- Integrates with `internal/system/` for language detection
- Priority: Important

**SysInfo Section** (`internal/sections/sysinfo.go`):
- CPU, memory, and disk usage
- Integrates with `internal/system/` for system monitoring
- Priority: Optional (hidden first on small terminals)

### Error Handling Architecture

The `internal/errors/` package provides comprehensive error handling:

- `MainRecovery()` - Top-level panic recovery
- `SafeGo()` - Goroutine wrapper with panic recovery
- `LogErrorWithLevel()` - Structured error logging
- Graceful degradation - system continues working even when data sources are unavailable

## Configuration

### Configuration File Location

Default: `~/.config/claude-hud/config.yaml`

Auto-created on first run with defaults if not present.

### Key Configuration Options

```yaml
# Refresh interval in milliseconds (100-5000)
refresh_interval_ms: 500

# Configurable layout system
layout:
  responsive:
    enabled: true
    small_breakpoint: 80
    medium_breakpoint: 120
    large_breakpoint: 160
  lines:
    - sections: [model, contextbar, duration]
      separator: " | "
    - sections: [workspace, status, beads]
      separator: " | "
    - sections: [tools, sysinfo]
      separator: " | "

# Section configuration
sections:
  model:
    enabled: true
    order: 1
  contextbar:
    enabled: true
    order: 2
  duration:
    enabled: true
    order: 3
  beads:
    enabled: true
    order: 4
  status:
    enabled: true
    order: 5
  workspace:
    enabled: true
    order: 6
  tools:
    enabled: true
    order: 7
  sysinfo:
    enabled: true
    order: 8

# Color customization (Catppuccin Mocha default)
colors:
  primary: "#89dceb"
  secondary: "#cba6f7"
  error: "#f38ba8"
  warning: "#fab387"
  info: "#b4befe"
  success: "#a6e3a1"
  muted: "#6c7086"

# Debug mode for verbose logging
debug: false
```

## Claude Code Integration

### StatusLine Configuration

Add to `~/.claude/settings.json`:

```json
{
  "statusLine": {
    "command": "~/.claude/claude-hud",
    "padding": 0,
    "type": "command"
  }
}
```

The `padding: 0` setting ensures the statusline extends to the edge of the terminal.

### Auto-Detection

The binary automatically detects when running in Claude Code statusline mode:
- If stdin has data (not a TTY), it assumes statusline mode
- Reads JSON context from stdin
- Changes to workspace directory if specified
- Renders once and exits

### Output Format Examples

**Large terminal (120+ cols) - Full layout:**
```
glm-4.7 2h15m | █████████░ 92% (in: 185k, cache: 12k) | ◐ Implement feature X | $0.45
🐹 Go | ~/claude-hud-enhanced | 🌿 main ±3 ⬆2
Read×12 | Edit×8 | Bash×5 | Grep×3 | Ask×2
💻 15% | 🎯 8.2/32GB | 💾 45GB
```

**Medium terminal (80-119 cols):**
```
glm-4.7 2h15m | █████████░ 92% | ◐ Implement feature X
🐹 Go | ~/claude-hud-enhanced | 🌿 main ±3 ⬆2
💻 15% | 🎯 8.2/32GB | 💾 45GB
```

**Small terminal (<80 cols):**
```
glm-4.7 | █████████░ 92% | Read×12 | Edit×8
🐹 Go | ~/claude-hud-enhanced
```

## Important Patterns and Conventions

### 1. Section Registration Pattern

Sections register themselves in `init()` functions:

```go
func init() {
    registry.Register("model", NewModelSection)
}
```

Import the sections package to trigger registration:
```go
import _ "github.com/ll931217/claude-hud-enhanced/internal/sections"
```

### 2. Error Handling

- Always use `errors.MainRecovery()` at top level
- Use `errors.SafeGo()` for goroutines
- Use `errors.Warn()`, `errors.Error()`, `errors.Info()` for logging
- Never crash - always return valid output even on errors

### 3. Concurrency

- All section rendering protected with `sync.RWMutex`
- Goroutines for background operations with proper shutdown
- Safe cancellation using `context.Context`

### 4. Base Section Pattern

Sections embed `BaseSection` for common functionality:

```go
type BaseSection struct {
    config *config.Config
    name   string
}

func (b *BaseSection) Enabled() bool {
    return b.config.IsSectionEnabled(b.name)
}

func (b *BaseSection) Order() int {
    return b.config.GetSectionOrder(b.name)
}
```

### 5. Performance Considerations

- 5-second TTL caching for expensive operations
- Streaming JSONL parsing with <50ms render latency target
- Non-blocking file operations where possible

## Directory Structure

```
├── cmd/claude-hud/          # Main application entry point
│   └── main.go              # Two-mode execution (standalone/statusline)
├── internal/
│   ├── beads/               # Beads issue tracker integration
│   ├── config/              # Configuration management
│   ├── errors/              # Error handling and recovery
│   ├── git/                 # Git status detection
│   ├── registry/            # Section registry factory
│   ├── sections/            # Section implementations
│   │   ├── model.go         # Model name section
│   │   ├── contextbar.go    # Context progress bar section
│   │   ├── duration.go      # Session duration section
│   │   ├── beads.go         # Beads section
│   │   ├── status.go        # Git status section
│   │   ├── workspace.go     # Workspace section
│   │   ├── tools.go         # Tools section
│   │   ├── sysinfo.go       # System info section
│   │   ├── base.go          # Base section with common functionality
│   │   ├── helpers.go       # Shared helper functions
│   │   └── init.go          # Package initialization
│   ├── statusline/          # Statusline orchestration
│   │   └── responsive.go    # Responsive layout engine
│   ├── system/              # System monitoring (CPU, RAM, disk)
│   ├── terminal/            # Terminal size detection
│   ├── theme/               # Color themes
│   ├── transcript/          # Transcript JSONL parsing
│   ├── version/             # Version information
│   └── watcher/             # File watching utilities
├── docs/                    # Additional documentation
├── examples/                # Example applications
└── Makefile                 # Build system
```

## Issue Tracking

This project uses **bd (beads)** for issue tracking.

**Quick reference:**
- `bd prime` - Show workflow context
- `bd ready` - Find unblocked work
- `bd create "Title" --type task --priority 2` - Create issue
- `bd close <id>` - Complete work
- `bd sync` - Sync with git (run at session end)

## When Working With This Codebase

### Key Things to Know

1. **Two Operation Modes**: The binary behaves differently based on stdin input
   - Standalone: Continuous refresh loop when run from terminal
   - Statusline: Single-shot render when stdin has JSON data

2. **Responsive Layout**: Adapts to terminal size with priority-based progressive disclosure
   - Small (<80 cols): Essential sections only (model, context, todos)
   - Medium (80-119 cols): Essential + Important sections
   - Large (120+ cols): Full layout including optional (tools, system info)

3. **Graceful Degradation**: The system continues working even when data sources (transcript, Git, Beads) are unavailable

4. **Configuration First**: All behavior controlled through YAML config - avoid hardcoded values

5. **Section-Based**: Each piece of functionality is a separate section that can be enabled/disabled via config

6. **Priority System**: Sections have Essential/Important/Optional priorities for responsive behavior

7. **Thread Safety**: All shared state is protected with mutexes - essential for concurrent refresh cycles

8. **Auto-Detection**: No need for wrapper scripts - the binary auto-detects Claude Code statusline mode

### Testing Philosophy

- Test files alongside source files (`*_test.go`)
- Integration tests for real scenarios
- Benchmarks for performance-critical paths (`*_bench_test.go`)
- CI-friendly test separation (fast vs full)

### Adding New Sections

1. Create section file in `internal/sections/`
2. Implement `Section` interface (including `Priority()` and `MinWidth()` methods)
3. Set priority via `base.SetPriority()` in constructor (Essential/Important/Optional)
4. Add `init()` function to register the section
5. Add default config in `internal/config/config.go`
6. Import section package in `cmd/claude-hud/main.go`
