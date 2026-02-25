package jsonq

import (
	"context"
	"encoding/json"

	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/tools"
)

// Tool evaluates a path against a JSON document.
type Tool struct{}

// New returns the json_query tool.
func New() tools.Handler { return Tool{} }

// Definition implements tools.Handler.
func (Tool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "json_query",
		Description: "Extract a value from a JSON document using a dotted path such as user.roles[0].name.",
		InputSchema: tools.Object(map[string]tools.Prop{
			"json": tools.Str("the JSON document as a string"),
			"path": tools.Str("dotted path expression, e.g. a.b[0].c"),
		}, "json", "path"),
		Annotations: &mcp.ToolAnnotations{Title: "JSON Query", ReadOnlyHint: true, IdempotentHint: true},
	}
}

// Call implements tools.Handler.
func (Tool) Call(_ context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct {
		JSON string `json:"json"`
		Path string `json:"path"`
	}
	if err := tools.Bind(args, &in); err != nil {
		return nil, err
	}
	var doc any
	if err := json.Unmarshal([]byte(in.JSON), &doc); err != nil {
		return mcp.ErrorResult("invalid JSON: " + err.Error()), nil
	}
	val, err := Query(doc, in.Path)
	if err != nil {
		return mcp.ErrorResult("query error: " + err.Error()), nil
	}
	out, err := json.MarshalIndent(val, "", "  ")
	if err != nil {
		return nil, err
	}
	return mcp.TextResult(string(out)), nil
}
