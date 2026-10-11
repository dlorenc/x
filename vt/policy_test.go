package vt

import (
	"fmt"
	"strings"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
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
		{name: "DEL dropped", input: "\x1b]8;;https://a\x7fb\x07", wantURL: "https://ab"},
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

func TestSanitizeOSC(t *testing.T) {
	for in, want := range map[string]string{
		"https://example.com/a;b": "https://example.com/a;b",
		"a\x1b]52;c;?\x07b":       "a]52;c;?b",
		"a\u009d52;c;?\u009cb":    "a52;c;?b",
		"a\x7fb":                  "ab",
		"a\xffb":                  "a\uFFFDb",
		"日本\u200d":                "日本\u200d",
	} {
		if got := SanitizeOSC(in); got != want {
			t.Errorf("SanitizeOSC(%q): got = %q, want = %q", in, got, want)
		}
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

func TestCursorVisibilityIsGlobal(t *testing.T) {
	e := NewEmulator(10, 3)
	_, _ = e.WriteString("\x1b[?1049h\x1b[?25l\x1b[4 q\x1b[?1049l")
	if c := e.MainScreen().Cursor(); !c.Hidden || c.Style != CursorUnderline || !c.Steady {
		t.Errorf("main screen cursor after leaving the alternate screen: got = %+v, want hidden steady underline", c)
	}
}

func TestCursorStyleDefault(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantStyle  CursorStyle
		wantSteady bool
	}{
		{name: "never set", input: "", wantStyle: CursorDefault},
		{name: "DECSCUSR 0 after a shape", input: "\x1b[4 q\x1b[0 q", wantStyle: CursorDefault},
		{name: "blinking block", input: "\x1b[1 q", wantStyle: CursorBlock},
		{name: "steady block", input: "\x1b[2 q", wantStyle: CursorBlock, wantSteady: true},
		{name: "blinking underline", input: "\x1b[3 q", wantStyle: CursorUnderline},
		{name: "steady bar", input: "\x1b[6 q", wantStyle: CursorBar, wantSteady: true},
		{name: "out of range is ignored", input: "\x1b[6 q\x1b[7 q", wantStyle: CursorBar, wantSteady: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEmulator(5, 2)
			_, _ = e.WriteString(tc.input)
			c := e.ActiveScreen().Cursor()
			if c.Style != tc.wantStyle || c.Steady != tc.wantSteady {
				t.Errorf("cursor: got = style %v steady %v, want = style %v steady %v", c.Style, c.Steady, tc.wantStyle, tc.wantSteady)
			}
		})
	}
}

func TestScrollOffOnlyFromTheTopRow(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "full screen", input: "\x1b[5;1Ha\nb\nc\n", want: 3},
		{name: "region from row 1", input: "\x1b[1;5r\x1b[5;1Ha\nb\nc\n", want: 3},
		{name: "region below a header", input: "\x1b[3;5r\x1b[5;1Ha\nb\nc\n", want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := NewEmulator(5, 5)
			var got int
			e.SetScrollOffFunc(func(uv.Line) { got++ })
			_, _ = e.WriteString(tc.input)
			if got != tc.want {
				t.Errorf("lines scrolled off: got = %d, want = %d", got, tc.want)
			}
		})
	}
}

func TestScrollbackOffByDefault(t *testing.T) {
	e := NewEmulator(10, 2)
	_, _ = e.WriteString("a\r\nb\r\nc\r\nd")
	if n := e.ScrollbackLen(); n != 0 {
		t.Errorf("ScrollbackLen: got = %d, want = 0", n)
	}
	if e.Scrollback() != nil {
		t.Error("Scrollback: got a buffer, want none until asked for")
	}
}

func TestRepliesDroppedByDefault(t *testing.T) {
	e := NewEmulator(10, 2)
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = e.WriteString("\x1b[5n\x1b[c\x1b[6n")
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Write blocked on a query reply by default")
	}
}
