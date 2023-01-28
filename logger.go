package slogctx

import (
	"context"

	"golang.org/x/exp/slog"
)

// Logger is a slog.Logger wrapper with a mandatory context argument.
type Logger struct {
	inner slog.Logger
}

// NewLogger creates a new Logger from a slog.Logger.
func NewLogger(logger *slog.Logger) *Logger {
	return &Logger{
		inner: *logger,
	}
}

// Enabled reports whether l emits log records at the given level.
func (l *Logger) Enabled(ctx context.Context, level slog.Level) bool {
	return l.inner.WithContext(ctx).Enabled(level)
}

// With returns a new Logger that includes the given arguments, like slog.Logger.With.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		inner: *l.inner.With(args...),
	}
}

// Debug logs at LevelDebug.
func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.inner.WithContext(ctx).LogDepth(2, slog.LevelDebug, msg, args...)
}

// Info logs at LevelInfo.
func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.inner.WithContext(ctx).LogDepth(2, slog.LevelInfo, msg, args...)
}

// Warn logs at LevelWarn.
func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.inner.WithContext(ctx).LogDepth(2, slog.LevelWarn, msg, args...)
}

// Error logs at LevelError.
// If err is non-nil, Error appends Any(ErrorKey, err)
// to the list of attributes.
func (l *Logger) Error(ctx context.Context, msg string, err error, args ...any) {
	if err != nil {
		args = append(args, slog.Any(slog.ErrorKey, err))
	}
	l.inner.WithContext(ctx).LogDepth(2, slog.LevelError, msg, args...)
}

// Log emits a log record, like slog.Logger.Log.
func (l *Logger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.inner.WithContext(ctx).LogDepth(2, level, msg, args...)
}

// Debug calls Logger.WithContext(ctx).Debug on the default logger.
func Debug(ctx context.Context, msg string, args ...any) {
	slog.Default().WithContext(ctx).LogDepth(2, slog.LevelDebug, msg, args...)
}

// Debug calls Logger.WithContext(ctx).Info on the default logger.
func Info(ctx context.Context, msg string, args ...any) {
	slog.Default().WithContext(ctx).LogDepth(2, slog.LevelInfo, msg, args...)
}

// Debug calls Logger.WithContext(ctx).Warn on the default logger.
func Warn(ctx context.Context, msg string, args ...any) {
	slog.Default().WithContext(ctx).LogDepth(2, slog.LevelWarn, msg, args...)
}

// Debug calls Logger.WithContext(ctx).Error on the default logger.
func Error(ctx context.Context, msg string, err error, args ...any) {
	if err != nil {
		args = append(args, slog.Any(slog.ErrorKey, err))
	}
	slog.Default().WithContext(ctx).LogDepth(2, slog.LevelError, msg, args...)
}
