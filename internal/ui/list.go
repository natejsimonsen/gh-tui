package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/natejsimonsen/gh-tui/internal/github"
)

const cardHeight = 3

type ListModel struct {
	prs    []github.PullRequest
	cursor int
	offset int
	width  int
	height int
}

func NewListModel() ListModel {
	return ListModel{}
}

func (m *ListModel) ApplyTheme(_ Theme) {}

func (m *ListModel) SetPRs(prs []github.PullRequest) {
	m.prs = prs
	m.cursor = 0
	m.offset = 0
}

func (m *ListModel) SetSize(w, h int) {
	m.width = w
	m.height = h - 2
}

func (m *ListModel) SelectedPR() *github.PullRequest {
	if m.cursor >= 0 && m.cursor < len(m.prs) {
		return &m.prs[m.cursor]
	}
	return nil
}

func (m *ListModel) VisiblePRs() []github.PullRequest {
	if len(m.prs) == 0 {
		return nil
	}
	vis := m.visibleCards()
	end := m.offset + vis
	if end > len(m.prs) {
		end = len(m.prs)
	}
	return m.prs[m.offset:end]
}

func (m *ListModel) GotoTop() {
	m.cursor = 0
	m.offset = 0
}

func (m *ListModel) GotoBottom() {
	if len(m.prs) == 0 {
		return
	}
	m.cursor = len(m.prs) - 1
	m.fixOffset()
}

func (m *ListModel) visibleCards() int {
	n := m.height / cardHeight
	if n < 1 {
		return 1
	}
	return n
}

func (m *ListModel) fixOffset() {
	vis := m.visibleCards()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+vis {
		m.offset = m.cursor - vis + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "j", "down":
			if m.cursor < len(m.prs)-1 {
				m.cursor++
				m.fixOffset()
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				m.fixOffset()
			}
		case "d", "ctrl+d":
			m.cursor += m.visibleCards() / 2
			if m.cursor >= len(m.prs) {
				m.cursor = len(m.prs) - 1
			}
			m.fixOffset()
		case "u", "ctrl+u":
			m.cursor -= m.visibleCards() / 2
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.fixOffset()
		}
	}
	return m, nil
}

func (m ListModel) View() string {
	if len(m.prs) == 0 {
		return metaStyle.Render("No pull requests.")
	}

	var b strings.Builder
	vis := m.visibleCards()
	end := m.offset + vis
	if end > len(m.prs) {
		end = len(m.prs)
	}

	for i := m.offset; i < end; i++ {
		pr := m.prs[i]
		selected := i == m.cursor

		var bg lipgloss.Color
		if selected {
			bg = activeTheme.SelectedBg
		}

		prefix := "  "
		if selected {
			prefix = lipgloss.NewStyle().Foreground(activeTheme.Purple).Background(bg).Render("▎") +
				lipgloss.NewStyle().Background(bg).Render(" ")
		}

		line1 := prefix + m.renderTitle(pr, selected, bg)
		line2Prefix := "  "
		if selected {
			line2Prefix = lipgloss.NewStyle().Background(bg).Render("  ")
		}
		line2 := line2Prefix + m.renderMeta(pr, bg)

		if selected {
			line1 = padWithBg(line1, m.width, bg)
			line2 = padWithBg(line2, m.width, bg)
		}

		b.WriteString(line1 + "\n")
		b.WriteString(line2 + "\n")
		if i < end-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}

func padWithBg(s string, width int, bg lipgloss.Color) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + lipgloss.NewStyle().Background(bg).Render(strings.Repeat(" ", width-w))
}

func styledWithBg(s lipgloss.Style, bg lipgloss.Color) lipgloss.Style {
	if bg != "" {
		return s.Background(bg)
	}
	return s
}

func (m ListModel) renderTitle(pr github.PullRequest, selected bool, bg lipgloss.Color) string {
	num := styledWithBg(metaStyle, bg).Render(fmt.Sprintf("#%d", pr.Number))
	titleW := m.width - 12
	if titleW < 20 {
		titleW = 20
	}
	title := truncate(pr.Title, titleW)
	ts := lipgloss.NewStyle().Bold(true)
	if selected {
		ts = ts.Foreground(titleStyle.GetForeground())
	}
	title = styledWithBg(ts, bg).Render(title)
	gap := "  "
	if bg != "" {
		gap = lipgloss.NewStyle().Background(bg).Render(gap)
	}
	return num + gap + title
}

func (m ListModel) renderMeta(pr github.PullRequest, bg lipgloss.Color) string {
	author := styledWithBg(authorStyle, bg).Render("@" + pr.Author)

	var state string
	if pr.IsDraft {
		state = styledWithBg(draftStyle, bg).Render("Draft")
	} else {
		switch pr.State {
		case "OPEN":
			state = styledWithBg(openStyle, bg).Render("Open")
		case "MERGED":
			state = styledWithBg(mergedStyle, bg).Render("Merged")
		case "CLOSED":
			state = styledWithBg(closedStyle, bg).Render("Closed")
		default:
			state = pr.State
		}
	}

	var ci string
	switch pr.CIStatus {
	case "SUCCESS":
		ci = styledWithBg(ciPassStyle, bg).Render("✓ pass")
	case "FAILURE", "ERROR":
		ci = styledWithBg(ciFailStyle, bg).Render("✗ fail")
	case "PENDING", "EXPECTED":
		ci = styledWithBg(ciPendStyle, bg).Render("○ pend")
	case "":
		ci = styledWithBg(metaStyle, bg).Render("—")
	default:
		ci = pr.CIStatus
	}

	updated := styledWithBg(timeStyle, bg).Render(timeAgo(pr.UpdatedAt))
	parts := []string{author, state, ci, updated}

	if labels := formatLabels(pr.Labels); labels != "" {
		parts = append(parts, styledWithBg(metaStyle, bg).Render(labels))
	}
	if reviewers := formatReviewers(pr.Reviewers); reviewers != "" {
		parts = append(parts, styledWithBg(metaStyle, bg).Render(reviewers))
	}

	sep := styledWithBg(metaStyle, bg).Render(" · ")
	return strings.Join(parts, sep)
}

func renderState(state string, isDraft bool) string {
	if isDraft {
		return draftStyle.Render("Draft")
	}
	switch state {
	case "OPEN":
		return openStyle.Render("Open")
	case "MERGED":
		return mergedStyle.Render("Merged")
	case "CLOSED":
		return closedStyle.Render("Closed")
	}
	return state
}

func renderCI(status string) string {
	switch status {
	case "SUCCESS":
		return ciPassStyle.Render("✓ pass")
	case "FAILURE", "ERROR":
		return ciFailStyle.Render("✗ fail")
	case "PENDING", "EXPECTED":
		return ciPendStyle.Render("○ pend")
	case "":
		return metaStyle.Render("—")
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
	return truncate(strings.Join(names, ", "), 30)
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
	return truncate(strings.Join(parts, ", "), 25)
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
