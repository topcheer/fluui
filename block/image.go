package block

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/internal/termcompat"
	"github.com/topcheer/fluui/theme"
)

// ImageBlock displays an inline image in the AI chat.
//
// When the terminal supports an image protocol (iTerm2, Kitty Graphics, or
// Sixel), ImageBlock generates the appropriate escape sequence via Sequence().
// The ChatApp or renderer can emit this sequence at the block's position to
// display the image directly in the terminal.
//
// Regardless of protocol support, Paint always renders an ASCII placeholder
// showing the image metadata (filename, dimensions, file size, protocol name).
// This ensures the chat is readable even on terminals without image support.
//
// Thread-safe via BaseBlock's RWMutex.
type ImageBlock struct {
	BaseBlock

	filename string // original filename
	format   string // "png", "jpeg", "gif", "rgba"
	imgW     int    // pixel width (0 = unknown)
	imgH     int    // pixel height (0 = unknown)
	data     []byte // raw image bytes (PNG/JPEG/GIF) or RGBA pixel data

	// display dimensions in terminal cells (0 = auto)
	displayW int
	displayH int

	// detected protocol and cached escape sequence
	protocol  termcompat.ImageProtocol
	sequence  string
	seqCached bool
}

// NewImageBlock creates an image block with the given filename and image data.
// The data should be raw image file bytes (PNG, JPEG, GIF, etc.).
// NewImageBlock auto-detects the best available terminal image protocol.
func NewImageBlock(id, filename string, data []byte) *ImageBlock {
	b := &ImageBlock{
		BaseBlock: NewBaseBlock(id, TypeImage),
		filename:  filename,
		data:      data,
		format:    detectImageFormat(filename, data),
	}
	b.Complete() // Images are complete on arrival
	return b
}

// NewImageBlockWithDims creates an image block with explicit pixel dimensions.
// Use this when the image dimensions are known separately from the data.
func NewImageBlockWithDims(id, filename string, data []byte, w, h int) *ImageBlock {
	b := NewImageBlock(id, filename, data)
	b.imgW = w
	b.imgH = h
	return b
}

// NewRGBAImageBlock creates an image block from raw RGBA pixel data.
// The width and height specify the pixel dimensions.
// On Sixel-capable terminals, this data is encoded directly.
func NewRGBAImageBlock(id, filename string, rgba []byte, w, h int) *ImageBlock {
	b := &ImageBlock{
		BaseBlock: NewBaseBlock(id, TypeImage),
		filename:  filename,
		data:      rgba,
		format:    "rgba",
		imgW:      w,
		imgH:      h,
	}
	b.Complete()
	return b
}

// ─── Public API ───

// Filename returns the image filename.
func (b *ImageBlock) Filename() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.filename
}

// Format returns the image format ("png", "jpeg", "gif", "rgba").
func (b *ImageBlock) Format() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.format
}

// Data returns the raw image data.
func (b *ImageBlock) Data() []byte {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.data
}

// Width returns the pixel width (0 if unknown).
func (b *ImageBlock) Width() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.imgW
}

// Height returns the pixel height (0 if unknown).
func (b *ImageBlock) Height() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.imgH
}

// DisplayWidth returns the display width in terminal cells (0 = auto).
func (b *ImageBlock) DisplayWidth() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.displayW
}

// DisplayHeight returns the display height in terminal cells (0 = auto).
func (b *ImageBlock) DisplayHeight() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.displayH
}

// SetDisplaySize sets the display dimensions in terminal cells.
// Use 0 for auto-sizing.
func (b *ImageBlock) SetDisplaySize(w, h int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.displayW = w
	b.displayH = h
	b.seqCached = false
	b.markDirtyLocked()
}

// Protocol returns the detected terminal image protocol.
func (b *ImageBlock) Protocol() termcompat.ImageProtocol {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.protocol == 0 && !b.seqCached {
		b.mu.RUnlock()
		b.mu.Lock()
		b.protocol = termcompat.DetectImageProtocol().Protocol
		b.mu.Unlock()
		b.mu.RLock()
	}
	return b.protocol
}

// ProtocolName returns a human-readable name for the image protocol.
func (b *ImageBlock) ProtocolName() string {
	return termcompat.ImageProtocolName(b.Protocol())
}

// Sequence returns the terminal escape sequence to display the image.
// Returns "" if no image protocol is supported.
// The sequence is computed once and cached.
func (b *ImageBlock) Sequence() string {
	b.mu.RLock()
	if b.seqCached {
		seq := b.sequence
		b.mu.RUnlock()
		return seq
	}
	b.mu.RUnlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.seqCached {
		return b.sequence
	}

	b.sequence = b.generateSequenceLocked()
	b.seqCached = true
	return b.sequence
}

// HasImageSupport returns true if the terminal supports any image protocol.
func (b *ImageBlock) HasImageSupport() bool {
	return b.Protocol() != termcompat.ImageNone
}

// FileSize returns the human-readable file size.
func (b *ImageBlock) FileSize() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return formatFileSize(len(b.data))
}

// ─── Block interface methods ───

// Measure returns the desired size of the image block placeholder.
func (b *ImageBlock) Measure(constraints component.Constraints) component.Size {
	b.mu.RLock()
	defer b.mu.RUnlock()

	maxWidth := constraints.MaxWidth
	// Minimum size: 2 rows (border + info), width = max(filename, metadata)
	info := b.infoLineLocked()
	meta := b.metaLineLocked()
	w := len(info) + 4 // padding + border
	if metaW := len(meta) + 4; metaW > w {
		w = metaW
	}
	if maxWidth > 0 && w > maxWidth {
		w = maxWidth
	}
	if w < 10 {
		w = 10
	}
	return component.Size{W: w, H: b.measureHeightLocked(w)}
}

// Paint renders an ASCII placeholder showing the image metadata.
func (b *ImageBlock) Paint(buf *buffer.Buffer) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	bounds := b.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	t := theme.Get()
	borderStyle := buffer.Style{
		Fg: t.Accent,
	}
	infoStyle := buffer.Style{
		Fg: t.AssistantFg,
	}
	metaStyle := buffer.Style{
		Fg: t.AssistantFg,
	}
	metaStyle.Flags |= buffer.Dim

	// Draw top border
	buf.DrawText(bounds.X, bounds.Y, "┌", borderStyle)
	for x := 1; x < bounds.W-1; x++ {
		buf.DrawText(bounds.X+x, bounds.Y, "─", borderStyle)
	}
	buf.DrawText(bounds.X+bounds.W-1, bounds.Y, "┐", borderStyle)

	// Draw content
	info := b.infoLineLocked()
	meta := b.metaLineLocked()
	innerW := bounds.W - 4 // left border + padding + right border + padding

	row := 1
	// Image icon + filename
	label := "  " + truncateText(info, innerW)
	buf.DrawText(bounds.X+1, bounds.Y+row, "│", borderStyle)
	buf.DrawText(bounds.X+2, bounds.Y+row, label, infoStyle)
	// Right border
	buf.DrawText(bounds.X+bounds.W-1, bounds.Y+row, "│", borderStyle)
	row++

	// Metadata line
	if row < bounds.H-1 {
		metaLabel := "  " + truncateText(meta, innerW)
		buf.DrawText(bounds.X+1, bounds.Y+row, "│", borderStyle)
		buf.DrawText(bounds.X+2, bounds.Y+row, metaLabel, metaStyle)
		buf.DrawText(bounds.X+bounds.W-1, bounds.Y+row, "│", borderStyle)
		row++
	}

	// Protocol indicator line
	if b.protocol != termcompat.ImageNone && row < bounds.H-1 {
		protoLabel := fmt.Sprintf("  Protocol: %s", termcompat.ImageProtocolName(b.protocol))
		buf.DrawText(bounds.X+1, bounds.Y+row, "│", borderStyle)
		buf.DrawText(bounds.X+2, bounds.Y+row, truncateText(protoLabel, innerW+2), metaStyle)
		buf.DrawText(bounds.X+bounds.W-1, bounds.Y+row, "│", borderStyle)
		row++
	}

	// Fill remaining rows with borders
	for ; row < bounds.H-1; row++ {
		buf.DrawText(bounds.X+1, bounds.Y+row, "│", borderStyle)
		buf.DrawText(bounds.X+bounds.W-1, bounds.Y+row, "│", borderStyle)
	}

	// Draw bottom border
	bottomY := bounds.Y + bounds.H - 1
	buf.DrawText(bounds.X, bottomY, "└", borderStyle)
	for x := 1; x < bounds.W-1; x++ {
		buf.DrawText(bounds.X+x, bottomY, "─", borderStyle)
	}
	buf.DrawText(bounds.X+bounds.W-1, bottomY, "┘", borderStyle)
}

// SerializeState serializes the image block's metadata.
// Note: image data is stored as base64; large images will produce large JSON.
func (b *ImageBlock) SerializeState() (json.RawMessage, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	data := struct {
		Filename   string `json:"filename"`
		Format     string `json:"format"`
		Width      int    `json:"width"`
		Height     int    `json:"height"`
		DisplayW   int    `json:"display_w"`
		DisplayH   int    `json:"display_h"`
		DataBase64 string `json:"data_b64"`
	}{
		Filename:   b.filename,
		Format:     b.format,
		Width:      b.imgW,
		Height:     b.imgH,
		DisplayW:   b.displayW,
		DisplayH:   b.displayH,
		DataBase64: base64.StdEncoding.EncodeToString(b.data),
	}
	return json.Marshal(data)
}

// DeserializeState restores the image block's state from JSON.
func (b *ImageBlock) DeserializeState(data json.RawMessage) error {
	var s struct {
		Filename   string `json:"filename"`
		Format     string `json:"format"`
		Width      int    `json:"width"`
		Height     int    `json:"height"`
		DisplayW   int    `json:"display_w"`
		DisplayH   int    `json:"display_h"`
		DataBase64 string `json:"data_b64"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.filename = s.Filename
	b.format = s.Format
	b.imgW = s.Width
	b.imgH = s.Height
	b.displayW = s.DisplayW
	b.displayH = s.DisplayH
	if s.DataBase64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(s.DataBase64)
		if err == nil {
			b.data = decoded
		}
	}
	b.seqCached = false
	b.markDirtyLocked()
	return nil
}

// ─── Internal helpers ───

// generateSequenceLocked generates the image escape sequence.
// Caller must hold b.mu.
func (b *ImageBlock) generateSequenceLocked() string {
	caps := termcompat.DetectImageProtocol()
	b.protocol = caps.Protocol
	if caps.Protocol == termcompat.ImageNone {
		return ""
	}

	b64Data := base64.StdEncoding.EncodeToString(b.data)

	switch caps.Protocol {
	case termcompat.ImageKitty:
		return termcompat.FormatKittyImage(b64Data, b.displayW, b.displayH)
	case termcompat.ImageIterm2:
		return termcompat.FormatIterm2Image(b64Data, b.filename, b.displayW, b.displayH)
	case termcompat.ImageSixel:
		// Sixel requires RGBA pixel data, not file format.
		// If we have raw RGBA data, encode directly.
		// Otherwise, we can't encode file-format images to Sixel without a decoder.
		if b.format == "rgba" && b.imgW > 0 && b.imgH > 0 {
			return term.EncodeSixel(b.data, b.imgW, b.imgH)
		}
		// Fallback: no Sixel encoding possible without pixel data
		return ""
	}
	return ""
}

// infoLineLocked returns the primary info line (icon + filename + dimensions).
func (b *ImageBlock) infoLineLocked() string {
	var sb strings.Builder
	sb.WriteString("[IMG] ")
	if b.filename != "" {
		sb.WriteString(b.filename)
	} else {
		sb.WriteString("image")
	}
	if b.imgW > 0 && b.imgH > 0 {
		sb.WriteString(fmt.Sprintf(" (%dx%d)", b.imgW, b.imgH))
	}
	return sb.String()
}

// metaLineLocked returns the metadata line (size + format + protocol status).
func (b *ImageBlock) metaLineLocked() string {
	var sb strings.Builder
	sb.WriteString(formatFileSize(len(b.data)))
	if b.format != "" {
		sb.WriteString(" · ")
		sb.WriteString(strings.ToUpper(b.format))
	}
	if b.protocol != termcompat.ImageNone {
		sb.WriteString(" · ")
		sb.WriteString(termcompat.ImageProtocolName(b.protocol))
	} else {
		sb.WriteString(" · no display")
	}
	return sb.String()
}

// measureHeightLocked computes the display height based on content.
func (b *ImageBlock) measureHeightLocked(w int) int {
	// Border (top + bottom) + info + meta + optional protocol = 3-5 rows
	h := 3 // minimum: border top, info, border bottom
	// Add meta line
	h++
	// Add protocol line if present
	if b.protocol != termcompat.ImageNone {
		h++
	}
	// Enforce display height if set
	if b.displayH > 0 && b.displayH > h {
		h = b.displayH
	}
	return h
}

// detectImageFormat guesses the image format from filename extension and data.
func detectImageFormat(filename string, data []byte) string {
	// Check by extension first
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".png"):
		return "png"
	case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
		return "jpeg"
	case strings.HasSuffix(lower, ".gif"):
		return "gif"
	case strings.HasSuffix(lower, ".bmp"):
		return "bmp"
	case strings.HasSuffix(lower, ".webp"):
		return "webp"
	}

	// Check by magic bytes
	if len(data) >= 4 {
		if data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G' {
			return "png"
		}
		if data[0] == 0xFF && data[1] == 0xD8 {
			return "jpeg"
		}
		if data[0] == 'G' && data[1] == 'I' && data[2] == 'F' {
			return "gif"
		}
		if data[0] == 'B' && data[1] == 'M' {
			return "bmp"
		}
	}

	return "unknown"
}

// formatFileSize returns a human-readable file size string.
func formatFileSize(bytes int) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// truncateText truncates text to fit within maxLen, adding ellipsis if needed.
func truncateText(text string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	if maxLen <= 1 {
		return "…"
	}
	return string(runes[:maxLen-1]) + "…"
}
