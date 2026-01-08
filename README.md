# Claude HUD Enhanced

A sophisticated statusline plugin for Claude Code sessions, providing real-time visibility into your development environment.

![Version](https://img.shields.io/badge/version-v0.1.0-blue)
![Go](https://img.shields.io/badge/Go-1.25.5+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)

## Features

- **Claude Code Integration**: Deep integration with Claude Code session transcripts
- **Color-Coded Context Bar**: Visual progress bar with green/yellow/red thresholds
- **Token Breakdown**: Shows detailed token usage at high context (≥85%)
- **Auto-Compact Buffer**: Accounts for 128k token buffer in calculations
- **Beads Issue Tracking**: Real-time display of your beads issue tracker status
- **Worktrunk Support**: Visualize your git worktree management
- **Git Status**: Show branch, dirty state, ahead/behind, and worktree info
- **System Monitoring**: CPU, memory, and disk usage at a glance
- **Todo Tracking**: Display todo progress from your session
- **Session Info**: Duration, cost calculation, and model information
- **Compact Layout**: Optimized 2-line output fits within 80 columns
- **Auto-Detection**: Works directly with Claude Code without wrapper script
- **Theming**: Beautiful Catppuccin Mocha color scheme
- **Nerd Font Icons**: Icon support with ASCII fallback
- **Performance**: Streaming JSONL parsing with <50ms render latency
- **Cross-Platform**: Builds for Linux, macOS, and Windows

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

# Compact mode (2-line layout vs 4-line layout)
compact_mode: true
max_lines: 2

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

1. **Session Info**: Duration, cost, tools, agents, todos
2. **Beads Status**: Open issues, in progress, blocked, current task
3. **Git Status**: Branch, dirty state, ahead/behind, worktree info
4. **Workspace**: CPU, RAM, disk, directory, language

Each section appears on its own line for maximum visibility (or 2 lines in compact mode).

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
- Color-coded **context progress bar** (green/yellow/red based on usage)
- Token breakdown at high usage (≥85%)
- Session duration
- Tool usage activity
- Agent activity
- Todo progress
- Estimated cost

**Context Progress Bar Colors:**
- 🟢 Green (<70%): Healthy context usage
- 🟡 Yellow (70-84%): Approaching limit
- 🔴 Red (≥85%): High usage with token breakdown

Example output (compact 2-line mode):
```
glm-4.7 1h41m | [████░░░░░░]60% | ◐ Implement feature X
☍ 40 total | ✓ 40 closed | 🌿 main * 2 changes ↓25

~/claude-hud-enhanced | 🐹 Go | 💻 1% | 🎯 12.6/23GB | 💾 10GB
```

At high context usage (≥85%), shows token breakdown:
```
glm-4.7 2h15m | [█████████░]92% (in: 185k, cache: 12k) | $0.45
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

System and workspace information:
- Current directory (truncated for fit)
- Detected programming language
- CPU usage percentage
- Memory usage (used/total)
- Disk available space

Example:
```
~/claude-hud-enhanced | 🐹 Go | 💻 1% | 🎯 12.6/23GB | 💾 10GB
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

### Section Order

Control the order and visibility of sections:

```yaml
sections:
  session:
    enabled: true
    order: 1
  beads:
    enabled: true
    order: 2
  status:
    enabled: false  # Disable this section
  workspace:
    enabled: true
    order: 3
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
