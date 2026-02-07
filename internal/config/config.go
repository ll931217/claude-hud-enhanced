package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/theme"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Colors           ColorsConfig   `yaml:"colors"`
	Layout           LayoutConfig   `yaml:"layout"`
	RefreshIntervalMs int            `yaml:"refresh_interval_ms"`
	Debug            bool           `yaml:"debug"`
	CompactMode      bool           `yaml:"compact_mode"`
	MaxLines         int            `yaml:"max_lines"`
}

// ColorsConfig holds color customization options
type ColorsConfig struct {
	Primary   string `yaml:"primary"`
	Secondary string `yaml:"secondary"`
	Error     string `yaml:"error"`
	Warning   string `yaml:"warning"`
	Info      string `yaml:"info"`
	Success   string `yaml:"success"`
	Muted     string `yaml:"muted"`
}

// LayoutConfig holds configuration for custom layouts
type LayoutConfig struct {
	Lines      []LineConfig    `yaml:"lines"`
	Responsive ResponsiveConfig `yaml:"responsive"`
}

// LineConfig defines sections on a single line with custom separator
type LineConfig struct {
	Sections  []string `yaml:"sections"` // Section names in order
	Separator string   `yaml:"separator"` // Custom separator for this line
	Wrap      bool     `yaml:"wrap"` // Allow wrapping to next line if too long
}

// ResponsiveConfig holds settings for responsive behavior
type ResponsiveConfig struct {
	Enabled bool `yaml:"enabled"` // Enable responsive behavior
	// Breakpoints for different terminal sizes
	Small  int `yaml:"small_breakpoint"`  // Default: 80 columns
	Medium int `yaml:"medium_breakpoint"` // Default: 120 columns
	Large  int `yaml:"large_breakpoint"`  // Default: 160 columns
}

// defaultConfig returns the embedded default configuration
func defaultConfig() *Config {
	// Use Catppuccin Mocha theme colors
	ct := theme.CatppuccinMocha()

	return &Config{
		Layout: DefaultLayout(),
		Colors: ColorsConfig{
			Primary:   ct.Primary,
			Secondary: ct.Secondary,
			Error:     ct.Error,
			Warning:   ct.Warning,
			Info:      ct.Info,
			Success:   ct.Success,
			Muted:     ct.Muted,
		},
		RefreshIntervalMs: 300,
		Debug:            false,
		CompactMode:      true,
		MaxLines:         2,
	}
}

// DefaultConfig returns the default configuration (public alias for defaultConfig)
func DefaultConfig() *Config {
	return defaultConfig()
}

// Load loads the configuration from the default config path
// Returns the default configuration if the file is missing or invalid
// Never crashes - always returns a valid config
func Load() *Config {
	configPath, err := getConfigPath()
	if err != nil {
		return defaultConfig()
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return defaultConfig()
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return defaultConfig()
	}

	// Parse YAML with graceful degradation
	config := defaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return defaultConfig()
	}

	// Validate and sanitize the config
	config.validate()

	return config
}

// LoadFromPath loads configuration from a specific path
// Useful for testing or custom config locations
// Never crashes - always returns a valid config
func LoadFromPath(path string) *Config {
	config := defaultConfig()

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return config
	}

	// Read the config file
	data, err := os.ReadFile(path)
	if err != nil {
		return config
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return config
	}

	// Validate and sanitize
	config.validate()

	return config
}

// getConfigPath returns the default configuration file path
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".config", "claude-hud", "config.yaml")
	return configPath, nil
}

// validate ensures the configuration is valid and applies sensible defaults
func (c *Config) validate() {
	// Get Catppuccin Mocha theme for defaults
	ct := theme.CatppuccinMocha()

	// Validate refresh interval (clamp between 100ms and 5000ms)
	if c.RefreshIntervalMs < 100 {
		c.RefreshIntervalMs = 100
	}
	if c.RefreshIntervalMs > 5000 {
		c.RefreshIntervalMs = 5000
	}

	// Validate colors - set defaults to Catppuccin Mocha if empty
	if c.Colors.Primary == "" {
		c.Colors.Primary = ct.Primary
	}
	if c.Colors.Secondary == "" {
		c.Colors.Secondary = ct.Secondary
	}
	if c.Colors.Error == "" {
		c.Colors.Error = ct.Error
	}
	if c.Colors.Warning == "" {
		c.Colors.Warning = ct.Warning
	}
	if c.Colors.Info == "" {
		c.Colors.Info = ct.Info
	}
	if c.Colors.Success == "" {
		c.Colors.Success = ct.Success
	}
	if c.Colors.Muted == "" {
		c.Colors.Muted = ct.Muted
	}
}

// GetEnabledSections returns a list of enabled section names in order from layout
// If layout is empty, returns all enabled sections in default order
func (c *Config) GetEnabledSections() []string {
	// If layout is configured, derive from layout.lines
	if len(c.Layout.Lines) > 0 {
		seen := make(map[string]bool)
		var result []string
		for _, line := range c.Layout.Lines {
			for _, sectionName := range line.Sections {
				if !seen[sectionName] {
					seen[sectionName] = true
					result = append(result, sectionName)
				}
			}
		}
		return result
	}

	// Fallback: return all enabled sections in default order
	defaultOrder := []string{"model", "contextbar", "duration", "beads", "status", "workspace", "tools", "sysinfo"}
	var result []string
	for _, name := range defaultOrder {
		if c.IsSectionEnabled(name) {
			result = append(result, name)
		}
	}
	return result
}

// IsSectionEnabled checks if a specific section is enabled
// A section is enabled if it appears in any layout.lines configuration
// If layout is empty, all sections are considered enabled (fallback behavior)
func (c *Config) IsSectionEnabled(sectionName string) bool {
	if len(c.Layout.Lines) == 0 {
		// No layout configured, all sections enabled
		return true
	}

	for _, line := range c.Layout.Lines {
		for _, name := range line.Sections {
			if name == sectionName {
				return true
			}
		}
	}
	return false
}

// GetRefreshInterval returns the refresh interval as a time.Duration
func (c *Config) GetRefreshInterval() time.Duration {
	return time.Duration(c.RefreshIntervalMs) * time.Millisecond
}

// Save writes the current configuration to the default config path
// Creates the config directory if it doesn't exist
func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ToYAML returns the YAML representation of the config
func (c *Config) ToYAML() (string, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DefaultLayout returns the default 3-line layout configuration
func DefaultLayout() LayoutConfig {
	return LayoutConfig{
		Lines: []LineConfig{
			{
				Sections:  []string{"model", "contextbar", "duration"},
				Separator: " | ",
			},
			{
				Sections:  []string{"workspace", "status", "beads"},
				Separator: " | ",
			},
			{
				Sections:  []string{"tools", "sysinfo"},
				Separator: " | ",
			},
		},
		Responsive: ResponsiveConfig{
			Enabled: true,
			Small:   80,
			Medium:  120,
			Large:   160,
		},
	}
}
