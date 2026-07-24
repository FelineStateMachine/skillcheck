// Package viewport computes which slice of a long list belongs on screen.
//
// Views own their own chrome, so they subtract the rows they need for headers
// and footers and ask for a window over what is left. Keeping the calculation
// here means every route scrolls the same way and the rule is testable without
// rendering anything.
package viewport

// Window returns the half-open range [start, end) of items to render so that
// cursor stays visible in height rows. The window keeps the cursor at the edge
// it collided with rather than recentring, which makes long presses feel like
// scrolling instead of jumping.
func Window(cursor, total, height int) (int, int) {
	if total <= 0 || height <= 0 {
		return 0, 0
	}
	if height >= total {
		return 0, total
	}
	cursor = min(max(cursor, 0), total-1)
	start := min(max(cursor-height/2, 0), total-height)
	return start, start + height
}

// Scrollbar describes the position of a window within a list, for views that
// want to show that content continues past the edge of the screen.
type Scrollbar struct {
	Above, Below int
}

func Bar(start, end, total int) Scrollbar {
	return Scrollbar{Above: max(start, 0), Below: max(total-end, 0)}
}
