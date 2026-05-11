package ui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type HelpModel struct {
	viewport viewport.Model
	width    int
	ready    bool
}

func NewHelpModel() HelpModel {
	return HelpModel{}
}

func (m *HelpModel) SetSize(w, h int) {
	m.width = w
	bodyH := h - 2
	if bodyH < 1 {
		bodyH = 1
	}
	if !m.ready {
		m.viewport = viewport.New(w, bodyH)
		m.viewport.KeyMap = viewportKeyMap()
		m.ready = true
	} else {
		m.viewport.Width = w
		m.viewport.Height = bodyH
	}
	m.renderContent()
}

func (m *HelpModel) renderContent() {
	if m.width == 0 {
		return
	}

	content := titleStyle.Render("gh-tui Help") + "\n\n"

	content += authorStyle.Render("List View") + "\n"
	content += helpEntry("j/k", "Navigate up/down")
	content += helpEntry("G/gg", "Jump to bottom/top")
	content += helpEntry("Enter", "View PR detail")
	content += helpEntry("o/m/x/a", "Filter: open/merged/closed/all")
	content += helpEntry("u", "Toggle my PRs")
	content += helpEntry("r", "Toggle review requests")
	content += helpEntry("b", "Open PR in browser")
	content += helpEntry("t", "Cycle color theme")
	content += helpEntry("?", "Toggle help")
	content += helpEntry("q", "Quit")
	content += "\n"

	content += authorStyle.Render("Detail View") + "\n"
	content += helpEntry("j/k", "Scroll up/down")
	content += helpEntry("d/u", "Half page down/up")
	content += helpEntry("G/gg", "Jump to bottom/top")
	content += helpEntry("c", "Conversation (comments + reviews)")
	content += helpEntry("s", "Checks / CI status")
	content += helpEntry("f", "Files changed (diff)")
	content += helpEntry("b", "Open in browser")
	content += helpEntry("Esc", "Back to list")
	content += "\n"

	content += authorStyle.Render("Sub Views (conversation/checks/files)") + "\n"
	content += helpEntry("j/k", "Scroll up/down")
	content += helpEntry("d/u", "Half page down/up")
	content += helpEntry("G/gg", "Jump to bottom/top")
	content += helpEntry("c/s/f", "Switch between views")
	content += helpEntry("b", "Open in browser")
	content += helpEntry("Esc", "Back to detail")
	content += "\n"

	content += authorStyle.Render("Theme Picker") + "\n"
	content += helpEntry("j/k", "Navigate themes")
	content += helpEntry("Enter", "Apply theme")
	content += helpEntry("Esc", "Cancel")

	m.viewport.SetContent(content)
}

func helpEntry(key, desc string) string {
	return "  " + headerStyle.Render(key) + "  " + metaStyle.Render(desc) + "\n"
}

func (m *HelpModel) GotoTop()    { m.viewport.GotoTop() }
func (m *HelpModel) GotoBottom() { m.viewport.GotoBottom() }

func (m HelpModel) Update(msg tea.Msg) (HelpModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m HelpModel) View() string {
	if !m.ready {
		return "Loading..."
	}
	return m.viewport.View()
}
