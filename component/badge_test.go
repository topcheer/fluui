package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// === Badge Construction ===

func TestBadgeDefaults(t *testing.T) {
	b := NewBadge("Active", BadgeSuccess)
	if b.Text() != "Active" {
		t.Errorf("Text = %q, want 'Active'", b.Text())
	}
	if b.Variant() != BadgeSuccess {
		t.Errorf("Variant = %v, want BadgeSuccess", b.Variant())
	}
	if b.Size() != BadgeSizeNormal {
		t.Errorf("Size = %v, want BadgeSizeNormal", b.Size())
	}
}

func TestBadgeEmpty(t *testing.T) {
	b := NewBadge("", BadgeNeutral)
	if b.Text() != "" {
		t.Errorf("Text = %q, want empty", b.Text())
	}
}

func TestBadgeWithSize(t *testing.T) {
	b := NewBadgeWithSize("X", BadgeInfo, BadgeSizeSmall)
	if b.Size() != BadgeSizeSmall {
		t.Errorf("Size = %v, want Small", b.Size())
	}
	b2 := NewBadgeWithSize("X", BadgeInfo, BadgeSizeLarge)
	if b2.Size() != BadgeSizeLarge {
		t.Errorf("Size = %v, want Large", b2.Size())
	}
}

func TestBadgeSetText(t *testing.T) {
	b := NewBadge("old", BadgeNeutral)
	b.SetText("new")
	if b.Text() != "new" {
		t.Errorf("Text = %q, want 'new'", b.Text())
	}
}

func TestBadgeSetVariant(t *testing.T) {
	b := NewBadge("X", BadgeInfo)
	b.SetVariant(BadgeError)
	if b.Variant() != BadgeError {
		t.Errorf("Variant = %v, want BadgeError", b.Variant())
	}
}

func TestBadgeSetSize(t *testing.T) {
	b := NewBadge("X", BadgeInfo)
	b.SetSize(BadgeSizeLarge)
	if b.Size() != BadgeSizeLarge {
		t.Errorf("Size = %v, want Large", b.Size())
	}
}

// === Convenience Constructors ===

func TestBadgeConstructors(t *testing.T) {
	tests := []struct {
		name    string
		fn      func(string) *Badge
		variant BadgeVariant
	}{
		{"info", NewInfoBadge, BadgeInfo},
		{"success", NewSuccessBadge, BadgeSuccess},
		{"warning", NewWarningBadge, BadgeWarning},
		{"error", NewErrorBadge, BadgeError},
		{"critical", NewCriticalBadge, BadgeCritical},
		{"neutral", NewNeutralBadge, BadgeNeutral},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := tc.fn("test")
			if b.Text() != "test" {
				t.Errorf("Text = %q", b.Text())
			}
			if b.Variant() != tc.variant {
				t.Errorf("Variant = %v, want %v", b.Variant(), tc.variant)
			}
		})
	}
}

// === Icon ===

func TestBadgeIcon(t *testing.T) {
	b := NewBadge("star", BadgeWarning)
	b.SetIcon("★")
	if b.Icon() != "★" {
		t.Errorf("Icon = %q, want '★'", b.Icon())
	}
}

func TestBadgeNoIconByDefault(t *testing.T) {
	b := NewBadge("X", BadgeInfo)
	if b.Icon() != "" {
		t.Errorf("Icon = %q, want empty", b.Icon())
	}
}

// === Measure ===

func TestBadgeMeasureNormal(t *testing.T) {
	b := NewBadge("OK", BadgeSuccess)
	size := b.Measure(Constraints{})
	// text width 2 + padding 1*2 = 4
	if size.W != 4 {
		t.Errorf("Width = %d, want 4", size.W)
	}
	if size.H != 1 {
		t.Errorf("Height = %d, want 1", size.H)
	}
}

func TestBadgeMeasureSmall(t *testing.T) {
	b := NewBadgeWithSize("OK", BadgeSuccess, BadgeSizeSmall)
	size := b.Measure(Constraints{})
	// text width 2 + padding 0*2 = 2
	if size.W != 2 {
		t.Errorf("Width = %d, want 2", size.W)
	}
}

func TestBadgeMeasureLarge(t *testing.T) {
	b := NewBadgeWithSize("OK", BadgeSuccess, BadgeSizeLarge)
	size := b.Measure(Constraints{})
	// text width 2 + padding 2*2 = 6
	if size.W != 6 {
		t.Errorf("Width = %d, want 6", size.W)
	}
}

func TestBadgeMeasureWithIcon(t *testing.T) {
	b := NewBadge("OK", BadgeSuccess)
	b.SetIcon("●")
	size := b.Measure(Constraints{})
	// icon 1 + space 1 + text 2 + padding 1*2 = 6
	if size.W != 6 {
		t.Errorf("Width = %d, want 6", size.W)
	}
}

func TestBadgeMeasureEmpty(t *testing.T) {
	b := NewBadge("", BadgeNeutral)
	size := b.Measure(Constraints{})
	if size.W < 1 {
		t.Errorf("Width = %d, should be >= 1", size.W)
	}
	if size.H != 1 {
		t.Errorf("Height = %d, want 1", size.H)
	}
}

func TestBadgeMeasureClamped(t *testing.T) {
	b := NewBadge("hello", BadgeInfo)
	size := b.Measure(Constraints{MaxWidth: 3, MaxHeight: 1})
	if size.W > 3 {
		t.Errorf("Width = %d, should be <= 3", size.W)
	}
}

// === Paint ===

func TestBadgePaint(t *testing.T) {
	b := NewBadge("OK", BadgeSuccess)
	b.SetBounds(Rect{X: 0, Y: 0, W: 4, H: 1})
	buf := buffer.NewBuffer(10, 1)
	b.Paint(buf)
	// Verify first cell (padding) is a space
	cell := buf.GetCell(0, 0)
	if cell.Rune != ' ' {
		t.Errorf("Cell(0,0) rune = %q, want space", string(cell.Rune))
	}
	// Second cell should be 'O'
	cell = buf.GetCell(1, 0)
	if cell.Rune != 'O' {
		t.Errorf("Cell(1,0) rune = %q, want 'O'", string(cell.Rune))
	}
}

func TestBadgePaintAllVariants(t *testing.T) {
	variants := []BadgeVariant{
		BadgeInfo, BadgeSuccess, BadgeWarning,
		BadgeError, BadgeCritical, BadgeNeutral,
	}
	for _, v := range variants {
		b := NewBadge("X", v)
		b.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})
		buf := buffer.NewBuffer(10, 1)
		b.Paint(buf) // should not panic
	}
}

func TestBadgePaintWithIcon(t *testing.T) {
	b := NewBadge("OK", BadgeSuccess)
	b.SetIcon("★")
	b.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 1})
	buf := buffer.NewBuffer(10, 1)
	b.Paint(buf)
	// Position 1 should be the icon
	cell := buf.GetCell(1, 0)
	if cell.Rune != '★' {
		t.Errorf("Cell(1,0) = %q, want '★'", string(cell.Rune))
	}
}

func TestBadgePaintSmall(t *testing.T) {
	b := NewBadgeWithSize("X", BadgeInfo, BadgeSizeSmall)
	b.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 1})
	buf := buffer.NewBuffer(5, 1)
	b.Paint(buf)
	cell := buf.GetCell(0, 0)
	if cell.Rune != 'X' {
		t.Errorf("Cell(0,0) = %q, want 'X'", string(cell.Rune))
	}
}

func TestBadgePaintTooSmall(t *testing.T) {
	b := NewBadge("longtext", BadgeError)
	b.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 1})
	buf := buffer.NewBuffer(1, 1)
	b.Paint(buf) // should not panic, truncate
}

func TestBadgePaintFillsBackground(t *testing.T) {
	b := NewBadge("A", BadgeSuccess)
	b.SetBounds(Rect{X: 0, Y: 0, W: 6, H: 1})
	buf := buffer.NewBuffer(6, 1)
	b.Paint(buf)
	// All cells should have non-default background
	for x := 0; x < 6; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Rune == 0 {
			t.Errorf("Cell %d has no rune", x)
		}
	}
}

func TestBadgePaintCustomStyle(t *testing.T) {
	b := NewBadge("X", BadgeInfo)
	custom := &buffer.Style{
		Fg: buffer.RGB(255, 0, 0),
		Bg: buffer.RGB(0, 0, 255),
	}
	b.SetStyle(custom)
	b.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	b.Paint(buf)
	// Verify custom colors are used
	cell := buf.GetCell(1, 0)
	if !cell.Fg.Equal(buffer.RGB(255, 0, 0)) {
		t.Error("Custom Fg not applied")
	}
}

// === Children ===

func TestBadgeChildren(t *testing.T) {
	b := NewBadge("X", BadgeInfo)
	if children := b.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

// === Variant Helpers ===

func TestVariantName(t *testing.T) {
	tests := []struct {
		v      BadgeVariant
		expect string
	}{
		{BadgeInfo, "info"},
		{BadgeSuccess, "success"},
		{BadgeWarning, "warning"},
		{BadgeError, "error"},
		{BadgeCritical, "critical"},
		{BadgeNeutral, "neutral"},
	}
	for _, tc := range tests {
		if got := VariantName(tc.v); got != tc.expect {
			t.Errorf("VariantName(%d) = %q, want %q", tc.v, got, tc.expect)
		}
	}
}

func TestParseVariant(t *testing.T) {
	tests := []struct {
		name   string
		expect BadgeVariant
	}{
		{"info", BadgeInfo},
		{"success", BadgeSuccess},
		{"warning", BadgeWarning},
		{"error", BadgeError},
		{"critical", BadgeCritical},
		{"neutral", BadgeNeutral},
		{"unknown", BadgeNeutral}, // fallback
	}
	for _, tc := range tests {
		if got := ParseVariant(tc.name); got != tc.expect {
			t.Errorf("ParseVariant(%q) = %v, want %v", tc.name, got, tc.expect)
		}
	}
}

func TestVariantRoundTrip(t *testing.T) {
	variants := []BadgeVariant{
		BadgeInfo, BadgeSuccess, BadgeWarning,
		BadgeError, BadgeCritical, BadgeNeutral,
	}
	for _, v := range variants {
		name := VariantName(v)
		parsed := ParseVariant(name)
		if parsed != v {
			t.Errorf("Round trip failed: %v -> %q -> %v", v, name, parsed)
		}
	}
}

func TestSizeName(t *testing.T) {
	if SizeName(BadgeSizeSmall) != "small" {
		t.Error("SizeName(Small) should be 'small'")
	}
	if SizeName(BadgeSizeNormal) != "normal" {
		t.Error("SizeName(Normal) should be 'normal'")
	}
	if SizeName(BadgeSizeLarge) != "large" {
		t.Error("SizeName(Large) should be 'large'")
	}
}

// === BadgeGroup ===

func TestBadgeGroupEmpty(t *testing.T) {
	g := NewBadgeGroup()
	if g.Count() != 0 {
		t.Errorf("Count = %d, want 0", g.Count())
	}
}

func TestBadgeGroupAdd(t *testing.T) {
	g := NewBadgeGroup()
	g.Add(NewInfoBadge("a"))
	g.Add(NewSuccessBadge("b"))
	if g.Count() != 2 {
		t.Errorf("Count = %d, want 2", g.Count())
	}
}

func TestBadgeGroupBadges(t *testing.T) {
	g := NewBadgeGroup()
	b1 := NewInfoBadge("a")
	b2 := NewSuccessBadge("b")
	g.Add(b1)
	g.Add(b2)
	badges := g.Badges()
	if len(badges) != 2 {
		t.Fatalf("len = %d, want 2", len(badges))
	}
	if badges[0] != b1 || badges[1] != b2 {
		t.Error("Badge order mismatch")
	}
}

func TestBadgeGroupClear(t *testing.T) {
	g := NewBadgeGroup()
	g.Add(NewInfoBadge("a"))
	g.Add(NewInfoBadge("b"))
	g.Clear()
	if g.Count() != 0 {
		t.Errorf("Count = %d after Clear, want 0", g.Count())
	}
}

func TestBadgeGroupSetSpacing(t *testing.T) {
	g := NewBadgeGroup()
	g.SetSpacing(3)
	g.Add(NewInfoBadge("x"))
}

func TestBadgeGroupMeasure(t *testing.T) {
	g := NewBadgeGroup()
	g.Add(NewInfoBadge("ab"))  // width: 2 + 2 = 4
	g.Add(NewSuccessBadge("c")) // width: 1 + 2 = 3
	// total: 4 + 1 (default spacing) + 3 = 8
	size := g.Measure(Constraints{})
	if size.W != 8 {
		t.Errorf("Width = %d, want 8", size.W)
	}
	if size.H != 1 {
		t.Errorf("Height = %d, want 1", size.H)
	}
}

func TestBadgeGroupMeasureEmpty(t *testing.T) {
	g := NewBadgeGroup()
	size := g.Measure(Constraints{})
	if size.W != 0 {
		t.Errorf("Width = %d, want 0", size.W)
	}
}

func TestBadgeGroupPaint(t *testing.T) {
	g := NewBadgeGroup()
	g.Add(NewInfoBadge("a"))
	g.Add(NewSuccessBadge("b"))
	g.Add(NewErrorBadge("c"))
	g.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	g.Paint(buf)
}

func TestBadgeGroupPaintWithSpacing(t *testing.T) {
	g := NewBadgeGroup()
	g.SetSpacing(3)
	g.Add(NewInfoBadge("x"))
	g.Add(NewInfoBadge("y"))
	g.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	g.Paint(buf)
}

func TestBadgeGroupChildren(t *testing.T) {
	g := NewBadgeGroup()
	g.Add(NewInfoBadge("a"))
	g.Add(NewSuccessBadge("b"))
	children := g.Children()
	if len(children) != 2 {
		t.Errorf("Children len = %d, want 2", len(children))
	}
}

func TestBadgeGroupChildrenEmpty(t *testing.T) {
	g := NewBadgeGroup()
	children := g.Children()
	if len(children) != 0 {
		t.Errorf("Children len = %d, want 0", len(children))
	}
}

// === Color Resolution ===

func TestBadgeResolveColorsAllVariants(t *testing.T) {
	theme.Get()
	variants := []BadgeVariant{
		BadgeInfo, BadgeSuccess, BadgeWarning,
		BadgeError, BadgeCritical, BadgeNeutral,
	}
	for _, v := range variants {
		b := NewBadge("X", v)
		b.mu.RLock()
		fg, bg := b.resolveColors()
		b.mu.RUnlock()
		_ = fg
		_ = bg
	}
}

func TestBadgeResolveColorsSuccess(t *testing.T) {
	theme.Get()
	b := NewBadge("OK", BadgeSuccess)
	b.mu.RLock()
	fg, bg := b.resolveColors()
	b.mu.RUnlock()
	th := theme.Get()
	if !fg.Equal(th.Bg) {
		t.Error("Success badge fg should be theme.Bg")
	}
	if !bg.Equal(th.Success) {
		t.Error("Success badge bg should be theme.Success")
	}
}

func TestBadgeResolveColorsCustom(t *testing.T) {
	customFg := buffer.RGB(10, 20, 30)
	customBg := buffer.RGB(40, 50, 60)
	b := NewBadge("X", BadgeInfo)
	b.SetStyle(&buffer.Style{Fg: customFg, Bg: customBg})
	b.mu.RLock()
	fg, bg := b.resolveColors()
	b.mu.RUnlock()
	if !fg.Equal(customFg) {
		t.Error("Custom fg not applied")
	}
	if !bg.Equal(customBg) {
		t.Error("Custom bg not applied")
	}
}

func TestBadgeResolveFlagsCritical(t *testing.T) {
	b := NewBadge("CRIT", BadgeCritical)
	b.mu.RLock()
	flags := b.resolveFlags()
	b.mu.RUnlock()
	if flags&buffer.Bold == 0 {
		t.Error("Critical should have Bold flag")
	}
	if flags&buffer.Reverse == 0 {
		t.Error("Critical should have Reverse flag")
	}
}

func TestBadgeResolveFlagsError(t *testing.T) {
	b := NewBadge("ERR", BadgeError)
	b.mu.RLock()
	flags := b.resolveFlags()
	b.mu.RUnlock()
	if flags&buffer.Bold == 0 {
		t.Error("Error should have Bold flag")
	}
}

func TestBadgeResolveFlagsInfo(t *testing.T) {
	b := NewBadge("i", BadgeInfo)
	b.mu.RLock()
	flags := b.resolveFlags()
	b.mu.RUnlock()
	if flags != 0 {
		t.Errorf("Info should have no flags, got %d", flags)
	}
}

// === Padding ===

func TestBadgePaddingSmall(t *testing.T) {
	b := NewBadgeWithSize("x", BadgeInfo, BadgeSizeSmall)
	b.mu.RLock()
	p := b.padding()
	b.mu.RUnlock()
	if p != 0 {
		t.Errorf("Small padding = %d, want 0", p)
	}
}

func TestBadgePaddingNormal(t *testing.T) {
	b := NewBadge("x", BadgeInfo)
	b.mu.RLock()
	p := b.padding()
	b.mu.RUnlock()
	if p != 1 {
		t.Errorf("Normal padding = %d, want 1", p)
	}
}

func TestBadgePaddingLarge(t *testing.T) {
	b := NewBadgeWithSize("x", BadgeInfo, BadgeSizeLarge)
	b.mu.RLock()
	p := b.padding()
	b.mu.RUnlock()
	if p != 2 {
		t.Errorf("Large padding = %d, want 2", p)
	}
}

// === Concurrency ===

func TestBadgeConcurrentAccess(t *testing.T) {
	b := NewBadge("test", BadgeInfo)
	var wg sync.WaitGroup

	// Writers
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			b.SetText("text")
			b.SetVariant(BadgeSuccess)
			b.SetSize(BadgeSizeLarge)
		}
	}()

	// Readers
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			_ = b.Text()
			_ = b.Variant()
			_ = b.Size()
		}
	}()

	wg.Wait()
}

func TestBadgeConcurrentPaint(t *testing.T) {
	b := NewBadge("concurrent", BadgeWarning)
	b.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				buf := buffer.NewBuffer(20, 1)
				b.Paint(buf)
			}
		}()
	}
	wg.Wait()
}

func TestBadgeGroupConcurrentAccess(t *testing.T) {
	g := NewBadgeGroup()
	var wg sync.WaitGroup

	// Writers
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			g.Add(NewInfoBadge("x"))
		}
	}()

	// Readers
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = g.Count()
			_ = g.Badges()
		}
	}()

	wg.Wait()
	if g.Count() != 100 {
		t.Errorf("Count = %d, want 100", g.Count())
	}
}
