package vt

import "github.com/charmbracelet/x/ansi"

// resetModes resets all modes to their default values.
func (e *Emulator) resetModes() {
	e.modes = ansi.Modes{
		// Recognized modes and their default values.
		ansi.ModeCursorKeys:          ansi.ModeReset, // ?1
		ansi.ModeOrigin:              ansi.ModeReset, // ?6
		ansi.ModeAutoWrap:            ansi.ModeSet,   // ?7
		ansi.ModeMouseX10:            ansi.ModeReset, // ?9
		ansi.ModeLineFeedNewLine:     ansi.ModeReset, // ?20
		ansi.ModeTextCursorEnable:    ansi.ModeSet,   // ?25
		ansi.ModeNumericKeypad:       ansi.ModeReset, // ?66
		ansi.ModeLeftRightMargin:     ansi.ModeReset, // ?69
		ansi.ModeMouseNormal:         ansi.ModeReset, // ?1000
		ansi.ModeMouseHighlight:      ansi.ModeReset, // ?1001
		ansi.ModeMouseButtonEvent:    ansi.ModeReset, // ?1002
		ansi.ModeMouseAnyEvent:       ansi.ModeReset, // ?1003
		ansi.ModeFocusEvent:          ansi.ModeReset, // ?1004
		ansi.ModeMouseExtSgr:         ansi.ModeReset, // ?1006
		ansi.ModeAltScreen:           ansi.ModeReset, // ?1047
		ansi.ModeSaveCursor:          ansi.ModeReset, // ?1048
		ansi.ModeAltScreenSaveCursor: ansi.ModeReset, // ?1049
		ansi.ModeBracketedPaste:      ansi.ModeReset, // ?2004
		ansi.ModeInsertReplace:       ansi.ModeReset, // 4
	}

	// Set mode effects.
	for mode, setting := range e.modes {
		e.setMode(mode, setting)
	}
}

// trackedModes are the modes the emulator records when a program sets or
// resets them: those it acts on, and those a screen snapshot restores on
// another terminal. Any other mode is ignored, so a program cannot grow the
// mode table without bound, and a mode that is unsafe to replay (DECCOLM
// clears the screen, DECANM leaves ANSI mode, KAM locks the keyboard, SRM
// turns on local echo) never reaches a snapshot.
var trackedModes = map[ansi.Mode]struct{}{
	ansi.ModeInsertReplace:   {}, // 4
	ansi.ModeLineFeedNewLine: {}, // 20
	ansi.DECMode(1):          {}, // DECCKM
	ansi.DECMode(5):          {}, // DECSCNM
	ansi.DECMode(6):          {}, // DECOM
	ansi.DECMode(7):          {}, // DECAWM
	ansi.DECMode(9):          {}, // X10 mouse
	ansi.DECMode(12):         {}, // cursor blink
	ansi.DECMode(25):         {}, // DECTCEM
	ansi.DECMode(47):         {}, // legacy alternate screen
	ansi.DECMode(66):         {}, // DECNKM
	ansi.DECMode(69):         {}, // DECLRMM
	ansi.DECMode(1000):       {}, // mouse
	ansi.DECMode(1001):       {}, // highlight mouse
	ansi.DECMode(1002):       {}, // button-event mouse
	ansi.DECMode(1003):       {}, // any-event mouse
	ansi.DECMode(1004):       {}, // focus events
	ansi.DECMode(1005):       {}, // UTF-8 mouse
	ansi.DECMode(1006):       {}, // SGR mouse
	ansi.DECMode(1007):       {}, // alternate scroll
	ansi.DECMode(1015):       {}, // urxvt mouse
	ansi.DECMode(1016):       {}, // SGR pixel mouse
	ansi.DECMode(1047):       {}, // alternate screen
	ansi.DECMode(1048):       {}, // save cursor
	ansi.DECMode(1049):       {}, // alternate screen with saved cursor
	ansi.DECMode(2004):       {}, // bracketed paste
	ansi.DECMode(2026):       {}, // synchronized output
	ansi.DECMode(2027):       {}, // grapheme clustering
	ansi.DECMode(2031):       {}, // color palette updates
	ansi.DECMode(2048):       {}, // in-band resize
}
