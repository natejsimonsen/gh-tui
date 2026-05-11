package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    lipgloss.Style
	openStyle     lipgloss.Style
	mergedStyle   lipgloss.Style
	closedStyle   lipgloss.Style
	draftStyle    lipgloss.Style
	ciPassStyle   lipgloss.Style
	ciFailStyle   lipgloss.Style
	ciPendStyle   lipgloss.Style
	headerStyle   lipgloss.Style
	helpStyle     lipgloss.Style
	statusStyle   lipgloss.Style
	metaStyle     lipgloss.Style
	authorStyle   lipgloss.Style
	timeStyle     lipgloss.Style
	dividerStyle  lipgloss.Style
	diffAddStyle        lipgloss.Style
	diffDelStyle        lipgloss.Style
	diffHunkStyle       lipgloss.Style
	diffFileStyle       lipgloss.Style
	diffFileBannerStyle lipgloss.Style
	diffGutterStyle     lipgloss.Style
	diffCtxStyle        lipgloss.Style
)

func init() {
	applyTheme(themes[0])
}

func viewportKeyMap() viewport.KeyMap {
	km := viewport.DefaultKeyMap()
	km.PageDown = key.NewBinding(key.WithKeys("pgdown", " ", "ctrl+f"))
	km.PageUp = key.NewBinding(key.WithKeys("pgup", "ctrl+b"))
	return km
}
