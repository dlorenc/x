// Package vt provides a virtual terminal implementation.
// SKIP: Fix typecheck errors - function signature mismatches and undefined types
package vt

import (
	"bytes"
	"image/color"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// handleOsc handles an OSC escape sequence.
func (e *Emulator) handleOsc(cmd int, data []byte) {
	e.flushGrapheme() // Flush any pending grapheme before handling OSC sequences.
	if !e.handlers.handleOsc(cmd, data) {
		e.logf("unhandled sequence: OSC %q", data)
	}
}

func (e *Emulator) handleTitle(cmd int, data []byte) {
	// A title may itself contain ';'.
	_, title, ok := bytes.Cut(data, []byte{';'})
	if !ok {
		// Invalid, ignore
		return
	}
	switch cmd {
	case 0: // Set window title and icon name
		name := string(title)
		e.iconName, e.title = name, name
		if e.cb.Title != nil {
			e.cb.Title(name)
		}
		if e.cb.IconName != nil {
			e.cb.IconName(name)
		}
	case 1: // Set icon name
		name := string(title)
		e.iconName = name
		if e.cb.IconName != nil {
			e.cb.IconName(name)
		}
	case 2: // Set window title
		name := string(title)
		e.title = name
		if e.cb.Title != nil {
			e.cb.Title(name)
		}
	}
}

func (e *Emulator) handleDefaultColor(cmd int, data []byte) {
	if cmd != 10 && cmd != 11 && cmd != 12 &&
		cmd != 110 && cmd != 111 && cmd != 112 {
		// Invalid, ignore
		return
	}

	parts := bytes.Split(data, []byte{';'})
	if len(parts) == 0 {
		// Invalid, ignore
		return
	}

	cb := func(c color.Color) {
		switch cmd {
		case 10, 110: // Foreground color
			e.SetForegroundColor(c)
		case 11, 111: // Background color
			e.SetBackgroundColor(c)
		case 12, 112: // Cursor color
			e.SetCursorColor(c)
		}
	}

	switch len(parts) {
	case 1: // Reset color
		cb(nil)
	case 2: // Set/Query color
		arg := string(parts[1])
		if arg == "?" {
			var xrgb ansi.XRGBColor
			switch cmd {
			case 10: // Query foreground color
				xrgb.Color = e.ForegroundColor()
				if xrgb.Color != nil {
					e.reply(ansi.SetForegroundColor(xrgb.String()))
				}
			case 11: // Query background color
				xrgb.Color = e.BackgroundColor()
				if xrgb.Color != nil {
					e.reply(ansi.SetBackgroundColor(xrgb.String()))
				}
			case 12: // Query cursor color
				xrgb.Color = e.CursorColor()
				if xrgb.Color != nil {
					e.reply(ansi.SetCursorColor(xrgb.String()))
				}
			}
		} else if c := ansi.XParseColor(arg); c != nil {
			cb(c)
		}
	}
}

func (e *Emulator) handleWorkingDirectory(cmd int, data []byte) {
	if cmd != 7 {
		// Invalid, ignore
		return
	}

	// The data is the working directory path, which may contain ';'.
	_, dir, ok := bytes.Cut(data, []byte{';'})
	if !ok {
		// Invalid, ignore
		return
	}

	path := string(dir)
	e.cwd = path

	if e.cb.WorkingDirectory != nil {
		e.cb.WorkingDirectory(path)
	}
}

func (e *Emulator) handleHyperlink(cmd int, data []byte) {
	// The URI may itself contain ';'.
	parts := bytes.SplitN(data, []byte{';'}, 3)
	if len(parts) != 3 || cmd != 8 {
		// Invalid, ignore
		return
	}

	params, uri := SanitizeOSC(string(parts[1])), SanitizeOSC(string(parts[2]))
	if len(uri) > maxLinkURI || len(params) > maxLinkParams {
		// Every cell the link covers holds it, and a repaint writes it out
		// again wherever it starts, so an oversized link is dropped, as VTE
		// does, rather than multiplied.
		e.scr.cur.Link = uv.Link{}
		return
	}
	e.scr.cur.Link.Params = params
	e.scr.cur.Link.URL = uri
}

// The limits on an OSC 8 hyperlink, VTE's.
const (
	maxLinkURI    = 2083
	maxLinkParams = 250
)

// SanitizeOSC returns s as valid UTF-8 with the control characters that
// would end or corrupt an OSC string dropped: C0, DEL and C1, which a
// terminal may also read in their UTF-8 form.
func SanitizeOSC(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r >= 0x7f && r <= 0x9f {
			return -1
		}
		return r
	}, strings.ToValidUTF8(s, "\uFFFD"))
}
