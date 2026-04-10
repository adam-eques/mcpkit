// Package shell implements the "shell_exec" tool. For safety the caller must
// pass the executable and its arguments separately (never a shell string), the
// executable must appear on an explicit allowlist, and every invocation runs
// under a timeout. With an empty allowlist the tool refuses to run anything.
package shell

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/tools"
)

// Tool runs allowlisted executables.
type Tool struct {
	allow   map[string]bool
	timeout time.Duration
	maxOut  int
}

// New returns the shell_exec tool. allow lists the executable names that may be
// run; an empty list disables execution.
func New(allow []string, timeout time.Duration) tools.Handler {
	set := make(map[string]bool, len(allow))
	for _, a := range allow {
		set[a] = true
	}
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &Tool{allow: set, timeout: timeout, maxOut: 64 << 10}
}

// Definition implements tools.Handler.
func (t *Tool) Definition() mcp.Tool {
	return mcp.Tool{
		Name:        "shell_exec",
		Description: "Run an allowlisted executable with explicit arguments and return its combined output. Arguments are never passed through a shell.",
		InputSchema: tools.Object(map[string]tools.Prop{
			"command": tools.Str("the executable name; must be allowlisted"),
			"args":    {Type: "array", Description: "arguments passed verbatim", Items: &tools.Prop{Type: "string"}},
		}, "command"),
		Annotations: &mcp.ToolAnnotations{Title: "Shell Exec", DestructiveHint: true},
	}
}

// Call implements tools.Handler.
func (t *Tool) Call(ctx context.Context, args json.RawMessage) (*mcp.CallToolResult, error) {
	var in struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}
	if err := tools.Bind(args, &in); err != nil {
		return nil, err
	}
	if in.Command == "" {
		return mcp.ErrorResult("command is required"), nil
	}
	if !t.allow[in.Command] {
		return mcp.ErrorResult(fmt.Sprintf("command %q is not allowlisted", in.Command)), nil
	}

	runCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, in.Command, in.Args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()

	text := out.String()
	if len(text) > t.maxOut {
		text = text[:t.maxOut] + "\n[output truncated]"
	}
	var b strings.Builder
	if runCtx.Err() == context.DeadlineExceeded {
		b.WriteString("[timed out]\n")
	}
	if err != nil {
		fmt.Fprintf(&b, "exit error: %v\n", err)
	}
	b.WriteString(text)
	res := mcp.TextResult(b.String())
	if err != nil {
		res.IsError = true
	}
	return res, nil
}
