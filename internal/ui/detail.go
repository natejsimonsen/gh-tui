package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/natejsimonsen/gh-tui/internal/github"
)

type DetailModel struct {
	viewport viewport.Model
	pr       *github.PRDetail
	width    int
	ready    bool
}

func NewDetailModel() DetailModel {
	return DetailModel{}
}

func (m *DetailModel) SetPR(pr *github.PRDetail) {
	m.pr = pr
	m.renderContent()
}

func (m *DetailModel) SetSize(w, h int) {
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

func (m *DetailModel) renderContent() {
	if !m.ready || m.pr == nil {
		return
	}
	pr := m.pr

	state := renderState(pr.State, pr.IsDraft)
	line1 := fmt.Sprintf("%s  %s",
		titleStyle.Render(fmt.Sprintf("#%d %s", pr.Number, pr.Title)),
		state,
	)

	line2 := fmt.Sprintf("%s → %s ← %s   %s  %s",
		metaStyle.Render(pr.Author),
		metaStyle.Render(pr.BaseBranch),
		metaStyle.Render(pr.HeadBranch),
		ciPassStyle.Render(fmt.Sprintf("+%d", pr.Additions)),
		ciFailStyle.Render(fmt.Sprintf("-%d", pr.Deletions)),
	)
	if pr.ChangedFiles > 0 {
		line2 += metaStyle.Render(fmt.Sprintf("  %d files", pr.ChangedFiles))
	}

	var meta []string
	if len(pr.Labels) > 0 {
		names := make([]string, len(pr.Labels))
		for i, l := range pr.Labels {
			names[i] = l.Name
		}
		meta = append(meta, "Labels: "+strings.Join(names, ", "))
	}
	if len(pr.Reviewers) > 0 {
		meta = append(meta, "Reviewers: "+formatReviewers(pr.Reviewers))
	}
	line3 := metaStyle.Render(strings.Join(meta, "   "))

	divider := dividerStyle.Render(strings.Repeat("─", m.width))

	var body string
	if pr.Body != "" {
		r := newRenderer(m.width)
		body = "\n" + renderMarkdown(r, pr.Body)
	} else {
		body = "\n" + metaStyle.Render("No description provided.") + "\n"
	}

	content := fmt.Sprintf("%s\n%s\n%s\n%s%s", line1, line2, line3, divider, body)
	m.viewport.SetContent(content)
}

func (m *DetailModel) GotoTop()    { m.viewport.GotoTop() }
func (m *DetailModel) GotoBottom() { m.viewport.GotoBottom() }

func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m DetailModel) View() string {
	if !m.ready || m.pr == nil {
		return "Loading..."
	}
	return m.viewport.View()
}
