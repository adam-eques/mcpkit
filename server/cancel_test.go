package server

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/adam-eques/mcpkit/jsonrpc"
	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/tools"
	"github.com/adam-eques/mcpkit/transport"
)

// slowTool blocks until its context is cancelled or a deadline elapses.
type slowTool struct{ started chan struct{} }

func (slowTool) Definition() mcp.Tool {
	return mcp.Tool{Name: "slow", InputSchema: tools.Object(nil)}
}

func (s slowTool) Call(ctx context.Context, _ json.RawMessage) (*mcp.CallToolResult, error) {
	close(s.started)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Second):
		return mcp.TextResult("completed"), nil
	}
}

func TestCancellationAbortsInFlight(t *testing.T) {
	reg := tools.NewRegistry()
	started := make(chan struct{})
	reg.MustRegister(slowTool{started: started})
	s := New(mcp.Implementation{Name: "t"}, reg)

	client, srv := transport.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.Serve(ctx, srv)

	send := func(v any) {
		frame, _ := json.Marshal(v)
		if err := client.Send(ctx, frame); err != nil {
			t.Fatal(err)
		}
	}

	send(jsonrpc.NewRequest(jsonrpc.Int64ID(1), mcp.MethodInitialize, mustJSON(mcp.InitializeParams{ProtocolVersion: mcp.ProtocolVersion})))
	if _, err := client.Receive(ctx); err != nil {
		t.Fatal(err)
	}
	send(jsonrpc.NewRequest(jsonrpc.Int64ID(2), mcp.MethodToolsCall, mustJSON(mcp.CallToolParams{Name: "slow"})))

	// Wait until the tool is running, then cancel it.
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("tool never started")
	}
	send(jsonrpc.NewNotification(mcp.NotificationCancelled, mustJSON(mcp.CancelledParams{RequestID: 2})))

	done := make(chan *jsonrpc.Response, 1)
	go func() {
		frame, err := client.Receive(ctx)
		if err != nil {
			return
		}
		var resp jsonrpc.Response
		json.Unmarshal(frame, &resp)
		done <- &resp
	}()

	select {
	case resp := <-done:
		if resp.Error == nil {
			t.Fatal("expected an error response for the cancelled call")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("cancellation did not abort the tool within the timeout")
	}
	client.Close()
}

func mustJSON(v any) json.RawMessage {
	raw, _ := json.Marshal(v)
	return raw
}
