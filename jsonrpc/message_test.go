package jsonrpc

import (
	"encoding/json"
	"testing"
)

func TestIDRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"int", `7`, "7"},
		{"string", `"abc"`, "abc"},
		{"null", `null`, "null"},
		{"zero", `0`, "0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var id ID
			if err := json.Unmarshal([]byte(tc.in), &id); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got := id.String(); got != tc.want {
				t.Fatalf("String()=%q want %q", got, tc.want)
			}
			out, err := json.Marshal(id)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(out) != tc.in {
				t.Fatalf("marshal=%s want %s", out, tc.in)
			}
		})
	}
}

func TestIDRejectsFractional(t *testing.T) {
	var id ID
	if err := json.Unmarshal([]byte(`1.5`), &id); err == nil {
		t.Fatal("expected error for fractional id")
	}
}

func TestRequestIsNotification(t *testing.T) {
	n := NewNotification("notifications/initialized", nil)
	if !n.IsNotification() {
		t.Fatal("expected notification")
	}
	r := NewRequest(Int64ID(1), "ping", nil)
	if r.IsNotification() {
		t.Fatal("expected request, got notification")
	}
	raw, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if json.Valid(raw) && contains(string(raw), `"id"`) {
		t.Fatalf("notification must not carry an id: %s", raw)
	}
}

func TestErrorImplementsError(t *testing.T) {
	var err error = NewError(CodeInvalidParams, "bad")
	if err.Error() == "" {
		t.Fatal("empty error string")
	}
	e := InvalidParams("").WithData(map[string]int{"n": 1})
	if len(e.Data) == 0 {
		t.Fatal("expected data attached")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
