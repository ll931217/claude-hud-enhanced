package errors

import (
	errs "errors"
	"fmt"
	"runtime/debug"
	"strings"
)

// HUDError represents a context-rich error in the Claude HUD application.
// It wraps underlying errors with operation context and human-readable messages.
type HUDError struct {
	Op  string // Operation that failed (e.g., "config.load", "statusline.render")
	Err error  // Underlying error (may be nil)
	Msg string // Human-readable message explaining what went wrong
}

// Error implements the error interface.
func (e *HUDError) Error() string {
	var sb strings.Builder

	if e.Op != "" {
		sb.WriteString(e.Op)
		sb.WriteString(": ")
	}

	if e.Msg != "" {
		sb.WriteString(e.Msg)
	}

	if e.Err != nil {
		if e.Msg != "" {
			sb.WriteString(": ")
		}
		sb.WriteString(e.Err.Error())
	}

	return sb.String()
}

// Unwrap returns the underlying error, compatible with errors.Is/As.
func (e *HUDError) Unwrap() error {
	return e.Err
}

// New creates a new HUDError with the given operation and message.
func New(op, msg string) *HUDError {
	return &HUDError{
		Op:  op,
		Msg: msg,
	}
}

// Wrap wraps an existing error with operation context and message.
// If err is nil, returns nil.
func Wrap(err error, op, msg string) error {
	if err == nil {
		return nil
	}
	return &HUDError{
		Op:  op,
		Err: err,
		Msg: msg,
	}
}

// Wrapf wraps an error with a formatted message.
func Wrapf(err error, op, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &HUDError{
		Op:  op,
		Err: err,
		Msg: fmt.Sprintf(format, args...),
	}
}

// ErrorOp returns the operation that caused the error.
// Returns the Op if it's a HUDError (even if empty), or "unknown" for other error types.
func ErrorOp(err error) string {
	var hudErr *HUDError
	if errs.As(err, &hudErr) {
		return hudErr.Op
	}
	return "unknown"
}

// ErrorMsg returns the human-readable message, or the error string if not a HUDError.
func ErrorMsg(err error) string {
	var hudErr *HUDError
	if errs.As(err, &hudErr) && hudErr.Msg != "" {
		return hudErr.Msg
	}
	return err.Error()
}

// StackTrace captures and returns the current stack trace.
func StackTrace() string {
	return string(debug.Stack())
}

// ErrorType categorizes different types of errors for better handling.
type ErrorType int

const (
	// TypeConfig is for configuration-related errors
	TypeConfig ErrorType = iota
	// TypeRender is for rendering/display errors
	TypeRender
	// TypeData is for data source errors
	TypeData
	// TypeFileSystem is for file system operations
	TypeFileSystem
	// TypeNetwork is for network operations
	TypeNetwork
	// TypePanic is for recovered panics
	TypePanic
)

// TypedError extends HUDError with error type classification.
type TypedError struct {
	*HUDError
	Type ErrorType
}

// Error implements the error interface for TypedError.
func (e *TypedError) Error() string {
	return e.HUDError.Error()
}

// NewTyped creates a new typed error.
func NewTyped(op, msg string, errType ErrorType) *TypedError {
	return &TypedError{
		HUDError: &HUDError{
			Op:  op,
			Msg: msg,
		},
		Type: errType,
	}
}

// WrapTyped wraps an error with type classification.
func WrapTyped(err error, op, msg string, errType ErrorType) error {
	if err == nil {
		return nil
	}
	return &TypedError{
		HUDError: &HUDError{
			Op:  op,
			Err: err,
			Msg: msg,
		},
		Type: errType,
	}
}

// ErrorTypeOf returns the type of the error, or TypeConfig as default.
func ErrorTypeOf(err error) ErrorType {
	var typedErr *TypedError
	if errs.As(err, &typedErr) {
		return typedErr.Type
	}
	return TypeConfig
}

// IsConfig returns true if this is a configuration error.
func IsConfig(err error) bool {
	return ErrorTypeOf(err) == TypeConfig
}

// IsRender returns true if this is a render error.
func IsRender(err error) bool {
	return ErrorTypeOf(err) == TypeRender
}

// IsData returns true if this is a data error.
func IsData(err error) bool {
	return ErrorTypeOf(err) == TypeData
}

// IsPanic returns true if this error represents a recovered panic.
func IsPanic(err error) bool {
	return ErrorTypeOf(err) == TypePanic
}

// Common error constructors for convenience

// ConfigError creates a configuration error.
func ConfigError(op, msg string) error {
	return NewTyped(op, msg, TypeConfig)
}

// RenderError creates a rendering error.
func RenderError(op, msg string) error {
	return NewTyped(op, msg, TypeRender)
}

// DataError creates a data source error.
func DataError(op, msg string) error {
	return NewTyped(op, msg, TypeData)
}

// WrapConfig wraps an error as a configuration error.
func WrapConfig(err error, op, msg string) error {
	return WrapTyped(err, op, msg, TypeConfig)
}

// WrapRender wraps an error as a rendering error.
func WrapRender(err error, op, msg string) error {
	return WrapTyped(err, op, msg, TypeRender)
}

// WrapData wraps an error as a data error.
func WrapData(err error, op, msg string) error {
	return WrapTyped(err, op, msg, TypeData)
}

// PanicError creates an error from a recovered panic.
func PanicError(op string, panicValue interface{}) error {
	msg := fmt.Sprintf("panic: %v", panicValue)
	return &TypedError{
		HUDError: &HUDError{
			Op:  op,
			Msg: msg,
		},
		Type: TypePanic,
	}
}
