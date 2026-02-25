// Package timeutil implements the "time_now" and "time_convert" tools for
// working with the current time and converting timestamps between time zones.
package timeutil

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/tools"
)

// Clock returns the current time; it is a field so tests can inject a fixed time.
type Clock func() time.Time

// Handlers returns the time tools driven by clock. A nil clock uses time.Now.
func Handlers(clock Clock) []tools.Handler {
	if clock == nil {
		clock = time.Now
	}
	return []tools.Handler{nowTool{clock}, convertTool{}}
}

type nowTool struct{ now Clock }

func (nowTool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "time_now",
		Description: "Return the current time in the requested IANA time zone (default UTC), formatted as RFC 3339.",
		InputSchema: tools.Object(map[string]tools.Prop{
			"timezone": {Type: "string", Description: "IANA zone, e.g. America/New_York", Default: "UTC"},
		}),
		Annotations: &mcp.ToolAnnotations{Title: "Current Time", ReadOnlyHint: true},
	}
}

func (t nowTool) Call(_ context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct {
		Timezone string `json:"timezone"`
	}
	if err := tools.Bind(args, &in); err != nil {
		return nil, err
	}
	loc, err := loadZone(in.Timezone)
	if err != nil {
		return mcp.ErrorResult(err.Error()), nil
	}
	return mcp.TextResult(t.now().In(loc).Format(time.RFC3339)), nil
}

type convertTool struct{}

func (convertTool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "time_convert",
		Description: "Convert an RFC 3339 timestamp from one time zone to another.",
		InputSchema: tools.Object(map[string]tools.Prop{
			"timestamp": tools.Str("RFC 3339 timestamp, e.g. 2025-10-01T12:00:00Z"),
			"to":        tools.Str("target IANA time zone"),
		}, "timestamp", "to"),
		Annotations: &mcp.ToolAnnotations{Title: "Convert Time", ReadOnlyHint: true},
	}
}

func (convertTool) Call(_ context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct {
		Timestamp string `json:"timestamp"`
		To        string `json:"to"`
	}
	if err := tools.Bind(args, &in); err != nil {
		return nil, err
	}
	ts, err := time.Parse(time.RFC3339, in.Timestamp)
	if err != nil {
		return mcp.ErrorResult("invalid timestamp: " + err.Error()), nil
	}
	loc, err := loadZone(in.To)
	if err != nil {
		return mcp.ErrorResult(err.Error()), nil
	}
	return mcp.TextResult(ts.In(loc).Format(time.RFC3339)), nil
}

func loadZone(name string) (*time.Location, error) {
	if name == "" || name == "UTC" {
		return time.UTC, nil
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("unknown time zone %q", name)
	}
	return loc, nil
}
