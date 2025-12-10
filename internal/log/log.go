// Package log provides a thin wrapper over log/slog tailored for an MCP server.
//
// A server that speaks MCP over stdio must never write logs to stdout, which is
// reserved for protocol frames. The constructors here default to stderr and make
// that constraint explicit.
package log

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Logger is the structured logger used across the server.
type Logger = slog.Logger

// Format selects the on-disk encoding of log records.
type Format string

const (
	// FormatText emits human-readable key=value lines.
	FormatText Format = "text"
	// FormatJSON emits one JSON object per line.
	FormatJSON Format = "json"
)

// Options configure a Logger.
type Options struct {
	Level  slog.Level
	Format Format
	Writer io.Writer // defaults to os.Stderr
}

// New builds a Logger from opts. The writer defaults to stderr to keep stdout
// clear for the stdio transport.
func New(opts Options) *Logger {
	w := opts.Writer
	if w == nil {
		w = os.Stderr
	}
	handlerOpts := &slog.HandlerOptions{Level: opts.Level}
	var h slog.Handler
	if opts.Format == FormatJSON {
		h = slog.NewJSONHandler(w, handlerOpts)
	} else {
		h = slog.NewTextHandler(w, handlerOpts)
	}
	return slog.New(h)
}

// Discard returns a Logger that drops every record. Useful in tests.
func Discard() *Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// ParseLevel converts a case-insensitive level name to a slog.Level. Unknown
// values default to info.
func ParseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
