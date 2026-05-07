package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/natejsimonsen/gh-tui/internal/github"
)

type DetailModel struct {
	viewport viewport.Model
	pr       *github.PRDetail
	width    int
	height   int
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
	m.height = h
	bodyH := h - 6
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
	if m.pr != nil {
		m.renderContent()
	}
}

func (m *DetailModel) renderContent() {
	if m.pr == nil || m.width == 0 {
		return
	}

	var content string
	if m.pr.Body != "" {
		rendered, err := glamour.Render(m.pr.Body, "auto")
		if err != nil {
			content = m.pr.Body
		} else {
			content = rendered
		}
	} else {
		content = metaStyle.Render("No description provided.")
	}

	m.viewport.SetContent(content)
}

func (m DetailModel) headerView() string {
	if m.pr == nil {
		return ""
	}
	pr := m.pr

	state := formatState(pr.State, pr.IsDraft)
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

	return fmt.Sprintf("%s\n%s\n%s\n%s", line1, line2, line3, divider)
}

func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m DetailModel) View() string {
	if !m.ready {
		return "Loading..."
	}
	return m.headerView() + "\n" + m.viewport.View()
}
