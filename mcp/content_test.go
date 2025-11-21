package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTextContentMarshalsType(t *testing.T) {
	raw, err := json.Marshal(Text("hello"))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), `"type":"text"`) {
		t.Fatalf("missing type discriminator: %s", raw)
	}
	if !strings.Contains(string(raw), `"text":"hello"`) {
		t.Fatalf("missing text: %s", raw)
	}
}

func TestDecodeContent(t *testing.T) {
	c, err := DecodeContent(json.RawMessage(`{"type":"text","text":"hi"}`))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	tc, ok := c.(TextContent)
	if !ok {
		t.Fatalf("want TextContent, got %T", c)
	}
	if tc.Text != "hi" {
		t.Fatalf("text=%q", tc.Text)
	}
}

func TestDecodeContentUnknown(t *testing.T) {
	if _, err := DecodeContent(json.RawMessage(`{"type":"video"}`)); err == nil {
		t.Fatal("expected error for unknown content type")
	}
}

func TestNegotiateVersion(t *testing.T) {
	if got := NegotiateVersion("2024-11-05"); got != "2024-11-05" {
		t.Fatalf("supported version not echoed: %s", got)
	}
	if got := NegotiateVersion("1999-01-01"); got != ProtocolVersion {
		t.Fatalf("fallback=%s want %s", got, ProtocolVersion)
	}
}

func TestCallToolResultHelpers(t *testing.T) {
	if ErrorResult("boom").IsError != true {
		t.Fatal("ErrorResult should set IsError")
	}
	if got := len(TextResult("x").Content); got != 1 {
		t.Fatalf("content len=%d", got)
	}
}
