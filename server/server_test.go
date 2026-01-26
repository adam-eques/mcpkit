package server

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/adam-eques/mcpkit/jsonrpc"
	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/tools"
	"github.com/adam-eques/mcpkit/transport"
)

type staticTool struct{}

func (staticTool) Definition() mcp.Tool {
	return mcp.Tool{Name: "greet", InputSchema: tools.Object(map[string]tools.Prop{
		"name": tools.Str("who to greet"),
	})}
}

func (staticTool) Call(_ context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct {
		Name string `json:"name"`
	}
	if err := tools.Bind(args, &in); err != nil {
		return nil, err
	}
	return mcp.TextResult("hello " + in.Name), nil
}

func newTestServer() *Server {
	reg := tools.NewRegistry()
	reg.MustRegister(staticTool{})
	return New(mcp.Implementation{Name: "test", Version: "0.0.1"}, reg)
}

func call(t *testing.T, s *Server, method string, params any) *jsonrpc.Response {
	t.Helper()
	var raw json.RawMessage
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			t.Fatal(err)
		}
		raw = b
	}
	req := jsonrpc.NewRequest(jsonrpc.Int64ID(1), method, raw)
	frame, _ := json.Marshal(req)
	respBytes, err := s.HandleMessage(context.Background(), frame)
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	var resp jsonrpc.Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return &resp
}

func TestInitializeThenCall(t *testing.T) {
	s := newTestServer()

	resp := call(t, s, mcp.MethodInitialize, mcp.InitializeParams{
		ProtocolVersion: mcp.ProtocolVersion,
		ClientInfo:      mcp.Implementation{Name: "client", Version: "1.0"},
	})
	if resp.Error != nil {
		t.Fatalf("initialize error: %v", resp.Error)
	}
	var initRes mcp.InitializeResult
	if err := json.Unmarshal(resp.Result, &initRes); err != nil {
		t.Fatal(err)
	}
	if initRes.Capabilities.Tools == nil {
		t.Fatal("tools capability not advertised")
	}

	resp = call(t, s, mcp.MethodToolsCall, mcp.CallToolParams{
		Name:      "greet",
		Arguments: json.RawMessage(`{"name":"ada"}`),
	})
	if resp.Error != nil {
		t.Fatalf("call error: %v", resp.Error)
	}
	var callRes mcp.CallToolResult
	if err := json.Unmarshal(resp.Result, &callRes); err != nil {
		t.Fatal(err)
	}
	if len(callRes.Content) != 1 {
		t.Fatalf("content len=%d", len(callRes.Content))
	}
}

func TestCallBeforeInitializeFails(t *testing.T) {
	s := newTestServer()
	resp := call(t, s, mcp.MethodToolsList, nil)
	if resp.Error == nil {
		t.Fatal("expected error before initialize")
	}
}

func TestUnknownMethod(t *testing.T) {
	s := newTestServer()
	call(t, s, mcp.MethodInitialize, mcp.InitializeParams{ProtocolVersion: mcp.ProtocolVersion})
	resp := call(t, s, "does/not/exist", nil)
	if resp.Error == nil || resp.Error.Code != jsonrpc.CodeMethodNotFound {
		t.Fatalf("expected method not found, got %+v", resp.Error)
	}
}

func TestServeOverPipe(t *testing.T) {
	s := newTestServer()
	client, srv := transport.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.Serve(ctx, srv)

	send := func(method string, params any) {
		b, _ := json.Marshal(params)
		req := jsonrpc.NewRequest(jsonrpc.Int64ID(1), method, b)
		frame, _ := json.Marshal(req)
		if err := client.Send(ctx, frame); err != nil {
			t.Fatal(err)
		}
	}
	recv := func() *jsonrpc.Response {
		frame, err := client.Receive(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var resp jsonrpc.Response
		if err := json.Unmarshal(frame, &resp); err != nil {
			t.Fatal(err)
		}
		return &resp
	}

	send(mcp.MethodInitialize, mcp.InitializeParams{ProtocolVersion: mcp.ProtocolVersion})
	if recv().Error != nil {
		t.Fatal("initialize failed over pipe")
	}
	send(mcp.MethodToolsList, nil)
	resp := recv()
	if resp.Error != nil {
		t.Fatalf("tools/list failed: %v", resp.Error)
	}
	var list mcp.ListToolsResult
	if err := json.Unmarshal(resp.Result, &list); err != nil {
		t.Fatal(err)
	}
	if len(list.Tools) != 1 || list.Tools[0].Name != "greet" {
		t.Fatalf("unexpected tools: %+v", list.Tools)
	}
	client.Close()
}
