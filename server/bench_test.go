package server

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/adam-eques/mcpkit/jsonrpc"
	"github.com/adam-eques/mcpkit/mcp"
)

func BenchmarkHandleToolsCall(b *testing.B) {
	s := newTestServer()
	s.HandleMessage(context.Background(), mustFrame(mcp.MethodInitialize, mcp.InitializeParams{ProtocolVersion: mcp.ProtocolVersion}))
	frame := mustFrame(mcp.MethodToolsCall, mcp.CallToolParams{Name: "greet", Arguments: json.RawMessage(`{"name":"bench"}`)})
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.HandleMessage(ctx, frame); err != nil {
			b.Fatal(err)
		}
	}
}

func mustFrame(method string, params any) []byte {
	raw, _ := json.Marshal(params)
	req := jsonrpc.NewRequest(jsonrpc.Int64ID(1), method, raw)
	frame, _ := json.Marshal(req)
	return frame
}
