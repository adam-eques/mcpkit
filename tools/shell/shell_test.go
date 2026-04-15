package shell

import (
	"context"
	"encoding/json"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/adam-eques/mcpkit/mcp"
)

func TestNotAllowlisted(t *testing.T) {
	tool := New(nil, time.Second)
	res, _ := tool.Call(context.Background(), json.RawMessage(`{"command":"rm","args":["-rf","/"]}`))
	if !res.IsError && !strings.Contains(res.Content[0].(mcp.TextContent).Text, "not allowlisted") {
		t.Fatalf("expected rejection, got %+v", res)
	}
}

func TestAllowlistedEcho(t *testing.T) {
	cmd := "echo"
	args := `["hello"]`
	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = `["/c","echo hello"]`
	}
	tool := New([]string{cmd}, 5*time.Second)
	res, err := tool.Call(context.Background(), json.RawMessage(`{"command":"`+cmd+`","args":`+args+`}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Content[0].(mcp.TextContent).Text, "hello") {
		t.Fatalf("unexpected output: %+v", res)
	}
}
