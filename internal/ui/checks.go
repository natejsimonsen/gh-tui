package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/natejsimonsen/gh-tui/internal/github"
)

type ChecksModel struct {
	viewport viewport.Model
	checks   []github.Check
	width    int
	ready    bool
}

func NewChecksModel() ChecksModel {
	return ChecksModel{}
}

func (m *ChecksModel) SetChecks(checks []github.Check) {
	m.checks = checks
	m.renderContent()
}

func (m *ChecksModel) SetSize(w, h int) {
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

func (m *ChecksModel) renderContent() {
	if m.width == 0 {
		return
	}
	if len(m.checks) == 0 {
		if m.ready {
			m.viewport.SetContent(metaStyle.Render("No checks."))
		}
		return
	}

	var b strings.Builder
	for _, c := range m.checks {
		icon := checkIcon(c.Conclusion, c.Status)
		b.WriteString(fmt.Sprintf("  %s  %s\n", icon, c.Name))
	}
	if m.ready {
		m.viewport.SetContent(b.String())
	}
}

func checkIcon(conclusion, status string) string {
	if status == "IN_PROGRESS" || status == "QUEUED" || status == "PENDING" {
		return ciPendStyle.Render("○")
	}
	switch strings.ToUpper(conclusion) {
	case "SUCCESS":
		return ciPassStyle.Render("✓")
	case "FAILURE", "TIMED_OUT", "CANCELLED", "ERROR":
		return ciFailStyle.Render("✗")
	case "NEUTRAL", "SKIPPED":
		return metaStyle.Render("—")
	default:
		return ciPendStyle.Render("○")
	}
}

func (m *ChecksModel) GotoTop()    { m.viewport.GotoTop() }
func (m *ChecksModel) GotoBottom() { m.viewport.GotoBottom() }

func (m ChecksModel) Update(msg tea.Msg) (ChecksModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m ChecksModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	pass, fail, pend := 0, 0, 0
	for _, c := range m.checks {
		switch strings.ToUpper(c.Conclusion) {
		case "SUCCESS":
			pass++
		case "FAILURE", "TIMED_OUT", "CANCELLED", "ERROR":
			fail++
		default:
			pend++
		}
	}

	header := titleStyle.Render(fmt.Sprintf("Checks (%d)  %s %s %s",
		len(m.checks),
		ciPassStyle.Render(fmt.Sprintf("✓%d", pass)),
		ciFailStyle.Render(fmt.Sprintf("✗%d", fail)),
		ciPendStyle.Render(fmt.Sprintf("○%d", pend)),
	))
	return header + "\n" + m.viewport.View()
}
