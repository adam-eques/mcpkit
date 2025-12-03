package transport

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestStdioRoundTrip(t *testing.T) {
	in := strings.NewReader("{\"a\":1}\n\n{\"b\":2}\n")
	var out bytes.Buffer
	tr := NewStdio(in, &out)
	defer tr.Close()

	ctx := context.Background()
	f1, err := tr.Receive(ctx)
	if err != nil {
		t.Fatalf("receive 1: %v", err)
	}
	if string(f1) != `{"a":1}` {
		t.Fatalf("frame1=%s", f1)
	}
	// Blank line is skipped, so the next frame is the second object.
	f2, err := tr.Receive(ctx)
	if err != nil {
		t.Fatalf("receive 2: %v", err)
	}
	if string(f2) != `{"b":2}` {
		t.Fatalf("frame2=%s", f2)
	}
	if _, err := tr.Receive(ctx); !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF, got %v", err)
	}

	if err := tr.Send(ctx, []byte(`{"ok":true}`)); err != nil {
		t.Fatalf("send: %v", err)
	}
	if out.String() != "{\"ok\":true}\n" {
		t.Fatalf("out=%q", out.String())
	}
}

func TestStdioRejectsEmbeddedNewline(t *testing.T) {
	tr := NewStdio(strings.NewReader(""), &bytes.Buffer{})
	defer tr.Close()
	if err := tr.Send(context.Background(), []byte("a\nb")); err == nil {
		t.Fatal("expected error for embedded newline")
	}
}

func TestReceiveHonoursContext(t *testing.T) {
	tr := NewStdio(&blockingReader{}, &bytes.Buffer{})
	defer tr.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := tr.Receive(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestPipe(t *testing.T) {
	client, server := Pipe()
	ctx := context.Background()
	if err := client.Send(ctx, []byte(`{"ping":1}`)); err != nil {
		t.Fatalf("send: %v", err)
	}
	got, err := server.Receive(ctx)
	if err != nil {
		t.Fatalf("receive: %v", err)
	}
	if string(got) != `{"ping":1}` {
		t.Fatalf("got=%s", got)
	}
	client.Close()
	if _, err := server.Receive(ctx); !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF after close, got %v", err)
	}
}

// blockingReader never returns, forcing Receive to rely on context cancellation.
type blockingReader struct{}

func (blockingReader) Read(p []byte) (int, error) {
	select {} //nolint:staticcheck // intentional block for the test
}
