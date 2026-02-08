package errors

import (
	"fmt"
	"runtime/debug"
	"sync"
)

// RecoveryHandler is a function that handles recovered panics.
type RecoveryHandler func(panicValue interface{}, stackTrace []byte)

// PanicRecovery manages panic recovery with custom handlers.
type PanicRecovery struct {
	mu            sync.Mutex
	handler       RecoveryHandler
	logByDefault  bool
	enabled       bool
	recoveryCount int
	maxRecoveries int
}

// NewPanicRecovery creates a new panic recovery manager.
func NewPanicRecovery() *PanicRecovery {
	return &PanicRecovery{
		logByDefault:  true,
		enabled:       true,
		maxRecoveries: -1, // Unlimited
		handler:       defaultRecoveryHandler,
	}
}

// SetHandler sets a custom recovery handler.
func (pr *PanicRecovery) SetHandler(handler RecoveryHandler) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.handler = handler
}

// SetLogByDefault enables or disables automatic logging of panics.
func (pr *PanicRecovery) SetLogByDefault(log bool) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.logByDefault = log
}

// SetEnabled enables or disables panic recovery.
func (pr *PanicRecovery) SetEnabled(enabled bool) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.enabled = enabled
}

// SetMaxRecoveries sets the maximum number of recoveries before giving up.
// Use -1 for unlimited (default).
func (pr *PanicRecovery) SetMaxRecoveries(max int) {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.maxRecoveries = max
}

// RecoveryCount returns the number of panics recovered.
func (pr *PanicRecovery) RecoveryCount() int {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	return pr.recoveryCount
}

// ResetCount resets the recovery counter.
func (pr *PanicRecovery) ResetCount() {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.recoveryCount = 0
}

// Recover catches a panic and handles it using the configured handler.
// Returns true if a panic was recovered, false otherwise.
// This should be called with defer.
func (pr *PanicRecovery) Recover(op string) bool {
	return pr.RecoverWithOperation(op)
}

// RecoverWithOperation catches a panic and handles it with operation context.
func (pr *PanicRecovery) RecoverWithOperation(op string) bool {
	pr.mu.Lock()
	if !pr.enabled {
		pr.mu.Unlock()
		return false
	}

	// Check if we've exceeded max recoveries
	if pr.maxRecoveries >= 0 && pr.recoveryCount >= pr.maxRecoveries {
		pr.mu.Unlock()
		// Re-panic if we've exceeded the limit
		panic(fmt.Sprintf("max panic recoveries (%d) exceeded in %s", pr.maxRecoveries, op))
	}
	pr.mu.Unlock()

	r := recover()
	if r == nil {
		return false
	}

	pr.mu.Lock()
	pr.recoveryCount++
	count := pr.recoveryCount
	pr.mu.Unlock()

	stack := debug.Stack()

	// Log by default if enabled
	if pr.logByDefault {
		err := PanicError(op, r)
		LogErrorWithLevel(err)
	}

	// Call custom handler if set
	if pr.handler != nil {
		pr.handler(r, stack)
	}

	// Log recovery count if it's getting high
	if count > 10 {
		Warn(op, "high panic recovery count: %d", count)
	}

	return true
}

// defaultRecoveryHandler is the default panic handler.
func defaultRecoveryHandler(panicValue interface{}, stackTrace []byte) {
	// Default behavior is just logging, which is already done in Recover()
	// This handler can be replaced with custom behavior
}

// Go runs a function in a goroutine with panic recovery.
func (pr *PanicRecovery) Go(op string, fn func()) {
	go func() {
		defer pr.Recover(op)
		fn()
	}()
}

// Global panic recovery instance.
var globalRecovery = NewPanicRecovery()

// SetGlobalRecoveryHandler sets the global panic recovery handler.
func SetGlobalRecoveryHandler(handler RecoveryHandler) {
	globalRecovery.SetHandler(handler)
}

// EnablePanicRecovery enables global panic recovery.
func EnablePanicRecovery() {
	globalRecovery.SetEnabled(true)
}

// DisablePanicRecovery disables global panic recovery.
func DisablePanicRecovery() {
	globalRecovery.SetEnabled(false)
}

// RecoverPanic catches a panic using the global recovery manager.
// This is the main function that should be used with defer.
func RecoverPanic(op string) {
	globalRecovery.Recover(op)
}

// RecoverPanicAndReturn catches a panic and returns an error.
// This is useful for converting panics to errors in functions.
func RecoverPanicAndReturn(op string) error {
	if r := recover(); r != nil {
		err := PanicError(op, r)
		LogErrorWithLevel(err)
		return err
	}
	return nil
}

// SafeGo runs a function in a goroutine with panic recovery using the global recovery.
func SafeGo(op string, fn func()) {
	globalRecovery.Go(op, fn)
}

// WithRecovery wraps a function with panic recovery.
func WithRecovery(op string, fn func()) {
	defer RecoverPanic(op)
	fn()
}

// WithRecoveryAndError wraps a function with panic recovery and returns any error.
func WithRecoveryAndError(op string, fn func() error) (err error) {
	defer RecoverPanic(op)
	return fn()
}

// WithRecoveryAndResult wraps a function with panic recovery and returns result and error.
func WithRecoveryAndResult[T any](op string, fn func() (T, error)) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = PanicError(op, r)
			LogErrorWithLevel(err)
		}
	}()
	return fn()
}

// MainRecovery is a specialized recovery for the main goroutine.
// It logs the panic and exits gracefully if recovery is not possible.
func MainRecovery() {
	if r := recover(); r != nil {
		err := PanicError("main", r)
		LogErrorWithLevel(err)
		Error("main", "fatal panic in main goroutine, application will exit")
		// In a real application, you might want to do cleanup here
		// For now, we re-panic to exit
		panic(r)
	}
}

// MainRecoveryWithHandler is a specialized recovery for the main goroutine with a custom handler.
func MainRecoveryWithHandler(handler func(error)) {
	if r := recover(); r != nil {
		err := PanicError("main", r)
		LogErrorWithLevel(err)
		if handler != nil {
			handler(err)
		} else {
			Error("main", "fatal panic in main goroutine, application will exit")
		}
		panic(r)
	}
}

// LoopRecovery is designed for use in render loops where you want to continue after panics.
// It returns true if the loop should continue, false if it should break.
func LoopRecovery(op string) bool {
	recovered := globalRecovery.Recover(op)
	if recovered {
		Warn(op, "recovered from panic, continuing loop")
	}
	return true // Always continue after recovery
}

// InfiniteRecovery runs a function continuously, recovering from panics.
// The function will be called repeatedly until it returns false or maxCalls is reached.
func InfiniteRecovery(op string, fn func() bool, maxCalls ...int) {
	callCount := 0
	max := -1
	if len(maxCalls) > 0 && maxCalls[0] > 0 {
		max = maxCalls[0]
	}

	for {
		if max > 0 && callCount >= max {
			Info(op, "reached max call count %d, stopping", max)
			break
		}

		recovered := false
		fnReturnedFalse := false
		func() {
			defer func() {
				recovered = globalRecovery.Recover(op)
			}()

			if !fn() {
				Info(op, "function returned false, stopping loop")
				fnReturnedFalse = true
				return
			}
			callCount++
		}()

		if fnReturnedFalse {
			break
		}

		if recovered && globalRecovery.RecoveryCount() > 100 {
			Error(op, "too many panics (%d), stopping loop", globalRecovery.RecoveryCount())
			return
		}
	}
}
