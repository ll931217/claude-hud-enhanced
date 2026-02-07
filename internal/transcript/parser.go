package transcript

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/errors"
)

// Constants for context window calculations
const (
	AUTOCOMPACT_BUFFER = 128000 // Tokens reserved for auto-compact
)

// Parser handles parsing Claude Code transcript JSONL files
type Parser struct {
	mu               sync.RWMutex
	state            *ParserState
	transcriptPath   string
	lastModified     time.Time
	lastFileSize     int64
	latestEvents     map[EventType]*Event
	toolActivity     map[string]*ToolInfo
	agentActivity    map[string]*AgentInfo
	contextWindow    *ContextWindow
	sessionStart     time.Time
	sessionEnd       time.Time
	totalInputTokens int
	totalOutputTokens int
	todos            map[string]*TodoInfo
}

// ParserState tracks the current state of the parser
type ParserState struct {
	LinesParsed     int
	ErrorsEncountered int
	LastParseTime   time.Time
}

// NewParser creates a new transcript parser
func NewParser(transcriptPath string) *Parser {
	return &Parser{
		transcriptPath: transcriptPath,
		latestEvents:   make(map[EventType]*Event),
		toolActivity:   make(map[string]*ToolInfo),
		agentActivity:  make(map[string]*AgentInfo),
		todos:          make(map[string]*TodoInfo),
		state:          &ParserState{},
	}
}

// Parse reads and parses the transcript file
// Uses streaming to avoid loading the entire file into memory
func (p *Parser) Parse(ctx context.Context) error {
	return errors.SafeCall(func() error {
		// Check if file exists
		if _, err := os.Stat(p.transcriptPath); os.IsNotExist(err) {
			return fmt.Errorf("transcript file not found: %s", p.transcriptPath)
		}

		// Get file info for change detection
		info, err := os.Stat(p.transcriptPath)
		if err != nil {
			return fmt.Errorf("failed to stat transcript: %w", err)
		}

		// Check if file has changed since last parse
		p.mu.Lock()
		modified := info.ModTime().After(p.lastModified) || info.Size() != p.lastFileSize
		p.lastModified = info.ModTime()
		p.lastFileSize = info.Size()
		p.mu.Unlock()

		if !modified && p.state.LinesParsed > 0 {
			// File hasn't changed, no need to reparse
			return nil
		}

		// Open the file
		file, err := os.Open(p.transcriptPath)
		if err != nil {
			return fmt.Errorf("failed to open transcript: %w", err)
		}
		defer file.Close()

		// Reset state for fresh parse
		p.resetState()

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

			// Parse the line
			if err := p.parseLine(line); err != nil {
				// Log error but continue parsing
				p.state.ErrorsEncountered++
				if p.state.ErrorsEncountered <= 10 {
					// Only log first 10 errors to avoid spam
					errors.Warn("transcript.parser", "line %d: %v", lineNum, err)
				}
			}

			p.state.LinesParsed++
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scanner error: %w", err)
		}

		p.state.LastParseTime = time.Now()
		return nil
	})
}

// parseLine parses a single JSONL line
func (p *Parser) parseLine(line []byte) error {
	defer errors.RecoverPanic("transcript.parseLine")

	var event Event
	event.Raw = line

	// First, detect event type
	eventType := ParseEventType(line)
	event.Type = eventType

	// Parse based on event type
	switch eventType {
	case EventTypeAssistantMessage, EventTypeUserMessage:
		var msg struct {
			Type      string      `json:"type"`
			Timestamp string      `json:"timestamp,omitempty"`
			Message   MessageInfo `json:"message"`
			ContextWindow *ContextWindow `json:"context_window,omitempty"`
		}
		if err := json.Unmarshal(line, &msg); err != nil {
			return err
		}
		event.Timestamp = msg.Timestamp
		event.Message = &msg.Message
		event.ContextWindow = msg.ContextWindow

		// Track context window from assistant messages
		if msg.ContextWindow != nil {
			p.contextWindow = msg.ContextWindow
		}

		// Track session start time from first message
		if p.sessionStart.IsZero() && msg.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339Nano, msg.Timestamp); err == nil {
				p.sessionStart = t
			}
		}

		// Track token usage
		if msg.Message.InputTokens > 0 {
			p.totalInputTokens += msg.Message.InputTokens
		}
		if msg.Message.OutputTokens > 0 {
			p.totalOutputTokens += msg.Message.OutputTokens
		}

	case EventTypeToolUse:
		var tool struct {
			Type      string   `json:"type"`
			Timestamp string   `json:"timestamp,omitempty"`
			ToolUse   ToolInfo `json:"tool_use"`
		}
		if err := json.Unmarshal(line, &tool); err != nil {
			return err
		}
		event.Timestamp = tool.Timestamp
		event.ToolUse = &tool.ToolUse

		// Track tool activity with timestamp
		if tool.ToolUse.ToolUseID != "" {
			// Parse timestamp for recency tracking
			if tool.Timestamp != "" {
				if t, err := time.Parse(time.RFC3339Nano, tool.Timestamp); err == nil {
					tool.ToolUse.LastUsed = t
				}
			}
			p.toolActivity[tool.ToolUse.ToolUseID] = &tool.ToolUse
		}

	case EventTypeToolResult:
		var result struct {
			Type        string     `json:"type"`
			Timestamp   string     `json:"timestamp,omitempty"`
			ToolResult  ToolResult `json:"tool_result"`
		}
		if err := json.Unmarshal(line, &result); err != nil {
			return err
		}
		event.Timestamp = result.Timestamp
		event.ToolResult = &result.ToolResult

	case EventTypeAgentRun:
		var agent struct {
			Type      string    `json:"type"`
			Timestamp string    `json:"timestamp,omitempty"`
			AgentRun  AgentInfo `json:"agent_run"`
		}
		if err := json.Unmarshal(line, &agent); err != nil {
			return err
		}
		event.Timestamp = agent.Timestamp
		event.AgentRun = &agent.AgentRun

		// Track agent activity
		if agent.AgentRun.AgentID != "" {
			p.agentActivity[agent.AgentRun.AgentID] = &agent.AgentRun
		}

	case EventTypeAgentMessage:
		var agentMsg struct {
			Type          string          `json:"type"`
			Timestamp     string          `json:"timestamp,omitempty"`
			AgentMessage  AgentMessageInfo `json:"agent_message"`
		}
		if err := json.Unmarshal(line, &agentMsg); err != nil {
			return err
		}
		event.Timestamp = agentMsg.Timestamp
		event.AgentMessage = &agentMsg.AgentMessage

	case EventTypeTaskStatus:
		var task struct {
			Type       string         `json:"type"`
			Timestamp  string         `json:"timestamp,omitempty"`
			TaskStatus TaskStatusInfo `json:"task_status"`
		}
		if err := json.Unmarshal(line, &task); err != nil {
			return err
		}
		event.Timestamp = task.Timestamp
		event.TaskStatus = &task.TaskStatus

	case EventTypeTodo:
		var todo struct {
			Type      string    `json:"type"`
			Timestamp string    `json:"timestamp,omitempty"`
			Todo      TodoInfo `json:"todo"`
		}
		if err := json.Unmarshal(line, &todo); err != nil {
			return err
		}
		event.Timestamp = todo.Timestamp
		event.Todo = &todo.Todo

	default:
		// For unknown types, just store the raw data
		var base struct {
			Type      string `json:"type"`
			Timestamp string `json:"timestamp,omitempty"`
		}
		if err := json.Unmarshal(line, &base); err != nil {
			return err
		}
		event.Timestamp = base.Timestamp
	}

	// Update latest event for this type
	p.mu.Lock()
	p.latestEvents[eventType] = &event

	// Track todos
	if event.Todo != nil && event.Todo.ID != "" {
		p.todos[event.Todo.ID] = event.Todo
	}

	p.mu.Unlock()

	return nil
}

// resetState clears parser state for a fresh parse
func (p *Parser) resetState() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.state = &ParserState{}
	p.latestEvents = make(map[EventType]*Event)
	p.toolActivity = make(map[string]*ToolInfo)
	p.agentActivity = make(map[string]*AgentInfo)
	p.todos = make(map[string]*TodoInfo)
	// Keep session start if we already found it
}

// GetState returns the current parser state
func (p *Parser) GetState() *ParserState {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.state
}

// GetLatestEvent returns the most recent event of the given type
func (p *Parser) GetLatestEvent(eventType EventType) *Event {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.latestEvents[eventType]
}

// GetToolActivity returns active tool usage
func (p *Parser) GetToolActivity() map[string]*ToolInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Return a copy
	result := make(map[string]*ToolInfo)
	for k, v := range p.toolActivity {
		result[k] = v
	}
	return result
}

// GetAgentActivity returns active agent runs
func (p *Parser) GetAgentActivity() map[string]*AgentInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Return a copy
	result := make(map[string]*AgentInfo)
	for k, v := range p.agentActivity {
		result[k] = v
	}
	return result
}

// GetContextWindow returns the latest context window information
func (p *Parser) GetContextWindow() *ContextWindow {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.contextWindow == nil {
		return nil
	}

	// Return a copy
	cw := *p.contextWindow
	return &cw
}

// GetSessionStart returns the session start time
func (p *Parser) GetSessionStart() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.sessionStart
}

// GetTotalTokens returns total token usage
func (p *Parser) GetTotalTokens() (input, output int) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.totalInputTokens, p.totalOutputTokens
}

// HasContextWindow returns true if context window info is available
func (p *Parser) HasContextWindow() bool {
	return p.GetContextWindow() != nil
}

// GetContextPercentage returns context usage as a percentage
// Includes auto-compact buffer in calculation for accuracy
func (p *Parser) GetContextPercentage() int {
	cw := p.GetContextWindow()
	if cw == nil || cw.ContextWindowSize == 0 {
		return 0
	}

	totalTokens := cw.CurrentUsage.TotalInput()

	// Include auto-compact buffer in calculation
	percentage := (totalTokens + AUTOCOMPACT_BUFFER) * 100 / cw.ContextWindowSize
	if percentage > 100 {
		return 100
	}
	if percentage < 0 {
		return 0
	}
	return percentage
}

// ActiveToolCount returns the number of active tools
func (p *Parser) ActiveToolCount() int {
	tools := p.GetToolActivity()
	return len(tools)
}

// ActiveAgentCount returns the number of active agents
func (p *Parser) ActiveAgentCount() int {
	agents := p.GetAgentActivity()
	return len(agents)
}

// ParseFromReader parses from an io.Reader (useful for testing)
func (p *Parser) ParseFromReader(ctx context.Context, r io.Reader) error {
	return errors.SafeCall(func() error {
		p.resetState()

		scanner := bufio.NewScanner(r)
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

			if err := p.parseLine(line); err != nil {
				p.state.ErrorsEncountered++
			}

			p.state.LinesParsed++
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scanner error: %w", err)
		}

		p.state.LastParseTime = time.Now()
		return nil
	})
}

// GetTodos returns all tracked todos
func (p *Parser) GetTodos() map[string]*TodoInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]*TodoInfo)
	for k, v := range p.todos {
		result[k] = v
	}
	return result
}

// GetTodoCount returns the total and completed todo counts
func (p *Parser) GetTodoCount() (total, completed int) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	total = len(p.todos)
	for _, todo := range p.todos {
		if todo.Status == "completed" {
			completed++
		}
	}
	return total, completed
}

// GetCurrentTodo returns the current in-progress todo
func (p *Parser) GetCurrentTodo() *TodoInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, todo := range p.todos {
		if todo.Status == "in_progress" {
			return todo
		}
	}
	return nil
}

// CalculateCost estimates the token cost based on model pricing
func (p *Parser) CalculateCost() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Pricing per million tokens (USD)
	// These are approximate prices for Claude models
	const (
		opusInputPrice  = 15.0
		opusOutputPrice = 75.0
		sonnetInputPrice  = 3.0
		sonnetOutputPrice = 15.0
		haikuInputPrice  = 0.25
		haikuOutputPrice = 1.25
	)

	// Get model from latest assistant message
	inputPrice, outputPrice := opusInputPrice, opusOutputPrice // default to Opus
	if event := p.latestEvents[EventTypeAssistantMessage]; event != nil && event.Message != nil {
		model := event.Message.Model
		switch {
		case strings.Contains(model, "opus"):
			inputPrice, outputPrice = opusInputPrice, opusOutputPrice
		case strings.Contains(model, "sonnet"):
			inputPrice, outputPrice = sonnetInputPrice, sonnetOutputPrice
		case strings.Contains(model, "haiku"):
			inputPrice, outputPrice = haikuInputPrice, haikuOutputPrice
		}
	}

	inputCost := (float64(p.totalInputTokens) / 1_000_000) * inputPrice
	outputCost := (float64(p.totalOutputTokens) / 1_000_000) * outputPrice

	return inputCost + outputCost
}

// GetDuration returns the formatted session duration
func (p *Parser) GetDuration() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.sessionStart.IsZero() {
		return "0s"
	}

	duration := time.Since(p.sessionStart)

	// Format duration in human-readable format
	switch {
	case duration < time.Minute:
		return duration.String()
	case duration < time.Hour:
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		mins := int(duration.Minutes()) % 60
		if mins > 0 {
			return fmt.Sprintf("%dh%dm", hours, mins)
		}
		return fmt.Sprintf("%dh", hours)
	default:
		days := int(duration.Hours() / 24)
		hours := int(duration.Hours()) % 24
		if hours > 0 {
			return fmt.Sprintf("%dd%dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}
}

// GetToolsByRecency returns tools aggregated by name, sorted by most recently used
func (p *Parser) GetToolsByRecency(maxTools int) []ToolUsage {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Aggregate tools by name
	toolMap := make(map[string]*ToolUsage)

	for _, tool := range p.toolActivity {
		if tool.Name == "" {
			continue
		}

		if existing, ok := toolMap[tool.Name]; ok {
			existing.Count++
			if tool.LastUsed.After(existing.LastUsed) {
				existing.LastUsed = tool.LastUsed
			}
		} else {
			toolMap[tool.Name] = &ToolUsage{
				Name:     tool.Name,
				Count:    1,
				LastUsed: tool.LastUsed,
			}
		}
	}

	// Convert to slice and sort by recency
	result := make([]ToolUsage, 0, len(toolMap))
	for _, usage := range toolMap {
		result = append(result, *usage)
	}

	// Sort by last used time (most recent first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].LastUsed.After(result[j].LastUsed)
	})

	// Limit to maxTools
	if maxTools > 0 && len(result) > maxTools {
		result = result[:maxTools]
	}

	return result
}
