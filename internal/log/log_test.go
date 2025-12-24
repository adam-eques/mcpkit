package log

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"INFO":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
		"":      slog.LevelInfo,
	}
	for in, want := range cases {
		if got := ParseLevel(in); got != want {
			t.Errorf("ParseLevel(%q)=%v want %v", in, got, want)
		}
	}
}

func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	l := New(Options{Level: slog.LevelInfo, Format: FormatJSON, Writer: &buf})
	l.Info("hello", "n", 1)
	if !strings.Contains(buf.String(), `"msg":"hello"`) {
		t.Fatalf("expected JSON log, got %q", buf.String())
	}
}

func TestDiscardDoesNotPanic(t *testing.T) {
	Discard().Info("ignored")
}
