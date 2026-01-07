package transcript

import (
	"context"
	"strings"
	"testing"
)

// Benchmark parsing single event
func BenchmarkParser_ParseLine_SingleEvent(b *testing.B) {
	p := NewParser("test.jsonl")
	line := []byte(`{"type": "tool_use", "tool_use": {"name": "Read", "tool_use_id": "test123"}}` + "\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.parseLine(line)
	}
}

// Benchmark parsing multiple events
func BenchmarkParser_ParseFromReader_MultipleEvents(b *testing.B) {
	input := strings.Builder{}
	for i := 0; i < 100; i++ {
		input.WriteString(`{"type": "tool_use", "tool_use": {"name": "Read", "tool_use_id": "test123"}}` + "\n")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser("test.jsonl")
		r := strings.NewReader(input.String())
		_ = p.ParseFromReader(ctx, r)
	}
}

// Benchmark parsing with different event types
func BenchmarkParser_ParseLine_EventTypes(b *testing.B) {
	events := []string{
		`{"type": "assistant_message", "message": {"role": "assistant", "model": "claude-opus-4-5-20251101"}}`,
		`{"type": "tool_use", "tool_use": {"name": "Read", "tool_use_id": "test123"}}`,
		`{"type": "tool_result", "tool_result": {"tool_use_id": "test123", "content": "result"}}`,
		`{"type": "agent_run", "agent_run": {"agent_id": "agent1", "agent_name": "test-agent"}}`,
		`{"type": "todo", "todo": {"id": "1", "status": "pending", "content": "Task 1"}}`,
	}

	for _, event := range events {
		line := []byte(event + "\n")
		b.Run(event, func(b *testing.B) {
			p := NewParser("test.jsonl")
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = p.parseLine(line)
			}
		})
	}
}

// Benchmark concurrent access
func BenchmarkParser_ConcurrentAccess(b *testing.B) {
	p := NewParser("test.jsonl")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = p.GetState()
			_ = p.GetContextWindow()
			_ = p.GetSessionStart()
			_ = p.ActiveToolCount()
			_ = p.ActiveAgentCount()
		}
	})
}

// Benchmark memory allocation
func BenchmarkParser_ParseFromReader_WithAllocations(b *testing.B) {
	input := strings.Builder{}
	for i := 0; i < 100; i++ {
		input.WriteString(`{"type": "tool_use", "tool_use": {"name": "Read", "tool_use_id": "test123"}}` + "\n")
	}

	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser("test.jsonl")
		r := strings.NewReader(input.String())
		_ = p.ParseFromReader(ctx, r)
	}
}

// Benchmark token cost calculation
func BenchmarkParser_CalculateCost(b *testing.B) {
	p := NewParser("test.jsonl")

	// Pre-populate with some data
	input := `{"type": "assistant_message", "message": {"role": "assistant", "model": "claude-opus-4-5-20251101", "input_tokens": 1000000, "output_tokens": 500000}}` + "\n"
	ctx := context.Background()
	r := strings.NewReader(input)
	_ = p.ParseFromReader(ctx, r)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.CalculateCost()
	}
}

// Benchmark todo operations
func BenchmarkParser_TodoOperations(b *testing.B) {
	p := NewParser("test.jsonl")

	// Pre-populate with todos
	for i := 0; i < 100; i++ {
		input := `{"type": "todo", "todo": {"id": "` + string(rune('0'+i)) + `", "status": "pending", "content": "Task ` + string(rune('0'+i)) + `"}}` + "\n"
		ctx := context.Background()
		r := strings.NewReader(input)
		_ = p.ParseFromReader(ctx, r)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.GetTodos()
		_, _ = p.GetTodoCount()
		_ = p.GetCurrentTodo()
	}
}
