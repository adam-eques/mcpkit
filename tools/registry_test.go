package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/adam-eques/mcpkit/mcp"
)

type echoTool struct{ name string }

func (e echoTool) Definition() mcp.Tool {
	return mcp.Tool{Name: e.name, InputSchema: Object(map[string]Prop{"msg": Str("message")})}
}

func (e echoTool) Call(_ context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct {
		Msg string `json:"msg"`
	}
	if err := Bind(args, &in); err != nil {
		return nil, err
	}
	return mcp.TextResult(in.Msg), nil
}

func TestRegistryRegisterAndList(t *testing.T) {
	r := NewRegistry()
	r.MustRegister(echoTool{"b"}, echoTool{"a"})
	if r.Len() != 2 {
		t.Fatalf("len=%d", r.Len())
	}
	list := r.List()
	if list[0].Name != "a" || list[1].Name != "b" {
		t.Fatalf("list not sorted: %+v", list)
	}
}

func TestRegistryDuplicate(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(echoTool{"x"}); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(echoTool{"x"}); err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestRegistryCall(t *testing.T) {
	r := NewRegistry()
	r.MustRegister(echoTool{"echo"})
	res, err := r.Call(context.Background(), "echo", json.RawMessage(`{"msg":"hi"}`))
	if err != nil {
		t.Fatal(err)
	}
	tc := res.Content[0].(mcp.TextContent)
	if tc.Text != "hi" {
		t.Fatalf("text=%q", tc.Text)
	}
}

func TestRegistryCallMissing(t *testing.T) {
	r := NewRegistry()
	if _, err := r.Call(context.Background(), "nope", nil); !errors.Is(err, ErrToolNotFound) {
		t.Fatalf("expected ErrToolNotFound, got %v", err)
	}
}

func TestObjectSchema(t *testing.T) {
	raw := Object(map[string]Prop{"n": Int("count")}, "n")
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed["type"] != "object" {
		t.Fatalf("type=%v", parsed["type"])
	}
	if _, ok := parsed["required"]; !ok {
		t.Fatal("required missing")
	}
}
