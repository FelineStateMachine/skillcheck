package grid

import (
	"github.com/mattn/go-runewidth"
	"strings"
)

type Grid struct {
	width, height int
	cells         [][]rune
}

func New(width, height int) *Grid {
	g := &Grid{width: width, height: height, cells: make([][]rune, height)}
	for y := range g.cells {
		g.cells[y] = []rune(strings.Repeat(" ", width))
	}
	return g
}
func (g *Grid) Set(x, y int, r rune) {
	if y < 0 || y >= g.height || x < 0 || x >= g.width {
		return
	}
	w := runewidth.RuneWidth(r)
	if w < 1 {
		return
	}
	g.cells[y][x] = r
	if w == 2 && x+1 < g.width {
		g.cells[y][x+1] = 0
	}
}
func (g *Grid) Text(x, y int, text string) {
	for _, r := range text {
		g.Set(x, y, r)
		x += runewidth.RuneWidth(r)
		if x >= g.width {
			return
		}
	}
}
func (g *Grid) String() string {
	lines := make([]string, g.height)
	for y, row := range g.cells {
		var b strings.Builder
		for _, r := range row {
			if r != 0 {
				b.WriteRune(r)
			}
		}
		lines[y] = strings.TrimRight(b.String(), " ")
	}
	return strings.Join(lines, "\n")
}
