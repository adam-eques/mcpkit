package kv

import (
	"context"
	"encoding/json"
	"path/filepath"
	"sync"
	"testing"

	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/tools"
)

func find(hs []tools.Handler, name string) tools.Handler {
	for _, h := range hs {
		if h.Definition().Name == name {
			return h
		}
	}
	return nil
}

func TestKVLifecycle(t *testing.T) {
	hs := Handlers(NewStore())
	ctx := context.Background()

	find(hs, "kv_set").Call(ctx, json.RawMessage(`{"key":"a","value":"1"}`))
	res, _ := find(hs, "kv_get").Call(ctx, json.RawMessage(`{"key":"a"}`))
	if res.Content[0].(mcp.TextContent).Text != "1" {
		t.Fatalf("get returned %+v", res)
	}
	res, _ = find(hs, "kv_get").Call(ctx, json.RawMessage(`{"key":"missing"}`))
	if !res.IsError {
		t.Fatal("expected error for missing key")
	}
	del, _ := find(hs, "kv_delete").Call(ctx, json.RawMessage(`{"key":"a"}`))
	if del.Content[0].(mcp.TextContent).Text != "deleted" {
		t.Fatalf("delete returned %+v", del)
	}
}

func TestPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store.json")
	s1, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s1.Set("k", "v"); err != nil {
		t.Fatal(err)
	}
	s2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := s2.Get("k"); !ok || v != "v" {
		t.Fatalf("reload failed: %q %v", v, ok)
	}
}

func TestConcurrentSet(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			s.Set("shared", "x")
		}(i)
	}
	wg.Wait()
	if _, ok := s.Get("shared"); !ok {
		t.Fatal("value missing after concurrent writes")
	}
}
