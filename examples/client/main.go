// Command example-client is a minimal MCP client. It launches the mcpkit server
// as a subprocess, performs the initialize handshake, lists the available tools
// and calls the calculator, printing each exchange. It demonstrates how little
// code is needed to speak the protocol against this server.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "example-client:", err)
		os.Exit(1)
	}
}

func run() error {
	// Launch the server: `go run ./cmd/mcpkit` from the repository root.
	cmd := exec.Command("go", "run", "./cmd/mcpkit")
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	defer cmd.Wait()
	defer stdin.Close()

	rpc := &conn{w: stdin, r: bufio.NewReader(stdout)}

	if err := rpc.call(1, "initialize", map[string]any{
		"protocolVersion": "2025-06-18",
		"clientInfo":      map[string]string{"name": "example-client", "version": "1.0"},
	}); err != nil {
		return err
	}
	if err := rpc.notify("notifications/initialized", nil); err != nil {
		return err
	}
	if err := rpc.call(2, "tools/list", nil); err != nil {
		return err
	}
	return rpc.call(3, "tools/call", map[string]any{
		"name":      "calculate",
		"arguments": map[string]string{"expression": "2 ^ 10 + sqrt(81)"},
	})
}

type conn struct {
	w io.Writer
	r *bufio.Reader
}

func (c *conn) call(id int, method string, params any) error {
	if err := c.send(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params}); err != nil {
		return err
	}
	line, err := c.r.ReadBytes('\n')
	if err != nil {
		return err
	}
	fmt.Printf("<- %s", line)
	return nil
}

func (c *conn) notify(method string, params any) error {
	return c.send(map[string]any{"jsonrpc": "2.0", "method": method, "params": params})
}

func (c *conn) send(msg any) error {
	raw, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	fmt.Printf("-> %s\n", raw)
	_, err = fmt.Fprintf(c.w, "%s\n", raw)
	return err
}
