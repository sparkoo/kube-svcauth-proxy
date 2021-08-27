package proxy

import (
	"testing"
)

func TestStripToken(t *testing.T) {
	if stripped := stripToken("bearer abc"); stripped != "abc" {
		t.Fatalf("expected 'abc', but got '%s'", stripped)
	}

	if stripped := stripToken("Bearer abc"); stripped != "abc" {
		t.Fatalf("expected 'abc', but got '%s'", stripped)
	}

	if stripped := stripToken("abc"); stripped != "abc" {
		t.Fatalf("expected 'abc', but got '%s'", stripped)
	}
}
