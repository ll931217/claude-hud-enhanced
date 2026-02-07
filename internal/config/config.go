package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/theme"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Sections         SectionsConfig `yaml:"sections"`
	Colors           ColorsConfig   `yaml:"colors"`
	Layout           LayoutConfig   `yaml:"layout"`
	RefreshIntervalMs int            `yaml:"refresh_interval_ms"`
	Debug            bool           `yaml:"debug"`
	CompactMode      bool           `yaml:"compact_mode"`
	MaxLines         int            `yaml:"max_lines"`
}

// SectionsConfig holds configuration for all HUD sections
type SectionsConfig struct {
	Session   SectionConfig `yaml:"session"`
	Beads     SectionConfig `yaml:"beads"`
	Status    SectionConfig `yaml:"status"`
	Workspace SectionConfig `yaml:"workspace"`
	Tools     SectionConfig `yaml:"tools"`
	SysInfo   SectionConfig `yaml:"sysinfo"`
}

// SectionConfig represents configuration for a single section
type SectionConfig struct {
	Enabled bool `yaml:"enabled"`
	Order   int  `yaml:"order"`
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
		Sections: SectionsConfig{
			Session: SectionConfig{
				Enabled: true,
				Order:   1,
			},
			Beads: SectionConfig{
				Enabled: true,
				Order:   2,
			},
			Status: SectionConfig{
				Enabled: true,
				Order:   3,
			},
			Workspace: SectionConfig{
				Enabled: true,
				Order:   4,
			},
			Tools: SectionConfig{
				Enabled: true,
				Order:   5,
			},
			SysInfo: SectionConfig{
				Enabled: true,
				Order:   6,
			},
		},
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

	// Ensure all section orders are unique and positive
	c.normalizeSectionOrders()
}

// normalizeSectionOrders ensures section orders are unique and start from 1
func (c *Config) normalizeSectionOrders() {
	sections := []struct {
		name   string
		config *SectionConfig
	}{
		{"session", &c.Sections.Session},
		{"beads", &c.Sections.Beads},
		{"status", &c.Sections.Status},
		{"workspace", &c.Sections.Workspace},
	}

	// Collect enabled sections with their orders
	type sectionOrder struct {
		name  string
		order int
	}
	var enabledSections []sectionOrder
	for _, s := range sections {
		if s.config.Enabled {
			// If order is 0 or negative, set to a default high value
			if s.config.Order <= 0 {
				s.config.Order = 999
			}
			enabledSections = append(enabledSections, sectionOrder{s.name, s.config.Order})
		}
	}

	// Sort by order
	sort.Slice(enabledSections, func(i, j int) bool {
		return enabledSections[i].order < enabledSections[j].order
	})

	// Reassign orders starting from 1
	for i, es := range enabledSections {
		switch es.name {
		case "session":
			c.Sections.Session.Order = i + 1
		case "beads":
			c.Sections.Beads.Order = i + 1
		case "status":
			c.Sections.Status.Order = i + 1
		case "workspace":
			c.Sections.Workspace.Order = i + 1
		}
	}
}

// GetEnabledSections returns a list of enabled section names in order
func (c *Config) GetEnabledSections() []string {
	type sectionOrder struct {
		name  string
		order int
	}

	var sections []sectionOrder

	if c.Sections.Session.Enabled {
		sections = append(sections, sectionOrder{"session", c.Sections.Session.Order})
	}
	if c.Sections.Beads.Enabled {
		sections = append(sections, sectionOrder{"beads", c.Sections.Beads.Order})
	}
	if c.Sections.Status.Enabled {
		sections = append(sections, sectionOrder{"status", c.Sections.Status.Order})
	}
	if c.Sections.Workspace.Enabled {
		sections = append(sections, sectionOrder{"workspace", c.Sections.Workspace.Order})
	}
	if c.Sections.Tools.Enabled {
		sections = append(sections, sectionOrder{"tools", c.Sections.Tools.Order})
	}
	if c.Sections.SysInfo.Enabled {
		sections = append(sections, sectionOrder{"sysinfo", c.Sections.SysInfo.Order})
	}

	sort.Slice(sections, func(i, j int) bool {
		return sections[i].order < sections[j].order
	})

	var result []string
	for _, s := range sections {
		result = append(result, s.name)
	}

	return result
}

// IsSectionEnabled checks if a specific section is enabled
func (c *Config) IsSectionEnabled(sectionName string) bool {
	switch sectionName {
	case "session":
		return c.Sections.Session.Enabled
	case "beads":
		return c.Sections.Beads.Enabled
	case "status":
		return c.Sections.Status.Enabled
	case "workspace":
		return c.Sections.Workspace.Enabled
	case "tools":
		return c.Sections.Tools.Enabled
	case "sysinfo":
		return c.Sections.SysInfo.Enabled
	default:
		return false
	}
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

// GetSectionOrder returns the order for a specific section
func (c *Config) GetSectionOrder(sectionName string) int {
	switch sectionName {
	case "session":
		return c.Sections.Session.Order
	case "beads":
		return c.Sections.Beads.Order
	case "status":
		return c.Sections.Status.Order
	case "workspace":
		return c.Sections.Workspace.Order
	case "tools":
		return c.Sections.Tools.Order
	case "sysinfo":
		return c.Sections.SysInfo.Order
	default:
		return 999
	}
}

// DefaultLayout returns the default 4-line layout configuration
func DefaultLayout() LayoutConfig {
	return LayoutConfig{
		Lines: []LineConfig{
			{
				Sections:  []string{"session"},
				Separator: " | ",
			},
			{
				Sections:  []string{"workspace", "status"},
				Separator: " | ",
			},
			{
				Sections:  []string{"tools"},
				Separator: " | ",
			},
			{
				Sections:  []string{"sysinfo"},
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
