package recorder

import "testing"

// P239: cover recorder NextDelay speed<=0 + Play delay>0 paths

func TestPlayer_NextDelay_SpeedZero_P239(t *testing.T) {
	rec := &Recording{
		Events: []RecordedEvent{
			{Timestamp: 0, Type: EventKey},
			{Timestamp: 1000000, Type: EventKey}, // 1ms later
		},
	}
	p := NewPlayer(rec)
	p.speed = 0 // speed<=0 → defaults to 1.0
	p.Advance() // cursor=1
	d := p.NextDelay()
	if d <= 0 {
		t.Error("NextDelay with speed=0 should still return positive delay")
	}
}

func TestPlayer_NextDelay_CursorZero_P239(t *testing.T) {
	rec := &Recording{
		Events: []RecordedEvent{
			{Timestamp: 0, Type: EventKey},
		},
	}
	p := NewPlayer(rec)
	d := p.NextDelay()
	if d != 0 {
		t.Error("NextDelay at cursor=0 should return 0")
	}
}

func TestPlayer_PlayWithDelay_P239(t *testing.T) {
	rec := &Recording{
		Events: []RecordedEvent{
			{Timestamp: 0, Type: EventKey},
			{Timestamp: 2000000, Type: EventKey}, // 2ms later
		},
	}
	p := NewPlayer(rec)
	p.speed = 100 // fast playback
	count := 0
	p.Play(func(ev RecordedEvent) bool {
		count++
		return true
	})
	if count != 2 {
		t.Errorf("Play should iterate all events, got %d", count)
	}
}
