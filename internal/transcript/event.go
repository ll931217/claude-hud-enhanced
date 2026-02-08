package transcript

import (
	"encoding/json"
	"time"
)

// EventType represents the type of transcript event
type EventType string

const (
	EventTypeError            EventType = "error"
	EventTypeWarning          EventType = "warning"
	EventTypeUserMessage      EventType = "user_message"
	EventTypeAssistantMessage EventType = "assistant_message"
	EventTypeToolUse          EventType = "tool_use"
	EventTypeToolResult       EventType = "tool_result"
	EventTypeAgentRun         EventType = "agent_run"
	EventTypeAgentMessage     EventType = "agent_message"
	EventTypeTaskStatus       EventType = "task_status"
	EventTypeTodo             EventType = "todo"
	EventTypeUnknown          EventType = "unknown"
)

// Event represents a single line in the Claude Code transcript JSONL
type Event struct {
	// Common fields
	Type      EventType `json:"type"`
	Timestamp string    `json:"timestamp,omitempty"`

	// Content fields
	Content string       `json:"content,omitempty"`
	Message *MessageInfo `json:"message,omitempty"`

	// Tool use fields
	ToolUse    *ToolInfo   `json:"tool_use,omitempty"`
	ToolResult *ToolResult `json:"tool_result,omitempty"`

	// Agent fields
	AgentRun     *AgentInfo        `json:"agent_run,omitempty"`
	AgentMessage *AgentMessageInfo `json:"agent_message,omitempty"`

	// Task/TODO fields
	TaskStatus *TaskStatusInfo `json:"task_status,omitempty"`
	Todo       *TodoInfo       `json:"todo,omitempty"`

	// Context window (from assistant messages)
	ContextWindow *ContextWindow `json:"context_window,omitempty"`

	// Raw bytes for unmarshaling
	Raw json.RawMessage `json:"-"`
}

// MessageInfo contains message metadata
type MessageInfo struct {
	Role         string `json:"role,omitempty"`
	ID           string `json:"id,omitempty"`
	Model        string `json:"model,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
	InputTokens  int    `json:"input_tokens,omitempty"`
	OutputTokens int    `json:"output_tokens,omitempty"`
}

// ToolInfo contains information about a tool being used
type ToolInfo struct {
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	Streaming bool            `json:"streaming,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	LastUsed  time.Time       `json:"-"` // Track last use time for recency sorting
	Target    string          `json:"-"` // Extracted target (file path, pattern, command)
	Status    string          `json:"-"` // running, completed, error
}

// ToolUsage represents aggregated tool usage statistics
type ToolUsage struct {
	Name     string
	Count    int
	LastUsed time.Time
	Target   string // Extracted target (file path, pattern, command)
	Status   string // running, completed, error
}

// ToolResult contains the result of a tool execution
type ToolResult struct {
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`
}

// AgentInfo contains information about a running agent
type AgentInfo struct {
	AgentID   string `json:"agent_id,omitempty"`
	AgentName string `json:"agent_name,omitempty"`
	Type      string `json:"type,omitempty"`
	Input     string `json:"input,omitempty"`
	Status    string `json:"status,omitempty"`
}

// AgentMessageInfo contains messages from agents
type AgentMessageInfo struct {
	AgentID string `json:"agent_id,omitempty"`
	Content string `json:"content,omitempty"`
	Status  string `json:"status,omitempty"`
}

// TaskStatusInfo contains task/TODO status information
type TaskStatusInfo struct {
	TodoID  string `json:"todo_id,omitempty"`
	Status  string `json:"status,omitempty"`
	Content string `json:"content,omitempty"`
}

// TodoInfo contains todo item information
type TodoInfo struct {
	ID       string `json:"id,omitempty"`
	Status   string `json:"status,omitempty"` // pending, in_progress, completed
	Content  string `json:"content,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

// ContextWindow contains context usage information
type ContextWindow struct {
	CurrentUsage      UsageInfo `json:"current_usage"`
	ContextWindowSize int       `json:"context_window_size"`
}

// UsageInfo contains token usage breakdown
type UsageInfo struct {
	InputTokens              int `json:"input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	OutputTokens             int `json:"output_tokens"`
}

// TotalInput returns the total input tokens including cache reads/writes
func (u *UsageInfo) TotalInput() int {
	return u.InputTokens + u.CacheCreationInputTokens + u.CacheReadInputTokens
}

// ClaudeCodeTranscriptLine represents a full line from the Claude Code transcript
type ClaudeCodeTranscriptLine struct {
	Type          string             `json:"type"` // user, assistant, summary, etc.
	Message       *ClaudeCodeMessage `json:"message,omitempty"`
	ToolUseResult *ToolResultExtra   `json:"toolUseResult,omitempty"`
	UUID          string             `json:"uuid,omitempty"`
	Timestamp     string             `json:"timestamp,omitempty"`
}

// ClaudeCodeMessage represents the full message structure from Claude Code
type ClaudeCodeMessage struct {
	Role       string         `json:"role"`
	Model      string         `json:"model,omitempty"`
	Content    []ContentBlock `json:"content"`
	Usage      *UsageInfo     `json:"usage,omitempty"`
	StopReason string         `json:"stop_reason,omitempty"`
}

// ContentBlock represents a content block within a message
type ContentBlock struct {
	Type       string          `json:"type"` // text, tool_use, tool_result, thinking, etc.
	ID         string          `json:"id,omitempty"`
	Name       string          `json:"name,omitempty"`        // for tool_use
	Input      json.RawMessage `json:"input,omitempty"`       // for tool_use
	ToolUseID  string          `json:"tool_use_id,omitempty"` // for tool_result
	Content    []ContentBlock  `json:"content,omitempty"`     // nested content array
	ContentStr string          `json:"-"`                     // Raw content string (extracted separately)
	Text       string          `json:"text,omitempty"`        // for text blocks
	IsError    bool            `json:"is_error,omitempty"`    // for tool_result error status
}

// ToolResultExtra contains extended tool result info
type ToolResultExtra struct {
	Status          string `json:"status"`
	AgentID         string `json:"agentId,omitempty"`
	TotalDurationMs int    `json:"totalDurationMs,omitempty"`
}

// ParseEventType attempts to determine event type from raw JSON
func ParseEventType(raw []byte) EventType {
	var base struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(raw, &base); err != nil {
		return EventTypeUnknown
	}

	switch EventType(base.Type) {
	case EventTypeError, EventTypeWarning, EventTypeUserMessage,
		EventTypeAssistantMessage, EventTypeToolUse, EventTypeToolResult,
		EventTypeAgentRun, EventTypeAgentMessage, EventTypeTaskStatus, EventTypeTodo:
		return EventType(base.Type)
	default:
		// Try to detect from other fields
		var detect struct {
			Message    map[string]interface{} `json:"message"`
			ToolUse    map[string]interface{} `json:"tool_use"`
			ToolResult map[string]interface{} `json:"tool_result"`
			AgentRun   map[string]interface{} `json:"agent_run"`
			TaskStatus map[string]interface{} `json:"task_status"`
			Todo       map[string]interface{} `json:"todo"`
		}

		if err := json.Unmarshal(raw, &detect); err == nil {
			if detect.Message != nil {
				if detect.Message["role"] != nil {
					if detect.Message["role"] == "assistant" {
						return EventTypeAssistantMessage
					}
					return EventTypeUserMessage
				}
			}
			if detect.ToolUse != nil {
				return EventTypeToolUse
			}
			if detect.ToolResult != nil {
				return EventTypeToolResult
			}
			if detect.AgentRun != nil {
				return EventTypeAgentRun
			}
			if detect.TaskStatus != nil {
				return EventTypeTaskStatus
			}
			if detect.Todo != nil {
				return EventTypeTodo
			}
		}

		return EventTypeUnknown
	}
}
