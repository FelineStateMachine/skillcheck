// Package text lays out terminal strings by display width rather than by rune
// count, and strips terminal control sequences out of text the tool did not
// author.
//
// Skill names and descriptions come from third-party SKILL.md files, so they
// carry CJK and emoji that occupy two columns each, names far longer than any
// column, and occasionally raw escape sequences. Formatting those with %-20s
// pads by rune count and shifts every later column out of alignment.
package text

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// Width reports how many terminal columns s occupies.
func Width(s string) int { return runewidth.StringWidth(s) }

// Clip truncates s to at most width columns, marking the truncation with an
// ellipsis when there is room for one.
func Clip(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= width {
		return s
	}
	if width <= 3 {
		return runewidth.Truncate(s, width, "")
	}
	return runewidth.Truncate(s, width, "...")
}

// Pad right-pads s with spaces to width columns, leaving it untouched when it
// is already at least that wide.
func Pad(s string, width int) string {
	gap := width - runewidth.StringWidth(s)
	if gap <= 0 {
		return s
	}
	return s + strings.Repeat(" ", gap)
}

// Cell renders s as a fixed-width column: clipped when too wide, padded when
// too narrow. Use it for anything that has to line up with the row above.
func Cell(s string, width int) string { return Pad(Clip(s, width), width) }

// Sanitize removes control characters and escape sequences so that untrusted
// text cannot restyle or reposition the terminal. Tabs become single spaces
// because their width depends on cursor position, which breaks column
// alignment; other C0 and C1 controls are dropped entirely.
func Sanitize(s string) string {
	if !needsSanitizing(s) {
		return s
	}
	const (
		normal = iota
		afterEscape
		inCSI
		inString
	)
	var b strings.Builder
	b.Grow(len(s))
	state := normal
	for _, r := range s {
		switch state {
		case afterEscape:
			// ESC [ opens a CSI sequence and ESC ] / P / X / ^ / _ open string
			// sequences; anything else is a complete two-character escape.
			switch r {
			case '[':
				state = inCSI
			case ']', 'P', 'X', '^', '_':
				state = inString
			default:
				state = normal
			}
		case inCSI:
			// Parameter and intermediate bytes precede a final byte in 0x40-0x7e.
			if r >= 0x40 && r <= 0x7e {
				state = normal
			}
		case inString:
			// String sequences run until BEL or a String Terminator.
			if r == 0x07 || r == 0x9c {
				state = normal
			}
		default:
			switch {
			case r == 0x1b:
				state = afterEscape
			case r == '\t':
				b.WriteRune(' ')
			case r < 0x20, r == 0x7f, r >= 0x80 && r <= 0x9f:
				// Newlines included: a description is a single line of chrome.
			default:
				b.WriteRune(r)
			}
		}
	}
	return strings.TrimSpace(b.String())
}

func needsSanitizing(s string) bool {
	for _, r := range s {
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) {
			return true
		}
	}
	return false
}
