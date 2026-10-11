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

func TestHyperlinkLimits(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantURL    string
		wantParams string
	}{
		{name: "ordinary", input: "\x1b]8;id=1;https://x/a;b\x07", wantURL: "https://x/a;b", wantParams: "id=1"},
		{name: "longest URI kept", input: "\x1b]8;;" + strings.Repeat("u", maxLinkURI) + "\x07", wantURL: strings.Repeat("u", maxLinkURI)},
		{name: "oversized URI dropped", input: "\x1b]8;;https://x\x07\x1b]8;;" + strings.Repeat("u", maxLinkURI+1) + "\x07"},
		{name: "oversized params dropped", input: "\x1b]8;" + strings.Repeat("p", maxLinkParams+1) + ";https://x\x07"},
		{name: "C1 in UTF-8 dropped", input: "\x1b]8;;https://a\u009d52;c;?\u009c\x07", wantURL: "https://a52;c;?"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEmulator(5, 2)
			_, _ = e.WriteString(tc.input)
			l := e.ActiveScreen().Cursor().Link
			if l.URL != tc.wantURL || l.Params != tc.wantParams {
				t.Errorf("link: got = %q params %q, want = %q params %q", l.URL, l.Params, tc.wantURL, tc.wantParams)
			}
		})
	}
}

func TestRepeatIsBounded(t *testing.T) {
	for _, n := range []int{29, 30, 31, 47, 1000, 1001} {
		capped, full := NewEmulator(10, 3), NewEmulator(10, 3)
		_, _ = capped.WriteString(fmt.Sprintf("xa\x1b[%db", n))
		_, _ = full.WriteString("x" + strings.Repeat("a", n+1))
		for y := range 3 {
			if got, want := capped.MainScreen().Line(y).String(), full.MainScreen().Line(y).String(); got != want {
				t.Errorf("n=%d row %d: got = %q, want = %q", n, y, got, want)
			}
		}
		if capped.CursorPosition() != full.CursorPosition() || capped.PendingWrap() != full.PendingWrap() {
			t.Errorf("n=%d cursor: got = %v wrap %v, want = %v wrap %v", n,
				capped.CursorPosition(), capped.PendingWrap(), full.CursorPosition(), full.PendingWrap())
		}
	}
}
