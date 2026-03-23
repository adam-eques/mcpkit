package fetch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/tools"
)

// Tool adapts a Fetcher to the tools.Handler interface.
type Tool struct{ f *Fetcher }

// New returns the http_fetch tool.
func New(opts ...Option) tools.Handler { return Tool{f: NewFetcher(opts...)} }

// Definition implements tools.Handler.
func (Tool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "http_fetch",
		Description: "Fetch an http(s) URL and return the status, selected headers and body. Private and loopback addresses are blocked by default.",
		InputSchema: tools.Object(map[string]tools.Prop{
			"url":    tools.Str("the absolute http(s) URL to fetch"),
			"method": {Type: "string", Description: "HTTP method", Enum: []string{"GET", "POST", "PUT", "DELETE", "HEAD"}, Default: "GET"},
			"body":   tools.Str("optional request body"),
		}, "url"),
		Annotations: &mcp.ToolAnnotations{Title: "HTTP Fetch", OpenWorldHint: true},
	}
}

// Call implements tools.Handler.
func (t Tool) Call(ctx context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct {
		URL     string            `json:"url"`
		Method  string            `json:"method"`
		Headers map[string]string `json:"headers"`
		Body    string            `json:"body"`
	}
	if err := tools.Bind(args, &in); err != nil {
		return nil, err
	}
	if in.URL == "" {
		return mcp.ErrorResult("url is required"), nil
	}
	res, err := t.f.Do(ctx, in.Method, in.URL, in.Headers, in.Body)
	if err != nil {
		return mcp.ErrorResult("fetch failed: " + err.Error()), nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "HTTP %d (%dms) %s\n", res.Status, res.DurationMS, res.FinalURL)
	for _, h := range []string{"Content-Type", "Content-Length", "Server"} {
		if v := res.Header.Get(h); v != "" {
			fmt.Fprintf(&b, "%s: %s\n", h, v)
		}
	}
	b.WriteString("\n")
	b.WriteString(res.Body)
	if res.Truncated {
		b.WriteString("\n\n[body truncated]")
	}
	return mcp.TextResult(b.String()), nil
}
