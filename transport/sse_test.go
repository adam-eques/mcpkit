package transport

import (
	"bytes"
	"strings"
	"testing"
)

func TestSSEWriterEvent(t *testing.T) {
	var buf bytes.Buffer
	w := NewSSEWriter(&buf)
	if err := w.WriteEvent("message", []byte(`{"a":1}`)); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "event: message\n") {
		t.Fatalf("missing event line: %q", out)
	}
	if !strings.Contains(out, "data: {\"a\":1}\n") {
		t.Fatalf("missing data line: %q", out)
	}
	if !strings.HasSuffix(out, "\n\n") {
		t.Fatalf("event must end with a blank line: %q", out)
	}
}

func TestSSEWriterMultilineData(t *testing.T) {
	var buf bytes.Buffer
	w := NewSSEWriter(&buf)
	w.WriteEvent("", []byte("line1\nline2"))
	out := buf.String()
	if strings.Count(out, "data: ") != 2 {
		t.Fatalf("expected two data lines: %q", out)
	}
}

func TestSSEComment(t *testing.T) {
	var buf bytes.Buffer
	w := NewSSEWriter(&buf)
	if err := w.WriteComment("keep-alive"); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(buf.String(), ": keep-alive") {
		t.Fatalf("unexpected comment: %q", buf.String())
	}
}
