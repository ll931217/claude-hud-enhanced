package transcript

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestParseEventType(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected EventType
	}{
		{
			name:     "tool use event",
			json:     `{"type": "tool_use", "tool_use": {"name": "Read"}}`,
			expected: EventTypeToolUse,
		},
		{
			name:     "agent run event",
			json:     `{"type": "agent_run", "agent_run": {"agent_id": "test"}}`,
			expected: EventTypeAgentRun,
		},
		{
			name:     "assistant message",
			json:     `{"type": "assistant_message", "message": {"role": "assistant"}}`,
			expected: EventTypeAssistantMessage,
		},
		{
			name:     "unknown type",
			json:     `{"type": "unknown_type"}`,
			expected: EventTypeUnknown,
		},
		{
			name:     "detect tool use without type field",
			json:     `{"tool_use": {"name": "Write"}}`,
			expected: EventTypeToolUse,
		},
		{
			name:     "detect agent run without type field",
			json:     `{"agent_run": {"agent_id": "test123"}}`,
			expected: EventTypeAgentRun,
		},
		{
			name:     "detect assistant message without type field",
			json:     `{"message": {"role": "assistant"}}`,
			expected: EventTypeAssistantMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseEventType([]byte(tt.json))
			if result != tt.expected {
				t.Errorf("ParseEventType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUsageInfo_TotalInput(t *testing.T) {
	tests := []struct {
		name string
		usage UsageInfo
		want int
	}{
		{
			name: "all tokens",
			usage: UsageInfo{
				InputTokens:             1000,
				CacheCreationInputTokens: 500,
				CacheReadInputTokens:     250,
			},
			want: 1750,
		},
		{
			name: "only input tokens",
			usage: UsageInfo{
				InputTokens: 1000,
			},
			want: 1000,
		},
		{
			name: "zero tokens",
			usage: UsageInfo{},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.usage.TotalInput(); got != tt.want {
				t.Errorf("TotalInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_ParseFromReader(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		input          string
		expectLines    int
		expectErrors   bool
		expectToolUse  bool
		expectAgentRun bool
	}{
		{
			name: "valid tool use",
			input: `{"type": "tool_use", "tool_name": "Read", "timestamp": "2026-01-11T03:26:59.508Z"}` + "\n",
			expectLines:   1,
			expectErrors:  false,
			expectToolUse: true,
		},
		{
			name: "valid agent run",
			input: `{"type": "agent_run", "agent_run": {"agent_id": "agent1", "agent_name": "test-agent"}}` + "\n",
			expectLines:    1,
			expectErrors:   false,
			expectAgentRun: true,
		},
		{
			name: "multiple events",
			input: `{"type": "tool_use", "tool_name": "Read", "timestamp": "2026-01-11T03:26:59.508Z"}` + "\n" +
				`{"type": "agent_run", "agent_run": {"agent_id": "a1"}}` + "\n" +
				`{"type": "tool_use", "tool_name": "Write", "timestamp": "2026-01-11T03:27:00.123Z"}` + "\n",
			expectLines:    3,
			expectErrors:   false,
			expectToolUse:  true,
			expectAgentRun: true,
		},
		{
			name: "invalid json - graceful handling",
			input: `{"type": "tool_use", "tool_name": {invalid}}` + "\n",
			expectLines:  1,
			expectErrors: true,
		},
		{
			name:           "empty lines",
			input:          "\n\n\n",
			expectLines:    0,
			expectErrors:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser("test.jsonl")
			r := strings.NewReader(tt.input)

			err := p.ParseFromReader(ctx, r)

			// Check for unexpected errors
			if err != nil && tt.expectErrors {
				// Expected to have parse errors but function should still succeed
			} else if err != nil && !tt.expectErrors {
				t.Errorf("ParseFromReader() unexpected error = %v", err)
			}

			state := p.GetState()
			if state.LinesParsed != tt.expectLines {
				t.Errorf("LinesParsed = %v, want %v", state.LinesParsed, tt.expectLines)
			}

			if tt.expectErrors && state.ErrorsEncountered == 0 {
				t.Errorf("Expected errors but got none")
			}

			if tt.expectToolUse && p.ActiveToolCount() == 0 {
				t.Errorf("Expected tool activity but got none")
			}

			if tt.expectAgentRun && p.ActiveAgentCount() == 0 {
				t.Errorf("Expected agent activity but got none")
			}
		})
	}
}

func TestParser_GetLatestEvent(t *testing.T) {
	ctx := context.Background()
	p := NewParser("test.jsonl")

	input := `{"type": "tool_use", "tool_name": "Read", "timestamp": "2026-01-11T03:26:59.508Z"}` + "\n" +
		`{"type": "tool_use", "tool_name": "Write", "timestamp": "2026-01-11T03:27:00.123Z"}` + "\n"

	r := strings.NewReader(input)
	if err := p.ParseFromReader(ctx, r); err != nil {
		t.Fatalf("ParseFromReader() error = %v", err)
	}

	event := p.GetLatestEvent(EventTypeToolUse)
	if event == nil {
		t.Fatal("GetLatestEvent() returned nil")
	}

	if event.ToolUse == nil {
		t.Fatal("GetLatestEvent().ToolUse is nil")
	}

	if event.ToolUse.Name != "Write" {
		t.Errorf("Got tool name %v, want Write", event.ToolUse.Name)
	}
}

func TestParser_ContextWindow(t *testing.T) {
	ctx := context.Background()
	p := NewParser("test.jsonl")

	input := `{"type": "assistant_message", "message": {"role": "assistant"}, "context_window": {"current_usage": {"input_tokens": 50000, "cache_creation_input_tokens": 5000, "cache_read_input_tokens": 0, "output_tokens": 1000}, "context_window_size": 200000}}` + "\n"

	r := strings.NewReader(input)
	if err := p.ParseFromReader(ctx, r); err != nil {
		t.Fatalf("ParseFromReader() error = %v", err)
	}

	if !p.HasContextWindow() {
		t.Error("Expected context window to be available")
	}

	percentage := p.GetContextPercentage()
	// Calculate expected with AUTOCOMPACT_BUFFER (128000 tokens)
	// (55000 + 128000) * 100 / 200000 = 91.5%
	expected := (55000 + AUTOCOMPACT_BUFFER) * 100 / 200000
	if percentage != expected {
		t.Errorf("GetContextPercentage() = %v, want %v", percentage, expected)
	}

	cw := p.GetContextWindow()
	if cw == nil {
		t.Fatal("GetContextWindow() returned nil")
	}

	if cw.ContextWindowSize != 200000 {
		t.Errorf("ContextWindowSize = %v, want 200000", cw.ContextWindowSize)
	}
}

func TestParser_SessionTracking(t *testing.T) {
	ctx := context.Background()
	p := NewParser("test.jsonl")

	input := `{"type": "assistant_message", "timestamp": "2026-01-07T12:00:00Z", "message": {"role": "assistant"}}` + "\n"

	r := strings.NewReader(input)
	if err := p.ParseFromReader(ctx, r); err != nil {
		t.Fatalf("ParseFromReader() error = %v", err)
	}

	start := p.GetSessionStart()
	if start.IsZero() {
		t.Error("Expected session start time to be set")
	}
}

func TestParser_TodoTracking(t *testing.T) {
	ctx := context.Background()
	p := NewParser("test.jsonl")

	input := `{"type": "todo", "todo": {"id": "1", "status": "pending", "content": "Task 1"}}` + "\n" +
		`{"type": "todo", "todo": {"id": "2", "status": "in_progress", "content": "Task 2 in progress"}}` + "\n" +
		`{"type": "todo", "todo": {"id": "3", "status": "completed", "content": "Task 3 done"}}` + "\n"

	r := strings.NewReader(input)
	if err := p.ParseFromReader(ctx, r); err != nil {
		t.Fatalf("ParseFromReader() error = %v", err)
	}

	// Test GetTodoCount
	total, completed := p.GetTodoCount()
	if total != 3 {
		t.Errorf("GetTodoCount() total = %v, want 3", total)
	}
	if completed != 1 {
		t.Errorf("GetTodoCount() completed = %v, want 1", completed)
	}

	// Test GetCurrentTodo
	current := p.GetCurrentTodo()
	if current == nil {
		t.Fatal("GetCurrentTodo() returned nil, expected todo 2")
	}
	if current.ID != "2" {
		t.Errorf("GetCurrentTodo().ID = %v, want 2", current.ID)
	}
	if current.Status != "in_progress" {
		t.Errorf("GetCurrentTodo().Status = %v, want in_progress", current.Status)
	}

	// Test GetTodos
	todos := p.GetTodos()
	if len(todos) != 3 {
		t.Errorf("GetTodos() length = %v, want 3", len(todos))
	}
	if todos["1"].Content != "Task 1" {
		t.Errorf("GetTodos()[\"1\"].Content = %v, want 'Task 1'", todos["1"].Content)
	}
}

func TestParser_CalculateCost(t *testing.T) {
	ctx := context.Background()
	p := NewParser("test.jsonl")

	// Test with Opus model
	input := `{"type": "assistant_message", "timestamp": "2026-01-07T12:00:00Z", "message": {"role": "assistant", "model": "claude-opus-4-5-20251101", "input_tokens": 1000000, "output_tokens": 500000}}` + "\n"

	r := strings.NewReader(input)
	if err := p.ParseFromReader(ctx, r); err != nil {
		t.Fatalf("ParseFromReader() error = %v", err)
	}

	cost := p.CalculateCost()
	// Opus: $15/M input, $75/M output
	// 1M input * $15/M = $15
	// 500K output * $75/M = $37.50
	// Total = $52.50
	expected := 52.50
	if cost < expected-0.01 || cost > expected+0.01 {
		t.Errorf("CalculateCost() = %v, want %v", cost, expected)
	}
}

func TestParser_CalculateCost_ModelDetection(t *testing.T) {
	tests := []struct {
		name         string
		model        string
		inputTokens  int
		outputTokens int
		expectedCost float64
	}{
		{
			name:         "Opus",
			model:        "claude-opus-4-5-20251101",
			inputTokens:  1000000,
			outputTokens: 1000000,
			expectedCost: 90.0, // $15 + $75
		},
		{
			name:         "Sonnet",
			model:        "claude-sonnet-4-5-20251101",
			inputTokens:  1000000,
			outputTokens: 1000000,
			expectedCost: 18.0, // $3 + $15
		},
		{
			name:         "Haiku",
			model:        "claude-haiku-4-5-20251101",
			inputTokens:  1000000,
			outputTokens: 1000000,
			expectedCost: 1.50, // $0.25 + $1.25
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			p := NewParser("test.jsonl")

			input := `{"type": "assistant_message", "timestamp": "2026-01-07T12:00:00Z", "message": {"role": "assistant", "model": "` + tt.model + `", "input_tokens": ` + fmt.Sprintf("%d", tt.inputTokens) + `, "output_tokens": ` + fmt.Sprintf("%d", tt.outputTokens) + `}}` + "\n"

			r := strings.NewReader(input)
			if err := p.ParseFromReader(ctx, r); err != nil {
				t.Fatalf("ParseFromReader() error = %v", err)
			}

			cost := p.CalculateCost()
			if cost < tt.expectedCost-0.01 || cost > tt.expectedCost+0.01 {
				t.Errorf("CalculateCost() = %v, want %v", cost, tt.expectedCost)
			}
		})
	}
}

func TestParser_GetDuration(t *testing.T) {
	ctx := context.Background()
	p := NewParser("test.jsonl")

	// Test with session start timestamp
	input := `{"type": "assistant_message", "timestamp": "2026-01-07T12:00:00Z", "message": {"role": "assistant"}}` + "\n"

	r := strings.NewReader(input)
	if err := p.ParseFromReader(ctx, r); err != nil {
		t.Fatalf("ParseFromReader() error = %v", err)
	}

	duration := p.GetDuration()
	if duration == "" {
		t.Error("GetDuration() returned empty string")
	}
	if duration == "0s" {
		t.Error("GetDuration() returned '0s', expected some duration")
	}
}
