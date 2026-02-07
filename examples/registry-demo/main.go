package main

import (
	"fmt"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
	_ "github.com/ll931217/claude-hud-enhanced/internal/sections" // Import to trigger init()
)

// This demonstrates the section registry factory pattern
func main() {
	fmt.Println("=== Section Registry Factory Pattern Demo ===")
	fmt.Println()

	// 1. List all registered sections
	fmt.Println("1. Registered sections:")
	sections := registry.List()
	for _, name := range sections {
		fmt.Printf("   - %s\n", name)
	}
	fmt.Println()

	// 2. Create sections with default config
	fmt.Println("2. Creating sections with default config:")
	cfg := config.DefaultConfig()
	for _, name := range sections {
		section, err := registry.Create(name, cfg)
		if err != nil {
			fmt.Printf("   ERROR creating %s: %v\n", name, err)
			continue
		}
		fmt.Printf("   - %s: enabled=%v, order=%d, render=%q\n",
			section.Name(), section.Enabled(), section.Order(), section.Render())
	}
	fmt.Println()

	// 3. Create sections with custom config (disabled section)
	fmt.Println("3. Creating section with custom config (disabled):")
	customCfg := &config.Config{}
	customCfg.Sections.Session.Enabled = false
	customCfg.Sections.Session.Order = 99

	section, err := registry.Create("session", customCfg)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   - %s: enabled=%v, order=%d\n",
			section.Name(), section.Enabled(), section.Order())
	}
	fmt.Println()

	// 4. Register custom section type
	fmt.Println("4. Registering custom section type:")
	customFactory := func(cfg interface{}) (registry.Section, error) {
		return &customSection{name: "custom", enabled: true, order: 100}, nil
	}
	registry.Register("custom", customFactory)

	customSection, err := registry.Create("custom", nil)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   - Created custom section: %s, render=%q\n",
			customSection.Name(), customSection.Render())
	}
	fmt.Println()

	// 5. Try to create unregistered section
	fmt.Println("5. Attempting to create unregistered section:")
	_, err = registry.Create("nonexistent", nil)
	if err != nil {
		fmt.Printf("   - Expected error: %v\n", err)
	}

	fmt.Println("\n=== Demo Complete ===")
}

// customSection is a simple custom section implementation
type customSection struct {
	name    string
	enabled bool
	order   int
}

func (c *customSection) Render() string {
	return "[Custom Section]"
}

func (c *customSection) Enabled() bool {
	return c.enabled
}

func (c *customSection) Order() int {
	return c.order
}

func (c *customSection) Name() string {
	return c.name
}

func (c *customSection) Priority() registry.Priority {
	return registry.PriorityImportant
}

func (c *customSection) MinWidth() int {
	return 0
}
