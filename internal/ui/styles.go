package ui

import "github.com/charmbracelet/lipgloss"

var (
	subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}

	titleStyle = lipgloss.NewStyle().Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"})

	openStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#238636"))
	mergedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#8957e5"))
	closedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#da3633"))
	draftStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#768390"))

	ciPassStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#238636"))
	ciFailStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#da3633"))
	ciPendStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#d29922"))

	headerStyle = lipgloss.NewStyle().Bold(true).Padding(0, 1)
	helpStyle   = lipgloss.NewStyle().Foreground(subtle)
	statusStyle = lipgloss.NewStyle().Foreground(subtle).Padding(0, 1)

	metaStyle    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#555555", Dark: "#aaaaaa"})
	authorStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"})
	timeStyle    = lipgloss.NewStyle().Foreground(subtle)
	dividerStyle = lipgloss.NewStyle().Foreground(subtle)
)
