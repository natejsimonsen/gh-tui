package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/natejsimonsen/gh-tui/internal/github"
)

type CommentsModel struct {
	viewport viewport.Model
	comments []github.Comment
	width    int
	ready    bool
}

func NewCommentsModel() CommentsModel {
	return CommentsModel{}
}

func (m *CommentsModel) SetComments(comments []github.Comment) {
	m.comments = comments
	m.renderContent()
}

func (m *CommentsModel) SetSize(w, h int) {
	m.width = w
	bodyH := h - 2
	if bodyH < 1 {
		bodyH = 1
	}
	if !m.ready {
		m.viewport = viewport.New(w, bodyH)
		m.ready = true
	} else {
		m.viewport.Width = w
		m.viewport.Height = bodyH
	}
	m.renderContent()
}

func (m *CommentsModel) renderContent() {
	if m.width == 0 {
		return
	}
	if len(m.comments) == 0 {
		m.viewport.SetContent(metaStyle.Render("No comments."))
		return
	}

	var b strings.Builder
	for i, c := range m.comments {
		b.WriteString(fmt.Sprintf("%s  %s\n",
			authorStyle.Render(c.Author),
			timeStyle.Render(timeAgo(c.CreatedAt)),
		))
		rendered, err := glamour.Render(c.Body, "auto")
		if err != nil {
			b.WriteString(c.Body + "\n")
		} else {
			b.WriteString(rendered)
		}
		if i < len(m.comments)-1 {
			b.WriteString(dividerStyle.Render(strings.Repeat("─", m.width)) + "\n")
		}
	}
	m.viewport.SetContent(b.String())
}

func (m CommentsModel) Update(msg tea.Msg) (CommentsModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m CommentsModel) View() string {
	if !m.ready {
		return "Loading..."
	}
	header := titleStyle.Render(fmt.Sprintf("Comments (%d)", len(m.comments)))
	return header + "\n" + m.viewport.View()
}
