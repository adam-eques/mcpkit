package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/tools"
)

// Tool adapts a Searcher to the tools.Handler interface.
type Tool struct{ s *Searcher }

// New returns the web_search tool.
func New(opts ...Option) tools.Handler { return Tool{s: NewSearcher(opts...)} }

// Definition implements tools.Handler.
func (Tool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "web_search",
		Description: "Search the web via the DuckDuckGo Instant Answer API and return an abstract plus related links.",
		InputSchema: tools.Object(map[string]tools.Prop{
			"query": tools.Str("the search query"),
			"limit": {Type: "integer", Description: "maximum related links", Default: 5},
		}, "query"),
		Annotations: &mcp.ToolAnnotations{Title: "Web Search", ReadOnlyHint: true, OpenWorldHint: true},
	}
}

// Call implements tools.Handler.
func (t Tool) Call(ctx context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := tools.Bind(args, &in); err != nil {
		return nil, err
	}
	if in.Query == "" {
		return mcp.ErrorResult("query is required"), nil
	}
	if in.Limit <= 0 {
		in.Limit = 5
	}
	abstract, results, err := t.s.Search(ctx, in.Query, in.Limit)
	if err != nil {
		return mcp.ErrorResult("search failed: " + err.Error()), nil
	}
	var b strings.Builder
	if abstract != "" {
		b.WriteString(abstract)
		b.WriteString("\n\n")
	}
	if len(results) == 0 && abstract == "" {
		return mcp.TextResult("no results found"), nil
	}
	for i, r := range results {
		fmt.Fprintf(&b, "%d. %s\n   %s\n", i+1, r.Title, r.URL)
	}
	return mcp.TextResult(strings.TrimSpace(b.String())), nil
}
