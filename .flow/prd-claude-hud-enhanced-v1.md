---
prd:
  version: v1
  feature_name: claude-hud-enhanced
  status: approved
git:
  branch: master
  branch_type: main
  created_at_commit: f05d62cf257e01e661e75d53bd202ba0c767067e
  updated_at_commit: f05d62cf257e01e661e75d53bd202ba0c767067e
worktree:
  is_worktree: false
  name: master
  path: ""
  repo_root: /home/ll931217/GitHub/dotfiles
metadata:
  created_at: 2026-01-06T13:14:46Z
  updated_at: 2026-01-06T13:14:46Z
  created_by: Liang-Shih Lin <liangshihlin@gmail.com>
  filename: prd-claude-hud-enhanced-v1.md
beads:
  related_issues: [dotfiles-66r.1,dotfiles-66r.2,dotfiles-66r.3,dotfiles-66r.4,dotfiles-66r.5,dotfiles-tye.1,dotfiles-tye.2,dotfiles-tye.3,dotfiles-tye.4,dotfiles-tye.5,dotfiles-tye.6,dotfiles-yzr.1,dotfiles-yzr.2,dotfiles-yzr.3,dotfiles-yzr.4,dotfiles-yzr.5,dotfiles-on4.1,dotfiles-on4.2,dotfiles-on4.3,dotfiles-on4.4,dotfiles-on4.5,dotfiles-on4.6,dotfiles-0zn.1,dotfiles-0zn.2,dotfiles-0zn.3,dotfiles-0zn.4,dotfiles-0zn.5,dotfiles-0zn.6,dotfiles-zov.1,dotfiles-zov.2,dotfiles-zov.3,dotfiles-zov.4,dotfiles-zov.5,dotfiles-zov.6]
  related_epics: [dotfiles-66r,dotfiles-tye,dotfiles-yzr,dotfiles-on4,dotfiles-0zn,dotfiles-zov]
code_references:
  - path: ".claude/statusline-enhanced.sh"
    lines: "1-450"
    reason: "Existing bash statusline with Claude Code integration patterns"
  - path: ".claude/settings.json"
    lines: "1-100"
    reason: "Claude Code plugin configuration and statusline setup"
  - path: ".config/worktrunk/config.toml"
    lines: "1-50"
    reason: "Worktrunk hooks and session management patterns"
  - path: ".config/starship/starship.toml"
    lines: "1-100"
    reason: "Starship configuration for color scheme and layout patterns"
priorities:
  enabled: true
  default: P2
  inference_method: ai_inference_with_review
  requirements:
    - id: FR-1
      text: "Parse Claude Code transcript for context, tools, agents, todos"
      priority: P1
      confidence: high
      inferred_from: "Core functionality from original claude-hud"
      user_confirmed: true
    - id: FR-2
      text: "Display current beads issue with status, priority, and completion"
      priority: P1
      confidence: high
      inferred_from: "User explicitly requested as must-have integration"
      user_confirmed: true
    - id: FR-3
      text: "Show worktrunk session name and git worktree status"
      priority: P1
      confidence: high
      inferred_from: "User explicitly requested as must-have integration"
      user_confirmed: true
    - id: FR-4
      text: "Display git branch, status, uncommitted changes, PR state"
      priority: P1
      confidence: high
      inferred_from: "User explicitly requested as must-have integration"
      user_confirmed: true
    - id: FR-5
      text: "Show system resources (CPU, memory, disk) with thresholds"
      priority: P2
      confidence: high
      inferred_from: "User explicitly requested but can be P2 for initial version"
      user_confirmed: true
    - id: FR-6
      text: "Display current working directory path with language detection"
      priority: P2
      confidence: high
      inferred_from: "User explicitly requested in custom input"
      user_confirmed: true
    - id: FR-7
      text: "Read .beads/issues.jsonl directly for issue data"
      priority: P1
      confidence: high
      inferred_from: "User specified direct read over shell commands"
      user_confirmed: true
    - id: FR-8
      text: "Implement file system watcher with polling fallback"
      priority: P2
      confidence: high
      inferred_from: "User specified hybrid approach for refresh"
      user_confirmed: true
    - id: FR-9
      text: "Support MCP server integration when available"
      priority: P3
      confidence: medium
      inferred_from: "Hybrid approach mentioned but MCP is optional enhancement"
      user_confirmed: true
    - id: FR-10
      text: "Configurable section layout and display options"
      priority: P2
      confidence: high
      inferred_from: "User selected configurable UI as primary choice"
      user_confirmed: true
---

# Product Requirements Document: Claude HUD Enhanced

## 1. Introduction/Overview

**Claude HUD Enhanced** is a comprehensive rewrite of [claude-hud](https://github.com/jarrodwatts/claude-hud) in Go, designed to provide a complete, configurable statusline for Claude Code sessions. It extends the original functionality with deep integration into the user's development workflow including beads issue tracking, worktrunk worktree management, git status visualization, and system resource monitoring.

### Goal

Create a production-ready, high-performance Claude Code statusline plugin that provides developers with complete situational awareness—context usage, active tools, running agents, todo progress, current beads issues, worktree status, git state, and system resources—all in a single, configurable display.

## 2. Goals

- Provide real-time visibility into Claude Code session state (context, tools, agents, todos)
- Integrate seamlessly with beads issue tracker for task context
- Display worktrunk worktree and git repository status
- Monitor system resources to prevent performance degradation
- Maintain high performance with <50ms refresh latency
- Support user-configurable layouts and section visibility
- Deploy as a single binary with no runtime dependencies

## 3. User Stories

- **As a developer using Claude Code**, I want to see my current beads issue so I know what task I'm working on
- **As a developer using worktrunk**, I want to see which worktree I'm in and its git status
- **As a developer**, I want to monitor my system resources so I can address issues before they impact my work
- **As a developer**, I want to see which files Claude is reading/editing so I understand its current focus
- **As a developer**, I want to configure which sections appear in my statusline so it fits my workflow
- **As a developer**, I want language detection in my current directory so I can see what tech stack I'm working with

## 4. Functional Requirements

| ID | Requirement | Priority | Notes |
|----|-------------|----------|-------|
| FR-1 | Parse Claude Code transcript for context, tools, agents, todos | P1 | Core functionality from original claude-hud |
| FR-2 | Display current beads issue with status, priority, and completion | P1 | Primary integration for task context |
| FR-3 | Show worktrunk session name and git worktree status | P1 | Worktree awareness for parallel development |
| FR-4 | Display git branch, status, uncommitted changes, PR state | P1 | Git state visibility |
| FR-5 | Show system resources (CPU, memory, disk) with thresholds | P2 | Resource monitoring with alerts |
| FR-6 | Display current working directory path with language detection | P2 | Path and tech stack awareness |
| FR-7 | Read .beads/issues.jsonl directly for issue data | P1 | Direct file access, no shell dependency |
| FR-8 | Implement file system watcher with polling fallback | P2 | Real-time updates with graceful degradation |
| FR-9 | Support MCP server integration when available | P3 | Future enhancement for richer data |
| FR-10 | Configurable section layout and display options | P2 | User customization via config file |
| FR-11 | Color-coded status indicators with catppuccin theme | P2 | Consistent with user's terminal aesthetic |
| FR-12 | Single binary deployment with embedded config defaults | P1 | Zero runtime dependencies |
| FR-13 | Graceful error handling for missing data sources | P1 | Degrade gracefully, don't crash |
| FR-14 | Session duration tracking and token cost estimation | P2 | Usage analytics like original claude-hud |
| FR-15 | Tool activity aggregation with counts | P2 | Show what Claude is doing |

## 5. Non-Goals (Out of Scope)

- Native mobile application (terminal-based only)
- Web-based dashboard (CLI tool only)
- Beads issue editing or creation (read-only display)
- Git operations (commit, push, pull) - display only
- Multi-language support (English only)
- Windows support (Linux/macOS only initially)
- Tmux integration beyond statusline display
- Alternative to Claude Code's native UI (complement, not replace)

## 6. Assumptions

- User has Claude Code v1.0.80+ installed with statusline API
- User has Go 1.21+ for building from source
- User's terminal supports ANSI color codes and Unicode symbols
- Claude Code transcript directory is accessible and readable
- .beads directory follows standard structure with issues.jsonl
- Git is available in system PATH for repository status
- File system watcher (inotify/fsevents) is available on target platform

## 7. Dependencies

### External Dependencies
- **Go 1.21+**: Build and runtime
- **Claude Code v1.0.80+**: Statusline API and transcript format
- **git**: For repository status detection
- **beads**: Issue tracker data structure (read-only)
- **worktrunk**: Worktree session management (read-only)

### Go Libraries
- `github.com/fsnotify/fsnotify`: File system watching
- `github.com/go-git/go-git`: Git operations (optional, can shell out)
- `gopkg.in/yaml.v3`: Configuration file parsing
- `github.com/stretchr/testify`: Testing framework

### Integration Points
- **Claude Code Settings**: `~/.claude/settings.json` for statusline command
- **Transcript Path**: Environment variable or auto-detection
- **Beads Data**: `.beads/issues.jsonl` in repository roots
- **Worktrunk Config**: `~/.config/worktrunk/config.toml` for session detection
- **Git Worktrees**: `git worktree list --porcelain` for worktree enumeration

## 8. Acceptance Criteria

### FR-1: Transcript Parsing
- Parse Claude Code JSONL transcript format correctly
- Extract context window usage with percentage calculation
- Detect active tools with file targets and spinners
- Track running subagents with descriptions and elapsed time
- Monitor todo items with completion status

### FR-2: Beads Integration
- Read `.beads/issues.jsonl` and parse all issues
- Identify current/active issue based on working directory
- Display issue ID, title, status, and priority
- Show completion progress (e.g., "3/5 todos complete")
- Handle missing .beads directory gracefully

### FR-3: Worktrunk Integration
- Detect git worktree via `.git/commondir` check
- Parse `git worktree list --porcelain` output
- Display worktree/session name derived from branch or path
- Show branch name and worktree path
- Distinguish main repo from worktree

### FR-4: Git Status
- Show current branch name with icon
- Display uncommitted changes count (modified, added, deleted)
- Indicate unpushed commits with ahead/behind count
- Detect and show PR state if available (requires API)
- Color-code status (clean, dirty, detached)

### FR-5: System Resources
- Monitor CPU usage percentage
- Track memory usage with absolute and percentage values
- Display disk usage for current partition
- Apply color thresholds (green <70%, yellow 70-90%, red >90%)
- Update every 5 seconds (separate from UI refresh)

### FR-6: Directory & Language
- Display current working directory path
- Truncate long paths intelligently (preserve meaningful parts)
- Detect primary programming language from file extensions
- Show language icon/name (Python 🐍, Rust 🦀, Go 🐹, etc.)
- Update on directory change

### FR-7: Direct Beads Reading
- Open and parse `.beads/issues.jsonl` line by line
- Handle JSON parsing errors gracefully
- Support streaming for large issue files
- Cache parsed issues with TTL (5 seconds)
- Fall back to shell `bd` commands if read fails

### FR-8: Hybrid Refresh
- Use fsnotify for file system watches on transcript and .beads
- Fall back to 300ms polling if watcher fails
- Detect and recover from watch failures
- Provide smooth updates without flickering

### FR-9: MCP Integration
- Detect available MCP servers from Claude Code config
- Query additional context when MCP is available
- Degrade gracefully if MCP is not responding
- Make MCP data optional (don't block on it)

### FR-10: Configurable Layout
- Support YAML configuration file (`~/.config/claude-hud/config.yaml`)
- Allow enabling/disabling each section
- Support section ordering customization
- Provide sensible defaults for missing config
- Support color scheme customization

## 9. Design Considerations

### UI Layout
```
┌─────────────────────────────────────────────────────────────┐
│ [Opus 4.5] ████░░░░░░ 19% | 2 CLAUDE.md | 8 rules | ⏱️ 1m  │
│ ✗ bd-auth: Fix JWT token validation | P1 | 3/5 todos      │
│ 🌿 feature-auth | 📁 src/auth | 🦀 Rust                 │
│ ✓ Read ×2 | ✓ Edit ×1 | 🔄 TaskOutput: src/auth.rs     │
│ 💻 CPU: 45% | RAM: 4.2GB/16GB | 💾 120GB free           │
└─────────────────────────────────────────────────────────────┘
```

### Color Scheme (Catppuccin Mocha)
- Background: `#1E1E2E` (base)
- Primary: `#89dceb` (sky)
- Success: `#a6e3a1` (green)
- Warning: `#fab387` (peach)
- Error: `#f38ba8` (red)
- Info: `#b4befe` (lavender)
- Mauve: `#cba6f7`
- Text: `#cdd6f4`

### Icons
- Git branch: `󰘬`
- Worktree: `🌿`
- Beads issue: `✗` (open), `✓` (closed), `◐` (in progress)
- Resources: `💻` CPU, `🎯` RAM, `💾` Disk
- Language: `🐍` Python, `🦀` Rust, `🐹` Go, `💎` Ruby, `🟨` JS

## 10. Technical Considerations

### Performance
- Target <50ms refresh latency for UI updates
- Use streaming JSON parsing for large files
- Cache parsed data with TTL to avoid repeated reads
- Separate resource monitoring (5s) from UI updates (300ms)
- Use efficient data structures (maps, slices) for lookups

### Error Handling
- Gracefully degrade when data sources are unavailable
- Log errors to stderr without interrupting statusline display
- Retry failed file operations with exponential backoff
- Provide debug mode for troubleshooting

### Memory Management
- Stream JSONL files instead of loading entirely
- Limit cache sizes for parsed issues and git status
- Use object pools for frequently allocated structures
- Profile and optimize memory hot paths

### Concurrency
- Goroutine for file system watching
- Separate goroutine for resource monitoring
- Channel-based communication between components
- Mutex protection for shared state

## 11. Architecture Patterns

### Pattern Checklist

**SOLID Principles:**
- [x] **Single Responsibility:** Each component has one clear purpose (parser, display, watcher)
- [x] **Open/Closed:** Extensible via section plugins without modifying core
- [x] **Liskov Substitution:** Data providers are swappable (fsnotify vs polling)
- [x] **Interface Segregation:** Small interfaces (DataProvider, UISection, Watcher)
- [x] **Dependency Inversion:** Depend on abstractions (interfaces) not concretions

**Creational Patterns:**
- [x] **Factory Pattern:** Create section instances from config (SessionSection, BeadsSection, etc.)
- [x] **Builder Pattern:** Build complex statusline layout from sections
- [ ] **Abstract Factory:** Not needed (single output format)

**Structural Patterns:**
- [x] **Registry Pattern:** Register section types by name for config-driven creation
- [ ] **Adapter Pattern:** Not needed (consistent data interfaces)
- [x] **Decorator Pattern:** Add color formatting to sections

**Inversion of Control / Dependency Injection:**
- [x] **Constructor Injection:** Inject dependencies (config, data providers) into sections
- [ ] **Setter Injection:** Not needed (immutable after creation)
- [ ] **Service Locator:** Not needed (DI is sufficient)
- [ ] **DI Container:** Not needed (simple constructor injection)

### Component Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Application                          │
│  ┌────────────────────────────────────────────────────────┐ │
│  │                  Config Manager                        │ │
│  │  - Load ~/.config/claude-hud/config.yaml              │ │
│  │  - Provide defaults for missing values                │ │
│  └────────────────────────────────────────────────────────┘ │
│                              │                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │              Section Registry (Factory)                │ │
│  │  register("session", SessionSection)                  │ │
│  │  register("beads", BeadsSection)                      │ │
│  │  register("git", GitSection)                          │ │
│  └────────────────────────────────────────────────────────┘ │
│                              │                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │                   Statusline                           │ │
│  │  - Manage section lifecycle                           │ │
│  │  - Orchestrate refresh cycles                         │ │
│  │  - Handle layout and formatting                       │ │
│  └────────────────────────────────────────────────────────┘ │
│         │                │                │                 │
│  ┌──────▼──────┐  ┌──────▼──────┐  ┌──────▼──────┐       │
│  │   Section   │  │   Section   │  │   Section   │       │
│  │  (Session)  │  │   (Beads)   │  │    (Git)    │       │
│  └─────────────┘  └─────────────┘  └─────────────┘       │
│         │                │                │                 │
└─────────┼────────────────┼────────────────┼─────────────────┘
          │                │                │
┌─────────▼────────────────▼────────────────▼─────────────────┐
│                      Data Providers                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Transcript   │  │   Beads      │  │     Git      │      │
│  │   Parser     │  │   Reader     │  │   Detector   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
          │                │                │
┌─────────▼────────────────▼────────────────▼─────────────────┐
│                      Watchers                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   fsnotify   │  │   Polling    │  │   Ticker     │      │
│  │  (transcript)│  │   (fallback) │  │ (resources)  │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

### Key Interfaces

```go
// DataProvider provides data for a section
type DataProvider interface {
    GetData() (interface{}, error)
    Watch(callback func()) (Stopper, error)
}

// Section renders a part of the statusline
type Section interface {
    Render(data interface{}) string
    Enabled() bool
    Order() int
}

// Stopper stops a watcher or goroutine
type Stopper interface {
    Stop() error
}
```

## 12. Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Claude Code API changes break transcript parsing | High | Medium | Version detection, graceful degradation, fallback to original claude-hud |
| File system watcher unavailable on some systems | Medium | Low | Polling fallback always available |
| Large .beads/issues.jsonl causes slow startup | Medium | Low | Streaming parser, lazy loading, cache with TTL |
| Go git library incompatible with worktree setup | Medium | Medium | Shell out to `git` command as fallback |
| Memory leak from unclosed goroutines | High | Low | Proper context cancellation, leak detection in tests |
| Race conditions in concurrent data access | High | Medium | Mutex protection, channel communication, race detector in CI |
| Performance degradation with many files | Medium | Low | Separate resource monitoring, throttled updates |

## 13. Success Metrics

- **Performance**: <50ms refresh latency, <50MB memory usage
- **Reliability**: >99% uptime (graceful degradation on errors)
- **Adoption**: User satisfaction with configurable sections
- **Integration**: Successfully displays all 5 data sources (transcript, beads, worktrunk, git, resources)
- **Compatibility**: Works on macOS and Linux without runtime dependencies
- **Maintainability**: Test coverage >80%, clear architecture documentation

## 14. Priority/Timeline

**Target**: Production-ready in 4-6 weeks

**Sprint 1 (Week 1-2)**: Core infrastructure
- Project setup, config system, section registry
- Transcript parser (FR-1)
- Basic statusline rendering

**Sprint 2 (Week 2-3)**: Primary integrations
- Beads reader (FR-2, FR-7)
- Git status detector (FR-3, FR-4)
- Worktrunk detection

**Sprint 3 (Week 3-4)**: Additional features
- System resource monitoring (FR-5)
- Directory and language detection (FR-6)
- File system watcher (FR-8)

**Sprint 4 (Week 4-5)**: Polish and MCP
- MCP integration (FR-9)
- Configurable layout (FR-10)
- Color scheme and icons

**Sprint 5 (Week 5-6)**: Production hardening
- Error handling (FR-13)
- Performance optimization
- Documentation and tests

## 15. Open Questions

1. Should we support reading beads issues from parent directories if not found in current repo?
2. What is the maximum acceptable memory usage for the plugin?
3. Should we cache git status results or always fetch fresh?
4. Do we need to support Claude Code running in remote/SSH sessions?
5. Should we provide a "compact mode" with single-line output?

## 16. Glossary

- **Beads**: Git-based issue tracker using JSONL storage
- **Worktrunk**: Git worktree management tool with tmux integration
- **Statusline**: A line of text displayed at the bottom of the terminal showing status information
- **Transcript**: Claude Code's JSONL file recording all session activity
- **MCP**: Model Context Protocol, for extending Claude Code capabilities
- **JSONL**: JSON Lines format, one JSON object per line

## 17. Changelog

| Version | Date | Summary of Changes |
|---------|------|-------------------|
| 1 | 2026-01-06 | Initial PRD created and approved |

## 18. Relevant Code References

| File Path | Lines | Purpose |
|-----------|-------|---------|
| `.claude/statusline-enhanced.sh` | 1-450 | Existing bash statusline with Claude Code integration patterns to replicate in Go |
| `.claude/settings.json` | 1-100 | Claude Code plugin configuration and statusline command setup |
| `.config/worktrunk/config.toml` | 1-50 | Worktrunk hooks and session management patterns for worktree detection |
| `.config/starship/starship.toml` | 1-100 | Starship configuration for color scheme (catppuccin_mocha) and layout patterns |
