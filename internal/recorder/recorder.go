// Package recorder provides session recording and playback for Fluui apps.
//
// Recorder captures terminal events (key, mouse, resize, focus) during a
// session. Player replays them with original timing for debugging, testing,
// and automated demos.
//
// Usage:
//
//	rec := recorder.NewRecorder()
//	app.OnKey(func(k *term.KeyEvent) {
//	    rec.RecordKey(k)
//	    // ... handle key
//	})
//	// After session:
//	rec.Save("session.json")
//
//	// Playback:
//	player, _ := recorder.Load("session.json")
//	player.Play(func(e recorder.RecordedEvent) {
//	    // Feed event to app
//	})
package recorder

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// EventType identifies the kind of recorded event.
type EventType uint8

const (
	EventKey EventType = iota
	EventMouse
	EventResize
	EventFocus
	EventPaste
)

// RecordedEvent is a single event captured during recording.
type RecordedEvent struct {
	Type      EventType `json:"type"`
	Timestamp int64     `json:"ts"` // nanoseconds since recording start

	// Key fields (for EventKey)
	KeyRune      rune   `json:"key_rune,omitempty"`
	KeyCode      int    `json:"key_code,omitempty"`
	KeyModifiers int    `json:"key_mods,omitempty"`

	// Mouse fields (for EventMouse)
	MouseX    int `json:"mx,omitempty"`
	MouseY    int `json:"my,omitempty"`
	MouseBtn  int `json:"mb,omitempty"`
	MouseAct  int `json:"ma,omitempty"`

	// Resize fields (for EventResize)
	Width  int `json:"w,omitempty"`
	Height int `json:"h,omitempty"`

	// Focus field (for EventFocus)
	Focused bool `json:"focused,omitempty"`

	// Paste field (for EventPaste)
	PasteText string `json:"paste,omitempty"`
}

// Recording is the serialized format for a session recording.
type Recording struct {
	Version   int             `json:"version"`
	Width     int             `json:"width"`
	Height    int             `json:"height"`
	CreatedAt time.Time       `json:"created_at"`
	Events    []RecordedEvent `json:"events"`
}

// Recorder captures events during a terminal session.
// It is thread-safe.
type Recorder struct {
	mu       sync.Mutex
	events   []RecordedEvent
	start    time.Time
	width    int
	height   int
	recording bool
}

// NewRecorder creates a new Recorder.
func NewRecorder() *Recorder {
	return &Recorder{
		events: make([]RecordedEvent, 0, 256),
	}
}

// Start begins recording. The timestamp for subsequent events is
// measured relative to this call.
func (r *Recorder) Start(width, height int) {
	r.mu.Lock()
	r.start = time.Now()
	r.width = width
	r.height = height
	r.recording = true
	r.events = r.events[:0]
	r.mu.Unlock()
}

// Stop ends recording.
func (r *Recorder) Stop() {
	r.mu.Lock()
	r.recording = false
	r.mu.Unlock()
}

// IsRecording returns whether the recorder is actively capturing.
func (r *Recorder) IsRecording() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.recording
}

// EventCount returns the number of captured events.
func (r *Recorder) EventCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.events)
}

// Events returns a copy of the recorded events.
func (r *Recorder) Events() []RecordedEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]RecordedEvent, len(r.events))
	copy(result, r.events)
	return result
}

func (r *Recorder) elapsedLocked() int64 {
	if r.start.IsZero() {
		return 0
	}
	return time.Since(r.start).Nanoseconds()
}

func (r *Recorder) record(e RecordedEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.recording {
		return
	}
	e.Timestamp = r.elapsedLocked()
	r.events = append(r.events, e)
}

// RecordKey records a key event.
func (r *Recorder) RecordKey(rune_ rune, keyCode, modifiers int) {
	r.record(RecordedEvent{
		Type:         EventKey,
		KeyRune:      rune_,
		KeyCode:      keyCode,
		KeyModifiers: modifiers,
	})
}

// RecordMouse records a mouse event.
func (r *Recorder) RecordMouse(x, y, button, action int) {
	r.record(RecordedEvent{
		Type:    EventMouse,
		MouseX:  x,
		MouseY:  y,
		MouseBtn: button,
		MouseAct: action,
	})
}

// RecordResize records a resize event.
func (r *Recorder) RecordResize(width, height int) {
	r.record(RecordedEvent{
		Type:   EventResize,
		Width:  width,
		Height: height,
	})
}

// RecordFocus records a focus change event.
func (r *Recorder) RecordFocus(focused bool) {
	r.record(RecordedEvent{
		Type:    EventFocus,
		Focused: focused,
	})
}

// RecordPaste records a paste event.
func (r *Recorder) RecordPaste(text string) {
	r.record(RecordedEvent{
		Type:      EventPaste,
		PasteText: text,
	})
}

// Save writes the recording to a file in JSON format.
func (r *Recorder) Save(path string) error {
	r.mu.Lock()
	rec := Recording{
		Version:   1,
		Width:     r.width,
		Height:    r.height,
		CreatedAt: time.Now(),
		Events:    make([]RecordedEvent, len(r.events)),
	}
	copy(rec.Events, r.events)
	r.mu.Unlock()

	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// SaveBytes returns the recording as JSON bytes.
func (r *Recorder) SaveBytes() ([]byte, error) {
	r.mu.Lock()
	rec := Recording{
		Version:   1,
		Width:     r.width,
		Height:    r.height,
		CreatedAt: time.Now(),
		Events:    make([]RecordedEvent, len(r.events)),
	}
	copy(rec.Events, r.events)
	r.mu.Unlock()

	return json.MarshalIndent(rec, "", "  ")
}

// Clear removes all recorded events.
func (r *Recorder) Clear() {
	r.mu.Lock()
	r.events = r.events[:0]
	r.mu.Unlock()
}

// Player replays a recorded session.
type Player struct {
	rec     *Recording
	cursor  int
	mu      sync.Mutex
	speed   float64 // playback speed multiplier (1.0 = real-time)
}

// NewPlayer creates a Player from a Recording.
func NewPlayer(rec *Recording) *Player {
	return &Player{
		rec:   rec,
		speed: 1.0,
	}
}

// Load reads a recording from a JSON file.
func Load(path string) (*Player, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rec Recording
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}
	return NewPlayer(&rec), nil
}

// LoadFromBytes creates a Player from JSON bytes.
func LoadFromBytes(data []byte) (*Player, error) {
	var rec Recording
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}
	return NewPlayer(&rec), nil
}

// SetSpeed sets the playback speed multiplier.
// 1.0 = real-time, 2.0 = 2x speed, 0.5 = half speed.
func (p *Player) SetSpeed(speed float64) {
	p.mu.Lock()
	p.speed = speed
	p.mu.Unlock()
}

// Speed returns the current playback speed.
func (p *Player) Speed() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.speed
}

// EventCount returns the total number of events in the recording.
func (p *Player) EventCount() int {
	return len(p.rec.Events)
}

// Width returns the terminal width from the recording.
func (p *Player) Width() int { return p.rec.Width }

// Height returns the terminal height from the recording.
func (p *Player) Height() int { return p.rec.Height }

// Remaining returns the number of events left to play.
func (p *Player) Remaining() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.rec.Events) - p.cursor
}

// HasNext returns whether there are more events to play.
func (p *Player) HasNext() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cursor < len(p.rec.Events)
}

// Next returns the next event without advancing.
// Returns ok=false if there are no more events.
func (p *Player) Next() (RecordedEvent, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cursor >= len(p.rec.Events) {
		return RecordedEvent{}, false
	}
	return p.rec.Events[p.cursor], true
}

// NextDelay returns the time delay until the next event should fire,
// relative to the previous event. Returns 0 for the first event.
func (p *Player) NextDelay() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cursor >= len(p.rec.Events) {
		return 0
	}
	if p.cursor == 0 {
		return 0
	}
	prev := p.rec.Events[p.cursor-1].Timestamp
	curr := p.rec.Events[p.cursor].Timestamp
	delay := time.Duration(curr-prev) * time.Nanosecond
	speed := p.speed
	if speed <= 0 {
		speed = 1.0
	}
	return time.Duration(float64(delay) / speed)
}

// Advance moves the cursor forward by one event.
func (p *Player) Advance() {
	p.mu.Lock()
	if p.cursor < len(p.rec.Events) {
		p.cursor++
	}
	p.mu.Unlock()
}

// Reset moves the cursor back to the beginning.
func (p *Player) Reset() {
	p.mu.Lock()
	p.cursor = 0
	p.mu.Unlock()
}

// Cursor returns the current playback position (0-indexed).
func (p *Player) Cursor() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cursor
}

// Play replays all events with original timing.
// The callback is called for each event. This function blocks until
// all events have been played or the callback returns false.
func (p *Player) Play(fn func(RecordedEvent) bool) {
	p.Reset()
	for p.HasNext() {
		delay := p.NextDelay()
		if delay > 0 {
			time.Sleep(delay)
		}
		ev, ok := p.Next()
		if !ok {
			break
		}
		p.Advance()
		if !fn(ev) {
			break
		}
	}
}

// PlayFast replays all events without delays (as fast as possible).
// The callback is called for each event. Returns false from callback to stop.
func (p *Player) PlayFast(fn func(RecordedEvent) bool) {
	p.Reset()
	for _, ev := range p.rec.Events {
		if !fn(ev) {
			break
		}
	}
}

// Events returns all events in the recording (read-only).
func (p *Player) Events() []RecordedEvent {
	return p.rec.Events
}

// TotalDuration returns the total duration of the recording.
func (p *Player) TotalDuration() time.Duration {
	if len(p.rec.Events) == 0 {
		return 0
	}
	last := p.rec.Events[len(p.rec.Events)-1].Timestamp
	return time.Duration(last) * time.Nanosecond
}
