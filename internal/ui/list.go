package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/natejsimonsen/gh-tui/internal/github"
)

type ListModel struct {
	table table.Model
	prs   []github.PullRequest
	width int
}

func NewListModel() ListModel {
	t := table.New(
		table.WithColumns(defaultColumns(80)),
		table.WithFocused(true),
		table.WithHeight(20),
	)
	applyTableStyles(&t)
	return ListModel{table: t}
}

func applyTableStyles(t *table.Model) {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(dividerStyle.GetForeground()).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
}

func (m *ListModel) ApplyTheme(t Theme) {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(t.Border)).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color(t.Fg)).
		Background(lipgloss.Color(t.BgHighlight)).
		Bold(false)
	m.table.SetStyles(s)
}

func (m *ListModel) SetPRs(prs []github.PullRequest) {
	m.prs = prs
	rows := make([]table.Row, len(prs))
	for i, pr := range prs {
		rows[i] = table.Row{
			fmt.Sprintf("#%d", pr.Number),
			truncate(pr.Title, m.titleWidth()),
			pr.Author,
			formatState(pr.State, pr.IsDraft),
			formatCI(pr.CIStatus),
			formatLabels(pr.Labels),
			formatReviewers(pr.Reviewers),
			timeAgo(pr.UpdatedAt),
		}
	}
	m.table.SetRows(rows)
}

func (m *ListModel) SetSize(w, h int) {
	m.width = w
	m.table.SetWidth(w)
	m.table.SetHeight(h - 4)
	m.table.SetColumns(defaultColumns(w))
}

func (m *ListModel) SelectedPR() *github.PullRequest {
	idx := m.table.Cursor()
	if idx >= 0 && idx < len(m.prs) {
		return &m.prs[idx]
	}
	return nil
}

func (m *ListModel) GotoTop()    { m.table.GotoTop() }
func (m *ListModel) GotoBottom() { m.table.GotoBottom() }

func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m ListModel) View() string {
	return m.table.View()
}

func (m ListModel) titleWidth() int {
	w := m.width - 82
	if w < 20 {
		return 20
	}
	return w
}

func defaultColumns(width int) []table.Column {
	titleW := width - 82
	if titleW < 20 {
		titleW = 20
	}
	return []table.Column{
		{Title: "#", Width: 7},
		{Title: "Title", Width: titleW},
		{Title: "Author", Width: 15},
		{Title: "Status", Width: 8},
		{Title: "CI", Width: 10},
		{Title: "Labels", Width: 20},
		{Title: "Reviewers", Width: 15},
		{Title: "Updated", Width: 7},
	}
}

func formatState(state string, isDraft bool) string {
	if isDraft {
		return "Draft"
	}
	switch state {
	case "OPEN":
		return "Open"
	case "MERGED":
		return "Merged"
	case "CLOSED":
		return "Closed"
	}
	return state
}

func formatCI(status string) string {
	switch status {
	case "SUCCESS":
		return "✓ pass"
	case "FAILURE", "ERROR":
		return "✗ fail"
	case "PENDING", "EXPECTED":
		return "○ pend"
	case "":
		return "—"
	}
	return status
}

func formatLabels(labels []github.Label) string {
	if len(labels) == 0 {
		return ""
	}
	names := make([]string, len(labels))
	for i, l := range labels {
		names[i] = l.Name
	}
	return truncate(strings.Join(names, ", "), 18)
}

func formatReviewers(reviewers []github.Reviewer) string {
	if len(reviewers) == 0 {
		return ""
	}
	parts := make([]string, len(reviewers))
	for i, r := range reviewers {
		switch r.State {
		case "APPROVED":
			parts[i] = r.Login + " ✓"
		case "CHANGES_REQUESTED":
			parts[i] = r.Login + " ✗"
		default:
			parts[i] = r.Login
		}
	}
	return truncate(strings.Join(parts, ", "), 13)
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw", int(d.Hours()/(24*7)))
	default:
		return fmt.Sprintf("%dmo", int(d.Hours()/(24*30)))
	}
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-1]) + "…"
}
