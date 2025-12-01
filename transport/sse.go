package transport

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
)

// SSEWriter serialises frames as Server-Sent Events for the HTTP gateway's
// server-to-client notification stream. Each frame is written as a single
// "message" event with the JSON payload split across data lines per the SSE
// grammar.
type SSEWriter struct {
	mu  sync.Mutex
	w   *bufio.Writer
	raw io.Writer
}

// httpFlusher matches the *http.ResponseWriter Flush method without importing
// net/http here, keeping the transport package dependency-free.
type httpFlusher interface{ Flush() }

// NewSSEWriter wraps w. If w implements an http.Flusher-style Flush method it is
// invoked after every event so clients receive frames promptly.
func NewSSEWriter(w io.Writer) *SSEWriter {
	return &SSEWriter{w: bufio.NewWriter(w), raw: w}
}

// WriteEvent writes a named SSE event carrying frame as its data.
func (s *SSEWriter) WriteEvent(event string, frame []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if event != "" {
		if _, err := fmt.Fprintf(s.w, "event: %s\n", event); err != nil {
			return err
		}
	}
	for _, line := range strings.Split(string(frame), "\n") {
		if _, err := fmt.Fprintf(s.w, "data: %s\n", line); err != nil {
			return err
		}
	}
	if _, err := s.w.WriteString("\n"); err != nil {
		return err
	}
	if err := s.w.Flush(); err != nil {
		return err
	}
	if f, ok := s.raw.(httpFlusher); ok {
		f.Flush()
	}
	return nil
}

// WriteComment writes an SSE comment line, commonly used as a keep-alive.
func (s *SSEWriter) WriteComment(text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := fmt.Fprintf(s.w, ": %s\n\n", text); err != nil {
		return err
	}
	return s.w.Flush()
}
