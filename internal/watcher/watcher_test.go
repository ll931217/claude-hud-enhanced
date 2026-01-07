package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewWatcher(t *testing.T) {
	w := NewWatcher()
	if w == nil {
		t.Fatal("NewWatcher() returned nil")
	}
	if w.eventChan == nil {
		t.Error("event channel not initialized")
	}
	if w.errorChan == nil {
		t.Error("error channel not initialized")
	}
	if w.watchPaths == nil {
		t.Error("watch paths map not initialized")
	}
}

func TestWatcher_AddWatch_NonExistent(t *testing.T) {
	w := NewWatcher()
	if err := w.AddWatch("/non/existent/path"); err != nil {
		t.Errorf("AddWatch() with non-existent path should not error, got: %v", err)
	}
}

func TestWatcher_AddWatch_Valid(t *testing.T) {
	// Create a temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	w := NewWatcher()
	if err := w.AddWatch(testFile); err != nil {
		t.Errorf("AddWatch() error = %v", err)
	}

	w.mu.Lock()
	if !w.watchPaths[testFile] {
		t.Error("path not added to watch list")
	}
	w.mu.Unlock()

	w.Stop()
}

func TestWatcher_PollingMode(t *testing.T) {
	// Create a temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	w := NewWatcher()
	w.SetPollingInterval(50 * time.Millisecond)
	w.AddWatch(testFile)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Modify the file
	time.Sleep(100 * time.Millisecond)
	if err := os.WriteFile(testFile, []byte("modified"), 0644); err != nil {
		t.Fatal(err)
	}

	// Check for event
	select {
	case event := <-w.Events():
		if event.Path != testFile {
			t.Errorf("expected path %s, got %s", testFile, event.Path)
		}
		if event.EventType != EventModified {
			t.Errorf("expected EventModified, got %v", event.EventType)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("did not receive file modification event")
	}

	w.Stop()
}

func TestWatcher_GetMode(t *testing.T) {
	w := NewWatcher()
	if w.GetMode() != ModeFsnotify {
		t.Errorf("expected initial mode ModeFsnotify, got %v", w.GetMode())
	}
	w.Stop()
}

func TestWatcher_SetIntervals(t *testing.T) {
	w := NewWatcher()

	w.SetPollingInterval(100 * time.Millisecond)
	w.mu.RLock()
	if w.pollingInterval != 100*time.Millisecond {
		t.Errorf("polling interval not set, got %v", w.pollingInterval)
	}
	w.mu.RUnlock()

	w.SetRecoveryInterval(10 * time.Second)
	w.mu.RLock()
	if w.recoveryInterval != 10*time.Second {
		t.Errorf("recovery interval not set, got %v", w.recoveryInterval)
	}
	w.mu.RUnlock()

	w.Stop()
}

func TestWatcher_StopIdempotent(t *testing.T) {
	w := NewWatcher()
	w.Stop()
	w.Stop() // Should not panic or error
}

func TestWatcher_MultipleFiles(t *testing.T) {
	// Create temp files
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	if err := os.WriteFile(file1, []byte("test1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("test2"), 0644); err != nil {
		t.Fatal(err)
	}

	w := NewWatcher()
	w.SetPollingInterval(50 * time.Millisecond)
	w.AddWatch(file1)
	w.AddWatch(file2)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Modify both files
	time.Sleep(100 * time.Millisecond)
	os.WriteFile(file1, []byte("modified1"), 0644)
	os.WriteFile(file2, []byte("modified2"), 0644)

	// Check for events
	events := make(map[string]bool)
	timeout := time.After(300 * time.Millisecond)
	for len(events) < 2 {
		select {
		case event := <-w.Events():
			events[event.Path] = true
		case <-timeout:
			t.Errorf("did not receive all events, got %d", len(events))
			w.Stop()
			return
		}
	}

	if !events[file1] {
		t.Error("did not receive event for file1")
	}
	if !events[file2] {
		t.Error("did not receive event for file2")
	}

	w.Stop()
}
