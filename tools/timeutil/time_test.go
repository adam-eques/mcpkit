package timeutil

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/adam-eques/mcpkit/mcp"
)

func TestNowFixedClock(t *testing.T) {
	fixed := time.Date(2025, 10, 1, 12, 0, 0, 0, time.UTC)
	hs := Handlers(func() time.Time { return fixed })
	res, err := hs[0].Call(context.Background(), json.RawMessage(`{"timezone":"UTC"}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := res.Content[0].(mcp.TextContent).Text; got != "2025-10-01T12:00:00Z" {
		t.Fatalf("now=%q", got)
	}
}

func TestConvert(t *testing.T) {
	hs := Handlers(nil)
	res, err := hs[1].Call(context.Background(), json.RawMessage(`{"timestamp":"2025-10-01T12:00:00Z","to":"UTC"}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(res.Content[0].(mcp.TextContent).Text, "2025-10-01T12:00:00") {
		t.Fatalf("convert=%+v", res)
	}
}

func TestUnknownZone(t *testing.T) {
	hs := Handlers(nil)
	res, _ := hs[0].Call(context.Background(), json.RawMessage(`{"timezone":"Mars/Olympus"}`))
	if !res.IsError {
		t.Fatal("expected error for unknown zone")
	}
}
