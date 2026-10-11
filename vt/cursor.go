package vt

import uv "github.com/charmbracelet/ultraviolet"

// CursorStyle represents a cursor style.
type CursorStyle int

// Cursor styles. CursorDefault is the shape the user configured in their
// terminal: no DECSCUSR yet, or DECSCUSR 0.
const (
	CursorDefault CursorStyle = iota
	CursorBlock
	CursorUnderline
	CursorBar
)

// Cursor represents a cursor in a terminal.
type Cursor struct {
	Pen  uv.Style
	Link uv.Link

	uv.Position

	Style  CursorStyle
	Steady bool // Not blinking
	Hidden bool
}
