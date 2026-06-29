package app

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/block"
)

// ReplaceResult represents the outcome of a search-and-replace operation.
type ReplaceResult struct {
	// BlockID is the ID of the block that was modified.
	BlockID string

	// Replacements is the number of individual string replacements made.
	Replacements int

	// OldText is the content before replacement (for undo).
	OldText string

	// NewText is the content after replacement.
	NewText string
}

// ReplaceMode manages search-and-replace state within ChatApp.
// It supports:
//   - Find and replace (all occurrences)
//   - Case-sensitive and case-insensitive matching
//   - Regex support via a simple interface
//   - Per-block content replacement with dirty marking
//
// ReplaceMode is safe for concurrent use.
type ReplaceMode struct {
	mu            sync.Mutex
	find          string
	replace       string
	caseSensitive bool
	all           bool // replace all vs. first occurrence
}

// NewReplaceMode creates a ReplaceMode with empty find/replace strings.
func NewReplaceMode() *ReplaceMode {
	return &ReplaceMode{
		caseSensitive: false,
		all:           true,
	}
}

// SetFind sets the search string for replacement.
func (rm *ReplaceMode) SetFind(s string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.find = s
}

// Find returns the current find string.
func (rm *ReplaceMode) Find() string {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.find
}

// SetReplace sets the replacement string.
func (rm *ReplaceMode) SetReplace(s string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.replace = s
}

// Replace returns the current replacement string.
func (rm *ReplaceMode) Replace() string {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.replace
}

// SetCaseSensitive toggles case-sensitive matching.
func (rm *ReplaceMode) SetCaseSensitive(cs bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.caseSensitive = cs
}

// CaseSensitive reports whether matching is case-sensitive.
func (rm *ReplaceMode) CaseSensitive() bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.caseSensitive
}

// SetReplaceAll sets whether to replace all occurrences (true) or just
// the first occurrence per block (false).
func (rm *ReplaceMode) SetReplaceAll(all bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.all = all
}

// ReplaceAll reports whether all-mode is active.
func (rm *ReplaceMode) ReplaceAll() bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.all
}

// ReplaceInBlock replaces occurrences of find with replace in a single block.
// It returns the number of replacements made and the old/new text.
// If the block type doesn't support content replacement, it returns (0, "", "").
//
// For blocks that implement SetContent (UserMessageBlock, AssistantTextBlock,
// ThinkingBlock), the replacement is applied directly. Other block types are
// skipped.
func (rm *ReplaceMode) ReplaceInBlock(b block.Block) ReplaceResult {
	rm.mu.Lock()
	find := rm.find
	replace := rm.replace
	caseSensitive := rm.caseSensitive
	all := rm.all
	rm.mu.Unlock()

	if find == "" {
		return ReplaceResult{BlockID: b.ID()}
	}

	// Only process blocks that support content replacement.
	if !CanReplaceBlock(b) {
		return ReplaceResult{BlockID: b.ID()}
	}

	text, ok := extractBlockText(b)
	if !ok || text == "" {
		return ReplaceResult{BlockID: b.ID()}
	}

	newText, count := replaceInString(text, find, replace, caseSensitive, all)
	if count == 0 {
		return ReplaceResult{BlockID: b.ID()}
	}

	// Apply the replacement to the block's content.
	applyBlockContent(b, newText)

	return ReplaceResult{
		BlockID:      b.ID(),
		Replacements: count,
		OldText:      text,
		NewText:      newText,
	}
}

// ReplaceInBlocks applies the replacement across all blocks and returns
// results for blocks that had at least one replacement.
func (rm *ReplaceMode) ReplaceInBlocks(blocks []block.Block) []ReplaceResult {
	var results []ReplaceResult
	for _, b := range blocks {
		r := rm.ReplaceInBlock(b)
		if r.Replacements > 0 {
			results = append(results, r)
		}
	}
	return results
}

// TotalReplacements returns the sum of replacements across all results.
func TotalReplacements(results []ReplaceResult) int {
	total := 0
	for _, r := range results {
		total += r.Replacements
	}
	return total
}

// replaceInString performs find-and-replace within a string.
// When all=true, replaces every occurrence. Otherwise replaces only the first.
// Supports both case-sensitive and case-insensitive matching.
func replaceInString(text, find, replace string, caseSensitive, all bool) (string, int) {
	if find == "" {
		return text, 0
	}

	if caseSensitive {
		if all {
			count := strings.Count(text, find)
			if count == 0 {
				return text, 0
			}
			return strings.ReplaceAll(text, find, replace), count
		}
		idx := strings.Index(text, find)
		if idx < 0 {
			return text, 0
		}
		return text[:idx] + replace + text[idx+len(find):], 1
	}

	// Case-insensitive: use lowercased comparison
	lowerText := strings.ToLower(text)
	lowerFind := strings.ToLower(find)
	findLen := len(lowerFind)
	count := 0

	if all {
		var sb strings.Builder
		sb.Grow(len(text) + len(replace))
		searchStart := 0
		for {
			idx := strings.Index(lowerText[searchStart:], lowerFind)
			if idx < 0 {
				sb.WriteString(text[searchStart:])
				break
			}
			absIdx := searchStart + idx
			sb.WriteString(text[searchStart:absIdx])
			sb.WriteString(replace)
			searchStart = absIdx + findLen
			count++
		}
		return sb.String(), count
	}

	// Case-insensitive, first only
	idx := strings.Index(lowerText, lowerFind)
	if idx < 0 {
		return text, 0
	}
	return text[:idx] + replace + text[idx+findLen:], 1
}

// applyBlockContent attempts to set content on a block.
// It uses a type switch on the concrete block types that support SetContent.
func applyBlockContent(b block.Block, content string) {
	switch blk := b.(type) {
	case *block.UserMessageBlock:
		blk.SetContent(content)
	case *block.AssistantTextBlock:
		blk.SetContent(content)
	case *block.ThinkingBlock:
		blk.SetContent(content)
	}
	// Other block types (tool calls, etc.) don't support content replacement
}

// CanReplaceBlock reports whether a block supports content replacement.
func CanReplaceBlock(b block.Block) bool {
	switch b.(type) {
	case *block.UserMessageBlock, *block.AssistantTextBlock, *block.ThinkingBlock:
		return true
	}
	return false
}

// ReplaceAll replaces all occurrences of query with replacement across all
// blocks. Returns the total number of replacements made.
// The search is case-insensitive.
func ReplaceAll(blocks []block.Block, query, replacement string) int {
	if query == "" {
		return 0
	}
	total := 0
	for _, b := range blocks {
		text, ok := extractBlockText(b)
		if !ok || text == "" {
			continue
		}
		newText, count := replaceInString(text, query, replacement, false, true)
		if count > 0 {
			applyBlockContent(b, newText)
			total += count
		}
	}
	return total
}

// ReplaceInBlock replaces the next occurrence of query in text starting at
// offset. Returns the new text, the offset past the replacement, and true
// if a replacement was made.
func ReplaceInBlock(text, query, replacement string, offset int) (string, int, bool) {
	if query == "" || offset < 0 || offset >= len(text) {
		return text, offset, false
	}
	lowerText := strings.ToLower(text[offset:])
	lowerQuery := strings.ToLower(query)
	idx := strings.Index(lowerText, lowerQuery)
	if idx < 0 {
		return text, offset, false
	}
	absIdx := offset + idx
	newText := text[:absIdx] + replacement + text[absIdx+len(query):]
	newOffset := absIdx + len(replacement)
	return newText, newOffset, true
}
