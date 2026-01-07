# Section Registry Factory Pattern Implementation

## Overview
This document describes the implementation of the Section Registry Factory Pattern for Claude HUD Enhanced.

## Architecture

### Section Interface (`internal/registry/section.go`)
The `Section` interface defines the contract for all HUD sections:
```go
type Section interface {
    Render() string      // Returns the rendered content
    Enabled() bool       // Returns whether the section is enabled
    Order() int          // Returns the display order
    Name() string        // Returns the section name
}
```

### Section Factory Type (`internal/registry/registry.go`)
The `SectionFactory` is a function type for creating sections:
```go
type SectionFactory func(config interface{}) (Section, error)
```

### SectionRegistry (`internal/registry/registry.go`)
The `SectionRegistry` manages section registration and creation:
- `Register(name string, factory SectionFactory)` - Register a section type
- `Create(name string, config interface{}) (Section, error)` - Create a section instance
- `List() []string` - List all registered sections

The registry is thread-safe using `sync.RWMutex`.

## Built-in Sections

Four section types are implemented as stubs in `internal/sections/stubs.go`:

1. **session** - Displays session/worktrunk information
2. **beads** - Displays beads issue tracking information
3. **status** - Displays git status information
4. **workspace** - Displays workspace information (resources/directory)

All sections are automatically registered in `internal/sections/init.go` via the `init()` function.

## Configuration Integration

Sections integrate with the existing `internal/config` package:
- `config.IsSectionEnabled(name)` - Check if a section is enabled
- `config.GetSectionOrder(name)` - Get the display order for a section
- Sections use `BaseSection` to get configuration from the global config

## Usage Example

```go
import (
    "github.com/ll931217/claude-hud-enhanced/internal/config"
    "github.com/ll931217/claude-hud-enhanced/internal/registry"
    _ "github.com/ll931217/claude-hud-enhanced/internal/sections" // Register built-in sections
)

// Create a section with config
cfg := config.DefaultConfig()
section, err := registry.Create("session", cfg)
if err != nil {
    log.Fatal(err)
}

// Use the section
if section.Enabled() {
    fmt.Println(section.Render())
}
```

## Extensibility

Custom sections can be registered at runtime:

```go
customFactory := func(cfg interface{}) (registry.Section, error) {
    return &MyCustomSection{}, nil
}
registry.Register("custom", customFactory)
```

## Testing

Comprehensive tests are in `internal/sections/sections_test.go`:
- Test section registration
- Test section creation with default config
- Test section creation with custom config (enable/disable)
- Test error handling for unregistered sections
- Test custom section registration

All tests pass:
```bash
go test ./internal/sections/... -v
```

## Acceptance Criteria

All acceptance criteria from the task have been met:

- [x] Section interface defined (Render, Enabled, Order, Name)
- [x] Register(sectionName, sectionFunc) works
- [x] Create(sectionName, config) returns Section
- [x] All 4 section types registered (session, beads, status, workspace)
- [x] Sections can be enabled/disabled via config
- [x] Sections ordered by config.Order()
- [x] Factory pattern properly implemented

Note: The original task specified 5 section types (session, beads, git, resources, directory),
but the actual config package defines 4 sections (session, beads, status, workspace).
The implementation matches the existing config structure.

## Files Created/Modified

### Created:
- `/home/ll931217/Projects/claude-hud-enhanced/internal/registry/section.go`
- `/home/ll931217/Projects/claude-hud-enhanced/internal/registry/registry.go`
- `/home/ll931217/Projects/claude-hud-enhanced/internal/sections/stubs.go`
- `/home/ll931217/Projects/claude-hud-enhanced/internal/sections/init.go`
- `/home/ll931217/Projects/claude-hud-enhanced/internal/sections/sections_test.go`
- `/home/ll931217/Projects/claude-hud-enhanced/examples/registry_demo.go`

### Modified:
- `/home/ll931217/Projects/claude-hud-enhanced/internal/sections/section.go` - Updated to use config methods

## Design Patterns Used

1. **Factory Pattern** - SectionFactory functions create Section instances
2. **Registry Pattern** - SectionRegistry manages type lookup and creation
3. **Interface Segregation** - Small, focused Section interface
4. **Dependency Injection** - Config passed to factory functions

## Thread Safety

The SectionRegistry uses `sync.RWMutex` to ensure thread-safe operations:
- Read operations (List, Create) use RLock
- Write operations (Register) use Lock
