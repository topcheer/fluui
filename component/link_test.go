package component

import (
	"strings"
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// --- URL Detection ---

func TestDetectLinks_HTTP(t *testing.T) {
	text := "Check https://example.com for info"
	links := DetectLinks(text, 0, 0)
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].URL != "https://example.com" {
		t.Errorf("URL = %q, want %q", links[0].URL, "https://example.com")
	}
	if links[0].StartX != 6 {
		t.Errorf("StartX = %d, want 6", links[0].StartX)
	}
	if links[0].EndX != 25 {
		t.Errorf("EndX = %d, want 25", links[0].EndX)
	}
}

func TestDetectLinks_HTTPS(t *testing.T) {
	text := "Check https://github.com/topcheer/fluui out"
	links := DetectLinks(text, 0, 0)
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].URL != "https://github.com/topcheer/fluui" {
		t.Errorf("URL = %q", links[0].URL)
	}
}

func TestDetectLinks_FTP(t *testing.T) {
	text := "ftp://files.example.com/path"
	links := DetectLinks(text, 0, 0)
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].URL != "ftp://files.example.com/path" {
		t.Errorf("URL = %q", links[0].URL)
	}
}

func TestDetectLinks_WWW(t *testing.T) {
	text := "Go to www.google.com now"
	links := DetectLinks(text, 0, 0)
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].URL != "https://www.google.com" {
		t.Errorf("URL = %q, want https://www.google.com", links[0].URL)
	}
}

func TestDetectLinks_Multiple(t *testing.T) {
	text := "https://a.com and https://b.com"
	links := DetectLinks(text, 0, 0)
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}
	if links[0].URL != "https://a.com" {
		t.Errorf("link[0].URL = %q", links[0].URL)
	}
	if links[1].URL != "https://b.com" {
		t.Errorf("link[1].URL = %q", links[1].URL)
	}
}

func TestDetectLinks_None(t *testing.T) {
	text := "No URLs here at all"
	links := DetectLinks(text, 0, 0)
	if links != nil {
		t.Errorf("expected nil, got %d links", len(links))
	}
}

func TestDetectLinks_Empty(t *testing.T) {
	links := DetectLinks("", 0, 0)
	if links != nil {
		t.Errorf("expected nil for empty text")
	}
}

func TestDetectLinks_WithLineIdx(t *testing.T) {
	links := DetectLinks("https://x.com", 3, 7)
	if len(links) != 1 {
		t.Fatalf("expected 1 link")
	}
	if links[0].LineIdx != 3 {
		t.Errorf("LineIdx = %d, want 3", links[0].LineIdx)
	}
	if links[0].Y != 7 {
		t.Errorf("Y = %d, want 7", links[0].Y)
	}
}

// --- LinkManager ---

func TestNewLinkManager(t *testing.T) {
	lm := NewLinkManager()
	if lm == nil {
		t.Fatal("NewLinkManager returned nil")
	}
	if lm.LinkCount() != 0 {
		t.Errorf("expected 0 links, got %d", lm.LinkCount())
	}
}

func TestLinkManager_ScanLine(t *testing.T) {
	lm := NewLinkManager()
	lm.ScanLine("Visit https://example.com today", 0, 0)
	if lm.LinkCount() != 1 {
		t.Fatalf("expected 1 link, got %d", lm.LinkCount())
	}
}

func TestLinkManager_ScanText(t *testing.T) {
	lm := NewLinkManager()
	lines := []string{
		"First https://a.com",
		"No URL here",
		"Second https://b.com",
	}
	lm.ScanText(lines)
	if lm.LinkCount() != 2 {
		t.Fatalf("expected 2 links, got %d", lm.LinkCount())
	}
}

func TestLinkManager_LinkAt(t *testing.T) {
	lm := NewLinkManager()
	lm.ScanLine("Go to https://example.com now", 0, 0)
	lr := lm.LinkAt(6, 0)
	if lr == nil {
		t.Fatal("expected link at (6,0)")
	}
	if lr.URL != "https://example.com" {
		t.Errorf("URL = %q", lr.URL)
	}
	lr2 := lm.LinkAt(0, 0)
	if lr2 != nil {
		t.Error("expected nil at (0,0)")
	}
}

func TestLinkManager_ClickLink(t *testing.T) {
	lm := NewLinkManager()
	clicked := ""
	lm.SetOnClick(func(url string) { clicked = url })
	lm.ScanLine("https://example.com", 0, 0)

	if !lm.ClickLink(0, 0) {
		t.Error("ClickLink returned false")
	}
	if clicked != "https://example.com" {
		t.Errorf("clicked = %q", clicked)
	}
}

func TestLinkManager_ClickLink_NoHit(t *testing.T) {
	lm := NewLinkManager()
	lm.SetOnClick(func(url string) { t.Error("should not be called") })
	lm.ScanLine("https://example.com", 0, 0)

	if lm.ClickLink(50, 50) {
		t.Error("ClickLink should return false for miss")
	}
}

func TestLinkManager_Clear(t *testing.T) {
	lm := NewLinkManager()
	lm.ScanLine("https://example.com", 0, 0)
	if lm.LinkCount() != 1 {
		t.Fatal("expected 1 link")
	}
	lm.Clear()
	if lm.LinkCount() != 0 {
		t.Errorf("expected 0 links after clear, got %d", lm.LinkCount())
	}
}

func TestLinkManager_AddLink(t *testing.T) {
	lm := NewLinkManager()
	lm.AddLink(LinkRange{
		URL:    "https://manual.com",
		Text:   "link",
		StartX: 10,
		EndX:   14,
		Y:      5,
	})
	if lm.LinkCount() != 1 {
		t.Fatalf("expected 1 link")
	}
	lr := lm.LinkAt(10, 5)
	if lr == nil {
		t.Error("expected link at (10,5)")
	}
}

func TestLinkManager_AnnotateBuffer(t *testing.T) {
	lm := NewLinkManager()
	lm.ScanLine("Go https://x.com", 0, 0)

	buf := buffer.NewBuffer(30, 3)
	buf.DrawText(0, 0, "Go https://x.com", buffer.DefaultStyle)

	lm.AnnotateBuffer(buf, 0, 0)

	cell := buf.GetCell(3, 0)
	if cell.Link == nil {
		t.Error("expected Link pointer on cell at x=3")
	}
	if cell.Link != nil && cell.Link.URL != "https://x.com" {
		t.Errorf("Link.URL = %q", cell.Link.URL)
	}
	cell0 := buf.GetCell(0, 0)
	if cell0.Link != nil {
		t.Error("expected nil Link on cell at x=0")
	}
}

func TestLinkManager_AnnotateBuffer_Nil(t *testing.T) {
	lm := NewLinkManager()
	lm.AddLink(LinkRange{URL: "https://x.com", StartX: 0, EndX: 5, Y: 0})
	lm.AnnotateBuffer(nil, 0, 0) // should not panic
}

func TestLinkManager_AnnotateBuffer_OutOfBounds(t *testing.T) {
	lm := NewLinkManager()
	lm.AddLink(LinkRange{URL: "https://x.com", StartX: 0, EndX: 100, Y: 0})
	buf := buffer.NewBuffer(10, 2)
	lm.AnnotateBuffer(buf, 0, 0) // should not panic
}

func TestLinkManager_HasLinks(t *testing.T) {
	lm := NewLinkManager()
	if lm.HasLinks() {
		t.Error("HasLinks should be false initially")
	}
	lm.ScanLine("https://example.com", 0, 0)
	if !lm.HasLinks() {
		t.Error("HasLinks should be true after adding a link")
	}
	lm.Clear()
	if lm.HasLinks() {
		t.Error("HasLinks should be false after Clear")
	}
}

func TestLinkManager_FindByURL(t *testing.T) {
	lm := NewLinkManager()
	lm.ScanLine("https://a.com and https://a.com", 0, 0)
	results := lm.FindByURL("https://a.com")
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestLinkManager_Links_ReturnsCopy(t *testing.T) {
	lm := NewLinkManager()
	lm.ScanLine("https://a.com https://b.com", 0, 0)
	links := lm.Links()
	if len(links) != 2 {
		t.Fatalf("expected 2 links")
	}
	links[0].URL = "modified"
	if lm.Links()[0].URL == "modified" {
		t.Error("Links() should return a copy")
	}
}

func TestLinkManager_SetStyle(t *testing.T) {
	lm := NewLinkManager()
	custom := LinkStyle{
		Normal:    buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Underline: buffer.Style{Fg: buffer.RGB(0, 255, 0)},
	}
	lm.SetStyle(custom)
	s := lm.Style()
	if !s.Normal.Fg.Equal(custom.Normal.Fg) {
		t.Error("style mismatch")
	}
}

func TestDefaultLinkStyle(t *testing.T) {
	s := DefaultLinkStyle()
	if !s.Normal.Fg.Equal(buffer.RGB(0x44, 0x8A, 0xFF)) {
		t.Error("expected blue link color")
	}
	if s.Normal.Flags&buffer.Underline == 0 {
		t.Error("expected underline flag")
	}
}

func TestLinkManager_String(t *testing.T) {
	lm := NewLinkManager()
	lm.ScanLine("https://x.com", 0, 0)
	s := lm.String()
	if !strings.Contains(s, "LinkManager") {
		t.Errorf("String should contain 'LinkManager', got %q", s)
	}
}

func TestLinkManager_SetOnClick_Nil(t *testing.T) {
	lm := NewLinkManager()
	lm.AddLink(LinkRange{URL: "https://x.com", StartX: 0, EndX: 6, Y: 0})
	if !lm.ClickLink(2, 0) {
		t.Error("ClickLink should return true even without OnClick callback")
	}
}

// --- Concurrency ---

func TestLinkManager_ConcurrentAccess(t *testing.T) {
	lm := NewLinkManager()
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				lm.ScanLine("https://example.com/"+itoa(j), n, j)
			}
		}(i)
	}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				lm.Links()
				lm.LinkCount()
				lm.HasLinks()
				lm.LinkAt(j%20, 0)
			}
		}()
	}

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				lm.ClickLink(j%20, 0)
			}
		}()
	}

	wg.Wait()
}
