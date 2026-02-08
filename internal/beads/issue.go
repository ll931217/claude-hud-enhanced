package beads

import (
	"time"
)

// IssueStatus represents the status of an issue
type IssueStatus string

const (
	StatusOpen       IssueStatus = "open"
	StatusInProgress IssueStatus = "in_progress"
	StatusClosed     IssueStatus = "closed"
	StatusBlocked    IssueStatus = "blocked"
)

// IssueType represents the type of issue
type IssueType string

const (
	TypeTask    IssueType = "task"
	TypeBug     IssueType = "bug"
	TypeFeature IssueType = "feature"
	TypeEpic    IssueType = "epic"
)

// Priority represents the priority level (0-4 or P0-P4)
type Priority int

const (
	PriorityCritical Priority = 0
	PriorityHigh     Priority = 1
	PriorityNormal   Priority = 2
	PriorityLow      Priority = 3
	PriorityLowest   Priority = 4
)

// String returns the priority as a string (P0-P4)
func (p Priority) String() string {
	return "P" + string(rune('0'+p))
}

// Icon returns the status icon for an issue
func (s IssueStatus) Icon() string {
	switch s {
	case StatusOpen:
		return "✗"
	case StatusInProgress:
		return "◐"
	case StatusClosed:
		return "✓"
	case StatusBlocked:
		return "✖"
	default:
		return "?"
	}
}

// Dependency represents a dependency relationship
type Dependency struct {
	IssueID     string    `json:"issue_id"`
	DependsOnID string    `json:"depends_on_id"`
	Type        string    `json:"type,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by,omitempty"`
}

// Issue represents a single issue from beads
type Issue struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Description  string       `json:"description,omitempty"`
	Status       IssueStatus  `json:"status"`
	Priority     Priority     `json:"priority"`
	IssueType    IssueType    `json:"issue_type"`
	CreatedAt    time.Time    `json:"created_at"`
	CreatedBy    string       `json:"created_by,omitempty"`
	UpdatedAt    time.Time    `json:"updated_at"`
	Labels       []string     `json:"labels,omitempty"`
	Dependencies []Dependency `json:"dependencies,omitempty"`
}

// IsEpic returns true if the issue is an epic
func (i *Issue) IsEpic() bool {
	return i.IssueType == TypeEpic
}

// IsInProgress returns true if the issue is in progress
func (i *Issue) IsInProgress() bool {
	return i.Status == StatusInProgress
}

// IsClosed returns true if the issue is closed
func (i *Issue) IsClosed() bool {
	return i.Status == StatusClosed
}

// GetPriorityLabel returns the priority label (P0-P4)
func (i *Issue) GetPriorityLabel() string {
	return i.Priority.String()
}

// GetStatusWithIcon returns the status with icon
func (i *Issue) GetStatusWithIcon() string {
	return string(i.Status.Icon()) + " " + string(i.Status)
}
