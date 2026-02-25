# Optional Sections

This document describes additional sections available in Claude HUD Enhanced that are **not enabled by default**. These sections provide extra functionality and insights that you can enable based on your workflow needs.

## Table of Contents

- [Agent Activity](#agent-activity)
- [Cost Tracker](#cost-tracker)
- [Todo Progress](#todo-progress)
- [Recent Errors](#recent-errors)
- [Test Coverage](#test-coverage)
- [Build Status](#build-status)
- [How to Enable Sections](#how-to-enable-sections)

---

## Agent Activity

**Section ID:** `agents`
**Priority:** Essential
**Minimum Width:** 30 columns

Displays currently running and recently completed Claude Code subagents.

### What it shows:

- **Running agents** (max 2) with ◐ spinner indicator
- **Completed agents** (max 3) with ✓ checkmark
- Agent count indicator if there are more than 5 agents

### Example output:

```
◐ Plan | ◐ Review | ✓ TDD | ✓ Sec | ✓ Build
```

### Shortened agent names:

| Full Name              | Short Name |
|------------------------|------------|
| planner                | Plan       |
| code-reviewer          | Review     |
| architect              | Arch       |
| tdd-guide              | TDD        |
| security-reviewer      | Sec        |
| build-error-resolver   | Build      |
| e2e-runner             | E2E        |
| refactor-cleaner       | Refactor   |
| doc-updater            | Docs       |
| debugger               | Debug      |
| general-purpose        | GP         |
| Explore                | Explore    |

---

## Cost Tracker

**Section ID:** `cost`
**Priority:** Important
**Minimum Width:** 10 columns

Displays accumulated API costs for the current Claude Code session.

### What it shows:

- **Session cost** in USD (formatted based on magnitude)
- **Cost rate per hour** (after 6 minutes of session time)

### Example output:

```
💰 $0.452 ($1.23/h)
```

### Pricing basis:

The cost tracker uses estimated pricing for Claude models:

| Model         | Input (per 1M tokens) | Output (per 1M tokens) |
|---------------|-----------------------|------------------------|
| Claude Opus   | $15.00                | $75.00                 |
| Claude Sonnet | $3.00                 | $15.00                 |
| Claude Haiku  | $0.25                 | $1.25                  |

**Note:** Prices are approximate and may not reflect your actual billing.

---

## Todo Progress

**Section ID:** `todoprogress`
**Priority:** Essential
**Minimum Width:** 20 columns

Displays progress for TodoWrite-based task lists in Claude Code.

### What it shows:

- **Task completion fraction** (e.g., 3/8 completed)
- **Current in-progress task** with ◐ indicator

### Example output:

```
📋 3/8 | ◐ Implementing Agent Activity section
```

**Note:** This section tracks tasks created via the `TodoWrite` tool, not Beads issues.

---

## Recent Errors

**Section ID:** `errors`
**Priority:** Important
**Minimum Width:** 15 columns

Displays error count from tool executions and transcript events.

### What it shows:

- **Total error count** with ⚠️ icon
- **Recent errors** (last 5 minutes)
- **Last error tool name**

### Example output:

```
⚠️ 3 (2 recent) [Bash]
```

### Error sources:

- Tool execution failures (tool_result with is_error: true)
- Explicit error events in transcript
- Build failures and test failures

---

## Test Coverage

**Section ID:** `testcoverage`
**Priority:** Important
**Minimum Width:** 15 columns

Displays test coverage percentage for the current project.

### What it shows:

- **Coverage percentage** with color-coded icon:
  - ✓ (green): >= 80% coverage
  - 🧪 (yellow): 60-79% coverage
  - ⚠️ (red): < 60% coverage

### Example output:

```
✓ 85.2%
```

### Supported languages:

| Language           | Command Used                        |
|--------------------|-------------------------------------|
| Go                 | `go test -cover ./...`              |
| JavaScript/TypeScript | `npm run test:coverage` or `jest --coverage` |
| Python             | `pytest --cov`                      |

### Cache behavior:

Coverage is cached for 30 seconds to avoid running expensive test suites on every render.

**Performance note:** Coverage commands may add latency to statusline rendering. Consider disabling this section if your test suite is slow.

---

## Build Status

**Section ID:** `buildstatus`
**Priority:** Important
**Minimum Width:** 12 columns

Displays build/type-check status for the current project.

### What it shows:

- **Build success:** 🔨 ✓ Build
- **Build failure:** 🔨 ✗ N errors

### Example output:

```
🔨 ✓ Build
```

or

```
🔨 ✗ 5 errors
```

### Supported languages:

| Language           | Command Used              |
|--------------------|---------------------------|
| Go                 | `go build ./...`          |
| JavaScript/TypeScript | `npx tsc --noEmit`     |
| Python             | `mypy .`                  |

### Cache behavior:

Build status is cached for 30 seconds to avoid running expensive build commands on every render.

**Performance note:** Build checks may add latency to statusline rendering. Consider disabling this section if your builds are slow.

---

## How to Enable Sections

To enable any of these optional sections, edit your `~/.config/claude-hud/config.yaml` file.

### Add sections to your layout

Simply add the section names to the `layout.lines` configuration where you want them to appear:

```yaml
layout:
  responsive:
    enabled: true
    small_breakpoint: 80
    medium_breakpoint: 120
    large_breakpoint: 160
  lines:
    # Line 1: Core status (model, context, duration)
    - sections: [model, contextbar, duration]
      separator: " | "

    # Line 2: Workspace and progress
    - sections: [workspace, status, beads, todoprogress]
      separator: " | "

    # Line 3: Tools and agents
    - sections: [tools, agents]
      separator: " | "

    # Line 4: Quality metrics
    - sections: [testcoverage, buildstatus, errors, cost, sysinfo]
      separator: " | "
```

That's it! Sections only appear if they're listed in the layout. If a section has no data to display, it will be hidden automatically.

### Restart Claude Code

After updating your configuration, restart your Claude Code session to see the changes.

---

## Example Full Layout

Here's a comprehensive 4-line layout with all optional sections:

```yaml
layout:
  responsive:
    enabled: true
    small_breakpoint: 80
    medium_breakpoint: 120
    large_breakpoint: 160
  lines:
    # Line 1: Model and Context
    - sections: [model, contextbar, duration, cost]
      separator: " | "

    # Line 2: Workspace and Git
    - sections: [workspace, status, beads]
      separator: " | "

    # Line 3: Activity (Tools and Agents)
    - sections: [tools, agents, todoprogress]
      separator: " | "

    # Line 4: Quality Metrics
    - sections: [testcoverage, buildstatus, errors, sysinfo]
      separator: " | "
```

### Expected Output (Large Terminal, 160+ cols):

```
glm-4.7 | █████████░ 92% (in: 185k, cache: 12k) | 2h15m | 💰 $0.452 ($1.23/h)
🐹 Go | ~/claude-hud-enhanced | 🌿 main ±3 ⬆2 | beads-123: Implement feature X
✓ Read×12 | ✓ Edit×8 | ◐ Plan | ◐ Review | 📋 3/8 | ◐ Writing tests
✓ 85.2% | 🔨 ✓ Build | ⚠️ 2 (1 recent) [Bash] | 💻 15% | 🎯 8.2/32GB | 💾 45GB
```

---

## Performance Considerations

Some optional sections may add latency to statusline rendering:

- **testcoverage**: Runs test suite (cached for 30s)
- **buildstatus**: Runs build command (cached for 30s)

If you experience slow statusline updates, consider:

1. Disabling these sections
2. Increasing cache duration (requires code modification)
3. Using faster build/test configurations

---

## Customization Tips

### Show cost only on large terminals

Use the responsive layout system to hide cost on small screens:

```yaml
layout:
  lines:
    # Show cost only when it fits
    - sections: [model, contextbar, duration, cost]
      separator: " | "
```

Then set the priority and min width appropriately in code (already done by default).

### Combine sections on one line

Group related sections together:

```yaml
layout:
  lines:
    - sections: [agents, todoprogress, errors]
      separator: " | "
```

This creates a "progress and issues" line.

---

## Troubleshooting

### Section not showing up

1. **Check enabled:** Ensure `enabled: true` in `sections:` config
2. **Check layout:** Verify section is listed in one of the `lines:`
3. **Check width:** Some sections hide on narrow terminals (check `MinWidth`)
4. **Check data:** Sections hide when they have no data to display

### Coverage/Build section always empty

1. **Check language:** Section may not support your language
2. **Check command:** Ensure the build/test command exists
3. **Check timeout:** Build might be taking too long (default 10s timeout)
4. **Check cache:** Clear cache by restarting Claude Code

### Performance issues

1. **Disable expensive sections:** testcoverage, buildstatus
2. **Increase cache duration:** Edit section code to cache longer
3. **Check terminal size:** Responsive layout hides sections on small terminals

---

## Contributing

Have ideas for new optional sections? See [CLAUDE.md](../CLAUDE.md) for development guidelines and section implementation patterns.
