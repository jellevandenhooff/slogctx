package slogctx

import (
	"context"

	"golang.org/x/exp/slog"
)

// ctxKey is the context key used by CtxHandler.
type ctxKey struct{}

// ctxInfo is the info stored in the context for CtxHandler.
type ctxInfo struct {
	attrs []slog.Attr

	hasLevel bool
	level    slog.Level
}

// CtxHandler wraps a slog.Handler with support for WithAttrs and
// WithMinimumLevel.
type CtxHandler struct {
	Inner slog.Handler
}

// NewCtxHandler creates a CtxHandler wrapping the given slog.Handler.
func NewCtxHandler(inner slog.Handler) *CtxHandler {
	return &CtxHandler{Inner: inner}
}

// Enabled implements Handler. It considers a level added to the context with
// WithMinimumLevel.
func (h *CtxHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if ctx != nil {
		if info, ok := ctx.Value(ctxKey{}).(*ctxInfo); ok && info.hasLevel {
			return level >= info.level
		}
	}
	return h.Inner.Enabled(ctx, level)
}

// Handle implements Handler. It adds attributes added to the context with
// WithAttrs.
func (h CtxHandler) Handle(r slog.Record) error {
	if r.Context != nil {
		if info, ok := r.Context.Value(ctxKey{}).(*ctxInfo); ok {
			r.AddAttrs(info.attrs...)
		}
	}
	return h.Inner.Handle(r)
}

// WithAttrs implements Handler. It forwards directly to the original handler.
func (h *CtxHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CtxHandler{Inner: h.Inner.WithAttrs(attrs)}
}

// WithGroup implements Handler. It forwards directly to the original handler.
func (h CtxHandler) WithGroup(name string) slog.Handler {
	return &CtxHandler{Inner: h.Inner.WithGroup(name)}
}

// WithAttrs attaches the given attributes (as in slog.Logger.With) to the
// context.
//
// Requires a slog.Handler wrapped with NewCtxHandler.
func WithAttrs(ctx context.Context, args ...any) context.Context {
	newAttrs := argsToAttrs(args)
	var newInfo ctxInfo
	if info, ok := ctx.Value(ctxKey{}).(*ctxInfo); ok {
		newInfo.attrs = make([]slog.Attr, len(info.attrs)+len(newAttrs))
		copy(newInfo.attrs, info.attrs)
		copy(newInfo.attrs[len(info.attrs):], newAttrs)
		newInfo.hasLevel = info.hasLevel
		newInfo.level = info.level
	} else {
		newInfo.attrs = newAttrs
	}
	return context.WithValue(ctx, ctxKey{}, &newInfo)
}

// WithMinimumLevel overrides the minimum logging level for all log calls using
// this context.
//
// Requires a slog.Handler wrapped with NewCtxHandler.
func WithMinimumLevel(ctx context.Context, level slog.Level) context.Context {
	var newInfo ctxInfo
	if info, ok := ctx.Value(ctxKey{}).(*ctxInfo); ok {
		newInfo.attrs = info.attrs
	}
	newInfo.hasLevel = true
	newInfo.level = level
	return context.WithValue(ctx, ctxKey{}, &newInfo)
}

// WrapDefaultLoggerWithCtxHandler wraps the handler used by slog.Default() with
// CtxHandler.
func WrapDefaultLoggerWithCtxHandler() {
	slog.SetDefault(slog.New(NewCtxHandler(slog.Default().Handler())))
}

// copied/modified from golang.org/x/exp/slog/record.go:

// argsToAttr turns a prefix of the nonempty args slice into an Attr
// and returns the unconsumed portion of the slice.
// If args[0] is an Attr, it returns it.
// If args[0] is a string, it treats the first two elements as
// a key-value pair.
// Otherwise, it treats args[0] as a value with a missing key.
func argsToAttr(args []any) (slog.Attr, []any) {
	const badKey = "!BADKEY"

	switch x := args[0].(type) {
	case string:
		if len(args) == 1 {
			return slog.String(badKey, x), nil
		}
		return slog.Any(x, args[1]), args[2:]

	case slog.Attr:
		return x, args[1:]

	default:
		return slog.Any(badKey, x), args[1:]
	}
}

func argsToAttrs(args []any) []slog.Attr {
	var attrs []slog.Attr
	for len(args) > 0 {
		var attr slog.Attr
		attr, args = argsToAttr(args)
		attrs = append(attrs, attr)
	}
	return attrs
}
