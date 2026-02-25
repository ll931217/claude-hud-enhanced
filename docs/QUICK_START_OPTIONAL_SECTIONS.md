# Quick Start: Optional Sections

## TL;DR

To enable optional sections, just add them to your `layout.lines` in `~/.config/claude-hud/config.yaml`:

```yaml
layout:
  responsive:
    enabled: true
    small_breakpoint: 80
    medium_breakpoint: 120
    large_breakpoint: 160
  lines:
    # Add optional sections to any line
    - sections: [model, contextbar, duration, cost]  # Added 'cost' here
      separator: " | "
    - sections: [workspace, status, beads, todoprogress]  # Added 'todoprogress' here
      separator: " | "
    - sections: [tools, agents]  # Added 'agents' here
      separator: " | "
    - sections: [testcoverage, buildstatus, errors, sysinfo]  # Added 3 new ones here
      separator: " | "
```

That's it! No `sections.*.enabled` configuration needed.

## Available Optional Sections

| Section ID       | What it shows                                |
|------------------|----------------------------------------------|
| `agents`         | Running and completed Claude Code subagents |
| `cost`           | Session API costs with hourly rate          |
| `todoprogress`   | TodoWrite task progress (3/8 tasks)         |
| `errors`         | Error count from tool executions            |
| `testcoverage`   | Test coverage % (Go/JS/Python)              |
| `buildstatus`    | Build/type-check status                     |

## How It Works

1. Sections appear **only if** they're listed in `layout.lines`
2. Order is determined by **position in the layout**, not a separate config
3. Sections auto-hide when they have no data to display
4. Responsive layout hides low-priority sections on small terminals

## Examples

### Minimal: Just cost and agents

```yaml
layout:
  lines:
    - sections: [model, contextbar, duration, cost]
      separator: " | "
    - sections: [workspace, status, beads]
      separator: " | "
    - sections: [tools, agents]
      separator: " | "
```

### Developer-focused: Coverage and build status

```yaml
layout:
  lines:
    - sections: [model, contextbar, duration]
      separator: " | "
    - sections: [workspace, status, beads, todoprogress]
      separator: " | "
    - sections: [testcoverage, buildstatus, errors]
      separator: " | "
```

### Everything

```yaml
layout:
  lines:
    - sections: [model, contextbar, duration, cost]
      separator: " | "
    - sections: [workspace, status, beads, todoprogress]
      separator: " | "
    - sections: [tools, agents]
      separator: " | "
    - sections: [testcoverage, buildstatus, errors, sysinfo]
      separator: " | "
```

## Performance Notes

⚠️ **Test Coverage** and **Build Status** sections run commands that may be slow:
- Test coverage: Runs your test suite (cached for 30s)
- Build status: Runs build/type-check (cached for 30s)

If you experience lag, remove these sections or increase cache duration in the code.

## See Also

- [Full documentation](OPTIONAL_SECTIONS.md) - Detailed descriptions and troubleshooting
- [Implementation details](IMPLEMENTATION_SUMMARY.md) - Technical architecture
