package component

import (
	"testing"
)

// TestP338_AppendTokenCount verifies all branches of appendTokenCount.
func TestP338_AppendTokenCount(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{42, "42"},
		{999, "999"},
		{1000, "1.0k"},
		{1500, "1.5k"},
		{10000, "10.0k"},
		{999999, "1000.0k"},
		{1000000, "1.0M"},
		{2500000, "2.5M"},
	}
	for _, tt := range tests {
		got := string(appendTokenCount(nil, tt.n))
		if got != tt.want {
			t.Errorf("appendTokenCount(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

// TestP338_AppendProgressBar verifies progress bar rendering edge cases.
func TestP338_AppendProgressBar(t *testing.T) {
	// 0% bar
	got := string(appendProgressBar(nil, 0, 5))
	if got != "\u2591\u2591\u2591\u2591\u2591" {
		t.Errorf("0%% bar = %q", got)
	}

	// 100% bar
	got = string(appendProgressBar(nil, 100, 5))
	if got != "\u2593\u2593\u2593\u2593\u2593" {
		t.Errorf("100%% bar = %q", got)
	}

	// 50% bar
	got = string(appendProgressBar(nil, 50, 4))
	if got != "\u2593\u2593\u2591\u2591" {
		t.Errorf("50%% bar = %q", got)
	}

	// Width < 1 should clamp to 1
	got = string(appendProgressBar(nil, 50, 0))
	if len([]rune(got)) != 1 {
		t.Errorf("width 0 clamped: runes = %d, want 1", len([]rune(got)))
	}

	// Negative pct → 0 filled
	got = string(appendProgressBar(nil, -10, 3))
	if got != "\u2591\u2591\u2591" {
		t.Errorf("negative pct bar = %q", got)
	}

	// Pct > 100 → all filled
	got = string(appendProgressBar(nil, 150, 3))
	if got != "\u2593\u2593\u2593" {
		t.Errorf("over 100 pct bar = %q", got)
	}

	// Large negative pct → filled clamped to 0 (hits filled<0 branch)
	got = string(appendProgressBar(nil, -50, 4))
	if got != "\u2591\u2591\u2591\u2591" {
		t.Errorf("large negative pct bar = %q", got)
	}
}

// TestP338_AppendCost verifies all cost formatting branches.
func TestP338_AppendCost(t *testing.T) {
	// < $0.01 → 4 decimal places
	got := string(appendCost(nil, 0.001))
	if got != "0.0010" {
		t.Errorf("micro cost = %q, want 0.0010", got)
	}

	// >= $0.01 → 2 decimal places
	got = string(appendCost(nil, 0.024))
	if got != "0.02" {
		t.Errorf("small cost = %q, want 0.02", got)
	}

	// Large cost
	got = string(appendCost(nil, 1.50))
	if got != "1.50" {
		t.Errorf("large cost = %q, want 1.50", got)
	}
}


