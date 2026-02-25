package calc

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/tools"
)

// Tool evaluates arithmetic expressions.
type Tool struct{}

// New returns the calculate tool.
func New() tools.Handler { return Tool{} }

// Definition implements tools.Handler.
func (Tool) Definition() mcp.Tool {
	return mcp.Tool{
		Name: "calculate",
		Description: "Evaluate a mathematical expression. Supports + - * / % ^, " +
			"parentheses, the constants pi/e/tau and functions such as sqrt, abs, " +
			"sin, cos, log, ln and round.",
		InputSchema: tools.Object(map[string]tools.Prop{
			"expression": tools.Str("the expression to evaluate, e.g. \"2 * (3 + 4) ^ 2\""),
		}, "expression"),
		Annotations: &mcp.ToolAnnotations{
			Title:        "Calculator",
			ReadOnlyHint: true,
			OpenWorldHint: false,
		},
	}
}

// Call implements tools.Handler.
func (Tool) Call(_ context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct {
		Expression string `json:"expression"`
	}
	if err := tools.Bind(args, &in); err != nil {
		return nil, err
	}
	if in.Expression == "" {
		return mcp.ErrorResult("expression is required"), nil
	}
	v, err := Eval(in.Expression)
	if err != nil {
		return mcp.ErrorResult("calculation error: " + err.Error()), nil
	}
	return mcp.TextResult(strconv.FormatFloat(v, 'g', -1, 64)), nil
}
