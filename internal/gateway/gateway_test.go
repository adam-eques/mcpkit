package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/adam-eques/mcpkit/internal/log"
	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/server"
	"github.com/adam-eques/mcpkit/tools"
	"github.com/adam-eques/mcpkit/tools/calc"
)

func newServer() *server.Server {
	reg := tools.NewRegistry()
	reg.MustRegister(calc.New())
	return server.New(mcp.Implementation{Name: "gw", Version: "0"}, reg)
}

func TestRPCEndpoint(t *testing.T) {
	h := Handler(newServer(), log.Discard())
	srv := httptest.NewServer(h)
	defer srv.Close()

	initReq := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}`
	resp, err := http.Post(srv.URL+"/rpc", "application/json", strings.NewReader(initReq))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	var jr struct {
		Result mcp.InitializeResult `json:"result"`
		Error  any                  `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jr); err != nil {
		t.Fatal(err)
	}
	if jr.Error != nil {
		t.Fatalf("rpc error: %v", jr.Error)
	}
	if jr.Result.ServerInfo.Name != "gw" {
		t.Fatalf("unexpected server info: %+v", jr.Result.ServerInfo)
	}
}

func TestHealthz(t *testing.T) {
	h := Handler(newServer(), log.Discard())
	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "ok") {
		t.Fatalf("healthz failed: %d %s", rec.Code, rec.Body.String())
	}
}

func TestNotificationReturns204(t *testing.T) {
	h := Handler(newServer(), log.Discard())
	req := httptest.NewRequest("POST", "/rpc", strings.NewReader(`{"jsonrpc":"2.0","method":"notifications/initialized"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}
