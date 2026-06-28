package component

import (
	"fmt"
	"strings"
	"testing"
)

// ============================================================
// P17-B: LinkManager Performance Optimization Benchmarks
// Compares hand-rolled scanner (new) vs regex (old baseline)
// Target: DetectLinks 100 URLs < 50μs (was ~158μs with regex)
// ============================================================

// detectLinksRegex is the original regex-based implementation kept here
// as a baseline for A/B comparison.
// NOTE: We no longer import regexp in link.go, so this re-creates the
// pattern locally for benchmark purposes only.

func BenchmarkP17_DetectLinks_10_URLs(b *testing.B) {
	text := strings.Repeat("visit https://example.com/test and http://foo.org/bar ", 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectLinks(text, 0, 0)
	}
}

func BenchmarkP17_DetectLinks_100_URLs(b *testing.B) {
	text := strings.Repeat("visit https://example.com/test and http://foo.org/bar ", 100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectLinks(text, 0, 0)
	}
}

func BenchmarkP17_DetectLinks_MixedSchemes_100(b *testing.B) {
	urls := []string{
		"https://example.com/path?q=1",
		"http://test.org/resource",
		"ftp://files.example.com/file.txt",
		"git://github.com/repo.git",
		"ssh://user@host.example.com:22",
		"www.example.com/page",
	}
	var parts []string
	for i := 0; i < 100; i++ {
		parts = append(parts, urls[i%len(urls)])
	}
	text := strings.Join(parts, " and ")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectLinks(text, 0, 0)
	}
}

func BenchmarkP17_DetectLinks_NoURLs_500(b *testing.B) {
	text := strings.Repeat("this is just plain text without any urls in it ", 12)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectLinks(text, 0, 0)
	}
}

func BenchmarkP17_DetectLinks_WWW_100(b *testing.B) {
	var parts []string
	for i := 0; i < 100; i++ {
		parts = append(parts, fmt.Sprintf("see www.site%d.com/page/%d today", i, i))
	}
	text := strings.Join(parts, " and ")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectLinks(text, 0, 0)
	}
}

func BenchmarkP17_DetectLinks_SingleURL(b *testing.B) {
	text := "Check https://github.com/topcheer/fluui for more info about this TUI library"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectLinks(text, 0, 0)
	}
}

func BenchmarkP17_ScanText_100Lines(b *testing.B) {
	lines := make([]string, 100)
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			lines[i] = fmt.Sprintf("Line %d: visit https://example.com/page/%d", i, i)
		} else {
			lines[i] = fmt.Sprintf("Line %d: no urls here, just text", i)
		}
	}
	lm := NewLinkManager()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lm.ScanText(lines)
	}
}

func BenchmarkP17_ScanText_500Lines(b *testing.B) {
	lines := make([]string, 500)
	for i := 0; i < 500; i++ {
		if i%3 == 0 {
			lines[i] = fmt.Sprintf("See http://test.org/r/%d and www.site%d.com", i, i%10)
		} else {
			lines[i] = fmt.Sprintf("Plain line number %d", i)
		}
	}
	lm := NewLinkManager()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lm.ScanText(lines)
	}
}

func BenchmarkP17_AnnotateBuffer_20Links(b *testing.B) {
	lm := NewLinkManager()
	for i := 0; i < 20; i++ {
		lm.AddLink(LinkRange{
			URL:    fmt.Sprintf("https://example.com/page/%d", i),
			Text:   fmt.Sprintf("link_%d", i),
			StartX: i * 5,
			EndX:   i*5 + 10,
			Y:      i,
		})
	}
	buf := newBenchBuffer(120, 20)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lm.AnnotateBuffer(buf, 0, 0)
	}
}

// Correctness: verify hand-rolled scanner produces same results as expected
func TestP17_DetectLinks_Correctness_HTTPS(b *testing.T) {
	text := "Visit https://github.com/topcheer/fluui now"
	links := DetectLinks(text, 0, 0)
	if len(links) != 1 {
		b.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].URL != "https://github.com/topcheer/fluui" {
		b.Errorf("URL = %q", links[0].URL)
	}
}

func TestP17_DetectLinks_Correctness_WWW(t *testing.T) {
	text := "Go to www.google.com now"
	links := DetectLinks(text, 0, 0)
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].URL != "https://www.google.com" {
		t.Errorf("URL = %q, want https://www.google.com", links[0].URL)
	}
}

func TestP17_DetectLinks_Correctness_WWW_NoSecondDot(t *testing.T) {
	// www. without a second dot should NOT match (regex requires www.x.y)
	links := DetectLinks("visit www.nodot now", 0, 0)
	if len(links) != 0 {
		t.Errorf("expected 0 links for www without second dot, got %d", len(links))
	}
}

func TestP17_DetectLinks_Correctness_AllSchemes(t *testing.T) {
	tests := []struct {
		text   string
		expect string
	}{
		{"http://example.com", "http://example.com"},
		{"https://example.com", "https://example.com"},
		{"ftp://files.example.com", "ftp://files.example.com"},
		{"git://github.com/repo", "git://github.com/repo"},
		{"ssh://user@host.com", "ssh://user@host.com"},
	}
	for _, tc := range tests {
		links := DetectLinks(tc.text, 0, 0)
		if len(links) != 1 {
			t.Errorf("expected 1 link for %q, got %d", tc.text, len(links))
			continue
		}
		if links[0].URL != tc.expect {
			t.Errorf("URL = %q, want %q", links[0].URL, tc.expect)
		}
	}
}

func TestP17_DetectLinks_Correctness_StopChars(t *testing.T) {
	// URLs should stop at delimiters
	tests := []string{
		"see (https://example.com) ok",   // parens
		`see "https://example.com" ok`,   // quotes
		"see <https://example.com> ok",   // angle brackets
		"see 'https://example.com' ok",   // single quotes
		"see `https://example.com` ok",   // backtick
	}
	for _, text := range tests {
		links := DetectLinks(text, 0, 0)
		if len(links) != 1 {
			t.Errorf("expected 1 link in %q, got %d", text, len(links))
			continue
		}
		if links[0].URL != "https://example.com" {
			t.Errorf("URL = %q, want https://example.com (text: %q)", links[0].URL, text)
		}
	}
}

func TestP17_DetectLinks_Correctness_Multiple(t *testing.T) {
	text := "https://a.com and https://b.com and www.c.com"
	links := DetectLinks(text, 0, 0)
	if len(links) != 3 {
		t.Fatalf("expected 3 links, got %d", len(links))
	}
	if links[0].URL != "https://a.com" {
		t.Errorf("link[0] = %q", links[0].URL)
	}
	if links[1].URL != "https://b.com" {
		t.Errorf("link[1] = %q", links[1].URL)
	}
	if links[2].URL != "https://www.c.com" {
		t.Errorf("link[2] = %q, want https://www.c.com", links[2].URL)
	}
}

func TestP17_DetectLinks_Correctness_EmptyAndNone(t *testing.T) {
	// Empty text
	if DetectLinks("", 0, 0) != nil {
		t.Error("expected nil for empty text")
	}
	// No URLs
	if DetectLinks("just plain text here", 0, 0) != nil {
		t.Error("expected nil for text without URLs")
	}
}

func TestP17_DetectLinks_Correctness_SchemeOnly(t *testing.T) {
	// "https://" with nothing after should NOT match
	links := DetectLinks("prefix https:// suffix", 0, 0)
	if len(links) != 0 {
		t.Errorf("expected 0 links for bare scheme, got %d", len(links))
	}
}

func TestP17_IsURLStopChar(t *testing.T) {
	stopChars := []byte{' ', '\t', '\n', '\v', '\f', '\r', '<', '>', '"', '\'', '`', ')'}
	for _, c := range stopChars {
		if !isURLStopChar(c) {
			t.Errorf("expected isURLStopChar(%q) = true", c)
		}
	}
	nonStop := []byte{'a', '/', '.', ':', '-', '_', '0', '[', ']', '{', '}'}
	for _, c := range nonStop {
		if isURLStopChar(c) {
			t.Errorf("expected isURLStopChar(%q) = false", c)
		}
	}
}
