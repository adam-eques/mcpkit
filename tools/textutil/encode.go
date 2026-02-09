package textutil

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/tools"
)

type uuidTool struct{}

func (uuidTool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "uuid",
		Description: "Generate a random RFC 4122 version-4 UUID.",
		InputSchema: tools.Object(nil),
		Annotations: &mcp.ToolAnnotations{Title: "UUID v4"},
	}
}

func (uuidTool) Call(_ context.Context, _ json.RawMessage) (*mcp.CallToolResult, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return nil, err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return mcp.TextResult(fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])), nil
}

type base64Tool struct{}

func (base64Tool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "base64",
		Description: "Base64-encode or -decode text using standard padded encoding.",
		InputSchema: tools.Object(map[string]tools.Prop{
			"text":      tools.Str("input text"),
			"operation": {Type: "string", Description: "direction", Enum: []string{"encode", "decode"}, Default: "encode"},
		}, "text"),
		Annotations: &mcp.ToolAnnotations{Title: "Base64", ReadOnlyHint: true, IdempotentHint: true},
	}
}

func (base64Tool) Call(_ context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct {
		Text      string `json:"text"`
		Operation string `json:"operation"`
	}
	if err := tools.Bind(args, &in); err != nil {
		return nil, err
	}
	switch in.Operation {
	case "", "encode":
		return mcp.TextResult(base64.StdEncoding.EncodeToString([]byte(in.Text))), nil
	case "decode":
		data, err := base64.StdEncoding.DecodeString(in.Text)
		if err != nil {
			return mcp.ErrorResult("invalid base64: " + err.Error()), nil
		}
		return mcp.TextResult(string(data)), nil
	default:
		return mcp.ErrorResult("unknown operation: " + in.Operation), nil
	}
}
