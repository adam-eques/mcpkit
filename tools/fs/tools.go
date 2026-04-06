package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/tools"
)

// Handlers returns the filesystem tools bound to sb. When the sandbox is
// read-only the write tool is omitted.
func Handlers(sb *Sandbox) []tools.Handler {
	hs := []tools.Handler{readTool{sb}, listTool{sb}}
	if !sb.ReadOnly {
		hs = append(hs, writeTool{sb})
	}
	return hs
}

type readTool struct{ sb *Sandbox }

func (readTool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "fs_read",
		Description: "Read a UTF-8 text file from the sandbox root. The path is relative to the configured root directory.",
		InputSchema: tools.Object(map[string]tools.Prop{
			"path": tools.Str("file path relative to the sandbox root"),
		}, "path"),
		Annotations: &mcp.ToolAnnotations{Title: "Read File", ReadOnlyHint: true},
	}
}

func (t readTool) Call(_ context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct {
		Path string `json:"path"`
	}
	if err := tools.Bind(args, &in); err != nil {
		return nil, err
	}
	abs, err := t.sb.Resolve(in.Path)
	if err != nil {
		return mcp.ErrorResult(err.Error()), nil
	}
	f, err := os.Open(abs)
	if err != nil {
		return mcp.ErrorResult("read failed: " + err.Error()), nil
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, t.sb.MaxBytes))
	if err != nil {
		return mcp.ErrorResult("read failed: " + err.Error()), nil
	}
	return mcp.TextResult(string(data)), nil
}

type writeTool struct{ sb *Sandbox }

func (writeTool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "fs_write",
		Description: "Write a UTF-8 text file inside the sandbox root, creating parent directories as needed.",
		InputSchema: tools.Object(map[string]tools.Prop{
			"path":    tools.Str("file path relative to the sandbox root"),
			"content": tools.Str("text content to write"),
		}, "path", "content"),
		Annotations: &mcp.ToolAnnotations{Title: "Write File"},
	}
}

func (t writeTool) Call(_ context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := tools.Bind(args, &in); err != nil {
		return nil, err
	}
	abs, err := t.sb.Resolve(in.Path)
	if err != nil {
		return mcp.ErrorResult(err.Error()), nil
	}
	if int64(len(in.Content)) > t.sb.MaxBytes {
		return mcp.ErrorResult("content exceeds size limit"), nil
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return mcp.ErrorResult("write failed: " + err.Error()), nil
	}
	if err := os.WriteFile(abs, []byte(in.Content), 0o644); err != nil {
		return mcp.ErrorResult("write failed: " + err.Error()), nil
	}
	return mcp.TextResult(fmt.Sprintf("wrote %d bytes to %s", len(in.Content), t.sb.Rel(abs))), nil
}

type listTool struct{ sb *Sandbox }

func (listTool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "fs_list",
		Description: "List the entries of a directory inside the sandbox root.",
		InputSchema: tools.Object(map[string]tools.Prop{
			"path": {Type: "string", Description: "directory relative to the sandbox root", Default: "."},
		}),
		Annotations: &mcp.ToolAnnotations{Title: "List Directory", ReadOnlyHint: true},
	}
}

func (t listTool) Call(_ context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct {
		Path string `json:"path"`
	}
	if err := tools.Bind(args, &in); err != nil {
		return nil, err
	}
	if in.Path == "" {
		in.Path = "."
	}
	abs, err := t.sb.Resolve(in.Path)
	if err != nil {
		return mcp.ErrorResult(err.Error()), nil
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return mcp.ErrorResult("list failed: " + err.Error()), nil
	}
	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		size := int64(0)
		if err == nil {
			size = info.Size()
		}
		kind := "f"
		if e.IsDir() {
			kind = "d"
		}
		lines = append(lines, fmt.Sprintf("%s %8d  %s", kind, size, e.Name()))
	}
	sort.Strings(lines)
	if len(lines) == 0 {
		return mcp.TextResult("(empty directory)"), nil
	}
	return mcp.TextResult(strings.Join(lines, "\n")), nil
}
