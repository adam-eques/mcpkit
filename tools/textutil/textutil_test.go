package textutil

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/adam-eques/mcpkit/mcp"
	"github.com/adam-eques/mcpkit/tools"
)

func byName(name string) tools.Handler {
	for _, h := range Handlers() {
		if h.Definition().Name == name {
			return h
		}
	}
	return nil
}

func text(res *mcp.CallToolResult) string { return res.Content[0].(mcp.TextContent).Text }

func TestHashSHA256(t *testing.T) {
	res, err := byName("hash").Call(context.Background(), json.RawMessage(`{"text":"abc"}`))
	if err != nil {
		t.Fatal(err)
	}
	want := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if text(res) != want {
		t.Fatalf("sha256(abc)=%s", text(res))
	}
}

func TestBase64RoundTrip(t *testing.T) {
	enc, _ := byName("base64").Call(context.Background(), json.RawMessage(`{"text":"hi","operation":"encode"}`))
	if text(enc) != "aGk=" {
		t.Fatalf("encode=%s", text(enc))
	}
	dec, _ := byName("base64").Call(context.Background(), json.RawMessage(`{"text":"aGk=","operation":"decode"}`))
	if text(dec) != "hi" {
		t.Fatalf("decode=%s", text(dec))
	}
}

func TestUUIDFormat(t *testing.T) {
	res, err := byName("uuid").Call(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if !re.MatchString(text(res)) {
		t.Fatalf("not a v4 uuid: %s", text(res))
	}
}
