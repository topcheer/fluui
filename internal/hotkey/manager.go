// Package hotkey provides a configurable hotkey (keyboard shortcut) management system.
//
// It supports:
//   - Key combination registration with conflict detection
//   - Scoped groups (Global, Local, Modal)
//   - Multi-key sequences (e.g., "g g" for goto top)
//   - Serialization for persistence
//   - Export for help/overlay display
package hotkey

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/topcheer/fluui/internal/term"
)

// === Errors ===

// ErrConflict is returned when a new binding conflicts with an existing one.
var ErrConflict = errors.New("hotkey: binding conflict")

// ErrNotFound is returned when a binding is not found.
var ErrNotFound = errors.New("hotkey: binding not found")

// ErrInvalidSequence is returned when a key sequence string cannot be parsed.
var ErrInvalidSequence = errors.New("hotkey: invalid key sequence")

// === Scope ===

// Scope determines when a hotkey binding is active.
type Scope int

const (
	// ScopeGlobal means the binding is always active.
	ScopeGlobal Scope = iota
	// ScopeLocal means the binding is active only for a specific component/context.
	ScopeLocal
	// ScopeModal means the binding is active only when a modal/overlay is open.
	ScopeModal
)

// String returns the human-readable name of the scope.
func (s Scope) String() string {
	switch s {
	case ScopeGlobal:
		return "Global"
	case ScopeLocal:
		return "Local"
	case ScopeModal:
		return "Modal"
	default:
		return fmt.Sprintf("Scope(%d)", int(s))
	}
}

// ParseScope converts a string to a Scope.
func ParseScope(s string) (Scope, error) {
	switch strings.ToLower(s) {
	case "global", "":
		return ScopeGlobal, nil
	case "local":
		return ScopeLocal, nil
	case "modal":
		return ScopeModal, nil
	default:
		return ScopeGlobal, fmt.Errorf("%w: unknown scope %q", ErrInvalidSequence, s)
	}
}

// === KeyCombo ===

// KeyCombo represents a single key combination (key + modifiers).
type KeyCombo struct {
	Key       term.KeyCode // non-zero for special keys (Enter, F1, etc.)
	Rune      rune         // non-zero for printable characters
	Modifiers term.ModMask // Ctrl/Alt/Shift bitmask
}

// Match checks if a KeyEvent matches this combo.
// With modifiers: case-insensitive (Ctrl+F matches Ctrl+f).
// Without modifiers: exact case ('G' != 'g', like vim).
func (c KeyCombo) Match(k *term.KeyEvent) bool {
	if k == nil {
		return false
	}
	if c.Key != term.KeyUnknown && k.Key == c.Key {
		return k.Modifiers == c.Modifiers
	}
	if c.Rune != 0 && k.Rune != 0 {
		if c.Modifiers != 0 || k.Modifiers != 0 {
			// With modifiers: case-insensitive
			if unicode.ToLower(k.Rune) == unicode.ToLower(c.Rune) {
				return k.Modifiers == c.Modifiers
			}
		} else {
			// No modifiers: exact case match
			if k.Rune == c.Rune {
				return true
			}
		}
	}
	// Also match by key code for space/escape which have both Key and Rune
	if c.Key != term.KeyUnknown && k.Rune != 0 {
		if c.Key == term.KeySpace && k.Rune == ' ' && k.Modifiers == c.Modifiers {
			return true
		}
	}
	return false
}

// Equal checks if two combos are equivalent.
// With modifiers: case-insensitive. Without: exact case.
func (c KeyCombo) Equal(other KeyCombo) bool {
	if c.Key != other.Key || c.Modifiers != other.Modifiers {
		return false
	}
	if c.Modifiers != 0 {
		return unicode.ToLower(c.Rune) == unicode.ToLower(other.Rune)
	}
	return c.Rune == other.Rune
}

// String returns a human-readable representation like "Ctrl+F" or "g".
func (c KeyCombo) String() string {
	var parts []string

	if c.Modifiers&term.ModCtrl != 0 {
		parts = append(parts, "Ctrl")
	}
	if c.Modifiers&term.ModAlt != 0 {
		parts = append(parts, "Alt")
	}
	if c.Modifiers&term.ModShift != 0 {
		parts = append(parts, "Shift")
	}

	// Key name
	if c.Key != term.KeyUnknown {
		parts = append(parts, c.Key.String())
	} else if c.Rune != 0 {
		if c.Rune >= 32 && c.Rune < 127 {
			// Printable ASCII
			parts = append(parts, string(c.Rune))
		} else {
			parts = append(parts, fmt.Sprintf("%q", c.Rune))
		}
	}

	return strings.Join(parts, "+")
}

// === KeySequence ===

// KeySequence represents one or more key combos pressed in sequence.
// For example, "g g" (press g then g) or just "Ctrl+F".
type KeySequence struct {
	Combos []KeyCombo
}

// Len returns the number of combos in the sequence.
func (ks KeySequence) Len() int {
	return len(ks.Combos)
}

// IsSingle returns true if this is a single-key shortcut.
func (ks KeySequence) IsSingle() bool {
	return len(ks.Combos) == 1
}

// IsMulti returns true if the sequence requires multiple key presses.
func (ks KeySequence) IsMulti() bool {
	return len(ks.Combos) > 1
}

// Equal checks if two sequences are identical.
func (ks KeySequence) Equal(other KeySequence) bool {
	if len(ks.Combos) != len(other.Combos) {
		return false
	}
	for i := range ks.Combos {
		if !ks.Combos[i].Equal(other.Combos[i]) {
			return false
		}
	}
	return true
}

// String returns a human-readable representation.
// Single combos: "Ctrl+F", multi combos: "g g".
func (ks KeySequence) String() string {
	parts := make([]string, len(ks.Combos))
	for i, c := range ks.Combos {
		parts[i] = c.String()
	}
	return strings.Join(parts, " ")
}

// HasPrefix checks if this sequence starts with the given prefix.
func (ks KeySequence) HasPrefix(prefix KeySequence) bool {
	if len(prefix.Combos) > len(ks.Combos) {
		return false
	}
	for i := range prefix.Combos {
		if !ks.Combos[i].Equal(prefix.Combos[i]) {
			return false
		}
	}
	return true
}

// === Binding ===

// Binding represents a registered keyboard shortcut.
type Binding struct {
	Action      string      // e.g., "search.find", "goto.top"
	Description string      // human-readable description for help display
	Group       string      // category for help grouping (e.g., "Navigation")
	Scope       Scope       // when this binding is active
	Sequence    KeySequence // the key sequence to trigger this action
	Enabled     bool        // whether this binding is currently active
}

// === MatchResult ===

// MatchResult represents the result of matching a key event.
type MatchResult int

const (
	// MatchNone means no binding matched.
	MatchNone MatchResult = iota
	// MatchComplete means a full sequence matched — action should be triggered.
	MatchComplete
	// MatchPartial means a prefix of a multi-key sequence matched — wait for more keys.
	MatchPartial
)

// === Manager ===

// Manager manages registered keyboard shortcuts with conflict detection
// and multi-key sequence support.
type Manager struct {
	mu              sync.RWMutex
	bindings        map[string]*Binding // action -> binding
	seqTimeout      time.Duration       // timeout for partial sequences
	pending         []KeyCombo          // partially matched combos
	pendingExpiry   time.Time           // when the pending sequence expires
	defaultGroup    string              // default group for new bindings
	allowOverride   bool                // if true, Register replaces conflicting bindings
}

// NewManager creates a new HotkeyManager with sensible defaults.
func NewManager() *Manager {
	return &Manager{
		bindings:     make(map[string]*Binding),
		seqTimeout:   1500 * time.Millisecond, // 1.5s window for multi-key sequences
		defaultGroup: "General",
	}
}

// SetSequenceTimeout sets the timeout for partial multi-key sequences.
// Default is 1500ms.
func (m *Manager) SetSequenceTimeout(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seqTimeout = d
}

// SetDefaultGroup sets the default group for new bindings.
func (m *Manager) SetDefaultGroup(group string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultGroup = group
}

// SetAllowOverride controls whether Register replaces conflicting bindings.
// When true, Register will silently replace existing bindings that conflict.
func (m *Manager) SetAllowOverride(allow bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.allowOverride = allow
}

// === Registration ===

// Register adds a new keyboard shortcut binding.
// Returns ErrConflict if the key sequence conflicts with an existing binding
// (unless allowOverride is set).
// Returns an error if the action name is already registered.
func (m *Manager) Register(action string, seq KeySequence, opts ...Option) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicate action name
	if _, exists := m.bindings[action]; exists {
		return fmt.Errorf("%w: action %q already registered", ErrConflict, action)
	}

	// Check for key sequence conflicts
	if existing := m.findConflictLocked(seq); existing != nil {
		if m.allowOverride {
			delete(m.bindings, existing.Action)
		} else {
			return fmt.Errorf("%w: %q conflicts with existing binding %q (%s)",
				ErrConflict, seq.String(), existing.Action, existing.Sequence.String())
		}
	}

	b := &Binding{
		Action:      action,
		Description: action,
		Group:       m.defaultGroup,
		Scope:       ScopeGlobal,
		Sequence:    seq,
		Enabled:     true,
	}

	for _, opt := range opts {
		opt(b)
	}

	m.bindings[action] = b
	return nil
}

// Unregister removes a binding by action name.
// Returns ErrNotFound if the action is not registered.
func (m *Manager) Unregister(action string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.bindings[action]; !ok {
		return fmt.Errorf("%w: action %q", ErrNotFound, action)
	}
	delete(m.bindings, action)
	return nil
}

// Enable enables a binding.
func (m *Manager) Enable(action string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, ok := m.bindings[action]
	if !ok {
		return fmt.Errorf("%w: action %q", ErrNotFound, action)
	}
	b.Enabled = true
	return nil
}

// Disable disables a binding without removing it.
func (m *Manager) Disable(action string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, ok := m.bindings[action]
	if !ok {
		return fmt.Errorf("%w: action %q", ErrNotFound, action)
	}
	b.Enabled = false
	return nil
}

// === Options ===

// Option configures a Binding during registration.
type Option func(*Binding)

// WithDescription sets the human-readable description.
func WithDescription(desc string) Option {
	return func(b *Binding) { b.Description = desc }
}

// WithGroup sets the category group.
func WithGroup(group string) Option {
	return func(b *Binding) { b.Group = group }
}

// WithScope sets the binding scope.
func WithScope(scope Scope) Option {
	return func(b *Binding) { b.Scope = scope }
}

// WithDisabled registers the binding in a disabled state.
func WithDisabled() Option {
	return func(b *Binding) { b.Enabled = false }
}

// === Matching ===

// Match checks if a key event matches any binding.
// For single-key shortcuts, returns (action, MatchComplete) or ("", MatchNone).
// For multi-key sequences, returns ("", MatchPartial) if a prefix matches.
func (m *Manager) Match(k *term.KeyEvent) (string, MatchResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	combo := keyEventToCombo(k)

	// Check for partial sequence timeout
	if len(m.pending) > 0 && time.Now().After(m.pendingExpiry) {
		m.pending = nil
	}

	// Build the candidate sequence
	candidate := append(m.pending, combo)

	// Check all enabled bindings
	var completeAction string
	hasPartial := false

	for _, b := range m.bindings {
		if !b.Enabled {
			continue
		}
		candidateSeq := KeySequence{Combos: candidate}
		if b.Sequence.Equal(candidateSeq) {
			completeAction = b.Action
			break
		}
		if b.Sequence.HasPrefix(candidateSeq) && candidateSeq.Len() < b.Sequence.Len() {
			hasPartial = true
		}
	}

	if completeAction != "" {
		m.pending = nil
		return completeAction, MatchComplete
	}

	if hasPartial {
		m.pending = candidate
		m.pendingExpiry = time.Now().Add(m.seqTimeout)
		return "", MatchPartial
	}

	// No match — reset pending
	m.pending = nil
	return "", MatchNone
}

// ResetPending clears any partially matched multi-key sequence.
func (m *Manager) ResetPending() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pending = nil
}

// HasPending returns true if a multi-key sequence is in progress.
func (m *Manager) HasPending() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.pending) > 0
}

// PendingKeys returns the currently accumulated key combos for a pending sequence.
func (m *Manager) PendingKeys() []KeyCombo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.pending) == 0 {
		return nil
	}
	result := make([]KeyCombo, len(m.pending))
	copy(result, m.pending)
	return result
}

// === Querying ===

// Bindings returns a copy of all registered bindings.
func (m *Manager) Bindings() []Binding {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Binding, 0, len(m.bindings))
	for _, b := range m.bindings {
		result = append(result, *b)
	}
	return result
}

// BindingsByGroup returns all bindings in the given group.
func (m *Manager) BindingsByGroup(group string) []Binding {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Binding
	for _, b := range m.bindings {
		if b.Group == group {
			result = append(result, *b)
		}
	}
	return result
}

// BindingsByScope returns all bindings with the given scope.
func (m *Manager) BindingsByScope(scope Scope) []Binding {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Binding
	for _, b := range m.bindings {
		if b.Scope == scope {
			result = append(result, *b)
		}
	}
	return result
}

// Get returns the binding for a given action.
func (m *Manager) Get(action string) (Binding, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	b, ok := m.bindings[action]
	if !ok {
		return Binding{}, false
	}
	return *b, true
}

// Groups returns a list of all group names with their bindings.
func (m *Manager) Groups() []Group {
	m.mu.RLock()
	defer m.mu.RUnlock()

	groupMap := make(map[string][]Binding)
	var groupOrder []string

	for _, b := range m.bindings {
		if _, exists := groupMap[b.Group]; !exists {
			groupOrder = append(groupOrder, b.Group)
		}
		groupMap[b.Group] = append(groupMap[b.Group], *b)
	}

	result := make([]Group, 0, len(groupOrder))
	for _, name := range groupOrder {
		result = append(result, Group{
			Name:     name,
			Bindings: groupMap[name],
		})
	}
	return result
}

// Count returns the number of registered bindings.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.bindings)
}

// HasConflict checks if a key sequence would conflict with an existing binding.
// Returns the conflicting binding, or nil if no conflict.
func (m *Manager) HasConflict(seq KeySequence) *Binding {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.findConflictLocked(seq)
}

// === Group type for export ===

// Group represents a named category of bindings for help display.
type Group struct {
	Name     string
	Bindings []Binding
}

// === Conflict Detection ===

func (m *Manager) findConflictLocked(seq KeySequence) *Binding {
	for _, b := range m.bindings {
		if b.Sequence.Equal(seq) {
			return b
		}
		// Check prefix overlap: if one sequence is a prefix of another,
		// the shorter one will always trigger first, making the longer unreachable.
		// We only flag exact duplicates as conflicts to allow flexible sequences.
	}
	return nil
}

// === Helpers ===

// keyEventToCombo converts a term.KeyEvent to a KeyCombo.
func keyEventToCombo(k *term.KeyEvent) KeyCombo {
	if k == nil {
		return KeyCombo{}
	}
	return KeyCombo{
		Key:       k.Key,
		Rune:      k.Rune,
		Modifiers: k.Modifiers,
	}
}

// === Parsing ===

var keyNameMap = map[string]term.KeyCode{
	"enter": term.KeyEnter, "return": term.KeyEnter,
	"tab": term.KeyTab, "backtab": term.KeyBacktab,
	"backspace": term.KeyBackspace, "bs": term.KeyBackspace,
	"delete": term.KeyDelete, "del": term.KeyDelete,
	"insert": term.KeyInsert, "ins": term.KeyInsert,
	"home": term.KeyHome, "end": term.KeyEnd,
	"pageup": term.KeyPageUp, "pgup": term.KeyPageUp,
	"pagedown": term.KeyPageDown, "pgdn": term.KeyPageDown, "pgdown": term.KeyPageDown,
	"up": term.KeyUp, "down": term.KeyDown,
	"left": term.KeyLeft, "right": term.KeyRight,
	"escape": term.KeyEscape, "esc": term.KeyEscape,
	"space": term.KeySpace,
	"f1": term.KeyF1, "f2": term.KeyF2, "f3": term.KeyF3, "f4": term.KeyF4,
	"f5": term.KeyF5, "f6": term.KeyF6, "f7": term.KeyF7, "f8": term.KeyF8,
	"f9": term.KeyF9, "f10": term.KeyF10, "f11": term.KeyF11, "f12": term.KeyF12,
}

// ParseCombo parses a single key combo string like "Ctrl+F" or "g".
func ParseCombo(s string) (KeyCombo, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return KeyCombo{}, fmt.Errorf("%w: empty combo", ErrInvalidSequence)
	}

	parts := strings.Split(s, "+")
	combo := KeyCombo{}

	for i, part := range parts {
		keyPart := strings.TrimSpace(part)
		lowerPart := strings.ToLower(keyPart) // only for modifier/special-key name matching
		switch lowerPart {
		case "ctrl", "control":
			combo.Modifiers |= term.ModCtrl
		case "alt", "option", "opt", "meta":
			combo.Modifiers |= term.ModAlt
		case "shift":
			combo.Modifiers |= term.ModShift
		default:
			if i != len(parts)-1 {
				return KeyCombo{}, fmt.Errorf("%w: %q is not a modifier", ErrInvalidSequence, keyPart)
			}
			// Last part must be the key
			if code, ok := keyNameMap[lowerPart]; ok {
				combo.Key = code
			} else if len(keyPart) == 1 {
				// Preserve original case for single-char keys.
				// Modifier-based normalization happens after the loop.
				combo.Rune = rune(keyPart[0])
			} else {
				return KeyCombo{}, fmt.Errorf("%w: unknown key %q", ErrInvalidSequence, keyPart)
			}
		}
	}

	if combo.Key == term.KeyUnknown && combo.Rune == 0 {
		return KeyCombo{}, fmt.Errorf("%w: no key specified in %q", ErrInvalidSequence, s)
	}

	// For modifier+key combos, normalize rune to lowercase.
	// Terminals handle Ctrl+letter case inconsistently (some send 'F', some 'f'),
	// so we normalize to lowercase for consistent matching.
	// For standalone keys, preserve case to distinguish 'G' from 'g'.
	if combo.Rune != 0 && combo.Modifiers != 0 {
		combo.Rune = unicode.ToLower(combo.Rune)
	}

	return combo, nil
}

// ParseSequence parses a key sequence string.
// Examples: "Ctrl+F", "g g", "Ctrl+X Ctrl+C".
func ParseSequence(s string) (KeySequence, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return KeySequence{}, fmt.Errorf("%w: empty sequence", ErrInvalidSequence)
	}

	// Split by spaces, but be careful: "Ctrl+F" has no spaces.
	// Each token separated by space is a combo.
	tokens := strings.Fields(s)
	combos := make([]KeyCombo, 0, len(tokens))

	for _, token := range tokens {
		combo, err := ParseCombo(token)
		if err != nil {
			return KeySequence{}, err
		}
		combos = append(combos, combo)
	}

	return KeySequence{Combos: combos}, nil
}

// MustParseSequence is like ParseSequence but panics on error.
func MustParseSequence(s string) KeySequence {
	seq, err := ParseSequence(s)
	if err != nil {
		panic(err)
	}
	return seq
}

// === Serialization ===

// Config represents a serializable hotkey configuration.
type Config struct {
	Bindings []ConfigBinding `json:"bindings"`
}

// ConfigBinding is a single binding in the serializable config.
type ConfigBinding struct {
	Action      string `json:"action"`
	Sequence    string `json:"sequence"`
	Description string `json:"description"`
	Group       string `json:"group"`
	Scope       string `json:"scope"`
	Enabled     bool   `json:"enabled"`
}

// ExportConfig returns a serializable representation of all bindings.
func (m *Manager) ExportConfig() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cfg := Config{
		Bindings: make([]ConfigBinding, 0, len(m.bindings)),
	}
	for _, b := range m.bindings {
		cfg.Bindings = append(cfg.Bindings, ConfigBinding{
			Action:      b.Action,
			Sequence:    b.Sequence.String(),
			Description: b.Description,
			Group:       b.Group,
			Scope:       b.Scope.String(),
			Enabled:     b.Enabled,
		})
	}
	return cfg
}

// ImportConfig imports bindings from a config, replacing all existing bindings.
func (m *Manager) ImportConfig(cfg Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.bindings = make(map[string]*Binding)

	for _, cb := range cfg.Bindings {
		seq, err := ParseSequence(cb.Sequence)
		if err != nil {
			return fmt.Errorf("import binding %q: %w", cb.Action, err)
		}
		scope, err := ParseScope(cb.Scope)
		if err != nil {
			return fmt.Errorf("import binding %q: %w", cb.Action, err)
		}
		m.bindings[cb.Action] = &Binding{
			Action:      cb.Action,
			Sequence:    seq,
			Description: cb.Description,
			Group:       cb.Group,
			Scope:       scope,
			Enabled:     cb.Enabled,
		}
	}
	return nil
}
