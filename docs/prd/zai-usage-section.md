# PRD: Z.ai Usage Section

**Status:** Completed
**Created:** 2026-03-02
**Priority:** Medium

## Overview

Add a new section to claude-hud-enhanced that displays Z.ai coding plan usage metrics. This section will monitor token quotas for session, weekly, and search usage by calling the Z.ai API.

## Requirements

### Functional Requirements

1. **API Integration**
   - Fetch usage data from `https://api.z.ai/api/monitor/usage/quota/limit`
   - Read API key from environment variable: `GLM_API_KEY` (primary) or `ZAI_API_KEY` (fallback)
   - Handle API errors gracefully (no key, network errors, invalid responses)

2. **Display Three Metrics**
   - **Session** (5-hour rolling window): TOKENS_LIMIT with unit=3, number=5
   - **Weekly** (5 sessions total): TOKENS_LIMIT with unit=6
   - **Search** (Monthly quota): TIME_LIMIT with unit=5, number=1

3. **Output Format**
   - Compact display with icons: `🔋 72% | 📊 45% | 🔍 30%`
   - Color coding based on percentage thresholds:
     - Green (< 70%)
     - Yellow (70-90%)
     - Red (> 90%)

4. **Responsive Behavior**
   - Priority: Important (show on medium+ terminals, 80+ cols)
   - Hide on small terminals

### Non-Functional Requirements

1. **Performance**
   - Cache API responses with 60-second TTL
   - Non-blocking API calls (don't block statusline rendering)
   - 10-second HTTP timeout

2. **Reliability**
   - Graceful degradation: show cached data or empty string on errors
   - Never crash the statusline due to API failures
   - Thread-safe data access

3. **Security**
   - API key read from environment variable only
   - No logging of API keys

## Technical Design

### File Structure

```
internal/
├── zai/                    # New package for Z.ai API client
│   ├── client.go          # HTTP client for API calls
│   ├── client_test.go     # Unit tests
│   └── types.go           # API response types
└── sections/
    └── zaiusage.go        # New section implementation
```

### API Response Structure

```json
{
  "success": true,
  "data": {
    "level": "pro",
    "limits": [
      {
        "type": "TOKENS_LIMIT",
        "unit": 3,
        "number": 5,
        "percentage": 72,
        "nextResetTime": 1709400000000
      }
    ]
  }
}
```

### Section Implementation

```go
type ZaiUsageSection struct {
    *BaseSection
    client *zai.Client
    cache  *zai.UsageCache
}
```

### Configuration

No additional config needed - uses environment variables for API key.

## Implementation Tasks

1. Create `internal/zai/` package with HTTP client
2. Implement API response parsing
3. Add caching with 60-second TTL
4. Create `internal/sections/zaiusage.go`
5. Register section in registry
6. Add to default layout configuration
7. Write unit tests
8. Write integration tests

## Acceptance Criteria

- [ ] Section displays all three usage metrics with icons
- [ ] Color coding works correctly (green/yellow/red)
- [ ] Gracefully handles missing API key
- [ ] Gracefully handles network errors
- [ ] Caches responses for 60 seconds
- [ ] Tests pass with > 80% coverage
- [ ] Works in both standalone and statusline modes

## Out of Scope

- Reset time display (future enhancement)
- Plan level display (future enhancement)
- Configurable refresh interval (uses global setting)
