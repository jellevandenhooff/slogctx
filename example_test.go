package slogctx_test

import (
	"context"
	"os"
	"time"

	"github.com/jellevandenhooff/slogctx"
	"golang.org/x/exp/slog"
)

func SetupSlogForExample() func() {
	original := slog.Default()

	slog.SetDefault(slog.New(slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "time" {
				return slog.Time("time", time.Date(2022, 1, 29, 15, 10, 0, 0, time.UTC))
			}
			return a
		},
	}.NewTextHandler(os.Stdout)))

	return func() {
		slog.SetDefault(original)
	}
}

func Example() {
	cleanup := SetupSlogForExample()
	defer cleanup()

	ctx := context.Background()

	// Setup slogctx for slog.Default(). This is required to make slog.WithAttrs
	// and slog.WithMinimumLevel work.
	slogctx.WrapDefaultLoggerWithCtxHandler()

	// The log functions slogctx.Info, slogctx.Error, etc. are alternatives to
	// slog.Info, slog.Error, etc. that take a context as first argument.
	// This can be useful to ensure all logger calls will supply a context.
	slogctx.Info(ctx, "doing a small log")

	// The slogctx.Logger is an alternative to slog.Logger where the log
	// functions take a context as first argument.
	// This can be useful to ensure all logger calls will supply a context.
	logger := slogctx.NewLogger(slog.Default())
	logger.Info(ctx, "from the logger")
	logger = logger.With("loggerAttr", "extra attr")
	logger.Info(ctx, "extra logger attr")

	// Using WithAttrs extra attributes can be included in a context.
	// This can be useful to include eg. a request ID with all future logs.
	ctx = slogctx.WithAttrs(ctx, "requestID", "1234")
	slogctx.Info(ctx, "extra context attr")

	// WithAttrs also works with the standard slog log functions.
	slog.Default().WithContext(ctx).Info("from default logger")

	// Using WithMinimumLevel the default log level can be overriden.
	// This can be useful to trace in detail what happens with a specific request.
	slogctx.Debug(ctx, "log ignored")
	ctx = slogctx.WithMinimumLevel(ctx, slog.LevelDebug)
	slogctx.Debug(ctx, "log not ignored")

	// Output:
	// time=2022-01-29T15:10:00.000Z level=INFO msg="doing a small log"
	// time=2022-01-29T15:10:00.000Z level=INFO msg="from the logger"
	// time=2022-01-29T15:10:00.000Z level=INFO msg="extra logger attr" loggerAttr="extra attr"
	// time=2022-01-29T15:10:00.000Z level=INFO msg="extra context attr" requestID=1234
	// time=2022-01-29T15:10:00.000Z level=INFO msg="from default logger" requestID=1234
	// time=2022-01-29T15:10:00.000Z level=DEBUG msg="log not ignored" requestID=1234
}
