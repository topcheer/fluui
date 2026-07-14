// Package compat provides drop-in compatibility layers for migrating
// projects from bubbletea/bubbles/lipgloss/glamour to fluui.
//
// Usage:
//
//	import (
//	    "github.com/topcheer/fluui/compat"
//	    tea "github.com/topcheer/fluui/compat/bubbletea"
//	    "github.com/topcheer/fluui/compat/bubbles/textarea"
//	    "github.com/topcheer/fluui/compat/bubbles/textinput"
//	    "github.com/topcheer/fluui/compat/bubbles/viewport"
//	    lipgloss "github.com/topcheer/fluui/compat/lipgloss"
//	    "github.com/topcheer/fluui/compat/lipgloss/compat"
//	)
//
// Each sub-package mirrors the original API exactly, so source files
// only need their import paths changed — no code modifications required.
package compat