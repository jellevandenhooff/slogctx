package slogctx

import (
	"context"

	"golang.org/x/exp/slog"
)

// ctxKey is the context key for ctxInfo
type ctxKey struct{}

// ctxInfo is the info stored in the context for ctxHandler
type ctxInfo struct {
	attrs []slog.Attr

	hasLevel bool
	level    slog.Level
}

// ctxHandler wraps a slog.Handler and includes ctxInfo when handling records
type ctxHandler struct {
	inner slog.Handler
}

// NewCtxHandler wraps a slog.Handler to include attributes and use a possible
// override level from the context when logging.
//
// Use WrapDefaultLoggerWithNewCtxHandler as a convenient shortcut to modify slog.Default().
func NewCtxHandler(handler slog.Handler) slog.Handler {
	return ctxHandler{inner: handler}
}

func (h ctxHandler) Enabled(level slog.Level) bool {
	/*
		// pending https://go-review.googlesource.com/c/exp/+/463935
		if ctx != nil {
			if info, ok := ctx.Value(ctxKey{}).(*ctxInfo); ok && info.hasLevel {
				return level >= info.level
			}
		}
	*/
	return h.inner.Enabled(level)
}

func (h ctxHandler) Handle(r slog.Record) error {
	if r.Context != nil {
		if info, ok := r.Context.Value(ctxKey{}).(*ctxInfo); ok {
			r.AddAttrs(info.attrs...)
		}
	}
	return h.inner.Handle(r)
}

func (h ctxHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return ctxHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h ctxHandler) WithGroup(name string) slog.Handler {
	return ctxHandler{inner: h.inner.WithGroup(name)}
}

// WrapDefaultLoggerWithNewCtxHandler wraps the handler used by slog.Default() with NewCtxHandler.
func WrapDefaultLoggerWithNewCtxHandler() {
	slog.SetDefault(slog.New(NewCtxHandler(slog.Default().Handler())))
}

// WithAttrs attaches the given attributes (as in slog.Logger.With) to the context.
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

/*
// pending https://go-review.googlesource.com/c/exp/+/463935

// WithMinimumLevel overrides the minimum logging level for all log calls using this
// context.
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
*/

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
