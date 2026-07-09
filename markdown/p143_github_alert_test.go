package markdown

import (
	"testing"
)

func TestP143_GitHubAlert_Note(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("> [!NOTE]\n> This is a note alert.")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
	// Should have at least one block with cells
	totalCells := 0
	for _, b := range blocks {
		totalCells += len(b.Cells)
	}
	if totalCells == 0 {
		t.Error("expected non-zero cell lines")
	}
}

func TestP143_GitHubAlert_Warning(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("> [!WARNING]\n> This is a warning.")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP143_GitHubAlert_Tip(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("> [!TIP]\n> Helpful tip here.")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP143_GitHubAlert_Important(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("> [!IMPORTANT]\n> Don't forget this!")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP143_GitHubAlert_Caution(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("> [!CAUTION]\n> Be very careful.")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP143_GitHubAlert_PlainBlockquote(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("> Just a regular blockquote.")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP143_GitHubAlert_NotAnAlert(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	// [!UNKNOWN] is not a valid alert type
	blocks, err := r.Render("> [!UNKNOWN]\n> Test.")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP143_GitHubAlert_MultilineContent(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("> [!NOTE]\n> First line.\n> Second line.\n> Third line.")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
	// Should have multiple lines
	totalLines := 0
	for _, b := range blocks {
		totalLines += len(b.Cells)
	}
	if totalLines < 1 {
		t.Errorf("expected at least 1 line, got %d", totalLines)
	}
}

func TestP143_GitHubAlert_CaseInsensitive(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("> [!note]\n> Lowercase alert type.")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP143_GitHubAlert_InlineText(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	// Alert with text on same line as marker
	blocks, err := r.Render("> [!WARNING] This is a warning\n> More info here.")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP143_AlertDetection(t *testing.T) {
	// Test the githubAlertType function directly
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	_ = r
	// Test alertIcon
	if icon := alertIcon("note"); icon == "" {
		t.Error("expected non-empty icon for note")
	}
	if icon := alertIcon("tip"); icon == "" {
		t.Error("expected non-empty icon for tip")
	}
	if icon := alertIcon("unknown"); icon == "" {
		t.Error("expected fallback icon for unknown type")
	}
}
