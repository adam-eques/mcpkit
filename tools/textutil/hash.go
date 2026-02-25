// Package textutil implements small text and encoding utilities exposed as the
// hash, uuid and base64 tools.
package textutil

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/tools"
)

// Handlers returns the textutil tools. rand is used by the uuid tool and may be
// nil to use crypto/rand.
func Handlers() []tools.Handler {
	return []tools.Handler{hashTool{}, uuidTool{}, base64Tool{}}
}

type hashTool struct{}

func (hashTool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "hash",
		Description: "Compute a cryptographic hash of the input text.",
		InputSchema: tools.Object(map[string]tools.Prop{
			"text":      tools.Str("text to hash"),
			"algorithm": {Type: "string", Description: "hash algorithm", Enum: []string{"sha256", "sha1", "md5"}, Default: "sha256"},
		}, "text"),
		Annotations: &mcp.ToolAnnotations{Title: "Hash", ReadOnlyHint: true, IdempotentHint: true},
	}
}

func (hashTool) Call(_ context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct {
		Text      string `json:"text"`
		Algorithm string `json:"algorithm"`
	}
	if err := tools.Bind(args, &in); err != nil {
		return nil, err
	}
	var sum []byte
	switch in.Algorithm {
	case "", "sha256":
		h := sha256.Sum256([]byte(in.Text))
		sum = h[:]
	case "sha1":
		h := sha1.Sum([]byte(in.Text))
		sum = h[:]
	case "md5":
		h := md5.Sum([]byte(in.Text))
		sum = h[:]
	default:
		return mcp.ErrorResult("unsupported algorithm: " + in.Algorithm), nil
	}
	return mcp.TextResult(hex.EncodeToString(sum)), nil
}
