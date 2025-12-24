package version

import "testing"

func TestString(t *testing.T) {
	Version = "1.2.3"
	Commit = "abc1234"
	if got := String(); got != "1.2.3+abc1234" {
		t.Fatalf("String()=%q", got)
	}
	Commit = "unknown"
	if got := String(); got != "1.2.3" {
		t.Fatalf("String()=%q", got)
	}
}
