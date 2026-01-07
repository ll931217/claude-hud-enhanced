package statusline

import (
	"context"
	"testing"
	"time"

	"github.com/ll931217/claude-hud-enhanced/internal/config"
	"github.com/ll931217/claude-hud-enhanced/internal/registry"
)

func BenchmarkStatusline_Render(b *testing.B) {
	cfg := config.DefaultConfig()
	reg := registry.DefaultRegistry()
	sl, _ := New(cfg, reg)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sl.Render()
	}
	_ = ctx
}

func BenchmarkStatusline_RenderWithSections(b *testing.B) {
	cfg := config.DefaultConfig()
	reg := registry.DefaultRegistry()
	sl, _ := New(cfg, reg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sl.Render()
	}
}

func BenchmarkStatusline_RenderSingle(b *testing.B) {
	cfg := config.DefaultConfig()
	reg := registry.DefaultRegistry()
	sl, _ := New(cfg, reg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sl.Render()
	}
}

func BenchmarkStatusline_Refresh(b *testing.B) {
	cfg := config.DefaultConfig()
	cfg.RefreshIntervalMs = 100
	reg := registry.DefaultRegistry()
	sl, _ := New(cfg, reg)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sl.Refresh()
	}
	_ = ctx
}

func BenchmarkStatusline_CreateAndRender(b *testing.B) {
	cfg := config.DefaultConfig()
	reg := registry.DefaultRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sl, _ := New(cfg, reg)
		_ = sl.Render()
	}
}

func BenchmarkConfig_DefaultConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.DefaultConfig()
	}
}

func BenchmarkConfig_Validate(b *testing.B) {
	cfg := config.DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a copy and validate
		c := *cfg
		_ = config.LoadFromPath("/dev/null")
		_ = c
	}
}

func BenchmarkConfig_GetEnabledSections(b *testing.B) {
	cfg := config.DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.GetEnabledSections()
	}
}

// Benchmark memory allocation with ReportAllocs
func BenchmarkStatusline_Render_WithAllocations(b *testing.B) {
	cfg := config.DefaultConfig()
	reg := registry.DefaultRegistry()
	sl, _ := New(cfg, reg)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sl.Render()
	}
}

// Benchmark timing to ensure we meet latency target
func BenchmarkStatusline_Render_Latency(b *testing.B) {
	cfg := config.DefaultConfig()
	reg := registry.DefaultRegistry()
	sl, _ := New(cfg, reg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		_ = sl.Render()
		elapsed := time.Since(start)

		// Fail if any iteration takes >50ms
		if elapsed > 50*time.Millisecond {
			b.Errorf("Render took too long: %v (>50ms target)", elapsed)
		}
	}
}
