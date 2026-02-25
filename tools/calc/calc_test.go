package calc

import (
	"context"
	"encoding/json"
	"math"
	"testing"

	"github.com/adam-eques/mcpkit/mcp"
)

func TestEval(t *testing.T) {
	cases := []struct {
		expr string
		want float64
	}{
		{"1 + 2 * 3", 7},
		{"(1 + 2) * 3", 9},
		{"2 ^ 3 ^ 2", 512}, // right associative
		{"-2 ^ 2", -4},      // unary binds looser than power here: -(2^2)
		{"10 % 3", 1},
		{"sqrt(16)", 4},
		{"abs(-5)", 5},
		{"round(3.6)", 4},
		{"2 * pi", 2 * math.Pi},
		{"ln(e)", 1},
		{"1.5e2", 150},
	}
	for _, tc := range cases {
		got, err := Eval(tc.expr)
		if err != nil {
			t.Errorf("Eval(%q) error: %v", tc.expr, err)
			continue
		}
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("Eval(%q)=%v want %v", tc.expr, got, tc.want)
		}
	}
}

func TestEvalErrors(t *testing.T) {
	for _, expr := range []string{"1 +", "(1", "1 / 0", "foo(2)", "2 @ 3", ""} {
		if _, err := Eval(expr); err == nil {
			t.Errorf("Eval(%q) expected error", expr)
		}
	}
}

func TestTool(t *testing.T) {
	res, err := New().Call(context.Background(), json.RawMessage(`{"expression":"6*7"}`))
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected error result: %+v", res)
	}
	if got := res.Content[0].(mcp.TextContent).Text; got != "42" {
		t.Fatalf("got %q", got)
	}
}

func TestToolReportsErrorResult(t *testing.T) {
	res, _ := New().Call(context.Background(), json.RawMessage(`{"expression":"1/0"}`))
	if !res.IsError {
		t.Fatal("expected IsError for division by zero")
	}
}
