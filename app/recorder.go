package app

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// EntryType identifies the kind of recorded event.
type EntryType string

const (
	EntryUserInput  EntryType = "user_input"
	EntryAIResponse EntryType = "ai_response"
	EntryToolCall   EntryType = "tool_call"
	EntryError      EntryType = "error"
)

// RecordingEntry represents one event in a recorded session.
type RecordingEntry struct {
	Timestamp time.Time       `json:"timestamp"`
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
}

// UserInputData is the JSON payload for EntryUserInput.
type UserInputData struct {
	Text string `json:"text"`
}

// AIResponseData is the JSON payload for EntryAIResponse.
type AIResponseData struct {
	Delta string `json:"delta"`
}

// ToolCallData is the JSON payload for EntryToolCall.
type ToolCallData struct {
	Name string `json:"name"`
	Args string `json:"args"`
}

// ErrorData is the JSON payload for EntryError.
type ErrorData struct {
	Message string `json:"message"`
}

// RecordedFile is the top-level JSON structure for a saved recording.
type RecordedFile struct {
	Version int               `json:"version"`
	Start   time.Time         `json:"start"`
	Entries []RecordingEntry  `json:"entries"`
}

// Recorder captures session events for later playback.
// It is safe for concurrent use.
type Recorder struct {
	entries []RecordingEntry
	start   time.Time
	active  bool
}

// NewRecorder creates a new, inactive Recorder.
func NewRecorder() *Recorder {
	return &Recorder{}
}

// Start begins recording. Resets any previous recording.
func (r *Recorder) Start() {
	r.entries = nil
	r.start = time.Now()
	r.active = true
}

// Stop ends recording. No further entries will be accepted.
func (r *Recorder) Stop() {
	r.active = false
}

// IsActive reports whether the recorder is currently recording.
func (r *Recorder) IsActive() bool {
	return r.active
}

// EntryCount returns the number of recorded entries.
func (r *Recorder) EntryCount() int {
	return len(r.entries)
}

// Entries returns a copy of all recorded entries.
func (r *Recorder) Entries() []RecordingEntry {
	result := make([]RecordingEntry, len(r.entries))
	copy(result, r.entries)
	return result
}

// record appends an entry if the recorder is active.
func (r *Recorder) record(entryType EntryType, data interface{}) {
	if !r.active {
		return
	}
	raw, _ := json.Marshal(data)
	r.entries = append(r.entries, RecordingEntry{
		Timestamp: time.Now(),
		Type:      string(entryType),
		Data:      raw,
	})
}

// RecordUserInput records a user message.
func (r *Recorder) RecordUserInput(text string) {
	r.record(EntryUserInput, UserInputData{Text: text})
}

// RecordAIResponse records a streaming AI response delta.
func (r *Recorder) RecordAIResponse(delta string) {
	r.record(EntryAIResponse, AIResponseData{Delta: delta})
}

// RecordToolCall records a tool invocation.
func (r *Recorder) RecordToolCall(name, args string) {
	r.record(EntryToolCall, ToolCallData{Name: name, Args: args})
}

// RecordError records an error that occurred during the session.
func (r *Recorder) RecordError(err error) {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	r.record(EntryError, ErrorData{Message: msg})
}

// Save serializes the recording to JSON and writes it to w.
// Returns an error if the recorder is still active.
func (r *Recorder) Save(w io.Writer) error {
	file := RecordedFile{
		Version: 1,
		Start:   r.start,
		Entries: r.entries,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(file); err != nil {
		return fmt.Errorf("encode recording: %w", err)
	}
	return nil
}

// LoadRecording reads a JSON recording from r and returns the parsed entries.
func LoadRecording(r io.Reader) ([]RecordingEntry, error) {
	var file RecordedFile
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&file); err != nil {
		return nil, fmt.Errorf("decode recording: %w", err)
	}
	if file.Version != 1 {
		return nil, fmt.Errorf("unsupported recording version %d", file.Version)
	}
	return file.Entries, nil
}

// --- Player ---

// Player replays a recorded session entry by entry.
type Player struct {
	entries []RecordingEntry
	current int
	speed   float64
}

// NewPlayer creates a Player from a slice of recording entries.
// Default playback speed is 1.0x.
func NewPlayer(entries []RecordingEntry) *Player {
	return &Player{
		entries: entries,
		speed:   1.0,
	}
}

// Next returns the next entry in the recording.
// Returns io.EOF when all entries have been consumed.
func (p *Player) Next() (*RecordingEntry, error) {
	if p.current >= len(p.entries) {
		return nil, io.EOF
	}
	entry := &p.entries[p.current]
	p.current++
	return entry, nil
}

// HasNext reports whether there are more entries to play.
func (p *Player) HasNext() bool {
	return p.current < len(p.entries)
}

// Reset moves the playhead back to the beginning.
func (p *Player) Reset() {
	p.current = 0
}

// SetSpeed sets the playback speed multiplier (1.0 = real-time, 2.0 = 2x).
func (p *Player) SetSpeed(s float64) {
	if s <= 0 {
		s = 1.0
	}
	p.speed = s
}

// Speed returns the current playback speed.
func (p *Player) Speed() float64 {
	return p.speed
}

// Progress returns the playback progress as a ratio from 0.0 to 1.0.
func (p *Player) Progress() float64 {
	if len(p.entries) == 0 {
		return 0
	}
	return float64(p.current) / float64(len(p.entries))
}

// CurrentIndex returns the current playback position (0-based).
func (p *Player) CurrentIndex() int {
	return p.current
}

// TotalEntries returns the total number of entries in the recording.
func (p *Player) TotalEntries() int {
	return len(p.entries)
}

// Peek returns the entry at the current position without advancing.
// Returns nil if at the end.
func (p *Player) Peek() *RecordingEntry {
	if p.current >= len(p.entries) {
		return nil
	}
	return &p.entries[p.current]
}

// Duration returns the total time span of the recording.
// For an empty recording, returns 0.
func (p *Player) Duration() time.Duration {
	if len(p.entries) < 2 {
		return 0
	}
	return p.entries[len(p.entries)-1].Timestamp.Sub(p.entries[0].Timestamp)
}

// Delay returns the time delay before the next entry should be played,
// adjusted by the current speed multiplier.
// For the first entry, returns 0.
// When at the end, returns 0.
func (p *Player) Delay() time.Duration {
	if p.current == 0 || p.current >= len(p.entries) {
		return 0
	}
	gap := p.entries[p.current].Timestamp.Sub(p.entries[p.current-1].Timestamp)
	if p.speed > 0 {
		gap = time.Duration(float64(gap) / p.speed)
	}
	return gap
}
