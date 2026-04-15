package fetch

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adam-eques/mcpkit/mcp"
)

func TestFetchAllowPrivateForLocalServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		w.Write([]byte("pong"))
	}))
	defer srv.Close()

	// The test server binds to loopback, so private connections must be enabled.
	f := NewFetcher(WithAllowPrivate(true))
	res, err := f.Do(context.Background(), "GET", srv.URL, nil, "")
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	if res.Status != 200 || res.Body != "pong" {
		t.Fatalf("unexpected response: %+v", res)
	}
}

func TestFetchRejectsNonHTTP(t *testing.T) {
	f := NewFetcher()
	if _, err := f.Do(context.Background(), "GET", "file:///etc/passwd", nil, ""); err == nil {
		t.Fatal("expected rejection of non-http scheme")
	}
}

func TestFetchBlocksLoopbackByDefault(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	f := NewFetcher() // guard enabled
	if _, err := f.Do(context.Background(), "GET", srv.URL, nil, ""); err == nil {
		t.Fatal("expected SSRF guard to block loopback")
	}
}

func TestIsBlocked(t *testing.T) {
	blocked := []string{"127.0.0.1", "10.0.0.5", "192.168.1.1", "169.254.1.1", "::1", "fc00::1"}
	for _, s := range blocked {
		if !isBlocked(net.ParseIP(s)) {
			t.Errorf("%s should be blocked", s)
		}
	}
	if isBlocked(net.ParseIP("8.8.8.8")) {
		t.Error("8.8.8.8 should be allowed")
	}
}

func TestToolTruncation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(strings.Repeat("x", 100)))
	}))
	defer srv.Close()
	tool := New(WithAllowPrivate(true), WithMaxBytes(10))
	res, err := tool.Call(context.Background(), json.RawMessage(`{"url":"`+srv.URL+`"}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Content[0].(mcp.TextContent).Text, "[body truncated]") {
		t.Fatal("expected truncation marker")
	}
}
