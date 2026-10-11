package vt

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// Tests for the policy changes the chainguard branch carries on top of the
// fixes proposed upstream.

func TestUntrackedModesIgnored(t *testing.T) {
	e := NewEmulator(5, 2)
	before := len(e.Modes())
	var b strings.Builder
	b.WriteString("\x1b[?3;2;5000;99999h\x1b[2;12h")
	for n := 10000; n < 12000; n++ {
		fmt.Fprintf(&b, "\x1b[?%dh", n)
	}
	_, _ = e.WriteString(b.String())
	for _, m := range []ansi.Mode{ansi.DECMode(3), ansi.DECMode(2), ansi.ANSIMode(2), ansi.ANSIMode(12)} {
		if got := e.Mode(m); got != ansi.ModeNotRecognized {
			t.Errorf("mode %v: got = %v, want = not recognized", m, got)
		}
	}
	if got := len(e.Modes()); got != before {
		t.Errorf("modes recorded: got = %d, want = %d", got, before)
	}
	_, _ = e.WriteString("\x1b[?2004h\x1b[4h")
	if !e.isModeSet(ansi.ModeBracketedPaste) || !e.isModeSet(ansi.ModeInsertReplace) {
		t.Error("tracked modes ?2004 and 4 were not set")
	}
}

func TestDefaultModesAreTracked(t *testing.T) {
	e := NewEmulator(5, 2)
	for m := range e.Modes() {
		if _, ok := trackedModes[m]; !ok {
			t.Errorf("mode %v has a default but is not tracked, so a program cannot change it", m)
		}
	}
}
