package transcript

import (
	"context"
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
			input: `{"type": "tool_use", "tool_use": {"name": "Read", "tool_use_id": "test123"}}` + "\n",
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
			input: `{"type": "tool_use", "tool_use": {"name": "Read", "tool_use_id": "t1"}}` + "\n" +
				`{"type": "agent_run", "agent_run": {"agent_id": "a1"}}` + "\n" +
				`{"type": "tool_use", "tool_use": {"name": "Write", "tool_use_id": "t2"}}` + "\n",
			expectLines:    3,
			expectErrors:   false,
			expectToolUse:  true,
			expectAgentRun: true,
		},
		{
			name: "invalid json - graceful handling",
			input: `{"type": "tool_use", "tool_use": {invalid}}` + "\n",
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

	input := `{"type": "tool_use", "tool_use": {"name": "Read", "tool_use_id": "t1"}}` + "\n" +
		`{"type": "tool_use", "tool_use": {"name": "Write", "tool_use_id": "t2"}}` + "\n"

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
	expected := 55000 * 100 / 200000 // 27.5%
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
