# Usage Guide

This guide covers how to use Claude HUD Enhanced.

## Basic Usage

### Starting the Statusline

Once installed and configured, simply run:

```bash
claude-hud
```

The statusline will begin updating based on your configured refresh interval.

### Command-Line Options

#### Show Version

```bash
claude-hud --version
```

Output:
```
claude-hud-enhanced version v0.1.0
```

#### Show Build Information

```bash
claude-hud --build-info
```

Output:
```
Claude HUD Enhanced Build Information
===================================
Version:   v0.1.0
Commit:    abc1234
Built At:  2024-01-15_10:30:00
Go Version: go1.25.5
```

## Output Interpretation

The statusline displays information in sections from left to right. Each section shows specific information about your development environment.

### Session Section

Shows information about your current Claude Code session.

**Example Output:**
```
Session: 2h15m | $0.45 | 3 tools | 1 agent | ✓ 5/7 todos
```

**Breakdown:**
- `2h15m` - Session duration
- `$0.45` - Estimated token cost
- `3 tools` - Number of active tools
- `1 agent` - Number of active agents
- `✓ 5/7 todos` - Todo progress (5 of 7 completed)

**When current todo is in progress:**
```
◐ Implement feature X
```

**When all todos complete:**
```
✓ All todos complete (5/5)
```

### Beads Section

Shows your beads issue tracker status.

**Example Output:**
```
Beads: 12 open | 3 in progress | 1 blocked | [dotfiles-abc.1] Implement feature X
```

**Breakdown:**
- `12 open` - Total open issues
- `3 in progress` - Issues currently being worked on
- `1 blocked` - Issues blocked by dependencies
- `[dotfiles-abc.1] Implement feature X` - Current task (if available)

**Issue Status Icons:**
- `✗` - Open issue
- `✓` - Closed issue
- `◐` - In progress
- `◌` - Blocked

### Status Section

Shows git repository information.

**Example Output:**
```
🌿 main ↑2↓1 * 3 changes
```

**Breakdown:**
- `🌿` - Branch icon
- `main` - Current branch name
- `↑2` - 2 commits ahead of remote
- `↓1` - 1 commit behind remote
- `* 3 changes` - 3 modified files

**Other States:**

Clean working directory:
```
🌿 main
```

Worktree:
```
🌿 main [feature-branch]
```

Stashed changes:
```
🌿 main + 2 stashed
```

### Workspace Section

Shows system and workspace information.

**Example Output:**
```
💻 45% CPU | 🎯 62% RAM (4.2/16GB) | 💾 78% Disk | ~/Projects/claude-hud-enhanced | Go
```

**Breakdown:**
- `💻 45% CPU` - CPU usage percentage
- `🎯 62% RAM (4.2/16GB)` - Memory usage
- `💾 78% Disk` - Disk usage
- `~/Projects/claude-hud-enhanced` - Current directory (truncated)
- `Go` - Detected programming language

**Color Coding:**
- **Green** - Usage < 70%
- **Yellow/Warning** - Usage 70-89%
- **Red/Critical** - Usage ≥ 90%

## Common Workflows

### Development Session

1. Start your Claude Code session
2. In a separate terminal, run:
   ```bash
   claude-hud
   ```

The statusline will show:
- Your current session duration
- Token cost as you work
- Active tools and agents
- Todo progress

### Monitoring System Resources

Keep the statusline running to monitor:
- CPU usage during compilation
- Memory usage during development
- Disk space availability

### Git Status Monitoring

See at a glance:
- Current branch
- Working directory state
- Sync status with remote
- Active worktree

### Issue Tracking Integration

Track your progress:
- See total open issues
- View in-progress tasks
- Check for blocked issues
- Display current task

## Customizing Display

### Changing Section Order

Edit `~/.config/claude-hud/config.yaml`:

```yaml
sections:
  status:
    enabled: true
    order: 1  # Show first
  beads:
    enabled: true
    order: 2
  session:
    enabled: true
    order: 3
  workspace:
    enabled: true
    order: 4
```

### Disabling Sections

Don't want to see certain information?

```yaml
sections:
  session:
    enabled: false  # Hide session info
  beads:
    enabled: true
  status:
    enabled: true
  workspace:
    enabled: false  # Hide workspace info
```

### Adjusting Refresh Rate

Need more frequent updates?

```yaml
refresh_interval_ms: 250  # Update 4 times per second
```

Or save CPU:

```yaml
refresh_interval_ms: 2000  # Update every 2 seconds
```

## Integration with Shell

### Bash

Add to your `~/.bashrc`:

```bash
# Start claude-hud automatically in new terminal
alias ch='claude-hud'
```

### Zsh

Add to your `~/.zshrc`:

```bash
# Start claude-hud automatically in new terminal
alias ch='claude-hud'
```

### Fish

Add to your `~/.config/fish/config.fish`:

```fish
# Start claude-hud automatically in new terminal
alias ch='claude-hud'
```

## Performance Considerations

### CPU Usage

The statusline is designed to be lightweight:
- Streaming JSONL parsing (line-by-line)
- 5-second TTL caching for expensive operations
- Efficient mutex-based thread safety

**Typical CPU usage:**
- Idle: < 1%
- Active: 1-3%
- With 250ms refresh: 3-5%

### Memory Usage

Memory usage is minimal:
- Base: ~20-30MB
- Per section: ~2-5MB
- Typical total: < 50MB

### Optimization Tips

1. **Increase refresh interval** if CPU usage is a concern:
   ```yaml
   refresh_interval_ms: 1000
   ```

2. **Disable unused sections**:
   ```yaml
   sections:
     workspace:
       enabled: false  # If you don't need system info
   ```

3. **Use custom colors** to avoid theme lookup overhead (minimal impact)

## Troubleshooting

### Statusline Shows Wrong Information

**Problem:** Information is outdated or incorrect.

**Solution:**
- The statusline refreshes based on `refresh_interval_ms`
- Force a refresh by restarting:
  ```bash
  # Ctrl+C to stop, then
  claude-hud
  ```

### High CPU Usage

**Problem:** Statusline using too much CPU.

**Solutions:**
1. Increase refresh interval:
   ```yaml
   refresh_interval_ms: 2000
   ```

2. Disable resource-intensive sections:
   ```yaml
   sections:
     workspace:
       enabled: false
   ```

3. Check debug mode is off:
   ```yaml
   debug: false
   ```

### Missing Information

**Problem:** Some sections show nothing or errors.

**Possible causes:**
1. Not in a git repository (status section)
2. No beads issues.jsonl file (beads section)
3. No active Claude Code session (session section)
4. No language detected (workspace section)

**Solution:** This is expected behavior. The statusline gracefully degrades when information is unavailable.

### Terminal Support

**Problem:** Icons not displaying correctly.

**Solution:**
- Install Nerd Fonts: https://www.nerdfonts.com/
- Configure your terminal to use a Nerd Font
- Some terminals may not support all Unicode characters

## Tips and Tricks

### Quick Session Check

Start the statusline in a separate terminal window to monitor your session while working.

### Minimal Display

Disable all sections except what you need:

```yaml
sections:
  status:
    enabled: true
    order: 1
  beads:
    enabled: true
    order: 2
```

### Development vs Production

Use different configs for different environments:

**Development (`~/.config/claude-hud/config-dev.yaml`):**
```yaml
refresh_interval_ms: 250
debug: true
sections:
  session:
    enabled: true
  workspace:
    enabled: true
```

**Production (`~/.config/claude-hud/config.yaml`):**
```yaml
refresh_interval_ms: 1000
debug: false
sections:
  status:
    enabled: true
  beads:
    enabled: true
```

Switch between them:
```bash
claude-hud # Uses default config
cp ~/.config/claude-hud/config-dev.yaml ~/.config/claude-hud/config.yaml
claude-hud # Uses dev config
```

## Next Steps

- [Installation Guide](INSTALLATION.md) - How to install
- [Configuration Guide](CONFIGURATION.md) - How to configure
- [Examples](../examples/) - Example configurations
