package component

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/fuzzy"
	"github.com/topcheer/fluui/internal/term"
)

// FileEntry represents a single file or directory in the picker.
type FileEntry struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	Mode    os.FileMode
	ModTime int64 // unix timestamp, simplified for testing
}

// FilePickerStyle holds colors for rendering the file picker.
type FilePickerStyle struct {
	Border      buffer.Style
	Normal      buffer.Style
	Selected    buffer.Style
	DirColor    buffer.Style
	FileColor   buffer.Style
	Header      buffer.Style
	FilterPrompt buffer.Style
	Highlight   buffer.Style
	Checkbox    buffer.Style
}

// DefaultFilePickerStyle returns a Dracula-themed style.
func DefaultFilePickerStyle() FilePickerStyle {
	normal := buffer.Style{Fg: buffer.RGB(0xF8, 0xF8, 0xF2), Bg: buffer.RGB(0x28, 0x2A, 0x36)}
	return FilePickerStyle{
		Border:       buffer.Style{Fg: buffer.RGB(0x62, 0xD6, 0xE8), Bg: buffer.RGB(0x28, 0x2A, 0x36)},
		Normal:       normal,
		Selected:     buffer.Style{Fg: buffer.RGB(0x28, 0x2A, 0x36), Bg: buffer.RGB(0xFF, 0x79, 0xC6), Flags: buffer.Bold},
		DirColor:     buffer.Style{Fg: buffer.RGB(0x8B, 0xE9, 0xFD), Bg: buffer.RGB(0x28, 0x2A, 0x36), Flags: buffer.Bold},
		FileColor:    buffer.Style{Fg: buffer.RGB(0xF1, 0xFA, 0x8C), Bg: buffer.RGB(0x28, 0x2A, 0x36)},
		Header:       buffer.Style{Fg: buffer.RGB(0xBD, 0x93, 0xF9), Bg: buffer.RGB(0x21, 0x22, 0x2C), Flags: buffer.Bold},
		FilterPrompt: buffer.Style{Fg: buffer.RGB(0x50, 0xFA, 0x7B), Bg: buffer.RGB(0x28, 0x2A, 0x36), Flags: buffer.Bold},
		Highlight:    buffer.Style{Fg: buffer.RGB(0xFF, 0xB8, 0x6C), Bg: buffer.RGB(0x28, 0x2A, 0x36), Flags: buffer.Underline},
		Checkbox:     buffer.Style{Fg: buffer.RGB(0x50, 0xFA, 0x7B), Bg: buffer.RGB(0x28, 0x2A, 0x36)},
	}
}

// FilePicker is a file browser widget with directory navigation,
// fuzzy filtering, and multi-select support.
type FilePicker struct {
	BaseComponent
	mu sync.RWMutex

	cwd       string
	entries   []FileEntry
	filtered  []int // indices into entries
	cursor    int
	scrollY   int
	selected  map[string]bool // map of file paths
	filter    string
	filtering bool
	style     FilePickerStyle
	matcher   *fuzzy.Matcher

	// Callbacks
	OnSelect  func(entry FileEntry)
	OnConfirm func(entry FileEntry)
	OnError   func(err error)

	// For testing: allows overriding directory reading
	dirReader func(dir string) ([]FileEntry, error)
}

// NewFilePicker creates a new FilePicker rooted at the given directory.
// If dir is empty, the current working directory is used.
func NewFilePicker(dir string) *FilePicker {
	if dir == "" {
		dir, _ = os.Getwd()
	}
	fp := &FilePicker{
		cwd:      dir,
		cursor:   0,
		scrollY:  0,
		selected: make(map[string]bool),
		style:    DefaultFilePickerStyle(),
		matcher:  fuzzy.NewMatcher(),
		dirReader: readDirReal,
	}
	fp.SetID(GenerateID("filepicker"))
	fp.loadDir(dir)
	return fp
}

// readDirReal reads a real directory using os.ReadDir.
func readDirReal(dir string) ([]FileEntry, error) {
	infos, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	entries := make([]FileEntry, 0, len(infos))
	for _, info := range infos {
		fi, err := info.Info()
		if err != nil {
			continue
		}
		entries = append(entries, FileEntry{
			Name:    info.Name(),
			Path:    filepath.Join(dir, info.Name()),
			IsDir:   info.IsDir(),
			Size:    fi.Size(),
			Mode:    fi.Mode(),
			ModTime: fi.ModTime().Unix(),
		})
	}
	return entries, nil
}

// loadDir loads entries from the given directory and resets cursor/scroll.
func (fp *FilePicker) loadDir(dir string) {
	entries, err := fp.dirReader(dir)
	fp.mu.Lock()
	defer fp.mu.Unlock()
	if err != nil {
		fp.entries = nil
		fp.filtered = nil
		fp.cursor = 0
		fp.scrollY = 0
		if fp.OnError != nil {
			fp.OnError(err)
		}
		return
	}
	// Sort: directories first (alphabetical), then files (alphabetical)
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	fp.cwd = dir
	fp.entries = entries
	fp.cursor = 0
	fp.scrollY = 0
	fp.applyFilterLocked()
}

// applyFilterLocked recomputes the filtered list. Must be called under lock.
func (fp *FilePicker) applyFilterLocked() {
	if fp.filter == "" {
		fp.filtered = make([]int, len(fp.entries))
		for i := range fp.entries {
			fp.filtered[i] = i
		}
		return
	}
	fp.filtered = fp.filtered[:0]
	for i, e := range fp.entries {
		if fp.matcher.IsMatch(fp.filter, e.Name) {
			fp.filtered = append(fp.filtered, i)
		}
	}
	if fp.cursor >= len(fp.filtered) {
		fp.cursor = max(0, len(fp.filtered)-1)
	}
	fp.ensureVisibleLocked()
}

// --- Navigation ---

// MoveDown moves the cursor down by one.
func (fp *FilePicker) MoveDown() {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	fp.moveCursorLocked(1)
}

// MoveUp moves the cursor up by one.
func (fp *FilePicker) MoveUp() {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	fp.moveCursorLocked(-1)
}

func (fp *FilePicker) moveCursorLocked(delta int) {
	n := len(fp.filtered)
	if n == 0 {
		fp.cursor = 0
		return
	}
	fp.cursor = (fp.cursor + delta + n) % n
	fp.ensureVisibleLocked()
}

func (fp *FilePicker) ensureVisibleLocked() {
	bounds := fp.Bounds()
	viewH := bounds.H - 2 // minus border
	if viewH < 1 {
		viewH = 1
	}
	// Account for filter line
	if fp.filtering {
		viewH--
	}
	if viewH < 1 {
		viewH = 1
	}
	if fp.cursor < fp.scrollY {
		fp.scrollY = fp.cursor
	}
	if fp.cursor >= fp.scrollY+viewH {
		fp.scrollY = fp.cursor - viewH + 1
	}
	if fp.scrollY < 0 {
		fp.scrollY = 0
	}
}

// CurrentEntry returns the currently highlighted entry, or zero value if empty.
func (fp *FilePicker) CurrentEntry() (FileEntry, bool) {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	return fp.currentEntryLocked()
}

func (fp *FilePicker) currentEntryLocked() (FileEntry, bool) {
	if fp.cursor < 0 || fp.cursor >= len(fp.filtered) {
		return FileEntry{}, false
	}
	idx := fp.filtered[fp.cursor]
	return fp.entries[idx], true
}

// EnterDir navigates into the currently highlighted directory.
// If the current entry is a file, fires OnConfirm.
func (fp *FilePicker) EnterDir() {
	entry, ok := fp.CurrentEntry()
	if !ok {
		return
	}
	fp.mu.RLock()
	cb := fp.OnConfirm
	fp.mu.RUnlock()
	if entry.IsDir {
		fp.loadDir(entry.Path)
	} else if cb != nil {
		cb(entry)
	}
}

// GoUp navigates to the parent directory.
func (fp *FilePicker) GoUp() {
	fp.mu.Lock()
	parent := filepath.Dir(fp.cwd)
	fp.mu.Unlock()
	if parent != fp.cwd {
		fp.loadDir(parent)
	}
}

// ToggleSelect toggles the selection of the current file.
// Directories are not selected (they are navigated into instead).
func (fp *FilePicker) ToggleSelect() {
	fp.mu.Lock()
	entry, ok := fp.currentEntryLocked()
	if !ok || entry.IsDir {
		fp.mu.Unlock()
		return
	}
	if fp.selected[entry.Path] {
		delete(fp.selected, entry.Path)
	} else {
		fp.selected[entry.Path] = true
	}
	cb := fp.OnSelect
	fp.mu.Unlock()
	if cb != nil {
		cb(entry)
	}
}

// SelectedFiles returns all selected file paths.
func (fp *FilePicker) SelectedFiles() []string {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	paths := make([]string, 0, len(fp.selected))
	for p := range fp.selected {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}

// IsSelected returns whether the given path is selected.
func (fp *FilePicker) IsSelected(path string) bool {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	return fp.selected[path]
}

// ClearSelection clears all selected files.
func (fp *FilePicker) ClearSelection() {
	fp.mu.Lock()
	fp.selected = make(map[string]bool)
	fp.mu.Unlock()
}

// --- Filter ---

// SetFilter sets the filter query and updates the filtered list.
func (fp *FilePicker) SetFilter(q string) {
	fp.mu.Lock()
	fp.filter = q
	fp.applyFilterLocked()
	fp.mu.Unlock()
}

// Filter returns the current filter query.
func (fp *FilePicker) Filter() string {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	return fp.filter
}

// SetFiltering toggles filter input mode.
func (fp *FilePicker) SetFiltering(on bool) {
	fp.mu.Lock()
	fp.filtering = on
	if !on {
		fp.filter = ""
		fp.applyFilterLocked()
	}
	fp.mu.Unlock()
}

// IsFiltering returns whether filter mode is active.
func (fp *FilePicker) IsFiltering() bool {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	return fp.filtering
}

// AppendFilter appends a rune to the filter query.
func (fp *FilePicker) AppendFilter(r rune) {
	fp.mu.Lock()
	fp.filter += string(r)
	fp.applyFilterLocked()
	fp.mu.Unlock()
}

// BackspaceFilter removes the last character from the filter query.
func (fp *FilePicker) BackspaceFilter() {
	fp.mu.Lock()
	if len(fp.filter) > 0 {
		fp.filter = fp.filter[:len(fp.filter)-1]
		fp.applyFilterLocked()
	}
	fp.mu.Unlock()
}

// --- State Queries ---

// Cwd returns the current directory.
func (fp *FilePicker) Cwd() string {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	return fp.cwd
}

// Entries returns a copy of the current directory entries.
func (fp *FilePicker) Entries() []FileEntry {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	return append([]FileEntry(nil), fp.entries...)
}

// FilteredCount returns the number of entries matching the current filter.
func (fp *FilePicker) FilteredCount() int {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	return len(fp.filtered)
}

// Cursor returns the current cursor index.
func (fp *FilePicker) Cursor() int {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	return fp.cursor
}

// SetCursor sets the cursor to the given index (clamped).
func (fp *FilePicker) SetCursor(idx int) {
	fp.mu.Lock()
	if idx < 0 {
		idx = 0
	}
	if idx >= len(fp.filtered) {
		idx = len(fp.filtered) - 1
	}
	fp.cursor = idx
	fp.ensureVisibleLocked()
	fp.mu.Unlock()
}

// --- Style ---

// SetStyle sets the style for the file picker.
func (fp *FilePicker) SetStyle(s FilePickerStyle) {
	fp.mu.Lock()
	fp.style = s
	fp.mu.Unlock()
}

// Style returns the current style.
func (fp *FilePicker) Style() FilePickerStyle {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	return fp.style
}

// --- Callbacks ---

// SetOnSelect sets the callback for when a file is selected/deselected.
func (fp *FilePicker) SetOnSelect(fn func(entry FileEntry)) {
	fp.mu.Lock()
	fp.OnSelect = fn
	fp.mu.Unlock()
}

// SetOnConfirm sets the callback for when a file is confirmed (Enter on file).
func (fp *FilePicker) SetOnConfirm(fn func(entry FileEntry)) {
	fp.mu.Lock()
	fp.OnConfirm = fn
	fp.mu.Unlock()
}

// SetOnError sets the callback for directory read errors.
func (fp *FilePicker) SetOnError(fn func(err error)) {
	fp.mu.Lock()
	fp.OnError = fn
	fp.mu.Unlock()
}

// SetDirReader sets a custom directory reader (for testing).
func (fp *FilePicker) SetDirReader(fn func(dir string) ([]FileEntry, error)) {
	fp.mu.Lock()
	fp.dirReader = fn
	fp.mu.Unlock()
}

// --- Component Interface ---

// Measure returns the ideal size for the file picker.
func (fp *FilePicker) Measure(cs Constraints) Size {
	fp.mu.RLock()
	defer fp.mu.RUnlock()

	maxW := 40 // minimum width
	for _, e := range fp.entries {
		w := len(e.Name) + 6 // icon + checkbox + padding
		if w > maxW {
			maxW = w
		}
	}
	h := len(fp.entries) + 3 // border + header + entries
	if cs.MaxWidth > 0 && maxW > cs.MaxWidth {
		maxW = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	if h < 5 {
		h = 5
	}
	return Size{W: maxW, H: h}
}

// SetBounds sets the bounds and is called by the layout system.
func (fp *FilePicker) SetBounds(r Rect) {
	fp.mu.Lock()
	fp.BaseComponent.SetBounds(r)
	fp.ensureVisibleLocked()
	fp.mu.Unlock()
}

// Paint renders the file picker into the buffer.
func (fp *FilePicker) Paint(buf *buffer.Buffer) {
	fp.mu.RLock()
	defer fp.mu.RUnlock()

	bounds := fp.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	style := fp.style
	x, y := bounds.X, bounds.Y
	w, h := bounds.W, bounds.H

	// Draw border
	drawBorderBox(buf, x, y, w, h, style.Border)

	// Draw header (current path)
	headerText := fp.cwd
	if len(headerText) > w-2 {
		headerText = "..." + headerText[len(headerText)-(w-5):]
	}
	for i, r := range headerText {
		if x+1+i >= x+w-1 {
			break
		}
		buf.SetCell(x+1+i, y, buffer.Cell{Rune: r, Width: 1, Fg: style.Header.Fg, Bg: style.Header.Bg, Flags: style.Header.Flags})
	}

	// Draw filter line if filtering
	lineY := y + 1
	if fp.filtering {
		filterText := "/" + fp.filter
		for i, r := range filterText {
			if x+1+i >= x+w-1 {
				break
			}
			buf.SetCell(x+1+i, lineY, buffer.Cell{Rune: r, Width: 1, Fg: style.FilterPrompt.Fg, Bg: style.FilterPrompt.Bg, Flags: style.FilterPrompt.Flags})
		}
		lineY++
	}

	// Draw entries
	viewH := y + h - 1 - lineY // bottom border - current line
	if viewH < 1 {
		viewH = 1
	}

	scrollY := fp.scrollY
	for i := 0; i < viewH; i++ {
		idx := scrollY + i
		if idx >= len(fp.filtered) {
			break
		}
		entryIdx := fp.filtered[idx]
		entry := fp.entries[entryIdx]

		rowY := lineY + i
		isCurrent := idx == fp.cursor

		// Determine cell style
		var cellStyle buffer.Style
		if isCurrent {
			cellStyle = style.Selected
		} else if entry.IsDir {
			cellStyle = style.DirColor
		} else {
			cellStyle = style.FileColor
		}

		// Draw checkbox (for files)
		checkboxRune := ' '
		if !entry.IsDir {
			if fp.selected[entry.Path] {
				checkboxRune = '✓'
			} else {
				checkboxRune = '☐'
			}
		}
		buf.SetCell(x+1, rowY, buffer.Cell{Rune: checkboxRune, Width: 1, Fg: style.Checkbox.Fg, Bg: cellStyle.Bg, Flags: buffer.Bold})

		// Draw icon
		iconRune := ' ' // assigned below but initialized for safety
		if entry.IsDir {
			iconRune = '▸'
		}
		buf.SetCell(x+2, rowY, buffer.Cell{Rune: iconRune, Width: 1, Fg: cellStyle.Fg, Bg: cellStyle.Bg, Flags: cellStyle.Flags})

		// Draw name
		nameText := entry.Name
		maxNameW := w - 5 // border + checkbox + icon + padding
		if maxNameW < 1 {
			maxNameW = 1
		}
		if len(nameText) > maxNameW {
			nameText = nameText[:maxNameW-1] + "…"
		}
		for i, r := range nameText {
			buf.SetCell(x+4+i, rowY, buffer.Cell{Rune: r, Width: 1, Fg: cellStyle.Fg, Bg: cellStyle.Bg, Flags: cellStyle.Flags})
		}

		// Fill rest of row with background
		for i := 4 + len(nameText); i < w-1; i++ {
			buf.SetCell(x+i, rowY, buffer.Cell{Rune: ' ', Width: 1, Fg: cellStyle.Fg, Bg: cellStyle.Bg, Flags: cellStyle.Flags})
		}
	}
}

// Children returns nil (leaf component).
func (fp *FilePicker) Children() []Component {
	return nil
}

// --- HandleKey ---

// HandleKey processes keyboard input.
// Returns true if the key was consumed.
//
// Key bindings:
//   - Up/k: move cursor up
//   - Down/j: move cursor down
//   - Enter: open directory or confirm file
//   - Backspace (not filtering): go to parent directory
//   - Backspace (filtering): delete last char
//   - Space: toggle file selection (not in filtering mode)
//   - / : enter filter mode
//   - Esc: exit filter mode or go to parent
//   - Home/g: jump to first
//   - End/G: jump to last
//   - PageUp: scroll up
//   - PageDown: scroll down
//   - h: go to parent directory (vim style)
//   - l: open directory/file (vim style)
func (fp *FilePicker) HandleKey(key *term.KeyEvent) bool {
	if key == nil {
		return false
	}

	fp.mu.RLock()
	filtering := fp.filtering
	fp.mu.RUnlock()

	if filtering {
		return fp.handleFilterKey(key)
	}

	switch key.Key {
	case term.KeyUp:
		if key.Rune == 'k' || key.Rune == 0 {
			fp.MoveUp()
			return true
		}
	case term.KeyDown:
		if key.Rune == 'j' || key.Rune == 0 {
			fp.MoveDown()
			return true
		}
	case term.KeyEnter:
		fp.EnterDir()
		return true
	case term.KeyBackspace:
		fp.GoUp()
		return true
	case term.KeySpace:
		fp.ToggleSelect()
		return true
	case term.KeyHome:
		fp.SetCursor(0)
		return true
	case term.KeyEnd:
		fp.mu.RLock()
		n := len(fp.filtered)
		fp.mu.RUnlock()
		fp.SetCursor(n - 1)
		return true
	case term.KeyPageUp:
		fp.mu.Lock()
		bounds := fp.Bounds()
		viewH := bounds.H - 3
		if viewH < 1 {
			viewH = 1
		}
		fp.moveCursorLocked(-viewH)
		fp.mu.Unlock()
		return true
	case term.KeyPageDown:
		fp.mu.Lock()
		bounds := fp.Bounds()
		viewH := bounds.H - 3
		if viewH < 1 {
			viewH = 1
		}
		fp.moveCursorLocked(viewH)
		fp.mu.Unlock()
		return true
	case term.KeyEscape:
		fp.GoUp()
		return true
	}

	// Check for rune-based shortcuts
	switch key.Rune {
	case 'j':
		fp.MoveDown()
		return true
	case 'k':
		fp.MoveUp()
		return true
	case 'h':
		fp.GoUp()
		return true
	case 'l':
		fp.EnterDir()
		return true
	case 'g':
		fp.SetCursor(0)
		return true
	case 'G':
		fp.mu.RLock()
		n := len(fp.filtered)
		fp.mu.RUnlock()
		fp.SetCursor(n - 1)
		return true
	case '/':
		fp.SetFiltering(true)
		return true
	}

	return false
}

// handleFilterKey processes keys while in filter mode.
func (fp *FilePicker) handleFilterKey(key *term.KeyEvent) bool {
	switch key.Key {
	case term.KeyEscape:
		fp.SetFiltering(false)
		return true
	case term.KeyEnter:
		fp.SetFiltering(false)
		// Enter on first match
		return true
	case term.KeyBackspace:
		fp.mu.RLock()
		f := fp.filter
		fp.mu.RUnlock()
		if len(f) == 0 {
			fp.SetFiltering(false)
		} else {
			fp.BackspaceFilter()
		}
		return true
	case term.KeyUp:
		fp.MoveUp()
		return true
	case term.KeyDown:
		fp.MoveDown()
		return true
	}

	// Printable rune
	if key.Rune != 0 && key.Rune >= 0x20 {
		fp.AppendFilter(key.Rune)
		return true
	}

	return false
}

// --- Helpers ---

// drawBorderBox draws a Unicode box border.
func drawBorderBox(buf *buffer.Buffer, x, y, w, h int, style buffer.Style) {
	if w < 2 || h < 2 {
		return
	}
	// Corners
	buf.SetCell(x, y, buffer.Cell{Rune: '┌', Width: 1, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags})
	buf.SetCell(x+w-1, y, buffer.Cell{Rune: '┐', Width: 1, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags})
	buf.SetCell(x, y+h-1, buffer.Cell{Rune: '└', Width: 1, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags})
	buf.SetCell(x+w-1, y+h-1, buffer.Cell{Rune: '┘', Width: 1, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags})

	// Horizontal lines
	for i := 1; i < w-1; i++ {
		buf.SetCell(x+i, y, buffer.Cell{Rune: '─', Width: 1, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags})
		buf.SetCell(x+i, y+h-1, buffer.Cell{Rune: '─', Width: 1, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags})
	}
	// Vertical lines
	for i := 1; i < h-1; i++ {
		buf.SetCell(x, y+i, buffer.Cell{Rune: '│', Width: 1, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags})
		buf.SetCell(x+w-1, y+i, buffer.Cell{Rune: '│', Width: 1, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags})
	}
}

// String returns a string representation of the file picker state.
func (fp *FilePicker) String() string {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	return fmt.Sprintf("FilePicker{cwd:%s, entries:%d, cursor:%d, filter:%q}", fp.cwd, len(fp.entries), fp.cursor, fp.filter)
}


