package errors

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"
)

// TestHUDError tests the HUDError type
func TestHUDError(t *testing.T) {
	tests := []struct {
		name     string
		err      *HUDError
		expected string
	}{
		{
			name:     "error with op and msg",
			err:      &HUDError{Op: "test.op", Msg: "test message"},
			expected: "test.op: test message",
		},
		{
			name:     "error with op, msg, and underlying error",
			err:      &HUDError{Op: "test.op", Msg: "test message", Err: fmt.Errorf("underlying")},
			expected: "test.op: test message: underlying",
		},
		{
			name:     "error with only msg",
			err:      &HUDError{Msg: "test message"},
			expected: "test message",
		},
		{
			name:     "error with only op",
			err:      &HUDError{Op: "test.op"},
			expected: "test.op: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestWrap tests the Wrap function
func TestWrap(t *testing.T) {
	original := fmt.Errorf("original error")
	wrapped := Wrap(original, "test.op", "wrapped message")

	if wrapped == nil {
		t.Fatal("Wrap() returned nil for non-nil error")
	}

	hudErr, ok := wrapped.(*HUDError)
	if !ok {
		t.Fatalf("Wrap() returned non-HUDError type: %T", wrapped)
	}

	if hudErr.Op != "test.op" {
		t.Errorf("Wrap() Op = %v, want %v", hudErr.Op, "test.op")
	}

	if hudErr.Msg != "wrapped message" {
		t.Errorf("Wrap() Msg = %v, want %v", hudErr.Msg, "wrapped message")
	}

	if hudErr.Err != original {
		t.Errorf("Wrap() Err = %v, want %v", hudErr.Err, original)
	}
}

// TestWrapNil tests wrapping nil error
func TestWrapNil(t *testing.T) {
	wrapped := Wrap(nil, "test.op", "message")
	if wrapped != nil {
		t.Errorf("Wrap(nil) = %v, want nil", wrapped)
	}
}

// TestNew tests the New function
func TestNew(t *testing.T) {
	hudErr := New("test.op", "test message")

	if hudErr.Op != "test.op" {
		t.Errorf("New() Op = %v, want %v", hudErr.Op, "test.op")
	}

	if hudErr.Msg != "test message" {
		t.Errorf("New() Msg = %v, want %v", hudErr.Msg, "test message")
	}
}

// TestErrorUnwrap tests the Unwrap method
func TestErrorUnwrap(t *testing.T) {
	original := fmt.Errorf("original")
	wrapped := &HUDError{Err: original}

	if unwrapped := unwrapError(wrapped); unwrapped != original {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, original)
	}
}

// unwrapError is a helper to unwrap errors
func unwrapError(err error) error {
	if u, ok := err.(interface{ Unwrap() error }); ok {
		return u.Unwrap()
	}
	return nil
}

// TestErrorOp tests the ErrorOp function
func TestErrorOp(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "HUDError with Op",
			err:      &HUDError{Op: "test.op"},
			expected: "test.op",
		},
		{
			name:     "HUDError without Op",
			err:      &HUDError{},
			expected: "",
		},
		{
			name:     "standard error",
			err:      fmt.Errorf("standard error"),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrorOp(tt.err); got != tt.expected {
				t.Errorf("ErrorOp() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestErrorTypeOf tests the ErrorTypeOf function
func TestErrorTypeOf(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorType
	}{
		{
			name:     "Config error",
			err:      NewTyped("config", "failed", TypeConfig),
			expected: TypeConfig,
		},
		{
			name:     "Render error",
			err:      NewTyped("render", "failed", TypeRender),
			expected: TypeRender,
		},
		{
			name:     "Data error",
			err:      NewTyped("data", "failed", TypeData),
			expected: TypeData,
		},
		{
			name:     "Panic error",
			err:      NewTyped("panic", "failed", TypePanic),
			expected: TypePanic,
		},
		{
			name:     "Standard error defaults to Config",
			err:      fmt.Errorf("standard"),
			expected: TypeConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrorTypeOf(tt.err); got != tt.expected {
				t.Errorf("ErrorTypeOf() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestIsTypeFunctions tests the Is* helper functions
func TestIsTypeFunctions(t *testing.T) {
	configErr := NewTyped("config", "failed", TypeConfig)
	renderErr := NewTyped("render", "failed", TypeRender)
	dataErr := NewTyped("data", "failed", TypeData)
	panicErr := NewTyped("panic", "failed", TypePanic)

	tests := []struct {
		name     string
		fn       func(error) bool
		err      error
		expected bool
	}{
		{"IsConfig config error", IsConfig, configErr, true},
		{"IsConfig render error", IsConfig, renderErr, false},
		{"IsRender render error", IsRender, renderErr, true},
		{"IsRender config error", IsRender, configErr, false},
		{"IsData data error", IsData, dataErr, true},
		{"IsData render error", IsData, renderErr, false},
		{"IsPanic panic error", IsPanic, panicErr, true},
		{"IsPanic config error", IsPanic, configErr, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(tt.err); got != tt.expected {
				t.Errorf("%s(%v) = %v, want %v", tt.name, tt.err, got, tt.expected)
			}
		})
	}
}

// TestPanicError tests the PanicError function
func TestPanicError(t *testing.T) {
	panicVal := "test panic"
	err := PanicError("test.op", panicVal)

	typedErr, ok := err.(*TypedError)
	if !ok {
		t.Fatalf("PanicError() returned non-TypedError: %T", err)
	}

	if typedErr.Type != TypePanic {
		t.Errorf("PanicError() Type = %v, want %v", typedErr.Type, TypePanic)
	}

	if typedErr.Op != "test.op" {
		t.Errorf("PanicError() Op = %v, want %v", typedErr.Op, "test.op")
	}
}

// TestSafeRender tests the SafeRender function
func TestSafeRender(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{
			name:     "successful render",
			fn:       func() string { return "success" },
			expected: "success",
		},
		{
			name:     "panic in render",
			fn:       func() string { panic("test panic") },
			expected: "",
		},
		{
			name:     "empty render",
			fn:       func() string { return "" },
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeRender(tt.fn)
			if got != tt.expected {
				t.Errorf("SafeRender() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestSafeRenderWithDefault tests the SafeRenderWithDefault function
func TestSafeRenderWithDefault(t *testing.T) {
	tests := []struct {
		name        string
		fn          func() string
		defaultVal  string
		expected    string
	}{
		{
			name:       "successful render",
			fn:         func() string { return "success" },
			defaultVal: "default",
			expected:   "success",
		},
		{
			name:       "panic in render returns default",
			fn:         func() string { panic("test panic") },
			defaultVal: "default",
			expected:   "default",
		},
		{
			name:       "empty render returns default",
			fn:         func() string { return "" },
			defaultVal: "default",
			expected:   "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeRenderWithDefault(tt.fn, tt.defaultVal)
			if got != tt.expected {
				t.Errorf("SafeRenderWithDefault() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestPlaceholder tests the Placeholder function
func TestPlaceholder(t *testing.T) {
	tests := []struct {
		name     string
		section  string
		reason   string
		expected string
	}{
		{
			name:     "with reason",
			section:  "Session",
			reason:   "unavailable",
			expected: "❓ Session (unavailable)",
		},
		{
			name:     "empty reason",
			section:  "Status",
			reason:   "",
			expected: "❓ Status (unavailable)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Placeholder(tt.section, tt.reason); got != tt.expected {
				t.Errorf("Placeholder() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestOrDefault tests the OrDefault function
func TestOrDefault(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		defaultVal string
		expected   string
	}{
		{
			name:       "non-zero value",
			value:      "actual",
			defaultVal: "default",
			expected:   "actual",
		},
		{
			name:       "zero value returns default",
			value:      "",
			defaultVal: "default",
			expected:   "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OrDefaultString(tt.value, tt.defaultVal)
			if got != tt.expected {
				t.Errorf("OrDefaultString() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestCoalesce tests the Coalesce function
func TestCoalesce(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		expected string
	}{
		{
			name:     "first non-zero",
			values:   []string{"", "", "value", "other"},
			expected: "value",
		},
		{
			name:     "all zero",
			values:   []string{"", "", ""},
			expected: "",
		},
		{
			name:     "first is non-zero",
			values:   []string{"first", "second"},
			expected: "first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Coalesce(tt.values...)
			if got != tt.expected {
				t.Errorf("Coalesce() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestSafeExecute tests the SafeExecute function
func TestSafeExecute(t *testing.T) {
	tests := []struct {
		name      string
		fn        func() (string, error)
		checkErr  bool
		expectErr bool
	}{
		{
			name: "successful execution",
			fn: func() (string, error) {
				return "success", nil
			},
			expectErr: false,
		},
		{
			name: "execution with error",
			fn: func() (string, error) {
				return "error", fmt.Errorf("test error")
			},
			expectErr: true,
		},
		{
			name: "panic in execution",
			fn: func() (string, error) {
				panic("test panic")
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SafeExecute(tt.fn)
			if tt.expectErr && err == nil {
				t.Errorf("SafeExecute() expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("SafeExecute() unexpected error: %v", err)
			}
			if !tt.expectErr && result == "" {
				t.Errorf("SafeExecute() returned empty result")
			}
		})
	}
}

// TestSafeIndex tests the SafeIndex function
func TestSafeIndex(t *testing.T) {
	slice := []string{"a", "b", "c"}

	tests := []struct {
		name     string
		index    int
		expected string
	}{
		{
			name:     "valid index",
			index:    1,
			expected: "b",
		},
		{
			name:     "negative index",
			index:    -1,
			expected: "",
		},
		{
			name:     "out of bounds",
			index:    10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeIndex(slice, tt.index)
			if got != tt.expected {
				t.Errorf("SafeIndex() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestSafeMapGet tests the SafeMapGet function
func TestSafeMapGet(t *testing.T) {
	m := map[string]string{"key": "value"}

	tests := []struct {
		name     string
		m        map[string]string
		key      string
		expected string
	}{
		{
			name:     "existing key",
			m:        m,
			key:      "key",
			expected: "value",
		},
		{
			name:     "non-existing key",
			m:        m,
			key:      "missing",
			expected: "",
		},
		{
			name:     "nil map",
			m:        nil,
			key:      "key",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeMapGet(tt.m, tt.key)
			if got != tt.expected {
				t.Errorf("SafeMapGet() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestRecoverPanic tests the RecoverPanic function
func TestRecoverPanic(t *testing.T) {
	// Test successful recovery
	didPanic := false
	func() {
		defer RecoverPanic("test")
		panic("test panic")
		didPanic = true
	}()

	// If we reach here, recovery worked
	if didPanic {
		t.Error("RecoverPanic() did not recover from panic")
	}
}

// TestLogger tests the Logger functionality
func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelDebug, true)
	logger.SetOutput(&buf)

	// Test logging at different levels
	logger.Debug("test", "debug message")
	logger.Info("test", "info message")
	logger.Warn("test", "warn message")
	logger.Error("test", "error message")

	output := buf.String()
	if output == "" {
		t.Error("Logger produced no output")
	}

	// Check that all log levels are present
	if !contains(output, "DEBUG") {
		t.Error("Logger output missing DEBUG level")
	}
	if !contains(output, "INFO") {
		t.Error("Logger output missing INFO level")
	}
	if !contains(output, "WARN") {
		t.Error("Logger output missing WARN level")
	}
	if !contains(output, "ERROR") {
		t.Error("Logger output missing ERROR level")
	}
}

// TestLoggerLevelFiltering tests that logger filters by level
func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelWarn, false) // Only Warn and Error should pass
	logger.SetOutput(&buf)

	logger.Debug("test", "debug message")
	logger.Info("test", "info message")
	logger.Warn("test", "warn message")
	logger.Error("test", "error message")

	output := buf.String()

	// Debug and Info should be filtered out
	if contains(output, "debug message") {
		t.Error("Logger logged debug message when level is Warn")
	}
	if contains(output, "info message") {
		t.Error("Logger logged info message when level is Warn")
	}

	// Warn and Error should be present
	if !contains(output, "warn message") {
		t.Error("Logger did not log warn message")
	}
	if !contains(output, "error message") {
		t.Error("Logger did not log error message")
	}
}

// TestSetDebugMode tests the global debug mode setting
func TestSetDebugMode(t *testing.T) {
	// Save original state
	originalLogger := GetGlobalLogger()

	SetDebugMode(true)
	if GetGlobalLogger().level != LevelDebug {
		t.Error("SetDebugMode(true) did not set level to Debug")
	}

	SetDebugMode(false)
	// Restore original logger
	SetGlobalLogger(originalLogger)
}

// TestInfiniteRecovery tests the InfiniteRecovery function
func TestInfiniteRecovery(t *testing.T) {
	callCount := 0
	maxCalls := 3

	done := make(chan bool)
	go func() {
		InfiniteRecovery("test", func() bool {
			callCount++
			return callCount < maxCalls
		}, maxCalls)
		done <- true
	}()

	// Should complete
	select {
	case <-done:
		if callCount != maxCalls {
			t.Errorf("InfiniteRecovery() call count = %v, want %v", callCount, maxCalls)
		}
	}
}

// contains is a helper to check if a string contains a substring
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

// BenchmarkSafeRender benchmarks the SafeRender function
func BenchmarkSafeRender(b *testing.B) {
	fn := func() string {
		return "test output"
	}

	for i := 0; i < b.N; i++ {
		SafeRender(fn)
	}
}

// BenchmarkSafeRenderWithPanic benchmarks SafeRender with panics
func BenchmarkSafeRenderWithPanic(b *testing.B) {
	// Introduce panics occasionally
	panicFn := func() string {
		if b.N%100 == 0 {
			panic("test panic")
		}
		return "test output"
	}

	for i := 0; i < b.N; i++ {
		SafeRender(panicFn)
	}
}

// TestStackTrace tests the StackTrace function
func TestStackTrace(t *testing.T) {
	trace := StackTrace()
	if trace == "" {
		t.Error("StackTrace() returned empty string")
	}
	// Stack trace should contain function names
	if !contains(trace, "TestStackTrace") {
		t.Error("StackTrace() does not contain calling function name")
	}
}

// TestMust tests the Must function
func TestMust(t *testing.T) {
	// Test successful case
	result := Must("success", nil)
	if result != "success" {
		t.Errorf("Must() = %v, want %v", result, "success")
	}

	// Test panic case
	defer func() {
		if r := recover(); r == nil {
			t.Error("Must() did not panic on error")
		}
	}()

	Must("value", fmt.Errorf("test error"))
}

// TestSuppressError tests the SuppressError function
func TestSuppressError(t *testing.T) {
	// SuppressError should not panic
	called := false
	SuppressError("test", func() error {
		called = true
		return fmt.Errorf("test error")
	})

	if !called {
		t.Error("SuppressError() did not call function")
	}
}

// TestFallback tests the Fallback function
func TestFallback(t *testing.T) {
	tests := []struct {
		name      string
		primary   string
		fallbacks []string
		expected  string
	}{
		{
			name:      "primary is valid",
			primary:   "primary",
			fallbacks: []string{"fallback1", "fallback2"},
			expected:  "primary",
		},
		{
			name:      "use first fallback",
			primary:   "",
			fallbacks: []string{"fallback1", "fallback2"},
			expected:  "fallback1",
		},
		{
			name:      "use second fallback",
			primary:   "",
			fallbacks: []string{"", "fallback2"},
			expected:  "fallback2",
		},
		{
			name:      "all empty",
			primary:   "",
			fallbacks: []string{"", ""},
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Fallback(tt.primary, tt.fallbacks...)
			if got != tt.expected {
				t.Errorf("Fallback() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestSafeGo tests the SafeGo function
func TestSafeGo(t *testing.T) {
	done := make(chan bool)
	SafeGo("test", func() {
		panic("test panic in goroutine")
		done <- true
	})

	// Should complete without crashing the test
	select {
	case <-done:
		// Success - goroutine completed
	case <-make(chan struct{}):
		// Timeout
		t.Error("SafeGo() goroutine did not complete")
	}
}

// TestIsNil tests the IsNil function
func TestIsNil(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"nil interface", nil, true},
		{"nil pointer", (*int)(nil), true},
		{"nil slice", ([]int)(nil), true},
		{"nil map", (map[string]int)(nil), true},
		{"non-nil pointer", intPtr(42), false},
		{"non-nil slice", []int{1}, false},
		{"non-nil map", map[string]int{"a": 1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNil(tt.value)
			if got != tt.expected {
				t.Errorf("IsNil() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Helper function
func intPtr(i int) *int {
	return &i
}

// TestLogErrorWithLevel tests LogErrorWithLevel
func TestLogErrorWithLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LevelDebug, true)
	logger.SetOutput(&buf)

	tests := []struct {
		name           string
		err            error
		shouldContain  string
		shouldNotLevel string
	}{
		{
			name:           "render error logs as WARN",
			err:            RenderError("render", "failed"),
			shouldContain:  "WARN",
			shouldNotLevel: "ERROR",
		},
		{
			name:           "data error logs as WARN",
			err:            DataError("data", "failed"),
			shouldContain:  "WARN",
			shouldNotLevel: "ERROR",
		},
		{
			name:           "config error logs as ERROR",
			err:            ConfigError("config", "failed"),
			shouldContain:  "ERROR",
			shouldNotLevel: "",
		},
		{
			name:           "panic error logs as ERROR",
			err:            PanicError("test", "panic"),
			shouldContain:  "ERROR",
			shouldNotLevel: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			logger.LogErrorWithLevel(tt.err)
			output := buf.String()

			if !contains(output, tt.shouldContain) {
				t.Errorf("LogErrorWithLevel() output should contain %v", tt.shouldContain)
			}

			if tt.shouldNotLevel != "" && contains(output, tt.shouldNotLevel) {
				t.Errorf("LogErrorWithLevel() output should not contain %v", tt.shouldNotLevel)
			}
		})
	}
}

// TestPanicRecoveryRecoveryCount tests panic recovery counting
func TestPanicRecoveryRecoveryCount(t *testing.T) {
	pr := NewPanicRecovery()
	pr.SetLogByDefault(false) // Disable logging in test

	if pr.RecoveryCount() != 0 {
		t.Errorf("NewPanicRecovery() count = %v, want 0", pr.RecoveryCount())
	}

	// Trigger a recovery
	func() {
		defer pr.Recover("test")
		panic("test")
	}()

	if pr.RecoveryCount() != 1 {
		t.Errorf("After one panic, count = %v, want 1", pr.RecoveryCount())
	}

	// Trigger another recovery
	func() {
		defer pr.Recover("test")
		panic("test2")
	}()

	if pr.RecoveryCount() != 2 {
		t.Errorf("After two panics, count = %v, want 2", pr.RecoveryCount())
	}

	pr.ResetCount()
	if pr.RecoveryCount() != 0 {
		t.Errorf("After ResetCount(), count = %v, want 0", pr.RecoveryCount())
	}
}

// TestPanicRecoveryMaxRecoveries tests max recovery limit
func TestPanicRecoveryMaxRecoveries(t *testing.T) {
	pr := NewPanicRecovery()
	pr.SetLogByDefault(false)
	pr.SetMaxRecoveries(2)

	// First two panics should be recovered
	for i := 0; i < 2; i++ {
		func() {
			defer pr.Recover("test")
			panic("test")
		}()
	}

	// Third panic should re-panic
	didPanic := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		defer pr.Recover("test")
		panic("test")
	}()

	if !didPanic {
		t.Error("Third panic should not be recovered (exceeded max)")
	}
}

// TestSafeGetValue tests SafeGetValue helper in graceful.go
func TestSafeGetValue(t *testing.T) {
	tests := []struct {
		name      string
		fn        func() string
		wantErr   bool
		expected  string
	}{
		{
			name:     "successful value",
			fn:       func() string { return "value" },
			wantErr:  false,
			expected: "value",
		},
		{
			name:     "panic returns empty string and error",
			fn:       func() string { panic("test") },
			wantErr:  true,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SafeValue(tt.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if result != tt.expected {
				t.Errorf("SafeValue() result = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestWithRecoveryWithResult tests WithRecoveryAndResult
func TestWithRecoveryAndResult(t *testing.T) {
	tests := []struct {
		name      string
		fn        func() (string, error)
		wantErr   bool
		checkPanic bool
	}{
		{
			name: "successful result",
			fn: func() (string, error) {
				return "success", nil
			},
			wantErr: false,
		},
		{
			name: "error result",
			fn: func() (string, error) {
				return "error", fmt.Errorf("test error")
			},
			wantErr: true,
		},
		{
			name: "panic in function",
			fn: func() (string, error) {
				panic("test panic")
			},
			wantErr:   true,
			checkPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := WithRecoveryAndResult("test", tt.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithRecoveryAndResult() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result == "" {
				t.Errorf("WithRecoveryAndResult() result should not be empty on success")
			}
			if tt.checkPanic && err == nil {
				t.Errorf("WithRecoveryAndResult() should return error on panic")
			}
		})
	}
}

// TestLoopRecovery tests LoopRecovery function
func TestLoopRecovery(t *testing.T) {
	// This test ensures LoopRecovery always returns true
	if !LoopRecovery("test") {
		t.Error("LoopRecovery() should always return true")
	}

	// Test with actual panic
	didPanic := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		if !LoopRecovery("test") {
			t.Error("LoopRecovery() should return true even after panic")
		}
		panic("test")
	}()

	if !didPanic {
		t.Error("Expected panic to occur before recovery")
	}
}

// TestPlaceholderWithValue tests PlaceholderWithValue
func TestPlaceholderWithValue(t *testing.T) {
	tests := []struct {
		name     string
		icon     string
		section  string
		reason   string
		expected string
	}{
		{
			name:     "custom icon",
			icon:     "⚠️",
			section:  "Section",
			reason:   "failed",
			expected: "⚠️ Section (failed)",
		},
		{
			name:     "default icon",
			icon:     "",
			section:  "Section",
			reason:   "failed",
			expected: "❓ Section (failed)",
		},
		{
			name:     "empty reason",
			icon:     "⚠️",
			section:  "Section",
			reason:   "",
			expected: "⚠️ Section (unavailable)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PlaceholderWithValue(tt.icon, tt.section, tt.reason)
			if got != tt.expected {
				t.Errorf("PlaceholderWithValue() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestDefaultHelpers tests default value helpers
func TestDefaultHelpers(t *testing.T) {
	t.Run("OrDefaultSlice", func(t *testing.T) {
		slice := []int{1, 2, 3}
		empty := []int{}
		defaultSlice := []int{4, 5}

		if got := OrDefaultSlice(slice, defaultSlice); len(got) != 3 {
			t.Errorf("OrDefaultSlice() returned wrong length")
		}
		if got := OrDefaultSlice(empty, defaultSlice); len(got) != 2 {
			t.Errorf("OrDefaultSlice() with empty should return default")
		}
	})

	t.Run("FirstNonNil", func(t *testing.T) {
		a, b := "a", "b"
		if got := FirstNonNil(nil, &a, nil, &b); got != &a {
			t.Errorf("FirstNonNil() should return first non-nil")
		}
		if got := FirstNonNil[string](nil, nil); got != nil {
			t.Errorf("FirstNonNil() with all nil should return nil")
		}
	})

	t.Run("FirstNonNilError", func(t *testing.T) {
		err1 := fmt.Errorf("error 1")

		if got := FirstNonNilError(nil, nil, err1); got != err1 {
			t.Errorf("FirstNonNilError() should return first non-nil error")
		}
		if got := FirstNonNilError(nil, nil); got != nil {
			t.Errorf("FirstNonNilError() with all nil should return nil")
		}
	})
}

// TestSafeMapGetWithDefault tests SafeMapGetWithDefault
func TestSafeMapGetWithDefault(t *testing.T) {
	m := map[string]string{"key": "value"}
	defaultVal := "default"

	tests := []struct {
		name     string
		m        map[string]string
		key      string
		defaultV string
		expected string
	}{
		{
			name:     "existing key",
			m:        m,
			key:      "key",
			defaultV: defaultVal,
			expected: "value",
		},
		{
			name:     "non-existing key",
			m:        m,
			key:      "missing",
			defaultV: defaultVal,
			expected: "default",
		},
		{
			name:     "nil map",
			m:        nil,
			key:      "key",
			defaultV: defaultVal,
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeMapGetWithDefault(tt.m, tt.key, tt.defaultV)
			if got != tt.expected {
				t.Errorf("SafeMapGetWithDefault() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestRecoverAndLog tests RecoverAndLog
func TestRecoverAndLog(t *testing.T) {
	// Test that RecoverAndLog catches panics
	didPanic := false
	func() {
		defer RecoverAndLog("test")
		panic("test panic")
		didPanic = true
	}()

	// If we reach here, recovery worked
	if didPanic {
		t.Error("RecoverAndLog() did not recover from panic")
	}
}

// TestSafeIndexWithDefault tests SafeIndexWithDefault
func TestSafeIndexWithDefault(t *testing.T) {
	slice := []string{"a", "b", "c"}
	defaultVal := "default"

	tests := []struct {
		name     string
		slice    []string
		index    int
		defaultV string
		expected string
	}{
		{
			name:     "valid index",
			slice:    slice,
			index:    1,
			defaultV: defaultVal,
			expected: "b",
		},
		{
			name:     "negative index",
			slice:    slice,
			index:    -1,
			defaultV: defaultVal,
			expected: "default",
		},
		{
			name:     "out of bounds",
			slice:    slice,
			index:    10,
			defaultV: defaultVal,
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeIndexWithDefault(tt.slice, tt.index, tt.defaultV)
			if got != tt.expected {
				t.Errorf("SafeIndexWithDefault() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestGoVersionChecks tests Go version-specific features
func TestGoVersionChecks(t *testing.T) {
	// This test ensures we're running on a compatible Go version
	if ver := runtime.Version(); ver == "" {
		t.Error("Go version is empty")
	}
}

// TestWrapf tests Wrapf function
func TestWrapf(t *testing.T) {
	original := fmt.Errorf("original")
	wrapped := Wrapf(original, "test.op", "formatted: %s %d", "value", 42)

	if wrapped == nil {
		t.Fatal("Wrapf() returned nil for non-nil error")
	}

	hudErr, ok := wrapped.(*HUDError)
	if !ok {
		t.Fatalf("Wrapf() returned non-HUDError type: %T", wrapped)
	}

	expectedMsg := "formatted: value 42"
	if hudErr.Msg != expectedMsg {
		t.Errorf("Wrapf() Msg = %v, want %v", hudErr.Msg, expectedMsg)
	}

	if hudErr.Err != original {
		t.Errorf("Wrapf() Err = %v, want %v", hudErr.Err, original)
	}
}

// TestLoggerConcurrent tests concurrent logging
func TestLoggerConcurrent(t *testing.T) {
	logger := NewLogger(LevelInfo, false)
	done := make(chan bool)

	// Log from multiple goroutines
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			logger.Info("test", "concurrent log %d", id)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
