package mcp

import (
	"encoding/json"
	"testing"
)

func TestInitializeResultRoundTrip(t *testing.T) {
	in := InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools:   &ToolsCapability{},
			Logging: &struct{}{},
		},
		ServerInfo:   Implementation{Name: "mcpkit", Version: "0.1.0"},
		Instructions: "hello",
	}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out InitializeResult
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out.ServerInfo.Name != "mcpkit" || out.Capabilities.Tools == nil {
		t.Fatalf("round trip mismatch: %+v", out)
	}
}

func TestServerCapabilitiesOmitEmpty(t *testing.T) {
	raw, _ := json.Marshal(ServerCapabilities{})
	if string(raw) != "{}" {
		t.Fatalf("empty capabilities should marshal to {}, got %s", raw)
	}
}
