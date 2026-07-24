package layout

type Point struct{ X, Y int }
type Box struct{ X, Y, Width, Height int }

func (b Box) Contains(p Point) bool {
	return p.X >= b.X && p.X < b.X+b.Width && p.Y >= b.Y && p.Y < b.Y+b.Height
}

type Node struct {
	ID, Label string
	Lane      int
	Bounds    Box
}
type Route struct {
	From, To string
	Points   []Point
}
type Plan struct {
	Width, Height int
	Nodes         []Node
	Routes        []Route
	Fallback      bool
}
