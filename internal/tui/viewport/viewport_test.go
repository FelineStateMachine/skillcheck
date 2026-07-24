package viewport

import "testing"

func TestWindowKeepsCursorVisible(t *testing.T) {
	const total, height = 100, 10
	for cursor := range total {
		start, end := Window(cursor, total, height)
		if cursor < start || cursor >= end {
			t.Fatalf("cursor %d outside window [%d,%d)", cursor, start, end)
		}
		if end-start != height {
			t.Fatalf("window [%d,%d) is %d rows, want %d", start, end, end-start, height)
		}
		if start < 0 || end > total {
			t.Fatalf("window [%d,%d) escapes [0,%d)", start, end, total)
		}
	}
}

func TestWindowAnchorsAtBothEnds(t *testing.T) {
	if start, end := Window(0, 100, 10); start != 0 || end != 10 {
		t.Fatalf("top window = [%d,%d), want [0,10)", start, end)
	}
	if start, end := Window(99, 100, 10); start != 90 || end != 100 {
		t.Fatalf("bottom window = [%d,%d), want [90,100)", start, end)
	}
}

func TestWindowShowsEverythingWhenItFits(t *testing.T) {
	if start, end := Window(3, 5, 40); start != 0 || end != 5 {
		t.Fatalf("window = [%d,%d), want the whole list [0,5)", start, end)
	}
}

func TestWindowHandlesDegenerateInput(t *testing.T) {
	for _, tc := range []struct{ cursor, total, height int }{
		{0, 0, 10}, {0, 10, 0}, {5, 0, 0}, {-3, 10, 4}, {99, 10, 4}, {0, 10, -1},
	} {
		start, end := Window(tc.cursor, tc.total, tc.height)
		if start < 0 || end < start || end > max(tc.total, 0) {
			t.Fatalf("Window(%d,%d,%d) = [%d,%d)", tc.cursor, tc.total, tc.height, start, end)
		}
	}
}

func TestBarCountsHiddenRows(t *testing.T) {
	if bar := Bar(10, 20, 100); bar.Above != 10 || bar.Below != 80 {
		t.Fatalf("bar = %#v", bar)
	}
	if bar := Bar(0, 5, 5); bar.Above != 0 || bar.Below != 0 {
		t.Fatalf("bar = %#v", bar)
	}
}
