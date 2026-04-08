package fs

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/tools"
)

func newSandbox(t *testing.T) *Sandbox {
	t.Helper()
	sb, err := New(t.TempDir(), false)
	if err != nil {
		t.Fatal(err)
	}
	return sb
}

func handler(hs []tools.Handler, name string) tools.Handler {
	for _, h := range hs {
		if h.Definition().Name == name {
			return h
		}
	}
	return nil
}

func TestSandboxRejectsTraversal(t *testing.T) {
	sb := newSandbox(t)
	for _, p := range []string{"../secret", "../../etc/passwd", "a/../../b"} {
		if _, err := sb.Resolve(p); err == nil {
			t.Errorf("Resolve(%q) should have failed", p)
		}
	}
	if _, err := sb.Resolve("ok/nested.txt"); err != nil {
		t.Errorf("Resolve of valid path failed: %v", err)
	}
}

func TestWriteThenRead(t *testing.T) {
	sb := newSandbox(t)
	hs := Handlers(sb)
	ctx := context.Background()

	w := handler(hs, "fs_write")
	if _, err := w.Call(ctx, json.RawMessage(`{"path":"dir/hello.txt","content":"hi there"}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(sb.Root, "dir", "hello.txt")); err != nil {
		t.Fatalf("file not written: %v", err)
	}

	r := handler(hs, "fs_read")
	res, err := r.Call(ctx, json.RawMessage(`{"path":"dir/hello.txt"}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := res.Content[0].(mcp.TextContent).Text; got != "hi there" {
		t.Fatalf("read back %q", got)
	}
}

func TestListDirectory(t *testing.T) {
	sb := newSandbox(t)
	hs := Handlers(sb)
	ctx := context.Background()
	handler(hs, "fs_write").Call(ctx, json.RawMessage(`{"path":"a.txt","content":"x"}`))
	handler(hs, "fs_write").Call(ctx, json.RawMessage(`{"path":"b.txt","content":"yy"}`))

	res, err := handler(hs, "fs_list").Call(ctx, json.RawMessage(`{"path":"."}`))
	if err != nil {
		t.Fatal(err)
	}
	text := res.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "a.txt") || !strings.Contains(text, "b.txt") {
		t.Fatalf("listing missing entries: %s", text)
	}
}

func TestReadOnlyOmitsWrite(t *testing.T) {
	sb, _ := New(t.TempDir(), true)
	if handler(Handlers(sb), "fs_write") != nil {
		t.Fatal("read-only sandbox must not expose fs_write")
	}
}
