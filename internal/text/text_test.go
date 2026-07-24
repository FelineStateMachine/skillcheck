package text

import "testing"

func TestCellAlignsWideRunesByDisplayWidth(t *testing.T) {
	// Each of these occupies a different number of runes but must occupy the
	// same number of columns, or the next column drifts.
	for _, name := range []string{"nzip", "日本語スキル", "🚀-rocket", "ascii-name"} {
		if got := Width(Cell(name, 20)); got != 20 {
			t.Fatalf("Cell(%q, 20) occupies %d columns, want 20", name, got)
		}
	}
}

func TestCellClipsNamesLongerThanTheColumn(t *testing.T) {
	long := "this-is-an-extremely-long-skill-name-that-far-exceeds-the-column"
	got := Cell(long, 20)
	if Width(got) != 20 {
		t.Fatalf("Cell width = %d, want 20", Width(got))
	}
	if got[len(got)-3:] != "..." {
		t.Fatalf("clipped name %q should end in an ellipsis", got)
	}
}

func TestClipNeverExceedsWidth(t *testing.T) {
	for _, s := range []string{"short", "日本語の説明文です", "🚀🚀🚀🚀🚀", "plain ascii sentence"} {
		for width := range 12 {
			if got := Width(Clip(s, width)); got > width {
				t.Fatalf("Clip(%q, %d) occupies %d columns", s, width, got)
			}
		}
	}
}

func TestSanitizeStripsEscapeSequences(t *testing.T) {
	cases := map[string]string{
		"desc with escape: \x1b[31mRED\x1b[0m": "desc with escape: RED",
		"clear\x1b[2J\x1b[Hhome":               "clearhome",
		"tab\there":                            "tab here",
		"line\nbreak":                          "linebreak",
		"bell\a and null\x00":                  "bell and null",
		"plain description":                    "plain description",
		"\x1b[38;5;196mfully styled\x1b[0m":    "fully styled",
	}
	for in, want := range cases {
		if got := Sanitize(in); got != want {
			t.Fatalf("Sanitize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSanitizePreservesUnicodeContent(t *testing.T) {
	for _, s := range []string{"日本語の説明", "emoji 🎉 description", "café naïve"} {
		if got := Sanitize(s); got != s {
			t.Fatalf("Sanitize(%q) = %q, want it unchanged", s, got)
		}
	}
}

func TestPadLeavesOverlongStringsAlone(t *testing.T) {
	if got := Pad("a-very-long-value", 4); got != "a-very-long-value" {
		t.Fatalf("Pad = %q", got)
	}
}
