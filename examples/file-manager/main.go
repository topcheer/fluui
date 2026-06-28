// Package main implements a dual-panel file manager using Fluui.
//
// This example showcases:
//   - FilePicker for browsing directories with fuzzy filtering
//   - TabBar for multi-tab directory navigation
//   - Tree for hierarchical file view
//   - StatusBar with current path and selection count
//   - Full keyboard navigation (vim-style + arrows)
//
// Keys:
//   j/k, Up/Down  — navigate files
//   Enter         — open directory / select file
//   Backspace, h  — go to parent directory
//   Space         — toggle file selection
//   Tab           — switch between panels (tree/list)
//   q/Esc         — quit
//   Ctrl+C        — quit
package main

import (
	"fmt"
	"os"
	"path/filepath"

	fluui "github.com/topcheer/fluui"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func main() {
	base, err := fluui.New()
	if err != nil {
		panic(err)
	}
	defer base.Close()

	cwd, _ := os.Getwd()

	// --- Components ---

	// File picker for browsing
	picker := component.NewFilePicker(cwd)

	// Tree for directory structure
	tree := buildDirectoryTree(cwd)

	// Tab bar
	tabBar := component.NewTabBar()
	tabBar.AddTab("tab-1", filepath.Base(cwd))
	tabBar.AddTab("tab-2", "Tree")
	tabBar.SetActive(0)

	// Status bar
	statusBar := component.NewStatusBar()
	statusBar.AddLeft("app", " Fluui File Manager")
	statusBar.AddCenter("path", " "+truncate(cwd, 50))
	statusBar.AddRight("sel", " 0 selected ")
	statusBar.AddRight("hint", " [q]uit [Tab]panel ")

	// --- State ---
	activePanel := 0 // 0=picker, 1=tree
	currentDir := cwd

	// --- Key handling ---
	base.OnKey(func(k *term.KeyEvent) {
		switch {
		case k.Key == term.KeyEscape || k.Rune == 'q':
			base.Quit()

		case k.Rune == 'c' && k.Modifiers&term.ModCtrl != 0:
			base.Quit()

		case k.Key == term.KeyTab:
			activePanel = (activePanel + 1) % 2
			tabBar.SetActive(activePanel)

		default:
			if activePanel == 0 {
				picker.HandleKey(k)
			} else {
				tree.HandleKey(k)
			}
		}

		// Update status bar
		statusBar.SetItemText("path", " "+truncate(currentDir, 50))
		statusBar.SetItemText("sel", fmt.Sprintf(" %d selected ", len(picker.SelectedFiles())))
		base.MarkDirty()
	})

	// --- Rendering ---
	base.OnPaint(func(buf *buffer.Buffer) {
		w, h := base.Size()

		// Layout: TabBar (1) + content (h-2) + StatusBar (1)
		tabBar.SetBounds(component.Rect{X: 0, Y: 0, W: w, H: 1})
		tabBar.Paint(buf)

		statusBar.SetBounds(component.Rect{X: 0, Y: h - 1, W: w, H: 1})
		statusBar.Paint(buf)

		contentY := 1
		contentH := h - 2

		if w >= 80 {
			// Dual-panel: picker (left) + tree (right)
			splitW := w / 2

			// Left panel: file picker
			drawPanelBorder(buf, 0, contentY, splitW, contentH, "Browser", activePanel == 0)
			picker.SetBounds(component.Rect{X: 1, Y: contentY + 1, W: splitW - 2, H: contentH - 2})
			picker.Paint(buf)

			// Right panel: directory tree
			drawPanelBorder(buf, splitW, contentY, w-splitW, contentH, "Tree", activePanel == 1)
			tree.SetBounds(component.Rect{X: splitW + 1, Y: contentY + 1, W: w - splitW - 2, H: contentH - 2})
			tree.Paint(buf)
		} else {
			// Single panel based on active tab
			if activePanel == 0 {
				drawPanelBorder(buf, 0, contentY, w, contentH, "Browser", true)
				picker.SetBounds(component.Rect{X: 1, Y: contentY + 1, W: w - 2, H: contentH - 2})
				picker.Paint(buf)
			} else {
				drawPanelBorder(buf, 0, contentY, w, contentH, "Tree", true)
				tree.SetBounds(component.Rect{X: 1, Y: contentY + 1, W: w - 2, H: contentH - 2})
				tree.Paint(buf)
			}
		}

		drawHLine(buf, h-2, w)
	})

	base.Run()
}

func drawPanelBorder(buf *buffer.Buffer, x, y, w, h int, title string, active bool) {
	var style buffer.Style
	if active {
		style = buffer.Style{Fg: buffer.RGB(0x7d, 0xd3, 0xfc), Flags: buffer.Bold}
	} else {
		style = buffer.Style{Fg: buffer.RGB(0x55, 0x55, 0x55)}
	}

	// Top border with title
	buf.SetCell(x, y, buffer.NewCell('┌', style))
	for i := 1; i < w-1; i++ {
		buf.SetCell(x+i, y, buffer.NewCell('─', style))
	}
	buf.SetCell(x+w-1, y, buffer.NewCell('┐', style))

	// Title
	titleText := " " + title + " "
	for i, r := range titleText {
		if x+2+i < x+w-1 {
			buf.SetCell(x+2+i, y, buffer.NewCell(r, style))
		}
	}

	// Side borders
	for i := 1; i < h-1; i++ {
		buf.SetCell(x, y+i, buffer.NewCell('│', style))
		buf.SetCell(x+w-1, y+i, buffer.NewCell('│', style))
	}

	// Bottom border
	buf.SetCell(x, y+h-1, buffer.NewCell('└', style))
	for i := 1; i < w-1; i++ {
		buf.SetCell(x+i, y+h-1, buffer.NewCell('─', style))
	}
	buf.SetCell(x+w-1, y+h-1, buffer.NewCell('┘', style))
}

func drawHLine(buf *buffer.Buffer, y, w int) {
	sepStyle := buffer.Style{Fg: buffer.RGB(0x55, 0x55, 0x55)}
	for x := 0; x < w; x++ {
		buf.SetCell(x, y, buffer.NewCell('─', sepStyle))
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return "..." + s[len(s)-maxLen+3:]
}

func buildDirectoryTree(dir string) *component.Tree {
	root := component.NewTreeNode(dir, filepath.Base(dir))

	entries, err := os.ReadDir(dir)
	if err != nil {
		return component.NewTree()
	}

	for _, entry := range entries {
		if entry.IsDir() && len(entry.Name()) > 0 && entry.Name()[0] == '.' {
			continue
		}
		child := component.NewTreeNode(
			filepath.Join(dir, entry.Name()),
			entry.Name(),
		)
		root.AddChild(child)

		if entry.IsDir() {
			// Add first level of subdirectories
			subPath := filepath.Join(dir, entry.Name())
			subEntries, err := os.ReadDir(subPath)
			if err == nil {
				count := 0
				for _, sub := range subEntries {
					if count >= 5 {
						break
					}
					if sub.IsDir() && len(sub.Name()) > 0 && sub.Name()[0] != '.' {
						grandchild := component.NewTreeNode(
							filepath.Join(subPath, sub.Name()),
							sub.Name(),
						)
						child.AddChild(grandchild)
						count++
					}
				}
			}
		}
	}

	tree := component.NewTree()
	tree.SetRoot(root)
	return tree
}
