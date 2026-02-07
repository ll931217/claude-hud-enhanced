package git

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ll931217/claude-hud-enhanced/internal/errors"
)

// Status represents the git status of a repository
type Status struct {
	Branch         string
	IsWorktree     bool
	WorktreeName   string
	Dirty          bool
	Modified       int
	Added          int
	Deleted        int
	Untracked      int
	Ahead          int
	Behind         int
	Stashed        int
}

// Detector handles git status and worktree detection
type Detector struct {
	mu        sync.RWMutex
	repoPath  string
	lastCheck int64
	status    *Status
}

// NewDetector creates a new git detector for the given path
func NewDetector(repoPath string) *Detector {
	// Resolve to absolute path
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		absPath = repoPath
	}

	return &Detector{
		repoPath: absPath,
	}
}

// Detect detects git status and worktree information
func (d *Detector) Detect(ctx context.Context) (*Status, error) {
	return errors.SafeExecute(func() (*Status, error) {
		// Check if we're in a git repository
		gitRoot, err := d.getGitRoot(ctx)
		if err != nil {
			return nil, fmt.Errorf("not a git repository: %w", err)
		}

		status := &Status{
			IsWorktree: d.isWorktree(gitRoot),
		}

		// Get branch name
		if branch, err := d.getCurrentBranch(ctx); err == nil {
			status.Branch = branch
		}

		// Get worktree info if applicable
		if status.IsWorktree {
			if name, err := d.getWorktreeName(ctx); err == nil {
				status.WorktreeName = name
			}
		}

		// Get status counts
		if err := d.getStatusCounts(ctx, status); err == nil {
			status.Dirty = status.Modified > 0 || status.Added > 0 ||
				status.Deleted > 0 || status.Untracked > 0
		}

		// Get ahead/behind
		if ahead, behind, err := d.getAheadBehind(ctx); err == nil {
			status.Ahead = ahead
			status.Behind = behind
		}

		// Get stash count
		if stashed, err := d.getStashCount(ctx); err == nil {
			status.Stashed = stashed
		}

		d.mu.Lock()
		d.status = status
		d.mu.Unlock()

		return status, nil
	})
}

// getGitRoot returns the git repository root directory
func (d *Detector) getGitRoot(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	cmd.Dir = d.repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// isWorktree checks if the current directory is a git worktree
func (d *Detector) isWorktree(gitRoot string) bool {
	// Check if .git/commondir exists (indicator of worktree)
	commondirPath := filepath.Join(gitRoot, ".git", "commondir")
	info, err := os.Stat(commondirPath)
	return err == nil && !info.IsDir()
}

// getCurrentBranch returns the current branch name
func (d *Detector) getCurrentBranch(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = d.repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	branch := strings.TrimSpace(string(output))
	if branch == "HEAD" {
		return "(detached)", nil
	}

	return branch, nil
}

// getWorktreeName derives the worktree name from branch or path
func (d *Detector) getWorktreeName(ctx context.Context) (string, error) {
	// Try to get worktree list
	cmd := exec.CommandContext(ctx, "git", "worktree", "list", "--porcelain")
	cmd.Dir = d.repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse worktree list to find current worktree
	absPath, _ := filepath.Abs(d.repoPath)

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	var worktreeBranch string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "worktree ") {
			worktreePath := strings.TrimPrefix(line, "worktree ")
			if filepath.Clean(worktreePath) == filepath.Clean(absPath) {
				// This is our worktree
			}
		} else if strings.HasPrefix(line, "branch ") {
			worktreeBranch = strings.TrimPrefix(line, "branch ")
			worktreeBranch = strings.TrimPrefix(worktreeBranch, "refs/heads/")
		}
	}

	// Use branch name or derive from path
	if worktreeBranch != "" {
		return worktreeBranch, nil
	}

	// Fallback: derive from directory name
	return filepath.Base(d.repoPath), nil
}

// getStatusCounts gets the count of changed files
func (d *Detector) getStatusCounts(ctx context.Context, status *Status) error {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = d.repoPath
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 2 {
			continue
		}

		// Git status format: XY filename
		// X = index status, Y = worktree status
		index := line[0]
		worktree := line[1]

		switch index {
		case 'M':
			status.Modified++
		case 'A':
			status.Added++
		case 'D':
			status.Deleted++
		}

		switch worktree {
		case 'M':
			status.Modified++
		case 'A':
			status.Added++
		case 'D':
			status.Deleted++
		case '?':
			status.Untracked++
		}
	}

	return nil
}

// getAheadBehind gets the ahead/behind count for the current branch
func (d *Detector) getAheadBehind(ctx context.Context) (ahead, behind int, err error) {
	cmd := exec.CommandContext(ctx, "git", "rev-list", "--left-right", "--count", "HEAD...@{u}")
	cmd.Dir = d.repoPath
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	// Output format: "ahead\tbehind"
	parts := strings.Fields(string(output))
	if len(parts) != 2 {
		return 0, 0, nil
	}

	fmt.Sscanf(parts[0], "%d", &ahead)
	fmt.Sscanf(parts[1], "%d", &behind)

	return ahead, behind, nil
}

// getStashCount returns the number of stashed changes
func (d *Detector) getStashCount(ctx context.Context) (int, error) {
	cmd := exec.CommandContext(ctx, "git", "stash", "list")
	cmd.Dir = d.repoPath
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	count := 0
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "stash@{") {
			count++
		}
	}

	return count, nil
}

// GetStatus returns the cached status or detects fresh status
func (d *Detector) GetStatus(ctx context.Context) (*Status, error) {
	d.mu.RLock()
	status := d.status
	d.mu.RUnlock()

	if status != nil {
		return status, nil
	}

	return d.Detect(ctx)
}

// GetBranchShort returns a shortened branch name
func (s *Status) GetBranchShort() string {
	if s.Branch == "" {
		return ""
	}

	// Shorten common prefixes
	branch := s.Branch
	branch = strings.TrimPrefix(branch, "feature/")
	branch = strings.TrimPrefix(branch, "fix/")
	branch = strings.TrimPrefix(branch, "bugfix/")
	branch = strings.TrimPrefix(branch, "hotfix/")
	branch = strings.TrimPrefix(branch, "release/")
	branch = strings.TrimPrefix(branch, "develop")

	return branch
}

// FormatStatus returns a formatted status string
func (s *Status) FormatStatus() string {
	if s.Branch == "" {
		return ""
	}

	var parts []string

	// Branch name
	parts = append(parts, "🌿", s.GetBranchShort())

	// Worktree indicator
	if s.IsWorktree && s.WorktreeName != "" {
		parts = append(parts, fmt.Sprintf("[%s]", s.WorktreeName))
	}

	// Dirty indicator (plus-minus symbol, universally understood as "changed")
	if s.Dirty {
		parts = append(parts, "±")
	}

	// Changes count (compact format)
	totalChanges := s.Modified + s.Added + s.Deleted + s.Untracked
	if totalChanges > 0 {
		parts = append(parts, fmt.Sprintf("%d", totalChanges))
	}

	// Ahead/Behind (using more visible directional arrows)
	if s.Ahead > 0 || s.Behind > 0 {
		if s.Ahead > 0 && s.Behind > 0 {
			// Diverged branches: use up-down arrow to clearly indicate divergence
			parts = append(parts, fmt.Sprintf("⇅ %d|%d", s.Ahead, s.Behind))
		} else if s.Ahead > 0 {
			parts = append(parts, fmt.Sprintf("⬆ %d", s.Ahead))
		} else if s.Behind > 0 {
			parts = append(parts, fmt.Sprintf("⬇ %d", s.Behind))
		}
	}

	return strings.Join(parts, " ")
}
