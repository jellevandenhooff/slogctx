// Package slogctx provides tools for working with contexts with the proposed
// log/slog package.
//
// The package provides alternatives to slog logging functions that require a
// context with a logs: slogctx.Info replaces slog.Info, slogctx.Error replaces
// slog.Error, etc.  The struct *slogctx.Logger replaces *slog.Logger. This is
// useful to ensure a context is always included. Usage:
//
//	slogctx.Info(ctx, "got a request")
//	logger := slogctx.Default() // or logger := slogctx.NewLogger(slog.Default())
//	logger.Info(ctx, "found something special")
//
// The package supports storing extra attributes in a context. All logs using
// the context created by slogctx.WithAttrs will include the extra attributes.
// This is useful to include a requestID with all logs. Usage:
//
//	ctx = slogctx.WithAttrs(ctx, "requestID", 1234)
//	slogctx.Info(ctx, "processing request")
//	slog.Default().WithContext(ctx).Info("processing more") // also works with plain slog
//
// The package supports overriding the minimum log level. All logs using the
// context created by slogctx.WithMinimumLevel will use the supplied level. This
// is useful to debug specific requests. Usage:
//
//	ctx = slogctx.WithMinimumLevel(ctx, slog.LevelDebug)
//	slogctx.Debug(ctx, "low-level information")
//
// Using WithAttrs and WithMinimumLevel requires wrapping the underlying
// slog.Handler using slogctx.CtxHandler. This can be done globally for the
// default logger using slogctx.WrapDefaultLoggerWithCtxHandler.
package slogctx
