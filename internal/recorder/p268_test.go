package recorder

import (
	"testing"
)

func TestPlayer_Play_EmptyEvents_P268(t *testing.T) {
	rec := &Recording{Events: []RecordedEvent{}}
	p := NewPlayer(rec)
	count := 0
	p.Play(func(ev RecordedEvent) bool {
		count++
		return true
	})
	if count != 0 {
		t.Errorf("expected 0 events, got %d", count)
	}
}

func TestPlayer_Play_StopsOnFalse_P268(t *testing.T) {
	events := []RecordedEvent{
		{Type: EventKey, KeyRune: 'a'},
		{Type: EventKey, KeyRune: 'b'},
		{Type: EventKey, KeyRune: 'c'},
	}
	rec := &Recording{Events: events}
	p := NewPlayer(rec)
	count := 0
	p.Play(func(ev RecordedEvent) bool {
		count++
		return count < 2
	})
	if count != 2 {
		t.Errorf("expected 2 events before stop, got %d", count)
	}
}

func TestPlayer_Next_AtEnd_P268(t *testing.T) {
	events := []RecordedEvent{
		{Type: EventKey, KeyRune: 'a'},
	}
	rec := &Recording{Events: events}
	p := NewPlayer(rec)
	p.Advance()
	_, ok := p.Next()
	if ok {
		t.Error("Next at end should return false")
	}
}

func TestPlayer_PlayFast_AllConsumed_P268(t *testing.T) {
	events := []RecordedEvent{
		{Type: EventKey, KeyRune: 'a'},
		{Type: EventKey, KeyRune: 'b'},
	}
	rec := &Recording{Events: events}
	p := NewPlayer(rec)
	count := 0
	p.PlayFast(func(ev RecordedEvent) bool {
		count++
		return true
	})
	if count != 2 {
		t.Errorf("expected 2 events, got %d", count)
	}
}

func TestPlayer_PlayFast_StopsOnFalse_P268(t *testing.T) {
	events := []RecordedEvent{
		{Type: EventKey, KeyRune: 'a'},
		{Type: EventKey, KeyRune: 'b'},
		{Type: EventKey, KeyRune: 'c'},
	}
	rec := &Recording{Events: events}
	p := NewPlayer(rec)
	count := 0
	p.PlayFast(func(ev RecordedEvent) bool {
		count++
		return false
	})
	if count != 1 {
		t.Errorf("expected 1 event before stop, got %d", count)
	}
}
