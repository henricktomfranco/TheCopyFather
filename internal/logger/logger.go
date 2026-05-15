package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Logger provides structured logging capabilities
type Logger struct {
	ctx context.Context
}

// NewLogger creates a new Logger instance
func NewLogger(ctx context.Context) *Logger {
	return &Logger{ctx: ctx}
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	runtime.LogInfo(l.ctx, message)
	// Also print to console for development
	fmt.Printf("[INFO] %s - %s\n", time.Now().Format(time.RFC3339), message)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	runtime.LogError(l.ctx, message)
	// Also print to console for development
	fmt.Printf("[ERROR] %s - %s\n", time.Now().Format(time.RFC3339), message)
}

// Warning logs a warning message
func (l *Logger) Warning(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	runtime.LogWarning(l.ctx, message)
	// Also print to console for development
	fmt.Printf("[WARN] %s - %s\n", time.Now().Format(time.RFC3339), message)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	runtime.LogDebug(l.ctx, message)
	// Also print to console for development
	fmt.Printf("[DEBUG] %s - %s\n", time.Now().Format(time.RFC3339), message)
}