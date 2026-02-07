package beads

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/errors"
	"github.com/ll931217/claude-hud-enhanced/internal/watcher"
)

// Reader reads and caches beads issues from .beads/issues.jsonl
type Reader struct {
	mu              sync.RWMutex
	repoPath        string
	issues          map[string]*Issue
	byStatus        map[IssueStatus][]*Issue
	lastModTime     time.Time
	lastCheck       time.Time
	cacheTTL        time.Duration
	watcher         *watcher.Watcher
	watcherStarted  bool
	forceReload     bool // Set to true when file changes are detected
	watcherCancel   context.CancelFunc
	watcherDone     chan struct{}
}

// NewReader creates a new beads reader for the given repository path
func NewReader(repoPath string) *Reader {
	return &Reader{
		repoPath:      repoPath,
		issues:        make(map[string]*Issue),
		byStatus:      make(map[IssueStatus][]*Issue),
		cacheTTL:      500 * time.Millisecond, // Faster initial load, will be improved with file watching
		watcher:       watcher.NewWatcher(),
		watcherDone:   make(chan struct{}),
	}
}

// GetIssuesPath returns the path to the issues.jsonl file
func (r *Reader) GetIssuesPath() string {
	return filepath.Join(r.repoPath, ".beads", "issues.jsonl")
}

// Exists checks if the beads directory exists
func (r *Reader) Exists() bool {
	issuesPath := r.GetIssuesPath()
	_, err := os.Stat(issuesPath)
	return err == nil
}

// Load loads (or reloads) the issues from the JSONL file
func (r *Reader) Load(ctx context.Context) error {
	return errors.SafeCall(func() error {
		// Start watcher on first load if not already started
		r.startWatcherOnce()

		// Check if we need to reload (either TTL expired or forceReload flag set)
		r.mu.RLock()
		needReload := r.forceReload || time.Since(r.lastCheck) > r.cacheTTL
		r.mu.RUnlock()

		if !needReload && len(r.issues) > 0 {
			return nil
		}

		// Clear forceReload flag
		r.mu.Lock()
		r.forceReload = false
		r.mu.Unlock()

		// Check if file exists
		issuesPath := r.GetIssuesPath()
		if _, err := os.Stat(issuesPath); os.IsNotExist(err) {
			return fmt.Errorf("beads issues file not found: %s", issuesPath)
		}

		// Get file modification time
		info, err := os.Stat(issuesPath)
		if err != nil {
			return fmt.Errorf("failed to stat issues file: %w", err)
		}

		// Check if file has been modified since last read
		r.mu.RLock()
		modified := info.ModTime().After(r.lastModTime)
		r.mu.RUnlock()

		if !modified && len(r.issues) > 0 {
			// File hasn't changed and we have cached data
			return nil
		}

		// Open the file
		file, err := os.Open(issuesPath)
		if err != nil {
			return fmt.Errorf("failed to open issues file: %w", err)
		}
		defer file.Close()

		// Clear cache
		r.mu.Lock()
		r.issues = make(map[string]*Issue)
		r.byStatus = make(map[IssueStatus][]*Issue)
		r.lastModTime = info.ModTime()
		r.lastCheck = time.Now()
		r.mu.Unlock()

		// Parse line by line
		scanner := bufio.NewScanner(file)
		lineNum := 0

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			lineNum++
			line := scanner.Bytes()

			if len(line) == 0 {
				continue
			}

			// Parse the issue
			var issue Issue
			if err := json.Unmarshal(line, &issue); err != nil {
				// Log error but continue parsing
				errors.Warn("beads.reader", "line %d: %v", lineNum, err)
				continue
			}

			// Add to cache
			r.mu.Lock()
			r.issues[issue.ID] = &issue
			r.byStatus[issue.Status] = append(r.byStatus[issue.Status], &issue)
			r.mu.Unlock()
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scanner error: %w", err)
		}

		return nil
	})
}

// GetAll returns all loaded issues
func (r *Reader) GetAll() map[string]*Issue {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy
	result := make(map[string]*Issue, len(r.issues))
	for k, v := range r.issues {
		result[k] = v
	}
	return result
}

// GetByID returns an issue by ID
func (r *Reader) GetByID(id string) *Issue {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.issues[id]
}

// GetByStatus returns issues filtered by status
func (r *Reader) GetByStatus(status IssueStatus) []*Issue {
	r.mu.RLock()
	defer r.mu.RUnlock()

	issues := r.byStatus[status]
	if issues == nil {
		return nil
	}

	// Return a copy
	result := make([]*Issue, len(issues))
	copy(result, issues)
	return result
}

// GetOpen returns all open issues
func (r *Reader) GetOpen() []*Issue {
	return r.GetByStatus(StatusOpen)
}

// GetInProgress returns all in-progress issues
func (r *Reader) GetInProgress() []*Issue {
	return r.GetByStatus(StatusInProgress)
}

// GetClosed returns all closed issues
func (r *Reader) GetClosed() []*Issue {
	return r.GetByStatus(StatusClosed)
}

// GetEpics returns all epic-type issues
func (r *Reader) GetEpics() []*Issue {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Issue
	for _, issue := range r.issues {
		if issue.IsEpic() {
			result = append(result, issue)
		}
	}

	return result
}

// Count returns the total number of issues
func (r *Reader) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.issues)
}

// CountByStatus returns the count of issues by status
func (r *Reader) CountByStatus(status IssueStatus) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	issues := r.byStatus[status]
	if issues == nil {
		return 0
	}
	return len(issues)
}

// GetCurrentIssue attempts to detect the current/working issue
// This is a heuristic - it looks for in-progress issues first,
// then falls back to the most recently updated open issue
func (r *Reader) GetCurrentIssue() *Issue {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// First, look for an in-progress issue
	if len(r.byStatus[StatusInProgress]) > 0 {
		// Return the most recently updated in-progress issue
		var latest *Issue
		for _, issue := range r.byStatus[StatusInProgress] {
			if latest == nil || issue.UpdatedAt.After(latest.UpdatedAt) {
				latest = issue
			}
		}
		return latest
	}

	// Fall back to the most recently updated open issue
	if len(r.byStatus[StatusOpen]) > 0 {
		var latest *Issue
		for _, issue := range r.byStatus[StatusOpen] {
			if latest == nil || issue.UpdatedAt.After(latest.UpdatedAt) {
				latest = issue
			}
		}
		return latest
	}

	return nil
}

// Refresh triggers a reload of the issues
func (r *Reader) Refresh(ctx context.Context) error {
	r.mu.Lock()
	r.lastCheck = time.Time{} // Force reload
	r.mu.Unlock()

	return r.Load(ctx)
}

// SetCacheTTL sets the cache time-to-live
func (r *Reader) SetCacheTTL(ttl time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cacheTTL = ttl
}

// startWatcherOnce starts the file watcher on first call (idempotent)
func (r *Reader) startWatcherOnce() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.watcherStarted {
		return
	}

	r.watcherStarted = true

	// Watch the issues file
	issuesPath := r.GetIssuesPath()
	if err := r.watcher.AddWatch(issuesPath); err != nil {
		errors.Warn("beads.reader", "failed to watch issues file: %v", err)
		return
	}

	// Start watcher in background
	ctx, cancel := context.WithCancel(context.Background())
	r.watcherCancel = cancel

	go func() {
		defer close(r.watcherDone)

		if err := r.watcher.Start(ctx); err != nil {
			errors.Warn("beads.reader", "watcher error: %v", err)
			return
		}

		// Handle file change events
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-r.watcher.Events():
				if event.Path == issuesPath {
					// File changed - invalidate cache
					r.mu.Lock()
					r.forceReload = true
					r.mu.Unlock()
					errors.Debug("beads.reader", "file changed, forcing reload")
				}
			case err := <-r.watcher.Errors():
				errors.Warn("beads.reader", "watcher error: %v", err)
			}
		}
	}()

	errors.Debug("beads.reader", "started watching %s", issuesPath)
}

// Stop stops the file watcher
func (r *Reader) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.watcherCancel != nil {
		r.watcherCancel()
		r.watcherCancel = nil
	}

	if r.watcher != nil {
		r.watcher.Stop()
	}

	// Wait for watcher goroutine to finish
	if r.watcherDone != nil {
		<-r.watcherDone
	}
}
