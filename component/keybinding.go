package component

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/term"
)

// KeyBinding represents a single keyboard shortcut registration.
type KeyBinding struct {
	Command   string         // unique command identifier, e.g. "save", "quit"
	Keys      string         // key description, e.g. "ctrl+s", "ctrl+x ctrl+s" (chord)
	Help      string         // human-readable help text
	Context   string         // scope: "global", "editor", "modal", etc.
	Handler   func() bool    // returns true if handled
	Enabled   bool           // can be toggled at runtime
}

// KeybindingManager provides declarative keyboard shortcut management
// with context scoping, chord sequences, conflict detection, and
// automatic help text generation.
//
// Features:
//   - Register bindings declaratively with human-readable key descriptions
//   - Context scopes (push/pop active context for modal/mode switching)
//   - Chord support (e.g., "ctrl+x ctrl+s" = Ctrl+X then Ctrl+S)
//   - Conflict detection on registration
//   - Automatic help text generation
//   - Runtime enable/disable of individual bindings
type KeybindingManager struct {
	mu          sync.RWMutex
	bindings    []*KeyBinding
	activeCtx   string   // currently active context stack top
	contextStack []string // stack for nested contexts
	chordPrefix string   // active chord prefix (e.g., "ctrl+x" waiting for next key)
	chordTimer  *chordTimer
}

type chordTimer struct {
	cancel chan struct{}
}

// NewKeybindingManager creates a new manager with "global" as the default context.
func NewKeybindingManager() *KeybindingManager {
	return &KeybindingManager{
		activeCtx: "global",
	}
}

// ─── Registration ───

// Register adds a key binding. Returns an error if the binding conflicts
// with an existing binding in the same context.
func (km *KeybindingManager) Register(command, keys, help string, handler func() bool) error {
	return km.RegisterIn("global", command, keys, help, handler)
}

// RegisterIn adds a key binding scoped to a specific context.
func (km *KeybindingManager) RegisterIn(context, command, keys, help string, handler func() bool) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	normalizedKeys := normalizeKeys(keys)

	// Conflict check: same keys in same or active context
	for _, b := range km.bindings {
		if b.Context == context && b.Keys == normalizedKeys && b.Enabled {
			return fmt.Errorf("keybinding conflict: %q (%s) already bound to %q in context %q",
				keys, normalizedKeys, b.Command, context)
		}
	}

	km.bindings = append(km.bindings, &KeyBinding{
		Command: command,
		Keys:    normalizedKeys,
		Help:    help,
		Context: context,
		Handler: handler,
		Enabled: true,
	})
	return nil
}

// Unregister removes a binding by command name.
func (km *KeybindingManager) Unregister(command string) bool {
	km.mu.Lock()
	defer km.mu.Unlock()
	for i, b := range km.bindings {
		if b.Command == command {
			km.bindings = append(km.bindings[:i], km.bindings[i+1:]...)
			return true
		}
	}
	return false
}

// Enable enables a binding by command name.
func (km *KeybindingManager) Enable(command string) {
	km.mu.Lock()
	defer km.mu.Unlock()
	for _, b := range km.bindings {
		if b.Command == command {
			b.Enabled = true
		}
	}
}

// Disable disables a binding by command name.
func (km *KeybindingManager) Disable(command string) {
	km.mu.Lock()
	defer km.mu.Unlock()
	for _, b := range km.bindings {
		if b.Command == command {
			b.Enabled = false
		}
	}
}

// ─── Context Management ───

// PushContext activates a new context scope. Bindings in this context
// take priority over global bindings.
func (km *KeybindingManager) PushContext(ctx string) {
	km.mu.Lock()
	defer km.mu.Unlock()
	km.contextStack = append(km.contextStack, ctx)
	km.activeCtx = ctx
}

// PopContext deactivates the current context scope.
func (km *KeybindingManager) PopContext() string {
	km.mu.Lock()
	defer km.mu.Unlock()
	if len(km.contextStack) == 0 {
		return ""
	}
	popped := km.contextStack[len(km.contextStack)-1]
	km.contextStack = km.contextStack[:len(km.contextStack)-1]
	if len(km.contextStack) > 0 {
		km.activeCtx = km.contextStack[len(km.contextStack)-1]
	} else {
		km.activeCtx = "global"
	}
	return popped
}

// ActiveContext returns the current context name.
func (km *KeybindingManager) ActiveContext() string {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.activeCtx
}

// ─── Matching ───

// Match checks if a key event matches any registered binding.
// Returns the matched command name and whether it was handled.
// Handles chords: if the key is the first part of a chord, it sets
// the chord prefix and returns ("", false) — the caller should wait
// for the next key.
func (km *KeybindingManager) Match(k *term.KeyEvent) (command string, handled bool) {
	if k == nil {
		return "", false
	}

	km.mu.Lock()
	defer km.mu.Unlock()

	keyDesc := keyEventToDesc(k)

	// If we're in a chord state, check for the second key
	if km.chordPrefix != "" {
		fullKey := km.chordPrefix + " " + keyDesc
		for _, b := range km.activeBindingsLocked() {
			if b.Keys == fullKey {
				km.chordPrefix = ""
				if b.Handler != nil {
					return b.Command, b.Handler()
				}
				return b.Command, true
			}
		}
		// Chord not completed — check if this key starts a new chord
		km.chordPrefix = ""
	}

	// Check for exact match first (non-chord)
	for _, b := range km.activeBindingsLocked() {
		if b.Keys == keyDesc {
			if b.Handler != nil {
				return b.Command, b.Handler()
			}
			return b.Command, true
		}
	}

	// Check if this key is the prefix of a chord
	for _, b := range km.activeBindingsLocked() {
		parts := strings.SplitN(b.Keys, " ", 2)
		if len(parts) == 2 && parts[0] == keyDesc {
			// Start chord — wait for next key
			km.chordPrefix = keyDesc
			return "", false // not handled yet, but consumed
		}
	}

	return "", false
}

// HandleKey is a convenience wrapper around Match that returns just the handled bool.
func (km *KeybindingManager) HandleKey(k *term.KeyEvent) bool {
	_, handled := km.Match(k)
	return handled
}

// IsChordActive returns true if the manager is waiting for the second
// key of a chord sequence.
func (km *KeybindingManager) IsChordActive() bool {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.chordPrefix != ""
}

// ChordPrefix returns the current chord prefix, or empty if none active.
func (km *KeybindingManager) ChordPrefix() string {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.chordPrefix
}

// CancelChord cancels any pending chord sequence.
func (km *KeybindingManager) CancelChord() {
	km.mu.Lock()
	defer km.mu.Unlock()
	km.chordPrefix = ""
}

// ─── Introspection ───

// Bindings returns a defensive copy of all registered bindings.
func (km *KeybindingManager) Bindings() []KeyBinding {
	km.mu.RLock()
	defer km.mu.RUnlock()
	result := make([]KeyBinding, len(km.bindings))
	for i, b := range km.bindings {
		result[i] = *b
	}
	return result
}

// BindingCount returns the number of registered bindings.
func (km *KeybindingManager) BindingCount() int {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return len(km.bindings)
}

// FindByCommand returns the binding for a given command, or nil.
func (km *KeybindingManager) FindByCommand(command string) *KeyBinding {
	km.mu.RLock()
	defer km.mu.RUnlock()
	for _, b := range km.bindings {
		if b.Command == command {
			cp := *b
			return &cp
		}
	}
	return nil
}

// FindByKeys returns the binding for a given key description in the
// active context, or nil.
func (km *KeybindingManager) FindByKeys(keys string) *KeyBinding {
	km.mu.RLock()
	defer km.mu.RUnlock()
	normalized := normalizeKeys(keys)
	for _, b := range km.bindings {
		if b.Keys == normalized && (b.Context == km.activeCtx || b.Context == "global") {
			cp := *b
			return &cp
		}
	}
	return nil
}

// ─── Help Generation ───

// HelpText generates formatted help text showing all active bindings.
// Bindings are grouped by context and sorted by key.
func (km *KeybindingManager) HelpText() string {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.helpTextLocked()
}

func (km *KeybindingManager) helpTextLocked() string {
	active := km.activeBindingsLocked()
	if len(active) == 0 {
		return "No keybindings registered."
	}

	// Group by context
	byContext := make(map[string][]*KeyBinding)
	for _, b := range active {
		byContext[b.Context] = append(byContext[b.Context], b)
	}

	// Sort contexts
	contexts := make([]string, 0, len(byContext))
	for ctx := range byContext {
		contexts = append(contexts, ctx)
	}
	sort.Strings(contexts)

	var sb strings.Builder
	for _, ctx := range contexts {
		bindings := byContext[ctx]
		// Sort by keys
		sort.Slice(bindings, func(i, j int) bool {
			return bindings[i].Keys < bindings[j].Keys
		})

		if ctx != "global" {
			fmt.Fprintf(&sb, "\n%s:\n", ctx)
		}
		maxKeyLen := 0
		for _, b := range bindings {
			if len(b.Keys) > maxKeyLen {
				maxKeyLen = len(b.Keys)
			}
		}
		for _, b := range bindings {
			pad := strings.Repeat(" ", maxKeyLen-len(b.Keys))
			if ctx != "global" {
				sb.WriteString("  ")
			}
			fmt.Fprintf(&sb, "  %s%s  %s\n", b.Keys, pad, b.Help)
		}
	}
	return sb.String()
}

// ─── Conflict Detection ───

// CheckConflicts returns a list of conflicting bindings (same keys, same context).
func (km *KeybindingManager) CheckConflicts() []string {
	km.mu.RLock()
	defer km.mu.RUnlock()

	seen := make(map[string]string) // "context|keys" -> command
	var conflicts []string

	for _, b := range km.bindings {
		if !b.Enabled {
			continue
		}
		key := b.Context + "|" + b.Keys
		if prevCmd, exists := seen[key]; exists {
			conflicts = append(conflicts,
				fmt.Sprintf("%q and %q both bound to %s in %s",
					prevCmd, b.Command, b.Keys, b.Context))
		} else {
			seen[key] = b.Command
		}
	}
	return conflicts
}

// ─── Internal helpers ───

// activeBindingsLocked returns bindings sorted by priority:
// active context first, then global. Must be called with mu held.
func (km *KeybindingManager) activeBindingsLocked() []*KeyBinding {
	var ctxBindings, globalBindings []*KeyBinding
	for _, b := range km.bindings {
		if !b.Enabled {
			continue
		}
		if b.Context == km.activeCtx {
			ctxBindings = append(ctxBindings, b)
		} else if b.Context == "global" {
			globalBindings = append(globalBindings, b)
		}
	}
	return append(ctxBindings, globalBindings...)
}

// normalizeKeys standardizes key descriptions.
// "Ctrl+S" -> "ctrl+s", "Ctrl+X Ctrl+S" -> "ctrl+x ctrl+s"
func normalizeKeys(keys string) string {
	parts := strings.Fields(keys)
	for i, p := range parts {
		parts[i] = strings.ToLower(p)
	}
	return strings.Join(parts, " ")
}

// keyEventToDesc converts a term.KeyEvent to a key description string.
func keyEventToDesc(k *term.KeyEvent) string {
	if k == nil {
		return ""
	}
	var parts []string

	if k.Modifiers&term.ModCtrl != 0 {
		parts = append(parts, "ctrl")
	}
	if k.Modifiers&term.ModAlt != 0 {
		parts = append(parts, "alt")
	}
	if k.Modifiers&term.ModShift != 0 {
		parts = append(parts, "shift")
	}

	// Determine the key name
	var keyName string
	switch k.Key {
	case term.KeyEnter:
		keyName = "enter"
	case term.KeyTab:
		keyName = "tab"
	case term.KeyBackspace:
		keyName = "backspace"
	case term.KeyEscape:
		keyName = "esc"
	case term.KeyUp:
		keyName = "up"
	case term.KeyDown:
		keyName = "down"
	case term.KeyLeft:
		keyName = "left"
	case term.KeyRight:
		keyName = "right"
	case term.KeyHome:
		keyName = "home"
	case term.KeyEnd:
		keyName = "end"
	case term.KeyPageUp:
		keyName = "pageup"
	case term.KeyPageDown:
		keyName = "pagedown"
	case term.KeyDelete:
		keyName = "delete"
	case term.KeySpace:
		keyName = "space"
	default:
		if k.Rune != 0 {
			keyName = string(k.Rune)
		} else {
			keyName = fmt.Sprintf("key%d", k.Key)
		}
	}
	parts = append(parts, keyName)

	// If only one part (no modifiers), return just the key name
	if len(parts) == 1 {
		return parts[0]
	}
	return strings.Join(parts, "+")
}

// ParseKeyDesc parses a key description string into a KeyShortcut.
// Useful for testing without creating real KeyEvent objects.
// Example: "ctrl+s" -> KeyShortcut{Rune: 's', Modifiers: ModCtrl}
func ParseKeyDesc(desc string) (term.KeyCode, term.ModMask, rune) {
	desc = strings.ToLower(strings.TrimSpace(desc))
	var mods term.ModMask

	parts := strings.Split(desc, "+")
	keyPart := parts[len(parts)-1]

	for _, p := range parts[:len(parts)-1] {
		switch strings.TrimSpace(p) {
		case "ctrl":
			mods |= term.ModCtrl
		case "alt":
			mods |= term.ModAlt
		case "shift":
			mods |= term.ModShift
		}
	}

	var keyCode term.KeyCode
	var runeVal rune
	switch keyPart {
	case "enter":
		keyCode = term.KeyEnter
	case "tab":
		keyCode = term.KeyTab
	case "esc", "escape":
		keyCode = term.KeyEscape
	case "up":
		keyCode = term.KeyUp
	case "down":
		keyCode = term.KeyDown
	case "left":
		keyCode = term.KeyLeft
	case "right":
		keyCode = term.KeyRight
	case "home":
		keyCode = term.KeyHome
	case "end":
		keyCode = term.KeyEnd
	case "pageup":
		keyCode = term.KeyPageUp
	case "pagedown":
		keyCode = term.KeyPageDown
	case "delete", "del":
		keyCode = term.KeyDelete
	case "backspace":
		keyCode = term.KeyBackspace
	case "space":
		keyCode = term.KeySpace
		runeVal = ' '
	default:
		if len(keyPart) == 1 {
			runeVal = rune(keyPart[0])
		}
	}
	return keyCode, mods, runeVal
}
