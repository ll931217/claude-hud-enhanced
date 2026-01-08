package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDetector_NewDetector(t *testing.T) {
	d := NewDetector(".")
	if d == nil {
		t.Fatal("NewDetector() returned nil")
	}
}

func TestDetector_Detect_NoGitRepo(t *testing.T) {
	tmpDir := t.TempDir()
	d := NewDetector(tmpDir)

	ctx := context.Background()
	status, err := d.Detect(ctx)

	// Should return error for non-git directory
	if err == nil {
		t.Error("Expected error for non-git directory, got nil")
	}
	if status != nil {
		t.Error("Expected nil status for non-git directory")
	}
}

func TestDetector_Detect_TestRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize a git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Cannot run git: %v", err)
	}

	// Configure git
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Cannot configure git: %v", err)
	}
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Cannot configure git: %v", err)
	}

	// Create initial commit
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Cannot add to git: %v", err)
	}
	cmd = exec.Command("git", "commit", "-m", "test")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skipf("Cannot commit: %v", err)
	}

	d := NewDetector(tmpDir)
	ctx := context.Background()

	status, err := d.Detect(ctx)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if status == nil {
		t.Fatal("Detect() returned nil status")
	}

	// Check branch name
	if status.Branch != "main" && status.Branch != "master" {
		t.Errorf("Expected branch 'main' or 'master', got %s", status.Branch)
	}

	// Check not dirty
	if status.Dirty {
		t.Error("Expected Dirty=false for clean repo")
	}
}

func TestStatus_FormatStatus(t *testing.T) {
	tests := []struct {
		name   string
		status *Status
		want   string
	}{
		{
			name:   "empty",
			status: &Status{},
			want:   "",
		},
		{
			name: "branch only",
			status: &Status{
				Branch: "main",
			},
			want: "🌿 main",
		},
		{
			name: "dirty",
			status: &Status{
				Branch:   "main",
				Dirty:    true,
				Modified: 1,
			},
			want: "🌿 main * 1",
		},
		{
			name: "ahead behind",
			status: &Status{
				Branch: "main",
				Ahead:  2,
				Behind: 1,
			},
			want: "🌿 main ↑2↓1",
		},
		{
			name: "worktree",
			status: &Status{
				Branch:       "main",
				IsWorktree:   true,
				WorktreeName: "feature-branch",
			},
			want: "🌿 main [feature-branch]",
		},
		{
			name: "stashed with changes",
			status: &Status{
				Branch:   "main",
				Stashed:  2,
				Modified: 1,
				Dirty:    true,
			},
			want: "🌿 main * 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.FormatStatus()
			if got != tt.want {
				t.Errorf("FormatStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}
