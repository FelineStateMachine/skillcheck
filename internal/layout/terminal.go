package layout

import "skilltrace/internal/workflow"

const NarrowWidth = 60

func Terminal(graph workflow.Graph, width int) Plan {
	if width < 1 {
		width = 1
	}
	plan := Plan{Width: width, Fallback: width < NarrowWidth}
	positions := map[string]Point{}
	for index, node := range graph.Nodes {
		lane := stageLane(node.Stage)
		x, y := 2+lane*18, 1+index*2
		if plan.Fallback {
			x, y = 2, 1+index*2
		}
		label := node.Action
		if len(label) > 14 {
			label = label[:14]
		}
		plan.Nodes = append(plan.Nodes, Node{ID: node.ID, Label: label, Lane: lane, Bounds: Box{X: x, Y: y, Width: len(label) + 2, Height: 1}})
		positions[node.ID] = Point{X: x, Y: y}
		if y+2 > plan.Height {
			plan.Height = y + 2
		}
	}
	for _, edge := range graph.Edges {
		a, aok := positions[edge.From]
		b, bok := positions[edge.To]
		if aok && bok {
			plan.Routes = append(plan.Routes, Route{From: edge.From, To: edge.To, Points: []Point{a, {X: b.X, Y: a.Y}, b}})
		}
	}
	return plan
}

func stageLane(stage workflow.Stage) int {
	for i, s := range workflow.StageOrder {
		if s == stage {
			return i
		}
	}
	return len(workflow.StageOrder)
}
