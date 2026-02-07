# Claude HUD Enhanced

A sophisticated statusline plugin for Claude Code sessions, providing real-time visibility into your development environment.

![Version](https://img.shields.io/badge/version-v0.1.0-blue)
![Go](https://img.shields.io/badge/Go-1.25.5+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)

## Features

- **Claude Code Integration**: Deep integration with Claude Code session transcripts
- **Context Progress Bar**: Visual progress bar with color-coded thresholds (yellow/red at high usage)
- **Token Breakdown**: Shows detailed token usage at high context (≥85%)
- **Responsive Layout**: Adapts to terminal size with priority-based progressive disclosure
- **Configurable Layout**: Each line is fully customizable with sections and separators
- **Tool Usage Tracking**: Display recently used tools with recency sorting
- **Session Info**: Duration, cost calculation, and model information
- **Todo Tracking**: Display todo progress from your session
- **Beads Issue Tracking**: Real-time display of your beads issue tracker status
- **Git Status**: Show branch, dirty state, ahead/behind, and worktree info
- **Language Detection**: Automatically detects project language with icon
- **System Monitoring**: CPU, memory, and disk usage (separate sysinfo section)
- **Auto-Detection**: Works directly with Claude Code without wrapper script
- **Theming**: Beautiful Catppuccin Mocha color scheme
- **Nerd Font Icons**: Icon support with ASCII fallback
- **Performance**: Streaming JSONL parsing with <50ms render latency
- **Cross-Platform**: Builds for Linux and macOS

## Quick Start

### One-Line Install

The easiest way to install Claude HUD Enhanced is with the install script:

```bash
curl -fsSL https://raw.githubusercontent.com/ll931217/claude-hud-enhanced/main/install.sh | sh
```

This will:
1. Detect your platform (Linux/macOS, amd64/arm64)
2. Download the latest binary from GitHub Releases
3. Install it to `~/.claude/claude-hud`
4. Update your Claude Code settings (if `jq` is installed)

### Manual Installation

#### From Source

```bash
git clone https://github.com/ll931217/claude-hud-enhanced.git
cd claude-hud-enhanced
make build
sudo cp bin/claude-hud /usr/local/bin/
```

#### From Release

Download the appropriate binary for your platform from the [releases page](https://github.com/ll931217/claude-hud-enhanced/releases):

```bash
# Linux AMD64
wget https://github.com/ll931217/claude-hud-enhanced/releases/latest/download/claude-hud-linux-amd64
chmod +x claude-hud-linux-amd64
sudo cp claude-hud-linux-amd64 /usr/local/bin/claude-hud

# macOS ARM64
wget https://github.com/ll931217/claude-hud-enhanced/releases/latest/download/claude-hud-darwin-arm64
chmod +x claude-hud-darwin-arm64
sudo cp claude-hud-darwin-arm64 /usr/local/bin/claude-hud
```

### Configuration

Create a configuration file at `~/.config/claude-hud/config.yaml`:

```yaml
# Claude HUD Enhanced Configuration

# Refresh interval in milliseconds (100-5000)
refresh_interval_ms: 500

# Configurable layout system
layout:
  # Enable responsive design (adapts to terminal size)
  responsive:
    enabled: true
    small_breakpoint: 80   # Columns < 80: minimal layout
    medium_breakpoint: 120 # Columns 80-119: balanced layout
    large_breakpoint: 160  # Columns >= 120: full layout

  # Define each line with sections and separators
  lines:
    - sections: [session]
      separator: " | "
    - sections: [workspace, status]
      separator: " | "
    - sections: [tools]
      separator: " | "
    - sections: [sysinfo]
      separator: " | "

# Section configuration
sections:
  session:
    enabled: true
    order: 1
  beads:
    enabled: true
    order: 2
  status:
    enabled: true
    order: 3
  workspace:
    enabled: true
    order: 4
  tools:
    enabled: true
    order: 5
  sysinfo:
    enabled: true
    order: 6

# Color customization (uses Catppuccin Mocha by default)
colors:
  primary: "#89dceb"
  secondary: "#cba6f7"
  error: "#f38ba8"
  warning: "#fab387"
  info: "#b4befe"
  success: "#a6e3a1"
  muted: "#6c7086"

# Debug mode
debug: false
```

### Usage

#### Standalone Mode

Run the statusline in standalone mode (continuous refresh):

```bash
claude-hud
```

Show version information:

```bash
claude-hud --version
claude-hud --build-info
```

#### Claude Code Statusline Mode

Run in single-shot mode for Claude Code integration:

```bash
claude-hud --statusline
```

### Claude Code Integration

Claude HUD Enhanced can be used as a custom statusline for Claude Code, providing real-time visibility into your development session.

#### Installation

Using the Makefile:

```bash
make install-statusline
```

Or manually:

1. Build the binary:
```bash
make build
```

2. Copy the binary to your `~/.claude` directory:
```bash
cp bin/claude-hud ~/.claude/claude-hud
chmod +x ~/.claude/claude-hud
```

#### Configuration

Add or update the `statusLine` section in your `~/.claude/settings.json`:

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

The binary will automatically read the JSON context from Claude Code's stdin and extract the workspace directory and transcript path.

#### Multiline Support

Claude Code's statusline supports multiline output, and Claude HUD Enhanced takes advantage of this by displaying:

1. **Session Info**: Model, context progress, duration, todos
2. **Workspace & Git**: Language, directory, branch status
3. **Tools**: Recently used tools with call counts (max 5)
4. **System Info**: CPU, RAM, disk usage

Each section appears on its own line for maximum visibility. The responsive layout adapts to terminal size:
- **Small (<80 cols)**: Shows essential sections (model, contextbar, duration, tools)
- **Medium (80-119 cols)**: Adds important sections (workspace, status, beads)
- **Large (120+ cols)**: Shows full layout including system info

#### Testing

Test your statusline setup with sample JSON input:

```bash
echo '{"model":{"display_name":"Opus"},"workspace":{"current_dir":"/home/test"}}' | ~/.claude/claude-hud --statusline
```

#### Customization

Configure which sections appear and their order in `~/.config/claude-hud/config.yaml`:

```yaml
sections:
  session:
    enabled: true
    order: 1
  beads:
    enabled: true
    order: 2
  status:
    enabled: true
    order: 3
  workspace:
    enabled: true
    order: 4
```

## Sections

### Session Section

Displays information about your current Claude Code session:
- Model name (e.g., "glm-4.7", "Claude Opus")
- Context progress bar (plain text at low usage, yellow/red at high usage)
- Token breakdown at high usage (≥85%)
- Session duration
- Tool usage activity
- Agent activity
- Todo progress
- Estimated cost

**Context Progress Bar Colors:**
- Plain text (<70%): Normal context usage
- 🟡 Yellow (70-84%): Approaching limit
- 🔴 Red (≥85%): High usage with token breakdown

Example output (4-line full mode):
```
glm-4.7 2h15m | █████████░ 92% (in: 185k, cache: 12k) | ◐ Implement feature X | $0.45
```

### Tools Section

Displays recently used Claude Code tools with call counts (max 5 tools):
- Tool name with usage count
- Sorted by most recently used
- MCP plugin names are shortened for readability

Example output:
```
Read×12 | Edit×8 | Bash×5 | Grep×3 | Ask×2
```

### Beads Section

Shows your beads issue tracker status:
- Total issue count
- In-progress issues
- Closed issues
- Current active task (if any)

Example:
```
☍ 40 | ✓ 40 closed
```

When working on an issue:
```
◐ [beads-123] Implement feature X (P2)
```

### Status Section

Git repository information:
- Current branch
- Dirty state (modified files)
- Ahead/behind remote
- Worktree info
- Stashed changes

Example:
```
🌿 main ↑2↓1 * 3 changes
```

### Workspace Section

Workspace information:
- Detected programming language (with icon)
- Current directory (truncated for fit)

Example:
```
🐹 Go | ~/claude-hud-enhanced
```

### SysInfo Section

System resource usage:
- CPU usage percentage
- Memory usage (used/total)
- Disk available space

Example:
```
💻 15% | 🎯 8.2/32GB | 💾 45GB
```

## Development

### Building

```bash
# Build for current platform
make build

# Build release for current platform
make release

# Build releases for all platforms
make release-all

# Create release archives
make archives
```

### Testing

```bash
# Run all tests
make test

# Run benchmarks
make benchmark
```

### Linting

```bash
make lint
```

## Configuration

### Layout System

The layout system allows you to customize which sections appear on each line:

```yaml
layout:
  responsive:
    enabled: true
    small_breakpoint: 80
    medium_breakpoint: 120
    large_breakpoint: 160
  lines:
    - sections: [session]
      separator: " | "
    - sections: [workspace, status]
      separator: " | "
    - sections: [tools]
      separator: " | "
    - sections: [sysinfo]
      separator: " | "
```

### Section Order

Control the order and visibility of sections:

```yaml
sections:
  session:
    enabled: true
    order: 1
  tools:
    enabled: true
    order: 2
  beads:
    enabled: true
    order: 3
  status:
    enabled: true
    order: 4
  workspace:
    enabled: true
    order: 5
  sysinfo:
    enabled: false  # Disable this section
```

### Colors

Customize the color scheme (defaults to Catppuccin Mocha):

```yaml
colors:
  primary: "#89dceb"    # Sky blue
  secondary: "#cba6f7"  # Mauve
  error: "#f38ba8"      # Red
  warning: "#fab387"    # Peach
  info: "#b4befe"       # Lavender
  success: "#a6e3a1"    # Green
  muted: "#6c7086"      # Gray
```

### Refresh Interval

Control how often the statusline updates (100-5000ms):

```yaml
refresh_interval_ms: 500  # Update every 500ms
```

## Architecture

- **Responsive Layout**: Adapts to terminal size with priority-based progressive disclosure
- **Streaming JSONL Parser**: Efficient transcript parsing with line-by-line processing
- **Factory Pattern**: Section registry for dynamic section creation
- **Graceful Degradation**: Continues working when data sources are unavailable
- **Thread-Safe**: All operations protected with mutexes
- **Context-Based Timeouts**: Safe cancellation of all operations
- **5-Second TTL Caching**: Optimized performance for expensive operations

## Documentation

- [Section Registry Implementation](docs/section-registry-implementation.md)
- [Statusline Implementation](STATUSLINE_IMPLEMENTATION.md)
- [Examples](examples/)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [Claude Code](https://claude.ai/code) - The AI-powered development environment
- [Beads](https://github.com/steveyegge/beads) - Git-based issue tracker
- [Worktrunk](https://worktrunk.dev/) - Git worktree management
- [Catppuccin](https://catppuccin.com/) - Beautiful color scheme
- [Nerd Fonts](https://www.nerdfonts.com/) - Iconic font aggregator
