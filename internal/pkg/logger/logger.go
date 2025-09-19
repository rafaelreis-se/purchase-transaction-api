package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// Logger levels
const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"` // "json" or "text"
}

// NewLogger creates a new structured logger
func NewLogger(cfg LoggerConfig) *Logger {
	var level slog.Level
	switch strings.ToUpper(cfg.Level) {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true, // Add file:line information
	}

	if strings.ToLower(cfg.Format) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	return &Logger{Logger: logger}
}

// WithContext adds context to logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{Logger: l.Logger.With("context", ctx)}
}

// WithField adds a single field to logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{Logger: l.Logger.With(key, value)}
}

// WithFields adds multiple fields to logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{Logger: l.Logger.With(args...)}
}

// Request logging helper
func (l *Logger) LogRequest(method, path, userAgent, clientIP string, statusCode int, duration string) {
	l.Info("Request completed",
		"method", method,
		"path", path,
		"status_code", statusCode,
		"duration", duration,
		"user_agent", userAgent,
		"client_ip", clientIP,
	)
}

// Error logging with stack trace context
func (l *Logger) LogError(err error, msg string, fields ...interface{}) {
	args := []interface{}{"error", err.Error()}
	args = append(args, fields...)
	l.Error(msg, args...)
}

// Business operation logging
func (l *Logger) LogOperation(operation string, entityID string, success bool, fields ...interface{}) {
	args := []interface{}{
		"operation", operation,
		"entity_id", entityID,
		"success", success,
	}
	args = append(args, fields...)

	if success {
		l.Info("Operation completed successfully", args...)
	} else {
		l.Warn("Operation failed", args...)
	}
}

// External service call logging
func (l *Logger) LogExternalCall(service, endpoint string, duration string, statusCode int, success bool) {
	l.Info("External service call",
		"service", service,
		"endpoint", endpoint,
		"duration", duration,
		"status_code", statusCode,
		"success", success,
	)
}
