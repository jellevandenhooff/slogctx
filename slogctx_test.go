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

func TestSlogctx(t *testing.T) {
	check := setTestDefault(t)

	// setup slogctx
	slogctx.WrapDefaultLoggerWithNewCtxHandler()

	// attaching attributes to context
	ctx := context.Background()
	ctx = slogctx.WithAttrs(ctx, "hello", "why not")

	// global handler using default logger
	slogctx.Info(ctx, "doing a small log")
	check(`level=INFO msg="doing a small log" hello="why not"`)

	slogctx.Error(ctx, "encountered an error", os.ErrClosed)
	check(`level=ERROR msg="encountered an error" err="file already closed" hello="why not"`)

	// converting a logger
	logger := slogctx.NewLogger(slog.Default())
	logger.Info(ctx, "hey", "count", 10)
	check(`level=INFO msg=hey count=10 hello="why not"`)
}

// copied/modified from golang.org/x/exp/slog/logger_test.go:

const timeRE = `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}(Z|[+-]\d{2}:\d{2})`

func setTestDefault(t *testing.T) func(want string) {
	var buf bytes.Buffer

	l := slog.New(slog.NewTextHandler(&buf))

	check := func(want string) {
		t.Helper()
		if want != "" {
			want = "time=" + timeRE + " " + want
		}
		checkLogOutput(t, buf.String(), want)
		buf.Reset()
	}

	before := slog.Default()
	slog.SetDefault(l)
	t.Cleanup(func() {
		slog.SetDefault(before)
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
