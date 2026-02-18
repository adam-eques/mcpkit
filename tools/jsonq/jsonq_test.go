package jsonq

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/adam-eques/mcpkit/mcp"
)

func decode(t *testing.T, s string) any {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestQuery(t *testing.T) {
	doc := decode(t, `{"user":{"name":"ada","roles":[{"name":"admin"},{"name":"dev"}]}}`)
	cases := []struct {
		path string
		want any
	}{
		{"user.name", "ada"},
		{"user.roles[0].name", "admin"},
		{"user.roles[1].name", "dev"},
		{`user["name"]`, "ada"},
	}
	for _, tc := range cases {
		got, err := Query(doc, tc.path)
		if err != nil {
			t.Errorf("Query(%q) error: %v", tc.path, err)
			continue
		}
		if got != tc.want {
			t.Errorf("Query(%q)=%v want %v", tc.path, got, tc.want)
		}
	}
}

func TestQueryErrors(t *testing.T) {
	doc := decode(t, `{"a":[1,2]}`)
	for _, p := range []string{"a[5]", "a.b", "missing", "a[x]"} {
		if _, err := Query(doc, p); err == nil {
			t.Errorf("Query(%q) expected error", p)
		}
	}
}

func TestTool(t *testing.T) {
	res, err := New().Call(context.Background(), json.RawMessage(`{"json":"{\"x\":[10,20]}","path":"x[1]"}`))
	if err != nil {
		t.Fatal(err)
	}
	if res.Content[0].(mcp.TextContent).Text != "20" {
		t.Fatalf("got %+v", res)
	}
}
