package layout

import "skilltrace/internal/workflow"

type ComparisonPlan struct {
	Left, Right Plan
	Anchors     []ComparisonAnchor
}
type ComparisonAnchor struct {
	DifferenceID              string
	Left, Right               Point
	LeftPresent, RightPresent bool
}

func Compare(left, right workflow.Graph, differences []workflow.Difference, width int) ComparisonPlan {
	half := width/2 - 1
	if half < 20 {
		half = width
	}
	result := ComparisonPlan{Left: Terminal(left, half), Right: Terminal(right, half)}
	for _, d := range differences {
		a := ComparisonAnchor{DifferenceID: d.ID}
		for _, n := range result.Left.Nodes {
			if n.ID == d.ID {
				a.Left, a.LeftPresent = Point{X: n.Bounds.X, Y: n.Bounds.Y}, true
			}
		}
		for _, n := range result.Right.Nodes {
			if n.ID == d.ID {
				a.Right, a.RightPresent = Point{X: n.Bounds.X, Y: n.Bounds.Y}, true
			}
		}
		result.Anchors = append(result.Anchors, a)
	}
	return result
}
