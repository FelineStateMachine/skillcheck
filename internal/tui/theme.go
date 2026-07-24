package tui

import "charm.land/lipgloss/v2"

type Theme struct {
	Title, Heading, Selected, Muted, Good, Warning lipgloss.Style
}

func DefaultTheme(color bool) Theme {
	if !color {
		return Theme{
			Title: lipgloss.NewStyle().Bold(true), Heading: lipgloss.NewStyle().Bold(true),
			Selected: lipgloss.NewStyle().Bold(true), Muted: lipgloss.NewStyle(),
			Good: lipgloss.NewStyle(), Warning: lipgloss.NewStyle().Bold(true),
		}
	}
	ink, paper := lipgloss.Color("#28251f"), lipgloss.Color("#f4efdf")
	return Theme{
		Title:    lipgloss.NewStyle().Bold(true).Foreground(ink).Background(paper),
		Heading:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8a5a13")),
		Selected: lipgloss.NewStyle().Bold(true).Foreground(paper).Background(ink),
		Muted:    lipgloss.NewStyle().Foreground(lipgloss.Color("#706a5d")),
		Good:     lipgloss.NewStyle().Foreground(lipgloss.Color("#39734c")),
		Warning:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#a8640b")),
	}
}
