package compat_test

import (
	"testing"

	tea "github.com/topcheer/fluui/compat/bubbletea"
	"github.com/topcheer/fluui/compat/bubbles/textarea"
	"github.com/topcheer/fluui/compat/bubbles/textinput"
	"github.com/topcheer/fluui/compat/bubbles/viewport"
	"github.com/topcheer/fluui/compat/glamour"
	"github.com/topcheer/fluui/compat/glamour/ansi"
	"github.com/topcheer/fluui/compat/glamour/styles"
	lg "github.com/topcheer/fluui/compat/lipgloss"
	"github.com/topcheer/fluui/compat/lipgloss/tree"
)

// TestGgcodeImportPattern reproduces ggcode's exact import + usage patterns.
// This is a COMPILE-TIME + RUNTIME verification that fluui compat is a drop-in.
//
// Migration note: lipgloss.Color now works natively (type Color string)("14") in ggcode → lipgloss.NewColor("14") in fluui
// (In charm.land, Color is type string; in fluui it's a struct for type safety)
func TestGgcodeImportPattern(t *testing.T) {
	// ── bubbletea ──
	var _ tea.Model = (*myModel)(nil)
	m := &myModel{}
	p := tea.NewProgram(m, tea.WithoutSignals(), tea.WithoutRenderer())
	_ = p
	v := tea.NewView("test")
	if v.Content != "test" {
		t.Error("View.Content mismatch")
	}

	// Keys
	_ = tea.KeyEnter
	_ = tea.KeyEsc
	_ = tea.KeyUp
	_ = tea.KeyDown
	_ = tea.ModAlt
	_ = tea.ErrInterrupted

	// ── lipgloss ── (Note: Color("14") → NewColor("14") in migration)
	s := lg.NewStyle().
		Foreground(lg.Color("14")).
		Background(lg.Color("#1a1a2e")).
		Bold(true).
		Italic(true).
		Underline(true).
		Faint(true).
		Width(80).
		Height(24).
		Padding(1, 2).
		Margin(0, 1).
		Border(lg.RoundedBorder())

	result := s.Render("hello")
	if result == "" {
		t.Error("Render should not be empty")
	}

	// Layout
	_ = lg.JoinHorizontal(lg.Top, "a", "b", "c")
	_ = lg.JoinVertical(lg.Left, "x", "y")
	_ = lg.PlaceHorizontal(40, lg.Center, "centered")
	if lg.Width("test") <= 0 || lg.Height("l1\nl2") != 2 {
		t.Error("Width/Height mismatch")
	}

	// AdaptiveColor
	_ = lg.AdaptiveColor{Light: "240", Dark: "250"}

	// ── textinput ──
	ti := textinput.New()
	ti.SetValue("hello")
	ti.SetPlaceholder("type here...")
	ti.SetPrompt("> ")
	ti.Focus()
	ti.Blur()
	_, _ = ti.Update(tea.KeyPressMsg{Rune: 'a'})
	_ = ti.View()
	ti.CursorEnd()
	ti.CursorStart()
	ti.InsertRune('\n')
	ti.SetEchoMode(textinput.EchoPassword)
	_ = textinput.Blink

	// ── textarea ──
	ta := textarea.New()
	ta.SetValue("line1\nline2")
	ta.SetWidth(60)
	ta.SetHeight(5)
	ta.Focus()
	_, _ = ta.Update(tea.KeyPressMsg{Rune: 'b'})
	_ = ta.View()
	ta.CursorEnd()
	ta.InsertRune('\n')
	taStyles := textarea.DefaultStyles(true)
	taStyles.Focused.Base = lg.NewStyle()
	taStyles.Focused.Text = lg.NewStyle().Bold(true)
	taStyles.Focused.Prompt = lg.NewStyle()
	taStyles.Focused.Placeholder = lg.NewStyle()
	taStyles.Blurred.Base = lg.NewStyle()
	taStyles.Blurred.Text = lg.NewStyle()
	ta.SetStyles(taStyles)
	_ = textarea.Blink()

	// ── viewport ──
	vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(24))
	vp.SetContent("line1\nline2\nline3")
	vp.SetSize(80, 10)
	vp.ScrollDown(3)
	vp.ScrollUp(1)
	vp.GotoBottom()
	_ = vp.View()
	_ = vp.YOffset()
	_ = vp.TotalLineCount()
	_ = vp.AtBottom()

	// ── glamour ──
	r, err := glamour.NewTermRenderer(glamour.WithStyles(styles.DarkStyleConfig), glamour.WithWordWrap(80))
	if err != nil {
		t.Fatalf("NewTermRenderer error: %v", err)
	}
	out, err := r.Render("# Hello\n\nworld")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	_ = out

	// ansi
	accent := "#7aa2f7"
	sc := ansi.StyleConfig{}
	sc.Code.Color = &accent
	_ = sc.Heading[0]

	// tree
	tr := tree.Root("root").Child("item1").Child("item2")
	_ = tr.String()
}

type myModel struct{}

func (m *myModel) Init() tea.Cmd                            { return nil }
func (m *myModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)  { return m, nil }
func (m *myModel) View() tea.View                           { return tea.NewView("") }
