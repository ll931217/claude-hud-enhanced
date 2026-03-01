// Package zai provides a client for fetching Z.ai usage quota information
package zai

import "time"

// APIResponse represents the top-level response from the Z.ai API
type APIResponse struct {
	Success bool     `json:"success"`
	Data    UsageData `json:"data"`
}

// UsageData contains the usage limits and plan level
type UsageData struct {
	Level  string  `json:"level"`
	Limits []Limit `json:"limits"`
}

// Limit represents a single usage limit/quota
type Limit struct {
	Type          string `json:"type"`           // "TOKENS_LIMIT" or "TIME_LIMIT"
	Unit          int    `json:"unit"`           // Limit unit identifier
	Number        int    `json:"number"`         // Limit number (e.g., 5 for 5-hour window)
	Percentage    int    `json:"percentage"`     // Usage percentage (0-100)
	NextResetTime int64  `json:"nextResetTime"`  // Unix timestamp in milliseconds
}

// UsageInfo contains the parsed usage information for display
type UsageInfo struct {
	SessionPercent int       // 5-hour rolling window usage
	WeeklyPercent  int       // Weekly aggregate usage
	SearchPercent  int       // Monthly search quota usage
	PlanLevel      string    // Plan level (e.g., "pro")
	SessionReset   time.Time // Session reset time
	WeeklyReset    time.Time // Weekly reset time
	FetchedAt      time.Time // When this data was fetched
}

// IsEmpty returns true if the usage info has no data
func (u *UsageInfo) IsEmpty() bool {
	return u.SessionPercent == 0 && u.WeeklyPercent == 0 && u.SearchPercent == 0
}

// Constants for limit type identification
const (
	LimitTypeTokens = "TOKENS_LIMIT"
	LimitTypeTime   = "TIME_LIMIT"

	// Unit identifiers based on API response
	UnitSession = 3  // 5-hour rolling window (with number=5)
	UnitWeekly  = 6  // Weekly aggregate
	UnitSearch  = 5  // Monthly search (with number=1)
)
