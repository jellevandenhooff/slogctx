package slogctx_test

import (
	"bytes"
	"context"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/jellevandenhooff/slogctx"
	"golang.org/x/exp/slog"
)

func TestLoggerEnabled(t *testing.T) {
	_ = setupTestSlogHandler(t, slog.HandlerOptions{})

	// setup slogctx
	slogctx.WrapDefaultLoggerWithCtxHandler()

	ctx := context.Background()

	logger := slogctx.Default()
	if logger.Enabled(ctx, slog.LevelDebug) {
		t.Error("expected DEBUG to be disabled")
	}
	if !logger.Enabled(ctx, slog.LevelInfo) {
		t.Error("expected INFO to be enabled")
	}

	overrideLevelCtx := slogctx.WithMinimumLevel(ctx, slog.LevelDebug)
	if !logger.Enabled(overrideLevelCtx, slog.LevelDebug) {
		t.Error("expected DEBUG to be enabled after override")
	}
}

func TestWithAttrs(t *testing.T) {
	check := setupTestSlogHandler(t, slog.HandlerOptions{})

	// setup slogctx
	slogctx.WrapDefaultLoggerWithCtxHandler()

	ctx := context.Background()
	ctx = slogctx.WithAttrs(ctx, "attr", 1, "buz", "boo")

	slogctx.Info(ctx, "hi")
	check(`level=INFO msg=hi attr=1 buz=boo`)

	extraAttrCtx := slogctx.WithAttrs(ctx, "extra", "foo")
	slogctx.Info(extraAttrCtx, "hi")
	check(`level=INFO msg=hi attr=1 buz=boo extra=foo`)

	slogctx.Info(ctx, "orig")
	check(`level=INFO msg=orig attr=1 buz=boo`)

	anotherAttrCtx := slogctx.WithAttrs(ctx, "extra", "two", "third", 3)
	slogctx.Info(anotherAttrCtx, "hi")
	check(`level=INFO msg=hi attr=1 buz=boo extra=two third=3`)

	slogctx.Info(extraAttrCtx, "back to foo")
	check(`level=INFO msg="back to foo" attr=1 buz=boo extra=foo`)

	slogctx.Debug(ctx, "ignored")
	check(``)

	badStrKeyCtx := slogctx.WithAttrs(ctx, "help")
	slogctx.Info(badStrKeyCtx, "help")
	check(`level=INFO msg=help attr=1 buz=boo !BADKEY=help`)

	badValKeyCtx := slogctx.WithAttrs(ctx, 1)
	slogctx.Info(badValKeyCtx, "help")
	check(`level=INFO msg=help attr=1 buz=boo !BADKEY=1`)

	attrCtx := slogctx.WithAttrs(ctx, slog.String("attr", "str"))
	slogctx.Info(attrCtx, "attr")
	check(`level=INFO msg=attr attr=1 buz=boo attr=str`)
}

func TestWithMinimumLevel(t *testing.T) {
	check := setupTestSlogHandler(t, slog.HandlerOptions{})

	// setup slogctx
	slogctx.WrapDefaultLoggerWithCtxHandler()

	ctx := context.Background()

	slogctx.Info(ctx, "hi")
	check(`level=INFO msg=hi`)

	slogctx.Debug(ctx, "hi")
	check(``)

	overrideLevelCtx := slogctx.WithMinimumLevel(ctx, slog.LevelDebug)
	slogctx.Debug(overrideLevelCtx, "hi")
	check(`level=DEBUG msg=hi`)

	slogctx.Debug(ctx, "still not")
	check(``)

	overrideLevelAgainCtx := slogctx.WithMinimumLevel(overrideLevelCtx, slog.LevelError)
	slogctx.Debug(overrideLevelAgainCtx, "no")
	check(``)

	slogctx.Info(overrideLevelAgainCtx, "also no")
	check(``)

	slogctx.Error(overrideLevelAgainCtx, "yes", os.ErrClosed)
	check(`level=ERROR msg=yes err="file already closed"`)

	slogctx.Debug(ctx, "still not")
	check(``)

	slogctx.Info(ctx, "still yes")
	check(`level=INFO msg="still yes"`)

	slogctx.Debug(overrideLevelCtx, "still yes also")
	check(`level=DEBUG msg="still yes also"`)
}

func TestWithAttrsAndMinimumLevel(t *testing.T) {
	check := setupTestSlogHandler(t, slog.HandlerOptions{})

	// setup slogctx
	slogctx.WrapDefaultLoggerWithCtxHandler()

	ctx := context.Background()

	slogctx.Info(ctx, "hi")
	check(`level=INFO msg=hi`)
	slogctx.Debug(ctx, "hi")
	check(``)

	ctx = slogctx.WithAttrs(slogctx.WithMinimumLevel(context.Background(), slog.LevelDebug), "hello", "foo")
	slogctx.Debug(ctx, "hi")
	check(`level=DEBUG msg=hi hello=foo`)

	ctx = slogctx.WithMinimumLevel(slogctx.WithAttrs(context.Background(), "hello", "foo"), slog.LevelDebug)
	slogctx.Debug(ctx, "hi")
	check(`level=DEBUG msg=hi hello=foo`)
}

// TestWrapperSourceAndContext verifies that the context is forwarded and
// correct source line is printed with all wrappers.
func TestWrapperSourceAndContext(t *testing.T) {
	check := setupTestSlogHandler(t, slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	})

	// setup slogctx
	slogctx.WrapDefaultLoggerWithCtxHandler()

	ctx := context.Background()
	ctx = slogctx.WithAttrs(ctx, "hi", "there")

	// logger methods
	logger := slogctx.Default()
	logger.Debug(ctx, "hello")
	check(`level=DEBUG source=.*/slogctx_test.go:.* msg=hello hi=there`)

	logger.Info(ctx, "hello")
	check(`level=INFO source=.*/slogctx_test.go:.* msg=hello hi=there`)

	logger.Warn(ctx, "hello")
	check(`level=WARN source=.*/slogctx_test.go:.* msg=hello hi=there`)

	logger.Error(ctx, "hello", os.ErrClosed)
	check(`level=ERROR source=.*/slogctx_test.go:.* msg=hello err="file already closed" hi=there`)

	logger.Log(ctx, slog.LevelDebug, "hello")
	check(`level=DEBUG source=.*/slogctx_test.go:.* msg=hello hi=there`)

	// top-level functions
	slogctx.Debug(ctx, "hello")
	check(`level=DEBUG source=.*/slogctx_test.go:.* msg=hello hi=there`)

	slogctx.Info(ctx, "hello")
	check(`level=INFO source=.*/slogctx_test.go:.* msg=hello hi=there`)

	slogctx.Warn(ctx, "hello")
	check(`level=WARN source=.*/slogctx_test.go:.* msg=hello hi=there`)

	slogctx.Error(ctx, "hello", os.ErrClosed)
	check(`level=ERROR source=.*/slogctx_test.go:.* msg=hello err="file already closed" hi=there`)
}

// copied/modified from golang.org/x/exp/slog/logger_test.go:

const timeRE = `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}(Z|[+-]\d{2}:\d{2})`

func setupTestSlogHandler(t *testing.T, opts slog.HandlerOptions) func(want string) {
	var buf bytes.Buffer

	l := slog.New(opts.NewTextHandler(&buf))

	check := func(want string) {
		t.Helper()
		if want != "" {
			want = "time=" + timeRE + " " + want
		}
		checkLogOutput(t, buf.String(), want)
		buf.Reset()
	}

	original := slog.Default()
	slog.SetDefault(l)
	t.Cleanup(func() {
		slog.SetDefault(original)
	})
	return check
}

// clean prepares log output for comparison.
func clean(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	return strings.ReplaceAll(s, "\n", "~")
}

func checkLogOutput(t *testing.T, got, wantRegexp string) {
	t.Helper()
	got = clean(got)
	wantRegexp = "^" + wantRegexp + "$"
	matched, err := regexp.MatchString(wantRegexp, got)
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Errorf("\ngot  %s\nwant %s", got, wantRegexp)
	}
}
