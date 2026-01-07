package errors

import (
	"encoding/json"
	errs "errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// As is a re-export of errors.As for convenience within this package.
func As(err error, target interface{}) bool {
	return errs.As(err, target)
}

// LogLevel represents the severity level of a log message.
type LogLevel int

const (
	// LevelDebug is for detailed debugging information.
	LevelDebug LogLevel = iota
	// LevelInfo is for general informational messages.
	LevelInfo
	// LevelWarn is for warning messages that don't require immediate action.
	LevelWarn
	// LevelError is for error messages.
	LevelError
)

// String returns the string representation of the log level.
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ANSI color codes for terminal output.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
)

// colorForLevel returns the ANSI color code for a given log level.
func colorForLevel(level LogLevel) string {
	switch level {
	case LevelDebug:
		return colorGray
	case LevelInfo:
		return colorGreen
	case LevelWarn:
		return colorYellow
	case LevelError:
		return colorRed
	default:
		return colorReset
	}
}

// Logger is a thread-safe logger with configurable levels and output.
type Logger struct {
	mu       sync.Mutex
	level    LogLevel
	output   io.Writer
	debug    bool
	useColor bool
}

// NewLogger creates a new logger with the specified configuration.
func NewLogger(level LogLevel, debug bool) *Logger {
	return &Logger{
		level:    level,
		output:   os.Stderr,
		debug:    debug,
		useColor: isTerminal(os.Stderr),
	}
}

// isTerminal checks if the writer is a terminal.
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		stat, _ := f.Stat()
		return (stat.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

// SetLevel changes the minimum log level.
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetOutput changes the output writer.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
	l.useColor = isTerminal(w)
}

// SetDebug toggles debug mode.
func (l *Logger) SetDebug(debug bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debug = debug
	if debug && l.level > LevelDebug {
		l.level = LevelDebug
	}
}

// shouldLog returns true if a message at the given level should be logged.
func (l *Logger) shouldLog(level LogLevel) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return level >= l.level
}

// formatMessage formats a log message with timestamp and context.
func (l *Logger) formatMessage(level LogLevel, op string, msg string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Format the message with arguments if provided
	message := msg
	if len(args) > 0 {
		message = fmt.Sprintf(msg, args...)
	}

	var colorStart string
	if l.useColor {
		colorStart = colorForLevel(level)
	}

	// Format: [timestamp] LEVEL [operation] message
	formatted := fmt.Sprintf("%s[%s]%s %s%-6s%s ",
		colorGray, timestamp, colorReset,
		colorStart, level.String(), colorReset,
	)

	if op != "" {
		formatted += fmt.Sprintf("[%s] ", op)
	}

	formatted += message

	return formatted
}

// log writes a log message at the specified level.
func (l *Logger) log(level LogLevel, op string, msg string, args ...interface{}) {
	if !l.shouldLog(level) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	message := l.formatMessage(level, op, msg, args...)
	fmt.Fprintln(l.output, message)
}

// logDirect writes a pre-formatted log message at the specified level.
// Use this when the message is already formatted or comes from user input.
func (l *Logger) logDirect(level LogLevel, op string, message string) {
	if !l.shouldLog(level) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	formatted := l.formatMessage(level, op, "%s", message)
	fmt.Fprintln(l.output, formatted)
}

// Debug logs a debug message.
func (l *Logger) Debug(op, msg string, args ...interface{}) {
	l.log(LevelDebug, op, msg, args...)
}

// Info logs an informational message.
func (l *Logger) Info(op, msg string, args ...interface{}) {
	l.log(LevelInfo, op, msg, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(op, msg string, args ...interface{}) {
	l.log(LevelWarn, op, msg, args...)
}

// Error logs an error message.
func (l *Logger) Error(op, msg string, args ...interface{}) {
	l.log(LevelError, op, msg, args...)
}

// Errorf logs an error message with formatting.
func (l *Logger) Errorf(op, format string, args ...interface{}) {
	l.log(LevelError, op, format, args...)
}

// LogError logs an error with its context.
func (l *Logger) LogError(err error) {
	if err == nil {
		return
	}

	op := ErrorOp(err)
	msg := err.Error()

	var hudErr *HUDError
	if As(err, &hudErr) && hudErr.Err != nil {
		// Log the full error chain
		l.logDirect(LevelError, op, msg)
		// Optionally log stack trace for errors
		if l.debug {
			l.log(LevelDebug, op, "stack trace: %s", StackTrace())
		}
	} else {
		l.logDirect(LevelError, op, msg)
	}
}

// LogErrorWithLevel logs an error at the specified level based on error type.
func (l *Logger) LogErrorWithLevel(err error) {
	if err == nil {
		return
	}

	op := ErrorOp(err)
	msg := err.Error()

	// Determine log level based on error type
	level := LevelError
	errType := ErrorTypeOf(err)

	switch errType {
	case TypeRender:
		// Render errors are warnings - the app continues
		level = LevelWarn
	case TypeData:
		// Data errors are warnings - show placeholder
		level = LevelWarn
	case TypePanic:
		// Panics are always errors
		level = LevelError
	case TypeConfig:
		// Config errors are errors
		level = LevelError
	}

	l.logDirect(level, op, msg)

	// Log stack trace for panics or in debug mode
	if errType == TypePanic || l.debug {
		if panicErr, ok := err.(*TypedError); ok && panicErr.Type == TypePanic {
			l.log(LevelError, op, "recovered from panic, stack trace: %s", StackTrace())
		}
	}
}

// LogJSON logs a message as JSON for structured logging.
func (l *Logger) LogJSON(level LogLevel, op string, data map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     level.String(),
		"operation": op,
	}

	for k, v := range data {
		entry[k] = v
	}

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(l.output, "{\"error\":\"failed to marshal log entry\"}\n")
		return
	}

	fmt.Fprintln(l.output, string(jsonBytes))
}

// Global logger instance.
var globalLogger = NewLogger(LevelInfo, false)

// SetGlobalLogger sets the global logger instance.
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance.
func GetGlobalLogger() *Logger {
	return globalLogger
}

// Convenience functions using the global logger.

// Debug logs a debug message to the global logger.
func Debug(op, msg string, args ...interface{}) {
	globalLogger.Debug(op, msg, args...)
}

// Info logs an info message to the global logger.
func Info(op, msg string, args ...interface{}) {
	globalLogger.Info(op, msg, args...)
}

// Warn logs a warning message to the global logger.
func Warn(op, msg string, args ...interface{}) {
	globalLogger.Warn(op, msg, args...)
}

// Error logs an error message to the global logger.
func Error(op, msg string, args ...interface{}) {
	globalLogger.Error(op, msg, args...)
}

// LogError logs an error to the global logger.
func LogError(err error) {
	globalLogger.LogError(err)
}

// LogErrorWithLevel logs an error with appropriate level to the global logger.
func LogErrorWithLevel(err error) {
	globalLogger.LogErrorWithLevel(err)
}

// SetDebugMode enables or disables debug mode globally.
func SetDebugMode(debug bool) {
	globalLogger.SetDebug(debug)
	if debug {
		globalLogger.SetLevel(LevelDebug)
	}
}
