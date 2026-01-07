# Configuration Guide

This guide covers all configuration options for Claude HUD Enhanced.

## Configuration File Location

Claude HUD Enhanced looks for configuration in the following locations (in order):

1. `~/.config/claude-hud/config.yaml`
2. `./config.yaml` (current working directory)
3. Built-in defaults

If no configuration file is found, sensible defaults are used.

## Quick Start Configuration

Create a minimal configuration:

```yaml
# ~/.config/claude-hud/config.yaml

# Refresh interval in milliseconds (100-5000)
refresh_interval_ms: 500

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
```

## Configuration Options

### Global Settings

#### `refresh_interval_ms`

Controls how often the statusline updates.

- **Type**: Integer
- **Range**: 100-5000 (milliseconds)
- **Default**: 500

```yaml
refresh_interval_ms: 500  # Update every 500ms
```

**Recommended values:**
- 100-250ms: Very responsive (higher CPU usage)
- 500ms: Balanced (recommended)
- 1000-2000ms: Lower CPU usage
- 3000-5000ms: Minimal CPU usage

#### `debug`

Enable debug logging.

- **Type**: Boolean
- **Default**: false

```yaml
debug: true  # Enable debug logging
```

### Section Configuration

Each section can be individually enabled or disabled and ordered.

#### Structure

```yaml
sections:
  <section_name>:
    enabled: boolean  # Enable/disable section
    order: integer    # Display order (lower numbers first)
```

#### Available Sections

##### Session Section

Displays Claude Code session information.

```yaml
sections:
  session:
    enabled: true
    order: 1
```

**Shows:**
- Session duration
- Estimated token cost
- Active tool count
- Agent count
- Todo progress

##### Beads Section

Displays beads issue tracker status.

```yaml
sections:
  beads:
    enabled: true
    order: 2
```

**Shows:**
- Total open issues
- Issues in progress
- Blocked issues
- Current task

##### Status Section

Displays git repository information.

```yaml
sections:
  status:
    enabled: true
    order: 3
```

**Shows:**
- Current branch
- Dirty state (modified files)
- Ahead/behind remote
- Worktree info
- Stashed changes

##### Workspace Section

Displays system and workspace information.

```yaml
sections:
  workspace:
    enabled: true
    order: 4
```

**Shows:**
- CPU usage percentage
- Memory usage
- Disk usage
- Current directory (truncated)
- Detected programming language

### Color Configuration

Customize the color scheme. Uses Catppuccin Mocha by default.

#### Structure

```yaml
colors:
  primary: "#hexcolor"
  secondary: "#hexcolor"
  error: "#hexcolor"
  warning: "#hexcolor"
  info: "#hexcolor"
  success: "#hexcolor"
  muted: "#hexcolor"
```

#### Color Options

##### `primary`

Primary accent color.

- **Type**: Hex color string
- **Default**: "#89dceb" (Sky blue)

```yaml
colors:
  primary: "#89dceb"
```

##### `secondary`

Secondary accent color.

- **Type**: Hex color string
- **Default**: "#cba6f7" (Mauve)

```yaml
colors:
  secondary: "#cba6f7"
```

##### `error`

Error color.

- **Type**: Hex color string
- **Default**: "#f38ba8" (Red)

```yaml
colors:
  error: "#f38ba8"
```

##### `warning`

Warning color.

- **Type**: Hex color string
- **Default**: "#fab387" (Peach)

```yaml
colors:
  warning: "#fab387"
```

##### `info`

Info color.

- **Type**: Hex color string
- **Default**: "#b4befe" (Lavender)

```yaml
colors:
  info: "#b4befe"
```

##### `success`

Success color.

- **Type**: Hex color string
- **Default**: "#a6e3a1" (Green)

```yaml
colors:
  success: "#a6e3a1"
```

##### `muted`

Muted/disabled text color.

- **Type**: Hex color string
- **Default**: "#6c7086" (Gray)

```yaml
colors:
  muted: "#6c7086"
```

## Example Configurations

### Minimal Configuration

```yaml
refresh_interval_ms: 500

sections:
  session:
    enabled: true
    order: 1
  beads:
    enabled: true
    order: 2
```

### Development Configuration

Faster refresh rate for active development:

```yaml
refresh_interval_ms: 250  # Update 4 times per second
debug: true

sections:
  session:
    enabled: true
    order: 1
  status:
    enabled: true
    order: 2
  workspace:
    enabled: true
    order: 3

colors:
  primary: "#89dceb"
  error: "#f38ba8"
  warning: "#fab387"
```

### Minimal Configuration

Only show essential information:

```yaml
refresh_interval_ms: 1000  # Update every second

sections:
  status:
    enabled: true
    order: 1
  beads:
    enabled: true
    order: 2
```

### Custom Color Scheme

Custom colors (Dracula theme example):

```yaml
colors:
  primary: "#BD93F9"    # Purple
  secondary: "#FF79C6"  # Pink
  error: "#FF5555"      # Red
  warning: "#F1FA8C"    # Yellow
  info: "#8BE9FD"       # Cyan
  success: "#50FA7B"    # Green
  muted: "#6272A4"      # Comment gray
```

### High Contrast Configuration

For better visibility:

```yaml
colors:
  primary: "#FFFFFF"    # White
  secondary: "#FFFF00"  # Yellow
  error: "#FF0000"      # Bright red
  warning: "#FFA500"    # Orange
  info: "#00FFFF"       # Cyan
  success: "#00FF00"    # Bright green
  muted: "#808080"      # Gray
```

## Configuration Validation

Claude HUD Enhanced validates configuration on startup:

- Invalid YAML will show an error and use defaults
- Out-of-range values are clamped to valid ranges
- Missing colors use Catppuccin Mocha defaults
- Invalid section names are ignored

## Reloading Configuration

To reload configuration after making changes:

1. Stop the running instance (Ctrl+C)
2. Edit configuration file
3. Start again:

```bash
claude-hud
```

## Environment Variables

Currently, environment variables are not supported. All configuration must be done via the YAML file.

## Troubleshooting

### Configuration Not Loading

If your configuration isn't being applied:

1. Check file location:
   ```bash
   ls -la ~/.config/claude-hud/config.yaml
   ```

2. Enable debug mode to see loading errors:
   ```yaml
   debug: true
   ```

3. Validate YAML syntax:
   ```bash
   # Install yamllint
   pip install yamllint

   # Check syntax
   yamllint ~/.config/claude-hud/config.yaml
   ```

### Invalid Values

If you see warnings about invalid values:

- Check that numeric values are within valid ranges
- Ensure hex colors are valid (6-digit hex codes)
- Verify section names are correct (session, beads, status, workspace)

## Default Configuration

The built-in default configuration is equivalent to:

```yaml
refresh_interval_ms: 500
debug: false

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

colors:
  primary: "#89dceb"
  secondary: "#cba6f7"
  error: "#f38ba8"
  warning: "#fab387"
  info: "#b4befe"
  success: "#a6e3a1"
  muted: "#6c7086"
```

## Next Steps

- [Usage Guide](USAGE.md) - How to use Claude HUD Enhanced
- [Examples](../examples/) - Example configurations
