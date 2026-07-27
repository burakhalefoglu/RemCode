// RemLinkAgent — AI coding agent for your machine, driven from your phone
// Copyright (C) 2026 Burak Halefoğlu
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Command rla-server is the RemLinkAgent relay.
//
// It moves events between a CLI host and its paired devices when they cannot
// reach each other directly. It is explicitly untrusted: payloads are
// end-to-end encrypted (ADR-004) and AI traffic never passes through it at all.
//
// This is the P0 skeleton — it serves health endpoints so deployment plumbing
// can be built and verified. The WebSocket gateway lands in P2.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/burakhalefoglu/RemLinkAgent/internal/version"
)

const (
	readHeaderTimeout = 10 * time.Second
	shutdownTimeout   = 15 * time.Second
)

func main() {
	addr := flag.String("addr", envOr("RLA_ADDR", ":8080"), "listen address")
	logLevel := flag.String("log-level", envOr("RLA_LOG_LEVEL", "info"), "debug|info|warn|error")
	flag.Parse()

	logger := newLogger(*logLevel)
	slog.SetDefault(logger)

	if err := run(*addr, logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(addr string, logger *slog.Logger) error {
	info := version.Get()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("GET /version", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(info); err != nil {
			logger.Error("encode version response", "error", err)
		}
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout, // slowloris guard
	}

	// Signal-driven shutdown, so `docker compose down` and Ctrl-C both drain
	// cleanly instead of severing connections.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("relay listening",
			"addr", addr,
			"version", info.Version,
			"protocol", info.Protocol,
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("listen: %w", err)
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received", "timeout", shutdownTimeout)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	logger.Info("relay stopped")
	return nil
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	// Fail loud on a bad value rather than silently defaulting: a typo in a
	// deployment config should be visible, not swallowed.
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		fmt.Fprintf(os.Stderr, "invalid -log-level %q: want debug|info|warn|error\n", level)
		os.Exit(2)
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
