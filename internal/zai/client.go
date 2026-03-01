package zai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	apiURL   = "https://api.z.ai/api/monitor/usage/quota/limit"
	timeout  = 10 * time.Second
	cacheTTL = 60 * time.Second
)

// Client handles fetching usage data from the Z.ai API
type Client struct {
	httpClient *http.Client
	cache      *UsageCache
}

// UsageCache provides thread-safe caching of usage data
type UsageCache struct {
	mu        sync.RWMutex
	data      *UsageInfo
	fetchedAt time.Time
	ttl       time.Duration
}

// NewCache creates a new usage cache with the specified TTL
func NewCache(ttl time.Duration) *UsageCache {
	return &UsageCache{
		ttl: ttl,
	}
}

// Get returns cached data if still valid, otherwise nil
func (c *UsageCache) Get() *UsageInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.data == nil || time.Since(c.fetchedAt) > c.ttl {
		return nil
	}
	return c.data
}

// Set updates the cache with new data
func (c *UsageCache) Set(data *UsageInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = data
	c.fetchedAt = time.Now()
}

// NewClient creates a new Z.ai API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		cache: NewCache(cacheTTL),
	}
}

// getAPIKey retrieves the API key from environment variables
// Checks GLM_API_KEY first, then ZAI_API_KEY as fallback
func getAPIKey() string {
	if key := os.Getenv("GLM_API_KEY"); key != "" {
		return key
	}
	return os.Getenv("ZAI_API_KEY")
}

// Fetch retrieves usage data from the API or returns cached data
func (c *Client) Fetch() *UsageInfo {
	// Check cache first
	if cached := c.cache.Get(); cached != nil {
		return cached
	}

	// Get API key
	apiKey := getAPIKey()
	if apiKey == "" {
		return nil
	}

	// Fetch from API
	data, err := c.fetchFromAPI(apiKey)
	if err != nil {
		return nil
	}

	// Cache the result
	c.cache.Set(data)
	return data
}

// fetchFromAPI makes the actual HTTP request to the Z.ai API
func (c *Client) fetchFromAPI(apiKey string) (*UsageInfo, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	return parseUsageData(&apiResp.Data), nil
}

// parseUsageData extracts usage information from the API response
func parseUsageData(data *UsageData) *UsageInfo {
	info := &UsageInfo{
		PlanLevel: data.Level,
		FetchedAt: time.Now(),
	}

	for _, limit := range data.Limits {
		switch {
		case limit.Type == LimitTypeTokens && limit.Unit == UnitSession && limit.Number == 5:
			// 5-hour rolling window (session)
			info.SessionPercent = limit.Percentage
			if limit.NextResetTime > 0 {
				info.SessionReset = time.UnixMilli(limit.NextResetTime)
			}
		case limit.Type == LimitTypeTokens && limit.Unit == UnitWeekly:
			// Weekly aggregate
			info.WeeklyPercent = limit.Percentage
			if limit.NextResetTime > 0 {
				info.WeeklyReset = time.UnixMilli(limit.NextResetTime)
			}
		case limit.Type == LimitTypeTime && limit.Unit == UnitSearch && limit.Number == 1:
			// Monthly search quota
			info.SearchPercent = limit.Percentage
		}
	}

	return info
}
