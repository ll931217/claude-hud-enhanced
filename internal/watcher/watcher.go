package watcher

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/ll931217/claude-hud-enhanced/internal/errors"
)

// EventType represents the type of change event
type EventType int

const (
	EventModified EventType = iota
	EventCreated
	EventDeleted
)

// Event represents a file change event
type Event struct {
	Path      string
	EventType EventType
}

// WatcherMode represents the current watching mode
type WatcherMode int

const (
	ModeFsnotify WatcherMode = iota
	ModePolling
)

// Watcher watches files for changes with fsnotify and polling fallback
type Watcher struct {
	mu               sync.RWMutex
	mode             WatcherMode
	fsnotifyWatcher  *fsnotify.Watcher
	pollingTicker    *time.Ticker
	watchPaths       map[string]bool
	eventChan        chan Event
	errorChan        chan error
	stopChan         chan struct{}
	wg               sync.WaitGroup
	recoveryInterval time.Duration
	pollingInterval  time.Duration
	lastModTimes     map[string]time.Time
	ctx              context.Context
	stopped          bool
}

// NewWatcher creates a new file watcher
func NewWatcher() *Watcher {
	return &Watcher{
		watchPaths:       make(map[string]bool),
		eventChan:        make(chan Event, 100),
		errorChan:        make(chan error, 10),
		stopChan:         make(chan struct{}),
		recoveryInterval: 30 * time.Second,
		pollingInterval:  300 * time.Millisecond,
		lastModTimes:     make(map[string]time.Time),
	}
}

// AddWatch adds a path to be watched
func (w *Watcher) AddWatch(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // Don't error, just don't watch non-existent files
	}

	// Add to watch list
	w.watchPaths[path] = true

	// Initialize last mod time
	if info, err := os.Stat(path); err == nil {
		w.lastModTimes[path] = info.ModTime()
	}

	// If fsnotify watcher is active, add the watch
	if w.mode == ModeFsnotify && w.fsnotifyWatcher != nil {
		// Watch the parent directory for file changes
		dir := filepath.Dir(path)
		if err := w.fsnotifyWatcher.Add(dir); err != nil {
			errors.Warn("watcher", "failed to watch directory %s: %v", dir, err)
			// Fall back to polling
			w.fallbackToPolling()
		}
	}

	return nil
}

// Events returns the event channel
func (w *Watcher) Events() <-chan Event {
	return w.eventChan
}

// Errors returns the error channel
func (w *Watcher) Errors() <-chan error {
	return w.errorChan
}

// Start begins watching files
func (w *Watcher) Start(ctx context.Context) error {
	return errors.SafeCall(func() error {
		w.mu.Lock()
		w.ctx = ctx
		w.mu.Unlock()

		// Try to start fsnotify watcher
		if err := w.startFsnotifyWatcher(); err != nil {
			errors.Warn("watcher", "fsnotify not available, using polling: %v", err)
			w.startPolling()
		}

		// Start recovery goroutine
		w.wg.Add(1)
		go w.recoveryLoop(ctx)

		// Start event processing
		w.wg.Add(1)
		go w.processEvents(ctx)

		return nil
	})
}

// startFsnotifyWatcher attempts to start the fsnotify watcher
func (w *Watcher) startFsnotifyWatcher() error {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	w.mu.Lock()
	w.fsnotifyWatcher = fsw
	w.mode = ModeFsnotify

	// Add watches for all paths
	for path := range w.watchPaths {
		dir := filepath.Dir(path)
		if err := fsw.Add(dir); err != nil {
			w.mu.Unlock()
			fsw.Close()
			return err
		}
	}
	w.mu.Unlock()

	// Start fsnotify event loop
	w.wg.Add(1)
	go w.fsnotifyEventLoop()

	return nil
}

// fsnotifyEventLoop processes fsnotify events
func (w *Watcher) fsnotifyEventLoop() {
	defer w.wg.Done()

	w.mu.RLock()
	ctx := w.ctx
	w.mu.RUnlock()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-w.fsnotifyWatcher.Events:
			if !ok {
				return
			}
			w.handleFsnotifyEvent(event)
		case err, ok := <-w.fsnotifyWatcher.Errors:
			if !ok {
				return
			}
			w.errorChan <- err
			// Fall back to polling on error
			w.fallbackToPolling()
		}
	}
}

// handleFsnotifyEvent handles a single fsnotify event
func (w *Watcher) handleFsnotifyEvent(event fsnotify.Event) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if this is a path we're watching
	for path := range w.watchPaths {
		if event.Name == path {
			// Check if file was actually modified
			if info, err := os.Stat(path); err == nil {
				lastMod := w.lastModTimes[path]
				if info.ModTime().After(lastMod) {
					w.lastModTimes[path] = info.ModTime()
					w.eventChan <- Event{Path: path, EventType: EventModified}
				}
			}
			break
		}
	}
}

// fallbackToPolling switches to polling mode
func (w *Watcher) fallbackToPolling() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.mode == ModePolling {
		return // Already in polling mode
	}

	// Close fsnotify watcher if open
	if w.fsnotifyWatcher != nil {
		w.fsnotifyWatcher.Close()
		w.fsnotifyWatcher = nil
	}

	w.mode = ModePolling
	w.startPolling()
	errors.Warn("watcher", "fell back to polling mode")
}

// startPolling starts the polling ticker
func (w *Watcher) startPolling() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.pollingTicker != nil {
		return // Already polling
	}

	w.pollingTicker = time.NewTicker(w.pollingInterval)

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.pollingLoop(context.Background())
	}()
}

// pollingLoop checks for file changes periodically
func (w *Watcher) pollingLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		case <-w.pollingTicker.C:
			w.checkForChanges()
		}
	}
}

// checkForChanges checks all watched files for modifications
func (w *Watcher) checkForChanges() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for path := range w.watchPaths {
		if info, err := os.Stat(path); err == nil {
			lastMod := w.lastModTimes[path]
			if info.ModTime().After(lastMod) {
				w.lastModTimes[path] = info.ModTime()
				w.eventChan <- Event{Path: path, EventType: EventModified}
			}
		}
	}
}

// recoveryLoop periodically attempts to recover fsnotify
func (w *Watcher) recoveryLoop(ctx context.Context) {
	defer w.wg.Done()

	recoveryTicker := time.NewTicker(w.recoveryInterval)
	defer recoveryTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-recoveryTicker.C:
			if w.mode == ModePolling {
				w.tryRecoverFsnotify()
			}
		}
	}
}

// tryRecoverFsnotify attempts to recover fsnotify mode
func (w *Watcher) tryRecoverFsnotify() {
	if err := w.startFsnotifyWatcher(); err == nil {
		errors.Info("watcher", "recovered fsnotify mode")
	}
}

// processEvents ensures events are processed
func (w *Watcher) processEvents(ctx context.Context) {
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		}
	}
}

// Stop stops the watcher
func (w *Watcher) Stop() {
	w.mu.Lock()
	if w.stopped {
		w.mu.Unlock()
		return
	}
	w.stopped = true
	w.mu.Unlock()

	close(w.stopChan)

	// Stop polling ticker
	w.mu.Lock()
	if w.pollingTicker != nil {
		w.pollingTicker.Stop()
		w.pollingTicker = nil
	}
	// Close fsnotify watcher
	if w.fsnotifyWatcher != nil {
		w.fsnotifyWatcher.Close()
		w.fsnotifyWatcher = nil
	}
	w.mu.Unlock()

	// Wait for all goroutines to finish
	w.wg.Wait()

	// Close channels
	close(w.eventChan)
	close(w.errorChan)
}

// GetMode returns the current watcher mode
func (w *Watcher) GetMode() WatcherMode {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.mode
}

// SetPollingInterval sets the polling interval (for testing)
func (w *Watcher) SetPollingInterval(interval time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.pollingInterval = interval
}

// SetRecoveryInterval sets the recovery interval (for testing)
func (w *Watcher) SetRecoveryInterval(interval time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.recoveryInterval = interval
}
