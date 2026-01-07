package errors

import (
	"fmt"
	"reflect"
)

// SafeRender executes a render function and catches any panics.
// If the function panics, it logs the panic and returns a placeholder.
func SafeRender(fn func() string) string {
	defer func() {
		if r := recover(); r != nil {
			err := PanicError("render", r)
			LogErrorWithLevel(err)
		}
	}()

	return fn()
}

// SafeRenderWithDefault executes a render function and catches any panics.
// If the function panics or returns an empty string, it returns the default value.
func SafeRenderWithDefault(fn func() string, defaultVal string) string {
	defer func() {
		if r := recover(); r != nil {
			err := PanicError("render", r)
			LogErrorWithLevel(err)
		}
	}()

	result := fn()
	if result == "" {
		return defaultVal
	}
	return result
}

// Placeholder returns a formatted placeholder string for a missing section.
// The format is: "❓ {section} ({reason})"
func Placeholder(section, reason string) string {
	if reason == "" {
		reason = "unavailable"
	}
	return fmt.Sprintf("❓ %s (%s)", section, reason)
}

// PlaceholderWithValue returns a formatted placeholder with a specific icon.
func PlaceholderWithValue(icon, section, reason string) string {
	if icon == "" {
		icon = "❓"
	}
	if reason == "" {
		reason = "unavailable"
	}
	return fmt.Sprintf("%s %s (%s)", icon, section, reason)
}

// OrDefault returns the value if it's non-zero, otherwise returns the default.
// This works with any comparable type.
func OrDefault[T comparable](value T, defaultVal T) T {
	var zero T
	if value == zero {
		return defaultVal
	}
	return value
}

// OrDefaultFunc returns the value if it's non-zero, otherwise calls the function to get the default.
func OrDefaultFunc[T comparable](value T, fn func() T) T {
	var zero T
	if value == zero {
		return fn()
	}
	return value
}

// OrDefaultSlice returns the slice if it's non-empty, otherwise returns the default.
func OrDefaultSlice[T any](value []T, defaultVal []T) []T {
	if len(value) == 0 {
		return defaultVal
	}
	return value
}

// OrDefaultString returns the string if it's non-empty, otherwise returns the default.
func OrDefaultString(value, defaultVal string) string {
	if value == "" {
		return defaultVal
	}
	return value
}

// Coalesce returns the first non-zero value from the provided options.
// If all values are zero, returns the zero value for the type.
func Coalesce[T comparable](values ...T) T {
	var zero T
	for _, v := range values {
		if v != zero {
			return v
		}
	}
	return zero
}

// Must panics if the error is non-nil, otherwise returns the value.
// This is similar to the pattern used in Go's standard library.
// Use this when you know the error cannot happen in a given context.
func Must[T any](value T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("Must failed: %v", err))
	}
	return value
}

// SafeCall executes a function and catches any panics, returning an error instead.
func SafeCall(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = PanicError("SafeCall", r)
		}
	}()

	return fn()
}

// SafeExecute executes a function and catches any panics, returning the result and error.
func SafeExecute[T any](fn func() (T, error)) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = PanicError("SafeExecute", r)
		}
	}()

	return fn()
}

// SafeValue executes a function and catches any panics, returning the zero value and error.
func SafeValue[T any](fn func() T) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = PanicError("SafeValue", r)
		}
	}()

	return fn(), nil
}

// IsNil checks if a value is nil, including interface nil checks.
func IsNil(value interface{}) bool {
	if value == nil {
		return true
	}

	// Check for nil interface with non-nil concrete type
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	}

	return false
}

// FirstNonNil returns the first non-nil value from the provided options.
func FirstNonNil[T any](values ...*T) *T {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}

// FirstNonNilError returns the first non-nil error from the provided options.
func FirstNonNilError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// SuppressError executes the function and returns nil error, logging any error that occurs.
func SuppressError(op string, fn func() error) {
	if err := fn(); err != nil {
		LogError(Wrap(err, op, "suppressed error"))
	}
}

// RecoverAndLog is a helper for deferred panic recovery.
// It should be used with defer: `defer RecoverAndLog("operation")`
func RecoverAndLog(op string) {
	if r := recover(); r != nil {
		err := PanicError(op, r)
		LogErrorWithLevel(err)
	}
}

// RecoverAndLogWithHandler is a helper for deferred panic recovery with custom handler.
// It should be used with defer: `defer RecoverAndLogWithHandler("operation", handler)`
func RecoverAndLogWithHandler(op string, handler func(error)) {
	if r := recover(); r != nil {
		err := PanicError(op, r)
		LogErrorWithLevel(err)
		if handler != nil {
			handler(err)
		}
	}
}

// Fallback returns the primary value if valid, otherwise tries fallback values in order.
func Fallback[T comparable](primary T, fallbacks ...T) T {
	var zero T
	if primary != zero {
		return primary
	}
	for _, fb := range fallbacks {
		if fb != zero {
			return fb
		}
	}
	return zero
}

// SafeIndex returns the element at index if it exists, otherwise returns zero value.
func SafeIndex[T any](slice []T, index int) T {
	if index >= 0 && index < len(slice) {
		return slice[index]
	}
	var zero T
	return zero
}

// SafeIndexWithDefault returns the element at index if it exists, otherwise returns defaultVal.
func SafeIndexWithDefault[T any](slice []T, index int, defaultVal T) T {
	if index >= 0 && index < len(slice) {
		return slice[index]
	}
	return defaultVal
}

// SafeMapGet returns the value for key if it exists, otherwise returns zero value.
func SafeMapGet[K comparable, V any](m map[K]V, key K) V {
	if m != nil {
		if val, ok := m[key]; ok {
			return val
		}
	}
	var zero V
	return zero
}

// SafeMapGetWithDefault returns the value for key if it exists, otherwise returns defaultVal.
func SafeMapGetWithDefault[K comparable, V any](m map[K]V, key K, defaultVal V) V {
	if m != nil {
		if val, ok := m[key]; ok {
			return val
		}
	}
	return defaultVal
}
