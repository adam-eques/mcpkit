// Package tools defines the Handler interface every MCP tool implements and a
// concurrency-safe registry that the server uses to advertise and dispatch
// tools. Individual tools live in subpackages and are wired in by the command.
package tools

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/adam-eques/mcpkit/mcp"
)

// Handler is a single callable tool.
type Handler interface {
	// Definition returns the tool's advertised schema. The Name must be unique
	// within a registry.
	Definition() mcp.Tool
	// Call executes the tool. args is the raw JSON arguments object, which may be
	// nil when the caller supplied none. A returned error is treated as an
	// internal protocol failure; tool-level failures should instead be reported
	// via a result with IsError set.
	Call(ctx context.Context, args json.RawMessage) (*mcp.CallToolResult, error)
}

// ErrToolNotFound is returned when a named tool is not registered.
var ErrToolNotFound = errors.New("tools: tool not found")

// Bind unmarshals args into v, tolerating a nil or empty arguments object.
func Bind(args json.RawMessage, v any) error {
	if len(args) == 0 || string(args) == "null" {
		return nil
	}
	return json.Unmarshal(args, v)
}
