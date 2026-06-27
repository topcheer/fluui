package app

import (
	"sort"
	"strings"
)

// CompletionItem represents a single completion suggestion.
type CompletionItem struct {
	Label  string // display text, e.g. "/theme"
	Detail string // description, e.g. "Change theme"
	Insert string // text to insert, e.g. "/theme "
}

// CompletionProvider returns candidates for tab completion.
type CompletionProvider interface {
	Candidates(prefix string) []CompletionItem
}

// MultiProvider combines multiple completion providers.
// It merges results from all providers, sorted alphabetically by Label.
type MultiProvider struct {
	providers []CompletionProvider
}

// NewMultiProvider creates a MultiProvider from the given providers.
func NewMultiProvider(providers ...CompletionProvider) *MultiProvider {
	return &MultiProvider{providers: providers}
}

// Candidates queries all providers and merges results.
func (m *MultiProvider) Candidates(prefix string) []CompletionItem {
	var all []CompletionItem
	for _, p := range m.providers {
		all = append(all, p.Candidates(prefix)...)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].Label < all[j].Label
	})
	return all
}

// --- SlashCommandProvider ---

// SlashCommandProvider provides /command completions.
type SlashCommandProvider struct {
	commands map[string]string // name (without /) → description
}

// NewSlashCommandProvider creates a SlashCommandProvider with default commands.
func NewSlashCommandProvider() *SlashCommandProvider {
	return &SlashCommandProvider{
		commands: map[string]string{
			"help":  "Show help",
			"clear": "Clear conversation",
			"theme": "Change theme (e.g. /theme dracula)",
			"copy":  "Copy last block to clipboard",
			"save":  "Save conversation to file",
			"load":  "Load conversation from file",
			"model": "Switch AI model",
			"quit":  "Quit application",
		},
	}
}

// Candidates returns matching slash commands for the given prefix.
// The prefix must start with "/" or be empty.
func (p *SlashCommandProvider) Candidates(prefix string) []CompletionItem {
	// Only complete if the prefix is "/" or starts with "/".
	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		return nil
	}

	name := strings.TrimPrefix(prefix, "/")

	var items []CompletionItem
	for cmd, desc := range p.commands {
		if strings.HasPrefix(cmd, name) {
			items = append(items, CompletionItem{
				Label:  "/" + cmd,
				Detail: desc,
				Insert: "/" + cmd + " ",
			})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Label < items[j].Label
	})
	return items
}

// AddCommand registers a custom slash command.
func (p *SlashCommandProvider) AddCommand(name, description string) {
	p.commands[name] = description
}

// RemoveCommand removes a slash command.
func (p *SlashCommandProvider) RemoveCommand(name string) {
	delete(p.commands, name)
}

// Commands returns a sorted list of all command names.
func (p *SlashCommandProvider) Commands() []string {
	names := make([]string, 0, len(p.commands))
	for cmd := range p.commands {
		names = append(names, cmd)
	}
	sort.Strings(names)
	return names
}

// --- MentionProvider ---

// MentionProvider provides @mention completions.
type MentionProvider struct {
	mentions map[string]string // name (without @) → description
}

// NewMentionProvider creates a MentionProvider with default mentions.
func NewMentionProvider() *MentionProvider {
	return &MentionProvider{
		mentions: map[string]string{
			"file":  "Attach a file",
			"url":   "Fetch URL content",
			"image": "Attach an image",
			"code":  "Insert code block",
		},
	}
}

// Candidates returns matching mentions for the given prefix.
// The prefix must start with "@" or be empty.
func (p *MentionProvider) Candidates(prefix string) []CompletionItem {
	if prefix != "" && !strings.HasPrefix(prefix, "@") {
		return nil
	}

	name := strings.TrimPrefix(prefix, "@")

	var items []CompletionItem
	for mention, desc := range p.mentions {
		if strings.HasPrefix(mention, name) {
			items = append(items, CompletionItem{
				Label:  "@" + mention,
				Detail: desc,
				Insert: "@" + mention + " ",
			})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Label < items[j].Label
	})
	return items
}

// AddMention registers a custom mention.
func (p *MentionProvider) AddMention(name, description string) {
	p.mentions[name] = description
}

// RemoveMention removes a mention.
func (p *MentionProvider) RemoveMention(name string) {
	delete(p.mentions, name)
}

// --- CompletionManager ---

// CompletionManager handles tab completion state.
// It tracks the current candidates, selected index, and active state.
type CompletionManager struct {
	provider CompletionProvider
	items    []CompletionItem // current candidates
	selected int              // currently highlighted item (0-based)
	active   bool             // popup visible
	prefix   string           // what triggered completion
}

// NewCompletionManager creates a CompletionManager with the given provider.
func NewCompletionManager(provider CompletionProvider) *CompletionManager {
	return &CompletionManager{provider: provider}
}

// SetProvider updates the completion provider.
func (m *CompletionManager) SetProvider(p CompletionProvider) {
	m.provider = p
	m.Cancel()
}

// Provider returns the current completion provider.
func (m *CompletionManager) Provider() CompletionProvider {
	return m.provider
}

// Active reports whether the completion popup is visible.
func (m *CompletionManager) Active() bool { return m.active }

// Items returns the current completion candidates.
func (m *CompletionManager) Items() []CompletionItem { return m.items }

// Selected returns the currently selected completion item.
// Returns false if completion is not active.
func (m *CompletionManager) Selected() (CompletionItem, bool) {
	if !m.active || m.selected < 0 || m.selected >= len(m.items) {
		return CompletionItem{}, false
	}
	return m.items[m.selected], true
}

// SelectedIndex returns the currently selected index.
func (m *CompletionManager) SelectedIndex() int { return m.selected }

// Prefix returns the prefix that triggered the current completion.
func (m *CompletionManager) Prefix() string { return m.prefix }

// Start triggers completion for the given prefix.
// Queries the provider and activates the popup if candidates exist.
// Returns true if the popup is now active.
func (m *CompletionManager) Start(prefix string) bool {
	if m.provider == nil {
		return false
	}
	m.prefix = prefix
	m.items = m.provider.Candidates(prefix)
	if len(m.items) == 0 {
		m.active = false
		m.selected = 0
		return false
	}
	m.active = true
	m.selected = 0
	return true
}

// CycleNext moves selection to the next candidate (wraps around).
// Returns the newly selected item and true if successful.
func (m *CompletionManager) CycleNext() (CompletionItem, bool) {
	if !m.active || len(m.items) == 0 {
		return CompletionItem{}, false
	}
	m.selected = (m.selected + 1) % len(m.items)
	return m.items[m.selected], true
}

// CyclePrev moves selection to the previous candidate (wraps around).
// Returns the newly selected item and true if successful.
func (m *CompletionManager) CyclePrev() (CompletionItem, bool) {
	if !m.active || len(m.items) == 0 {
		return CompletionItem{}, false
	}
	m.selected = (m.selected - 1 + len(m.items)) % len(m.items)
	return m.items[m.selected], true
}

// Accept returns the currently selected item and deactivates the popup.
// Returns false if completion is not active.
func (m *CompletionManager) Accept() (CompletionItem, bool) {
	if !m.active {
		return CompletionItem{}, false
	}
	item := CompletionItem{}
	if m.selected >= 0 && m.selected < len(m.items) {
		item = m.items[m.selected]
	}
	m.active = false
	return item, item.Insert != ""
}

// Cancel dismisses the completion popup without accepting.
func (m *CompletionManager) Cancel() {
	m.active = false
	m.selected = 0
	m.items = nil
	m.prefix = ""
}

// ExtractCompletionPrefix finds the completion prefix at the cursor position.
// If the cursor is at the end of a word starting with "/" or "@", returns that word.
// Otherwise returns empty string (no completion).
func ExtractCompletionPrefix(text string, cursor int) string {
	if cursor <= 0 || cursor > len(text) {
		return ""
	}

	// Find the start of the current word.
	start := cursor - 1
	for start >= 0 && text[start] != ' ' {
		start--
	}
	start++

	word := text[start:cursor]
	if len(word) == 0 {
		return ""
	}

	// Only complete words starting with / or @.
	if word[0] == '/' || word[0] == '@' {
		return word
	}

	return ""
}

// ReplacePrefix replaces the completion prefix in text with the insertion text.
// Returns the new text and the new cursor position.
func ReplacePrefix(text string, cursor int, prefix, insertion string) (string, int) {
	if cursor <= 0 || cursor > len(text) {
		return text, cursor
	}

	// Find where the prefix starts.
	start := cursor - len(prefix)
	if start < 0 {
		return text, cursor
	}

	// Verify the prefix matches at this position.
	if text[start:cursor] != prefix {
		return text, cursor
	}

	// Replace prefix with insertion.
	newText := text[:start] + insertion + text[cursor:]
	newCursor := start + len(insertion)
	return newText, newCursor
}
