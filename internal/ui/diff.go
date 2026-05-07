package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type DiffModel struct {
	viewport viewport.Model
	raw      string
	fileCount int
	width    int
	ready    bool
}

func NewDiffModel() DiffModel {
	return DiffModel{}
}

func (m *DiffModel) SetDiff(content string) {
	m.raw = content
	m.fileCount = strings.Count(content, "diff --git")
	m.renderContent()
}

func (m *DiffModel) SetSize(w, h int) {
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

func (m *DiffModel) renderContent() {
	if m.width == 0 {
		return
	}
	if m.raw == "" {
		if m.ready {
			m.viewport.SetContent(metaStyle.Render("No changes."))
		}
		return
	}

	var b strings.Builder
	for _, line := range strings.Split(m.raw, "\n") {
		switch {
		case strings.HasPrefix(line, "diff --git"):
			b.WriteString(diffFileStyle.Render(line))
		case strings.HasPrefix(line, "index "):
			b.WriteString(metaStyle.Render(line))
		case strings.HasPrefix(line, "--- "):
			b.WriteString(diffFileStyle.Render(line))
		case strings.HasPrefix(line, "+++ "):
			b.WriteString(diffFileStyle.Render(line))
		case strings.HasPrefix(line, "@@"):
			b.WriteString(diffHunkStyle.Render(line))
		case strings.HasPrefix(line, "+"):
			b.WriteString(diffAddStyle.Render(line))
		case strings.HasPrefix(line, "-"):
			b.WriteString(diffDelStyle.Render(line))
		default:
			b.WriteString(line)
		}
		b.WriteByte('\n')
	}
	if m.ready {
		m.viewport.SetContent(b.String())
	}
}

func (m *DiffModel) HasContent() bool { return m.raw != "" }

func (m *DiffModel) GotoTop()    { m.viewport.GotoTop() }
func (m *DiffModel) GotoBottom() { m.viewport.GotoBottom() }

func (m DiffModel) Update(msg tea.Msg) (DiffModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m DiffModel) View() string {
	if !m.ready {
		return "Loading..."
	}
	header := titleStyle.Render(fmt.Sprintf("Files Changed (%d)", m.fileCount))
	return header + "\n" + m.viewport.View()
}
