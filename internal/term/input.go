package term

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// Event types.
type EventType uint8

const (
	EventKey EventType = iota
	EventMouse
	EventPaste
	EventResize
	EventClipboard // OSC52 clipboard response
)

// Event is a parsed terminal input event.
type Event struct {
	Type      EventType
	Key       *KeyEvent
	Mouse     *MouseEvent
	Paste     string
	Clipboard string // for EventClipboard (OSC52 response)
	Width     int    // for EventResize
	Height    int    // for EventResize
}

// KeyEvent represents a keyboard input.
type KeyEvent struct {
	Key       KeyCode
	Modifiers ModMask
	Rune      rune // valid for printable characters
}

// KeyCode identifies a non-printable key.
type KeyCode uint16

const (
	KeyUnknown KeyCode = iota
	KeyEnter
	KeyTab
	KeyBacktab
	KeyBackspace
	KeyDelete
	KeyInsert
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown
	KeyUp
	KeyDown
	KeyRight
	KeyLeft
	KeyEscape
	KeySpace
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
)

// ModMask is a bitmask of modifier keys.
type ModMask uint8

const (
	ModCtrl  ModMask = 1 << iota
	ModAlt
	ModShift
)

// String returns a human-readable description.
func (m ModMask) String() string {
	s := ""
	if m&ModCtrl != 0 {
		s += "Ctrl+"
	}
	if m&ModAlt != 0 {
		s += "Alt+"
	}
	if m&ModShift != 0 {
		s += "Shift+"
	}
	return s
}

// MouseButton identifies a mouse button.
type MouseButton uint8

const (
	MouseLeft MouseButton = iota
	MouseRight
	MouseMiddle
	MouseWheelUp
	MouseWheelDown
	MouseNone
)

// MouseAction describes what happened with the mouse.
type MouseAction uint8

const (
	MouseDown MouseAction = iota
	MouseUp
	MouseMove
	MouseDrag
	MouseWheel
)

// MouseEvent represents a mouse input.
type MouseEvent struct {
	X, Y      int
	Button    MouseButton
	Modifiers ModMask
	Action    MouseAction
}

// --- Parser ---

// parseState tracks the parser state machine.
type parseState uint8

const (
	stateNormal   parseState = iota
	stateEscape              // after ESC
	stateCSI                 // after ESC [
	stateSS3                 // after ESC O
	statePaste               // inside bracketed paste
	statePasteESC            // inside paste, got ESC
	stateUTF8                // accumulating UTF-8 continuation bytes
	stateOSC                 // after ESC ] (OSC sequence)
	stateOSCESC              // inside OSC, got ESC (possible ST)
)

// Parser parses a raw byte stream into structured Events.
// It handles partial sequences gracefully (they arrive split across reads).
type Parser struct {
	state parseState
	buf   []byte
	paste []byte

	// UTF-8 accumulation (BUG 1 fix)
	utf8Buf    [4]byte
	utf8Len    int
	utf8Expect int

	// ESC timeout tracking (BUG 2 fix)
	escAt     time.Time
	escTimeout time.Duration // configurable per-instance, default 50ms

	// Thread safety: Feed runs in readInput goroutine,
	// FeedTimeout runs in the event loop goroutine (BUG 2 fix)
	mu sync.Mutex

	// Paste CSI containment (BUG 3 fix)
	inPaste bool
}

// DefaultESCTimeout is the default duration after which a lone ESC is
// treated as KeyEscape.
const DefaultESCTimeout = 50 * time.Millisecond

// escTimeoutForTest allows tests to override the ESC timeout duration.
// Default is DefaultESCTimeout; set to 0 for immediate timeout.
var escTimeoutForTest = DefaultESCTimeout

// NewParser creates a new input parser.
func NewParser() *Parser {
	return &Parser{state: stateNormal, escTimeout: escTimeoutForTest}
}

// Feed processes raw bytes and returns parsed events.
// Thread-safe: locks the parser mutex to avoid races with FeedTimeout.
func (p *Parser) Feed(data []byte) []Event {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.feed(data)
}

// feed is the unlocked inner implementation of Feed.
// Callers must hold p.mu.
func (p *Parser) feed(data []byte) []Event {
	var events []Event

	for _, b := range data {
		switch p.state {
		case stateNormal:
			if b == 0x1b { // ESC
				p.state = stateEscape
				p.buf = p.buf[:0]
				p.escAt = time.Now()
			} else if b >= 0x20 && b != 0x7f {
				// Printable character — may be ASCII or UTF-8 lead byte
				if ev := p.handlePrintable(b); ev != nil {
					events = append(events, *ev)
				}
			} else if b == 0x0d {
				events = append(events, Event{Type: EventKey, Key: &KeyEvent{Key: KeyEnter}})
			} else if b == 0x0a {
				events = append(events, Event{Type: EventKey, Key: &KeyEvent{Key: KeyEnter}})
			} else if b == 0x09 {
				events = append(events, Event{Type: EventKey, Key: &KeyEvent{Key: KeyTab}})
			} else if b == 0x7f {
				events = append(events, Event{Type: EventKey, Key: &KeyEvent{Key: KeyBackspace}})
			} else if b < 0x20 {
				events = append(events, *p.handleCtrl(b))
			}

		case stateUTF8:
			// Accumulating UTF-8 continuation bytes (BUG 1 fix)
			// Check for invalid continuation byte
			if b&0xC0 != 0x80 {
				// Invalid continuation — emit replacement, reprocess b in normal state
				events = append(events, Event{Type: EventKey, Key: &KeyEvent{Rune: 0xFFFD}})
				p.utf8Len = 0
				p.utf8Expect = 0
				p.state = stateNormal
				// Reprocess this byte in normal state (call unlocked feed to avoid deadlock)
				events = append(events, p.feed([]byte{b})...)
				break
			}
			p.utf8Buf[p.utf8Len] = b
			p.utf8Len++
			if p.utf8Len >= p.utf8Expect {
				r, _ := utf8.DecodeRune(p.utf8Buf[:p.utf8Len])
				if r == utf8.RuneError {
					events = append(events, Event{Type: EventKey, Key: &KeyEvent{Rune: 0xFFFD}})
				} else {
					events = append(events, Event{Type: EventKey, Key: &KeyEvent{Rune: r}})
				}
				p.utf8Len = 0
				p.utf8Expect = 0
				p.state = stateNormal
			}

		case stateEscape:
			if b == 0x1b {
				// Double ESC: first was standalone ESC, reprocess second
				events = append(events, Event{Type: EventKey, Key: &KeyEvent{Key: KeyEscape}})
				p.state = stateNormal
				p.buf = p.buf[:0]
				// Reprocess this ESC byte
				p.state = stateEscape
				p.escAt = time.Now()
			} else if b == '[' {
				p.state = stateCSI
				p.buf = p.buf[:0]
			} else if b == ']' {
				// OSC sequence: ESC ] ... (terminated by BEL or ESC \)
				p.state = stateOSC
				p.buf = p.buf[:0]
			} else if b == 'O' {
				p.state = stateSS3
				p.buf = p.buf[:0]
			} else if b >= 0x20 && b < 0x7f {
				// Alt+key: ESC followed by a character
				events = append(events, Event{Type: EventKey, Key: &KeyEvent{
					Rune:      rune(b),
					Modifiers: ModAlt,
				}})
				p.state = stateNormal
			} else if b == 0x0d || b == 0x0a {
				events = append(events, Event{Type: EventKey, Key: &KeyEvent{
					Key:       KeyEnter,
					Modifiers: ModAlt,
				}})
				p.state = stateNormal
			} else {
				// Unknown escape sequence, go back to normal
				p.state = stateNormal
			}

		case stateCSI:
			// Collect parameters until a final byte (0x40-0x7E)
			p.buf = append(p.buf, b)
			if b >= 0x40 && b <= 0x7E {
				oldState := p.state
				ev := p.parseCSI(p.buf)
				if ev != nil {
					events = append(events, *ev)
				}
				if p.state == oldState {
					p.state = stateNormal
				}
			}

		case stateSS3:
			p.buf = append(p.buf, b)
			if b >= 0x40 && b <= 0x7E {
				ev := p.parseSS3(p.buf)
				if ev != nil {
					events = append(events, *ev)
				}
				p.state = stateNormal
			}

		case statePaste:
			if b == 0x1b {
				p.state = statePasteESC
			} else {
				p.paste = append(p.paste, b)
			}

		case statePasteESC:
			if b == '[' {
				// Possibly the end of paste (\e[201~).
				// Switch to CSI to collect the rest.
				// Don't append '[' — consumed by this transition.
				p.inPaste = true // BUG 3 fix: mark we're in paste context
				p.state = stateCSI
				p.buf = p.buf[:0]
			} else {
				// Not end of paste, treat as literal
				p.paste = append(p.paste, 0x1b, b)
				p.state = statePaste
			}

		case stateOSC:
			// Collect OSC payload until BEL (0x07) or ST (ESC \)
			if b == 0x07 {
				// BEL terminator — process complete OSC
				if ev := p.parseOSC(p.buf); ev != nil {
					events = append(events, *ev)
				}
				p.state = stateNormal
				p.buf = p.buf[:0]
			} else if b == 0x1b {
				// Possible ST (ESC \)
				p.state = stateOSCESC
			} else {
				p.buf = append(p.buf, b)
			}

		case stateOSCESC:
			// Inside OSC, got ESC — check for ST (ESC \) or literal ESC
			if b == '\\' {
				// ST terminator — process complete OSC
				if ev := p.parseOSC(p.buf); ev != nil {
					events = append(events, *ev)
				}
				p.state = stateNormal
				p.buf = p.buf[:0]
			} else {
				// Not ST — append ESC and continue collecting
				p.buf = append(p.buf, 0x1b, b)
				p.state = stateOSC
			}
		}
	}

	return events
}

// FeedTimeout handles the ESC timeout (BUG 2 fix).
// Call this periodically (e.g., every 10ms) from the event loop.
// If the parser is in stateEscape and ESCTimeout has elapsed since
// the ESC byte arrived, it emits a KeyEscape event.
func (p *Parser) FeedTimeout() []Event {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.escTimeout == 0 {
		// 0 means immediate
		if p.state == stateEscape {
			p.state = stateNormal
			return []Event{{Type: EventKey, Key: &KeyEvent{Key: KeyEscape}}}
		}
		return nil
	}
	if p.state == stateEscape && time.Since(p.escAt) >= p.escTimeout {
		p.state = stateNormal
		return []Event{{Type: EventKey, Key: &KeyEvent{Key: KeyEscape}}}
	}
	return nil
}

// handlePrintable handles a printable byte — may be ASCII or UTF-8 lead byte.
// BUG 1 fix: properly accumulate multi-byte UTF-8 sequences.
// Returns nil for multi-byte lead bytes (state switches to stateUTF8).
func (p *Parser) handlePrintable(b byte) *Event {
	if b < 0x80 {
		// Plain ASCII
		return &Event{Type: EventKey, Key: &KeyEvent{Rune: rune(b)}}
	}

	// UTF-8 lead byte — determine expected sequence length
	expect := utf8ByteLen(b)
	if expect == 0 {
		// Invalid lead byte, emit replacement
		return &Event{Type: EventKey, Key: &KeyEvent{Rune: 0xFFFD}}
	}

	// Start accumulating; return nil (no event yet)
	p.utf8Buf[0] = b
	p.utf8Len = 1
	p.utf8Expect = expect
	p.state = stateUTF8
	return nil
}

// utf8ByteLen returns the expected total bytes for a UTF-8 lead byte.
func utf8ByteLen(b byte) int {
	switch {
	case b < 0x80:
		return 1
	case b < 0xC0:
		return 0 // continuation byte, not a valid lead
	case b < 0xE0:
		return 2
	case b < 0xF0:
		return 3
	case b < 0xF8:
		return 4
	default:
		return 0
	}
}

// handleCtrl converts a control byte to a Ctrl+key event.
// BUG 5 fix: correct mapping for 0x1C-0x1F.
func (p *Parser) handleCtrl(b byte) *Event {
	switch b {
	case 0x1c:
		return &Event{Type: EventKey, Key: &KeyEvent{Rune: '\\', Modifiers: ModCtrl}}
	case 0x1d:
		return &Event{Type: EventKey, Key: &KeyEvent{Rune: ']', Modifiers: ModCtrl}}
	case 0x1e:
		return &Event{Type: EventKey, Key: &KeyEvent{Rune: '^', Modifiers: ModCtrl}}
	case 0x1f:
		return &Event{Type: EventKey, Key: &KeyEvent{Rune: '_', Modifiers: ModCtrl}}
	default:
		// 0x01-0x1A → 'a'-'z'
		key := rune(b) + 'a' - 1
		return &Event{Type: EventKey, Key: &KeyEvent{
			Rune:      key,
			Modifiers: ModCtrl,
		}}
	}
}

// parseOSC parses a completed OSC sequence payload (after ESC ] payload).
// Currently handles OSC52 clipboard responses.
func (p *Parser) parseOSC(buf []byte) *Event {
	payload := string(buf)
	// OSC52 response: "52;c;<base64>" or "52;p;<base64>"
	if strings.HasPrefix(payload, "52;") {
		text, ok := ParseOSC52Response("\x1b]" + payload)
		if ok {
			return &Event{Type: EventClipboard, Clipboard: text}
		}
	}
	return nil
}

// parseCSI parses a completed CSI sequence (after ESC [).
func (p *Parser) parseCSI(buf []byte) *Event {
	if len(buf) == 0 {
		return nil
	}

	final := buf[len(buf)-1]
	params := string(buf[:len(buf)-1])

	// Check for bracketed paste start: ESC [ 200 ~
	// BUG 6 fix: don't process paste start if already in paste mode
	if params == "200" && final == '~' {
		if p.inPaste {
			// Stray paste-start inside paste content — treat as literal
			p.appendPasteCSI(buf)
			p.state = statePaste
			return nil
		}
		p.state = statePaste
		p.paste = p.paste[:0]
		return nil
	}

	// Check for paste end: ESC [ 201 ~
	// BUG 3/6 fix: only emit paste event if we're actually in paste mode.
	// A stray 201~ outside paste mode (e.g., double paste-end) is ignored.
	if params == "201" && final == '~' {
		if !p.inPaste {
			// Stray paste-end marker — not in paste mode, ignore
			return nil
		}
		ev := &Event{Type: EventPaste, Paste: string(p.paste)}
		p.paste = p.paste[:0]
		p.inPaste = false
		return ev
	}

	// BUG 3 fix: if we're inside paste mode, any non-201~ CSI is literal paste content
	if p.inPaste {
		p.appendPasteCSI(buf)
		p.state = statePaste
		return nil
	}

	// Parse SGR mouse: ESC [ < button ; x ; y M/m
	if len(buf) > 0 && buf[0] == '<' {
		return p.parseSGRMouse(buf)
	}

	// Parse parameters
	nums := parseCSIParams(params)

	switch final {
	case 'A': // Up
		return keyEventCSI(nums, KeyUp, false)
	case 'B': // Down
		return keyEventCSI(nums, KeyDown, false)
	case 'C': // Right
		return keyEventCSI(nums, KeyRight, false)
	case 'D': // Left
		return keyEventCSI(nums, KeyLeft, false)
	case 'H': // Home
		return keyEventCSI(nums, KeyHome, false)
	case 'F': // End
		return keyEventCSI(nums, KeyEnd, false)
	case 'Z': // Shift+Tab
		return &Event{Type: EventKey, Key: &KeyEvent{Key: KeyBacktab, Modifiers: ModShift}}
	case '~':
		// BUG 4 fix: tilde sequences use different modifier extraction.
		// For ~ sequences, the modifier is in the second parameter
		// regardless of the first parameter value.
		switch firstNum(nums) {
		case 1:
			return keyEventCSI(nums, KeyHome, true)
		case 2:
			return keyEventCSI(nums, KeyInsert, true)
		case 3:
			return keyEventCSI(nums, KeyDelete, true)
		case 4:
			return keyEventCSI(nums, KeyEnd, true)
		case 5:
			return keyEventCSI(nums, KeyPageUp, true)
		case 6:
			return keyEventCSI(nums, KeyPageDown, true)
		case 15:
			return keyEventCSI(nums, KeyF5, true)
		case 17:
			return keyEventCSI(nums, KeyF6, true)
		case 18:
			return keyEventCSI(nums, KeyF7, true)
		case 19:
			return keyEventCSI(nums, KeyF8, true)
		case 20:
			return keyEventCSI(nums, KeyF9, true)
		case 21:
			return keyEventCSI(nums, KeyF10, true)
		case 23:
			return keyEventCSI(nums, KeyF11, true)
		case 24:
			return keyEventCSI(nums, KeyF12, true)
		}
	}

	return nil
}

// appendPasteCSI appends a raw CSI sequence to the paste buffer (BUG 3 fix).
func (p *Parser) appendPasteCSI(buf []byte) {
	p.paste = append(p.paste, 0x1b, '[')
	p.paste = append(p.paste, buf...)
}

// parseSS3 parses SS3 sequences (ESC O X) for F1-F4.
func (p *Parser) parseSS3(buf []byte) *Event {
	if len(buf) == 0 {
		return nil
	}
	switch buf[0] {
	case 'P':
		return &Event{Type: EventKey, Key: &KeyEvent{Key: KeyF1}}
	case 'Q':
		return &Event{Type: EventKey, Key: &KeyEvent{Key: KeyF2}}
	case 'R':
		return &Event{Type: EventKey, Key: &KeyEvent{Key: KeyF3}}
	case 'S':
		return &Event{Type: EventKey, Key: &KeyEvent{Key: KeyF4}}
	}
	return nil
}

// parseSGRMouse parses an SGR mouse sequence: <button;x;y{M|m}
func (p *Parser) parseSGRMouse(buf []byte) *Event {
	s := string(buf[1:]) // strip leading '<'

	var actionByte byte = 'M'
	if len(s) > 0 && (s[len(s)-1] == 'M' || s[len(s)-1] == 'm') {
		actionByte = s[len(s)-1]
		s = s[:len(s)-1]
	}

	parts := splitSemicolons(s)
	if len(parts) < 3 {
		return nil
	}

	btn := atoi(parts[0])
	x := atoi(parts[1]) - 1
	y := atoi(parts[2]) - 1

	var button MouseButton
	var modifiers ModMask
	var action MouseAction

	decodeMouseButton(btn, &button, &modifiers, &action)

	if actionByte == 'm' {
		action = MouseUp
	}

	return &Event{
		Type: EventMouse,
		Mouse: &MouseEvent{
			X:         x,
			Y:         y,
			Button:    button,
			Modifiers: modifiers,
			Action:    action,
		},
	}
}

func decodeMouseButton(code int, button *MouseButton, mods *ModMask, action *MouseAction) {
	*mods = 0
	if code&4 != 0 {
		*mods |= ModShift
	}
	if code&8 != 0 {
		*mods |= ModAlt
	}
	if code&16 != 0 {
		*mods |= ModCtrl
	}

	btn := code & 3
	isMotion := code&32 != 0

	if code&64 != 0 {
		if btn == 0 {
			*button = MouseWheelUp
		} else {
			*button = MouseWheelDown
		}
		*action = MouseWheel
		return
	}

	switch btn {
	case 0:
		*button = MouseLeft
	case 1:
		*button = MouseMiddle
	case 2:
		*button = MouseRight
	default:
		*button = MouseNone
	}

	if isMotion {
		if *button == MouseNone {
			*action = MouseMove
		} else {
			*action = MouseDrag
		}
	} else {
		*action = MouseDown
	}
}

// keyEventCSI creates a key event from CSI parameters, extracting modifiers.
// BUG 4 fix: for tilde (~) sequences, modifiers are in the second parameter
// regardless of the first value. For non-tilde (A/B/C/D/H/F), the first
// parameter must be 1 (or absent) for modifiers to be in the second param.
func keyEventCSI(nums []int, baseKey KeyCode, isTilde bool) *Event {
	mods := ModMask(0)
	if isTilde {
		// Tilde sequences: modifier is always in second param if present
		if len(nums) >= 2 {
			mods = decodeCSIModifier(nums[1])
		}
	} else {
		// Arrow/Home/End: modifier in second param, first must be 1
		if len(nums) >= 2 && nums[0] == 1 {
			mods = decodeCSIModifier(nums[1])
		}
	}
	return &Event{Type: EventKey, Key: &KeyEvent{Key: baseKey, Modifiers: mods}}
}

// decodeCSIModifier converts xterm modifier code to ModMask.
// Encoding: 2=Shift, 3=Alt, 4=Shift+Alt, 5=Ctrl, 6=Shift+Ctrl, 7=Alt+Ctrl, 8=Shift+Alt+Ctrl
// Subtract 1 to get bitmask: bit0=Shift, bit1=Alt, bit2=Ctrl.
func decodeCSIModifier(code int) ModMask {
	mods := ModMask(0)
	bits := code - 1
	if bits&1 != 0 {
		mods |= ModShift
	}
	if bits&2 != 0 {
		mods |= ModAlt
	}
	if bits&4 != 0 {
		mods |= ModCtrl
	}
	return mods
}

// --- helpers ---

func parseCSIParams(s string) []int {
	if s == "" {
		return nil
	}
	parts := splitSemicolons(s)
	result := make([]int, len(parts))
	for i, p := range parts {
		result[i] = atoi(p)
	}
	return result
}

func splitSemicolons(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ';' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n * 10 + int(c-'0')
	}
	return n
}

func firstNum(nums []int) int {
	if len(nums) > 0 {
		return nums[0]
	}
	return 0
}

// String returns a human-readable key name.
func (k KeyCode) String() string {
	names := map[KeyCode]string{
		KeyEnter: "Enter", KeyTab: "Tab", KeyBacktab: "BackTab",
		KeyBackspace: "Backspace", KeyDelete: "Delete", KeyInsert: "Insert",
		KeyHome: "Home", KeyEnd: "End", KeyPageUp: "PageUp", KeyPageDown: "PageDown",
		KeyUp: "Up", KeyDown: "Down", KeyLeft: "Left", KeyRight: "Right",
		KeyEscape: "Escape", KeySpace: "Space",
		KeyF1: "F1", KeyF2: "F2", KeyF3: "F3", KeyF4: "F4",
		KeyF5: "F5", KeyF6: "F6", KeyF7: "F7", KeyF8: "F8",
		KeyF9: "F9", KeyF10: "F10", KeyF11: "F11", KeyF12: "F12",
	}
	if name, ok := names[k]; ok {
		return name
	}
	return fmt.Sprintf("Key(%d)", k)
}
