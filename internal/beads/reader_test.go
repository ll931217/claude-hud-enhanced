package beads

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIssueStatus_Icon(t *testing.T) {
	tests := []struct {
		status   IssueStatus
		expected string
	}{
		{StatusOpen, "✗"},
		{StatusInProgress, "◐"},
		{StatusClosed, "✓"},
		{StatusBlocked, "✖"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.Icon(); got != tt.expected {
				t.Errorf("Icon() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPriority_String(t *testing.T) {
	tests := []struct {
		priority Priority
		expected string
	}{
		{PriorityCritical, "P0"},
		{PriorityHigh, "P1"},
		{PriorityNormal, "P2"},
		{PriorityLow, "P3"},
		{PriorityLowest, "P4"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.priority.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestReader_Load(t *testing.T) {
	ctx := context.Background()

	// Create temporary directory
	tmpDir := t.TempDir()
	beadsDir := filepath.Join(tmpDir, ".beads")
	os.MkdirAll(beadsDir, 0755)

	// Create test issues.jsonl
	issuesPath := filepath.Join(beadsDir, "issues.jsonl")
	content := `{"id":"test-1","title":"Test Issue 1","status":"open","priority":1,"issue_type":"task","created_at":"2026-01-07T12:00:00Z","updated_at":"2026-01-07T12:00:00Z"}
{"id":"test-2","title":"Test Issue 2","status":"in_progress","priority":2,"issue_type":"bug","created_at":"2026-01-07T12:00:00Z","updated_at":"2026-01-07T12:01:00Z"}
{"id":"test-3","title":"Test Epic","status":"open","priority":1,"issue_type":"epic","created_at":"2026-01-07T12:00:00Z","updated_at":"2026-01-07T12:02:00Z"}
`
	if err := os.WriteFile(issuesPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create reader
	reader := NewReader(tmpDir)

	// Load issues
	if err := reader.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check total count
	if count := reader.Count(); count != 3 {
		t.Errorf("Count() = %v, want 3", count)
	}

	// Check GetByID
	issue := reader.GetByID("test-1")
	if issue == nil {
		t.Fatal("GetByID(test-1) returned nil")
	}
	if issue.Title != "Test Issue 1" {
		t.Errorf("Title = %v, want 'Test Issue 1'", issue.Title)
	}

	// Check GetByStatus
	openIssues := reader.GetByStatus(StatusOpen)
	if len(openIssues) != 2 {
		t.Errorf("GetByStatus(open) = %v, want 2", len(openIssues))
	}

	inProgressIssues := reader.GetByStatus(StatusInProgress)
	if len(inProgressIssues) != 1 {
		t.Errorf("GetByStatus(in_progress) = %v, want 1", len(inProgressIssues))
	}

	// Check GetEpics
	epics := reader.GetEpics()
	if len(epics) != 1 {
		t.Errorf("GetEpics() = %v, want 1", len(epics))
	}

	// Check GetCurrentIssue (should return in_progress)
	current := reader.GetCurrentIssue()
	if current == nil {
		t.Fatal("GetCurrentIssue() returned nil")
	}
	if current.ID != "test-2" {
		t.Errorf("GetCurrentIssue().ID = %v, want test-2", current.ID)
	}
}

func TestReader_NotExists(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	reader := NewReader(tmpDir)

	if reader.Exists() {
		t.Error("Exists() returned true for non-existent beads")
	}

	err := reader.Load(ctx)
	if err == nil {
		t.Error("Load() should return error for non-existent file")
	}
}

func TestReader_Caching(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	beadsDir := filepath.Join(tmpDir, ".beads")
	os.MkdirAll(beadsDir, 0755)

	// Create test issues.jsonl
	issuesPath := filepath.Join(beadsDir, "issues.jsonl")
	content := `{"id":"test-1","title":"Test","status":"open","priority":2,"issue_type":"task","created_at":"2026-01-07T12:00:00Z","updated_at":"2026-01-07T12:00:00Z"}
`
	if err := os.WriteFile(issuesPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	reader := NewReader(tmpDir)
	reader.SetCacheTTL(1 * time.Second)

	// First load
	if err := reader.Load(ctx); err != nil {
		t.Fatalf("First Load() error = %v", err)
	}

	// Second load should use cache (check file not read again)
	if err := reader.Load(ctx); err != nil {
		t.Fatalf("Second Load() error = %v", err)
	}

	// Wait for cache to expire
	time.Sleep(1100 * time.Millisecond)

	// Third load should check file again
	if err := reader.Load(ctx); err != nil {
		t.Fatalf("Third Load() error = %v", err)
	}
}

func TestReader_GracefulErrorHandling(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	beadsDir := filepath.Join(tmpDir, ".beads")
	os.MkdirAll(beadsDir, 0755)

	// Create issues.jsonl with one valid and one invalid line
	issuesPath := filepath.Join(beadsDir, "issues.jsonl")
	content := `{"id":"test-1","title":"Valid","status":"open","priority":2,"issue_type":"task","created_at":"2026-01-07T12:00:00Z","updated_at":"2026-01-07T12:00:00Z"}
{"invalid json line}
{"id":"test-2","title":"Another Valid","status":"closed","priority":1,"issue_type":"bug","created_at":"2026-01-07T12:00:00Z","updated_at":"2026-01-07T12:00:00Z"}
`
	if err := os.WriteFile(issuesPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	reader := NewReader(tmpDir)

	// Should not error, just skip invalid lines
	if err := reader.Load(ctx); err != nil {
		t.Fatalf("Load() should handle errors gracefully, got: %v", err)
	}

	// Should have loaded the 2 valid issues
	if count := reader.Count(); count != 2 {
		t.Errorf("Count() = %v, want 2 (invalid lines should be skipped)", count)
	}
}

func TestReader_ParseFromString(t *testing.T) {
	ctx := context.Background()

	// Create temporary directory with test file
	tmpDir := t.TempDir()
	beadsDir := filepath.Join(tmpDir, ".beads")
	os.MkdirAll(beadsDir, 0755)

	issuesPath := filepath.Join(beadsDir, "issues.jsonl")
	content := strings.NewReader(`{"id":"test-1","title":"Test","status":"open","priority":1,"issue_type":"task","created_at":"2026-01-07T12:00:00Z","updated_at":"2026-01-07T12:00:00Z"}` + "\n")

	// Create reader and load from string
	file, err := os.Create(issuesPath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer file.Close()

	if _, err := file.ReadFrom(content); err != nil {
		t.Fatalf("Failed to write to test file: %v", err)
	}

	reader := NewReader(tmpDir)
	if err := reader.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	issue := reader.GetByID("test-1")
	if issue == nil {
		t.Fatal("GetByID(test-1) returned nil")
	}

	if issue.Priority.String() != "P1" {
		t.Errorf("Priority = %v, want P1", issue.Priority.String())
	}
}
