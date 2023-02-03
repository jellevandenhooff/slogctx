package slogctx_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jellevandenhooff/slogctx"
	"golang.org/x/exp/slog"
)

func NewRequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.NewString()
		ctx := r.Context()
		ctx = slogctx.WithAttrs(ctx, "requestID", requestID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

type WebServer struct {
	// WebServer always has a context when logging. Use a *slogctx.Logger for convenience.
	logger *slogctx.Logger

	db *DB
}

func (s *WebServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	s.logger.Info(ctx, "got messages request", "url", r.URL.String(), "remoteAddr", r.RemoteAddr)

	response, err := s.db.QueryMessages(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(response)
}

type DB struct {
	// DB sometimes has a context when logging, sometimes does not. Use a *slog.Logger.
	logger *slog.Logger
}

func (db *DB) QueryMessages(ctx context.Context) ([]string, error) {
	db.logger.WithContext(ctx).Debug("querying db")

	return []string{"hello world", "another message"}, nil
}

func (db *DB) PrintStats() {
	t := time.NewTicker(10 * time.Second)

	for range t.C {
		db.logger.Info("stats", "numDBConnections", 10)
	}
}

func Example_whole_program() {
	// This example shows a way of passing around *slog(ctx).Loggers as well and
	// extra attributes (or trace IDs) with a context.
	//
	// At program start time, the main function sets up the logger(s), DB
	// clients, server structs, and glues them altogether.
	//
	// At runtime, handlers pass around context.Context which gets passed down
	// to the loggers.

	// Setup slog
	slog.SetDefault(slog.New(slog.HandlerOptions{
		AddSource: true,
	}.NewTextHandler(os.Stderr)))
	slogctx.WrapDefaultLoggerWithCtxHandler()

	// Build DB client.
	db := &DB{
		logger: slog.Default().With("dbURL", "sqlite://foo"),
	}
	go db.PrintStats()

	// Build web server.
	webServer := &WebServer{
		logger: slogctx.NewLogger(slog.Default()),
		db:     db,
	}

	// Wrap web server with middleware.
	handler := NewRequestIDMiddleware(webServer)

	// Run the server.
	http.ListenAndServe(":8080", handler)
}
