package kv

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/tools"
)

// Handlers returns the four kv tools backed by store.
func Handlers(store *Store) []tools.Handler {
	return []tools.Handler{
		setTool{store}, getTool{store}, listTool{store}, deleteTool{store},
	}
}

type setTool struct{ s *Store }

func (setTool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "kv_set",
		Description: "Store a string value under a key in the persistent key/value store.",
		InputSchema: tools.Object(map[string]tools.Prop{
			"key":   tools.Str("the key"),
			"value": tools.Str("the value to store"),
		}, "key", "value"),
		Annotations: &mcp.ToolAnnotations{Title: "KV Set", IdempotentHint: true},
	}
}

func (t setTool) Call(_ context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct{ Key, Value string }
	if err := tools.Bind(args, &in); err != nil {
		return nil, err
	}
	if in.Key == "" {
		return mcp.ErrorResult("key is required"), nil
	}
	if err := t.s.Set(in.Key, in.Value); err != nil {
		return nil, err
	}
	return mcp.TextResult("ok"), nil
}

type getTool struct{ s *Store }

func (getTool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "kv_get",
		Description: "Retrieve the value stored under a key.",
		InputSchema: tools.Object(map[string]tools.Prop{"key": tools.Str("the key")}, "key"),
		Annotations: &mcp.ToolAnnotations{Title: "KV Get", ReadOnlyHint: true},
	}
}

func (t getTool) Call(_ context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct{ Key string }
	if err := tools.Bind(args, &in); err != nil {
		return nil, err
	}
	v, ok := t.s.Get(in.Key)
	if !ok {
		return mcp.ErrorResult("key not found"), nil
	}
	return mcp.TextResult(v), nil
}

type listTool struct{ s *Store }

func (listTool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "kv_list",
		Description: "List all keys in the key/value store.",
		InputSchema: tools.Object(nil),
		Annotations: &mcp.ToolAnnotations{Title: "KV List", ReadOnlyHint: true},
	}
}

func (t listTool) Call(_ context.Context, _ json.RawMessage) (*mcp.CallToolResult, error) {
	keys := t.s.Keys()
	if len(keys) == 0 {
		return mcp.TextResult("(no keys)"), nil
	}
	return mcp.TextResult(strings.Join(keys, "\n")), nil
}

type deleteTool struct{ s *Store }

func (deleteTool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "kv_delete",
		Description: "Delete a key from the store.",
		InputSchema: tools.Object(map[string]tools.Prop{"key": tools.Str("the key")}, "key"),
		Annotations: &mcp.ToolAnnotations{Title: "KV Delete", IdempotentHint: true},
	}
}

func (t deleteTool) Call(_ context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct{ Key string }
	if err := tools.Bind(args, &in); err != nil {
		return nil, err
	}
	existed, err := t.s.Delete(in.Key)
	if err != nil {
		return nil, err
	}
	if !existed {
		return mcp.TextResult("key did not exist"), nil
	}
	return mcp.TextResult("deleted"), nil
}
