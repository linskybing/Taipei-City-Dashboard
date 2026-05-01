package assistant

import "testing"

func TestResolvedSearchLimitDefaultsAndClamps(t *testing.T) {
	if got := resolvedSearchLimit(nil); got != 5 {
		t.Fatalf("default limit = %d, want 5", got)
	}
	below := 0
	if got := resolvedSearchLimit(&below); got != 1 {
		t.Fatalf("clamped low limit = %d, want 1", got)
	}
	above := 20
	if got := resolvedSearchLimit(&above); got != 10 {
		t.Fatalf("clamped high limit = %d, want 10", got)
	}
}

func TestResolvedSearchScoreThresholdPreservesExplicitZero(t *testing.T) {
	if got := resolvedSearchScoreThreshold(nil); got != 0.78 {
		t.Fatalf("default score threshold = %v, want 0.78", got)
	}
	zero := 0.0
	if got := resolvedSearchScoreThreshold(&zero); got != 0 {
		t.Fatalf("explicit zero score threshold = %v, want 0", got)
	}
	below := -1.0
	if got := resolvedSearchScoreThreshold(&below); got != 0 {
		t.Fatalf("clamped low score threshold = %v, want 0", got)
	}
	above := 2.0
	if got := resolvedSearchScoreThreshold(&above); got != 1 {
		t.Fatalf("clamped high score threshold = %v, want 1", got)
	}
}