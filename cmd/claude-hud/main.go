package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/errors"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
	"github.com/ll931217/claude-hud-enhanced/internal/statusline"
	"github.com/ll931217/claude-hud-enhanced/internal/version"
	_ "github.com/ll931217/claude-hud-enhanced/internal/sections" // Register sections via init()
)

var (
	showVersion = flag.Bool("version", false, "Show version information")
	showBuild   = flag.Bool("build-info", false, "Show detailed build information")
	statuslineMode = flag.Bool("statusline", false, "Run in Claude Code statusline mode (single shot, multiline output)")
)

func main() {
	// Auto-detect statusline mode: if stdin has data (not a TTY), assume statusline mode
	// This allows the binary to work directly with Claude Code without the --statusline flag
	if !isStdinTTY() && !hasExplicitFlags() {
		// Parse JSON from stdin and run in statusline mode
		if err := runStatuslineMode(); err != nil {
			// Silent failure for statusline mode
			os.Exit(0)
		}
		os.Exit(0)
	}

	// Parse flags
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Println(version.FullVersionInfo())
		os.Exit(0)
	}

	// Handle build info flag
	if *showBuild {
		info := version.BuildInfo()
		fmt.Println("Claude HUD Enhanced Build Information")
		fmt.Println("===================================")
		fmt.Printf("Version:   %s\n", info["version"])
		fmt.Printf("Commit:    %s\n", info["commit"])
		fmt.Printf("Built At:  %s\n", info["built_at"])
		fmt.Printf("Go Version: %s\n", info["go_version"])
		os.Exit(0)
	}

	// Handle statusline mode - single shot output for Claude Code
	if *statuslineMode {
		if err := runStatuslineMode(); err != nil {
			// Silent failure for statusline mode
			os.Exit(0)
		}
		os.Exit(0)
	}
	// Set up panic recovery at the top level
	defer errors.MainRecovery()

	// Load configuration with error handling
	cfg := config.Load()
	if cfg == nil {
		cfg = config.DefaultConfig()
		errors.Warn("main", "using default configuration")
	}

	// Configure logging based on config
	if cfg.Debug {
		errors.SetDebugMode(true)
		errors.Info("main", "debug mode enabled")
	}

	// Log startup
	errors.Info("main", "Claude HUD Enhanced starting")
	errors.Info("main", "refresh interval: %dms", cfg.RefreshIntervalMs)

	// Create and run the application
	app, err := NewApplication(cfg)
	if err != nil {
		errors.LogErrorWithLevel(err)
		errors.Error("main", "failed to create application")
		os.Exit(1)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start the application in a goroutine with panic recovery
	errors.SafeGo("app.run", func() {
		if err := app.Run(); err != nil {
			errors.LogErrorWithLevel(err)
			errors.Error("main", "application error")
		}
	})

	// Wait for shutdown signal
	<-sigChan
	errors.Info("main", "shutdown signal received")

	// Stop the application with error handling
	if err := app.Stop(); err != nil {
		errors.LogErrorWithLevel(err)
		errors.Error("main", "error during shutdown")
	}

	errors.Info("main", "Claude HUD Enhanced stopped")
}

// runStatuslineMode runs the statusline in single-shot mode for Claude Code
func runStatuslineMode() error {
	// Read JSON from stdin (non-blocking if no input)
	input, err := readStdinJSON()
	if err != nil {
		// Invalid JSON, but continue anyway
		input = nil
	}

	// Load configuration
	cfg := config.Load()
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Set global context from JSON input
	if input != nil {
		// Change to workspace directory if specified
		if input.Workspace.CurrentDir != "" {
			if err := os.Chdir(input.Workspace.CurrentDir); err == nil {
				statusline.SetContext(
					input.TranscriptPath,
					input.Workspace.CurrentDir,
					input.Model.DisplayName,
				)
			}
		}
	}

	// Create statusline with registry
	sl, err := statusline.New(cfg, registry.DefaultRegistry())
	if err != nil {
		return fmt.Errorf("failed to create statusline: %w", err)
	}

	// Create sections from config
	enabledSections := cfg.GetEnabledSections()
	for _, sectionName := range enabledSections {
		section, err := registry.Create(sectionName, cfg)
		if err != nil {
			continue
		}
		sl.AddSection(section)
	}

	// Render once and exit (no continuous refresh)
	return sl.RenderStatuslineMode()
}

// Application represents the main application
type Application struct {
	config     *config.Config
	statusline *statusline.Statusline
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewApplication creates a new application instance with error handling
func NewApplication(cfg *config.Config) (*Application, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Create statusline with registry
	sl, err := statusline.New(cfg, registry.DefaultRegistry())
	if err != nil {
		return nil, fmt.Errorf("failed to create statusline: %w", err)
	}

	// Create sections from config
	enabledSections := cfg.GetEnabledSections()
	for _, sectionName := range enabledSections {
		section, err := registry.Create(sectionName, cfg)
		if err != nil {
			errors.Warn("app", "failed to create section %s: %v", sectionName, err)
			continue
		}
		sl.AddSection(section)
	}

	ctx, cancel := context.WithCancel(context.Background())

	app := &Application{
		config:     cfg,
		statusline: sl,
		ctx:        ctx,
		cancel:     cancel,
	}

	errors.Info("app", "application created with %d sections", len(enabledSections))

	return app, nil
}

// Run starts the main application loop with panic recovery
func (a *Application) Run() error {
	errors.Info("app", "starting application")

	// Run the statusline refresh loop
	if err := a.statusline.Run(a.ctx); err != nil && err != context.Canceled {
		errors.LogErrorWithLevel(err)
		errors.Error("app", "statusline error")
		return err
	}

	errors.Info("app", "application stopped")
	return nil
}

// Stop stops the application gracefully with error handling
func (a *Application) Stop() error {
	errors.Info("app", "stopping application")

	// Cancel the context to stop the statusline
	a.cancel()

	// Stop the statusline
	a.statusline.Stop()

	errors.Info("app", "application stopped successfully")
	return nil
}

// isStdinTTY checks if stdin is a terminal (has no piped input)
func isStdinTTY() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice != 0
}

// hasExplicitFlags checks if any command-line flags were provided
func hasExplicitFlags() bool {
	// Check if any arguments were passed (beyond program name)
	return len(os.Args) > 1
}
