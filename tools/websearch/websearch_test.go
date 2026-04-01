package websearch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adam-eques/mcpkit/mcp"
)

const sample = `{
  "AbstractText": "Go is a programming language.",
  "AbstractURL": "https://example.com/go",
  "RelatedTopics": [
    {"Text": "Goroutines", "FirstURL": "https://example.com/goroutines"},
    {"Text": "Channels", "FirstURL": "https://example.com/channels"},
    {"Text": "", "FirstURL": "https://example.com/skip"}
  ]
}`

func testServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") == "" {
			t.Error("missing query parameter")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(sample))
	}))
}

func TestSearch(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()
	s := NewSearcher(WithEndpoint(srv.URL + "/"))
	abstract, results, err := s.Search(context.Background(), "go language", 5)
	if err != nil {
		t.Fatal(err)
	}
	if abstract == "" {
		t.Fatal("expected abstract")
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results (blank skipped), got %d", len(results))
	}
}

func TestSearchRespectsLimit(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()
	_, results, _ := NewSearcher(WithEndpoint(srv.URL + "/")).Search(context.Background(), "x", 1)
	if len(results) != 1 {
		t.Fatalf("limit not applied: %d", len(results))
	}
}

func TestTool(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()
	tool := New(WithEndpoint(srv.URL + "/"))
	res, err := tool.Call(context.Background(), json.RawMessage(`{"query":"go"}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Content[0].(mcp.TextContent).Text, "Goroutines") {
		t.Fatalf("unexpected output: %+v", res)
	}
}
