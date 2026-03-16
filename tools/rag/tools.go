package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/tools"
)

// Handlers returns the rag_index and rag_search tools sharing store.
func Handlers(store *Store) []tools.Handler {
	return []tools.Handler{indexTool{store}, searchTool{store}}
}

type indexTool struct{ s *Store }

func (indexTool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "rag_index",
		Description: "Add one or more text passages to the in-memory vector index so they can later be retrieved by semantic similarity.",
		InputSchema: tools.Object(map[string]tools.Prop{
			"documents": {
				Type:        "array",
				Description: "passages to index",
				Items: &tools.Prop{Type: "object", Properties: map[string]tools.Prop{
					"id":   tools.Str("optional stable id"),
					"text": tools.Str("passage text"),
				}},
			},
		}, "documents"),
		Annotations: &mcp.ToolAnnotations{Title: "RAG Index"},
	}
}

func (t indexTool) Call(_ context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct {
		Documents []struct {
			ID       string            `json:"id"`
			Text     string            `json:"text"`
			Metadata map[string]string `json:"metadata"`
		} `json:"documents"`
	}
	if err := tools.Bind(args, &in); err != nil {
		return nil, err
	}
	if len(in.Documents) == 0 {
		return mcp.ErrorResult("at least one document is required"), nil
	}
	ids := make([]string, 0, len(in.Documents))
	for _, d := range in.Documents {
		if strings.TrimSpace(d.Text) == "" {
			continue
		}
		ids = append(ids, t.s.Add(d.ID, d.Text, d.Metadata))
	}
	return mcp.TextResult(fmt.Sprintf("indexed %d document(s); %d total in store\nids: %s",
		len(ids), t.s.Len(), strings.Join(ids, ", "))), nil
}

type searchTool struct{ s *Store }

func (searchTool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "rag_search",
		Description: "Retrieve the passages most semantically similar to a query from the vector index, with cosine similarity scores.",
		InputSchema: tools.Object(map[string]tools.Prop{
			"query": tools.Str("the search query"),
			"k":     {Type: "integer", Description: "number of results", Default: 3},
		}, "query"),
		Annotations: &mcp.ToolAnnotations{Title: "RAG Search", ReadOnlyHint: true},
	}
}

func (t searchTool) Call(_ context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct {
		Query string `json:"query"`
		K     int    `json:"k"`
	}
	if err := tools.Bind(args, &in); err != nil {
		return nil, err
	}
	if in.Query == "" {
		return mcp.ErrorResult("query is required"), nil
	}
	if in.K <= 0 {
		in.K = 3
	}
	hits := t.s.Search(in.Query, in.K)
	if len(hits) == 0 {
		return mcp.TextResult("no documents indexed yet"), nil
	}
	var b strings.Builder
	for i, h := range hits {
		fmt.Fprintf(&b, "%d. [%s] score=%.3f\n   %s\n", i+1, h.Document.ID, h.Score, truncate(h.Document.Text, 200))
	}
	return mcp.TextResult(b.String()), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
