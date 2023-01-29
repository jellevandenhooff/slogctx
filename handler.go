package slogctx

import (
	"context"

	"golang.org/x/exp/slices"
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

// pendingGroup is a work-in-progress slog.Group attribute.
//
// It is used by ctxHandler to support outputting slogctx.WithAttrs attributes
// at the top-level while still supporting Handler.WithGroup.
type pendingGroup struct {
	name  string
	attrs []slog.Attr
}

// ctxHandler wraps a slog.Handler with support for WithAttrs and
// WithMinimumLevel.
type ctxHandler struct {
	inner slog.Handler

	// groups is a set of pending slog.Group attributes. Each element will
	// become a slog.Group nested in the previous group.
	groups []pendingGroup
}

// WrapWithCtxHandler wraps a slog.Handler with support for WithAttrs
// and WithMinimumLevel.
//
// Use WrapDefaultLoggerWithCtxHandler to wrap the handler used by slog.Default.
func WrapWithCtxHandler(inner slog.Handler) slog.Handler {
	return &ctxHandler{inner: inner}
}

// Enabled implements Handler. It considers a level added to the context with
// WithMinimumLevel.
func (h *ctxHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if ctx != nil {
		if info, ok := ctx.Value(ctxKey{}).(*ctxInfo); ok && info.hasLevel {
			return level >= info.level
		}
	}
	return h.inner.Enabled(ctx, level)
}

// Handle implements Handler. It adds attributes added to the context with
// WithAttrs.
func (h *ctxHandler) Handle(r slog.Record) error {
	if h.groups != nil {
		last := h.groups[len(h.groups)-1]
		attrs := make([]slog.Attr, len(last.attrs)+r.NumAttrs())
		copy(attrs, last.attrs)
		i := len(last.attrs)
		r.Attrs(func(a slog.Attr) {
			attrs[i] = a
			i++
		})
		attr := slog.Group(last.name, attrs...)
		for i := len(h.groups) - 2; i >= 0; i-- {
			cur := h.groups[i]
			attrs := make([]slog.Attr, len(cur.attrs)+1)
			copy(attrs, cur.attrs)
			attrs[len(cur.attrs)] = attr
			attr = slog.Group(cur.name, attrs...)
		}
		r = slog.NewRecord(r.Time, r.Level, r.Message, r.PC, r.Context)
		r.AddAttrs(attr)
	}

	if r.Context != nil {
		if info, ok := r.Context.Value(ctxKey{}).(*ctxInfo); ok {
			r.AddAttrs(info.attrs...)
		}
	}
	return h.inner.Handle(r)
}

// WithAttrs implements Handler. It forwards directly to the original handler if h.groups is nil.
func (h *ctxHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if h.groups == nil {
		return &ctxHandler{inner: h.inner.WithAttrs(attrs), groups: nil}
	} else {
		cur := h.groups[len(h.groups)-1]
		newAttrs := make([]slog.Attr, len(cur.attrs)+len(attrs))
		copy(newAttrs, cur.attrs)
		copy(newAttrs[len(cur.attrs):], attrs)
		newGroups := slices.Clone(h.groups)
		newGroups[len(newGroups)-1].attrs = newAttrs
		return &ctxHandler{inner: h.inner, groups: newGroups}
	}
}

// WithGroup implements Handler.
func (h *ctxHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]pendingGroup, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(newGroups)-1].name = name
	return &ctxHandler{inner: h.inner, groups: newGroups}
}

// WithAttrs attaches the given attributes (as in slog.Logger.With) to the
// context.
//
// Requires a slog.Handler wrapped with WrapWithCtxHandler.
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
// Requires a slog.Handler wrapped with WrapWithCtxHandler.
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
// WrapWithCtxHandler.
func WrapDefaultLoggerWithCtxHandler() {
	slog.SetDefault(slog.New(WrapWithCtxHandler(slog.Default().Handler())))
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
