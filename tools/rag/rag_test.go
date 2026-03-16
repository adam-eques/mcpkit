package rag

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/adam-eques/mcpkit/mcp"
)

func TestEmbedIsUnitLength(t *testing.T) {
	v := Embed("the quick brown fox")
	var sum float64
	for _, x := range v {
		sum += float64(x) * float64(x)
	}
	if math.Abs(sum-1) > 1e-5 {
		t.Fatalf("embedding not normalised: |v|^2=%v", sum)
	}
}

func TestSearchRanksRelevantHigher(t *testing.T) {
	s := NewStore()
	s.Add("go", "Go is a statically typed compiled programming language designed at Google", nil)
	s.Add("python", "Python is an interpreted high-level general-purpose programming language", nil)
	s.Add("cooking", "A recipe for baking sourdough bread with a wild yeast starter", nil)

	hits := s.Search("golang compiled language by google", 3)
	if len(hits) != 3 {
		t.Fatalf("expected 3 hits, got %d", len(hits))
	}
	if hits[0].Document.ID != "go" {
		t.Fatalf("expected 'go' ranked first, got %q (%.3f)", hits[0].Document.ID, hits[0].Score)
	}
	if hits[0].Score <= hits[2].Score {
		t.Fatal("scores not in descending order")
	}
}

func TestTools(t *testing.T) {
	s := NewStore()
	hs := Handlers(s)
	ctx := context.Background()

	_, err := hs[0].Call(ctx, json.RawMessage(`{"documents":[{"id":"a","text":"vector databases store embeddings"},{"text":"the ocean is deep and blue"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if s.Len() != 2 {
		t.Fatalf("store len=%d", s.Len())
	}

	res, err := hs[1].Call(ctx, json.RawMessage(`{"query":"embedding storage","k":1}`))
	if err != nil {
		t.Fatal(err)
	}
	text := res.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "[a]") {
		t.Fatalf("expected document a to rank first: %s", text)
	}
}
