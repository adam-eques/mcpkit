// Package transport moves raw JSON-RPC frames between the server and a peer. A
// frame is a single, complete JSON-RPC message with no embedded newlines. The
// package provides a newline-delimited stdio transport, an in-memory pipe for
// tests, and Server-Sent Events helpers for the HTTP gateway.
package transport

import "context"

// Transport is a bidirectional stream of JSON-RPC frames.
type Transport interface {
	// Receive blocks until the next frame arrives, ctx is cancelled, or the peer
	// closes the stream (reported as io.EOF).
	Receive(ctx context.Context) ([]byte, error)

	// Send writes a single frame. Implementations must be safe for concurrent use
	// by multiple goroutines.
	Send(ctx context.Context, frame []byte) error

	// Close releases the underlying resources and unblocks any pending Receive.
	Close() error
}
