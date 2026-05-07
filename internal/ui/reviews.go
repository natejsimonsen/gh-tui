package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/natejsimonsen/gh-tui/internal/github"
)

type ReviewsModel struct {
	viewport viewport.Model
	reviews  []github.Review
	width    int
	ready    bool
}

func NewReviewsModel() ReviewsModel {
	return ReviewsModel{}
}

func (m *ReviewsModel) SetReviews(reviews []github.Review) {
	m.reviews = reviews
	m.renderContent()
}

func (m *ReviewsModel) SetSize(w, h int) {
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

func (m *ReviewsModel) renderContent() {
	if m.width == 0 {
		return
	}
	if len(m.reviews) == 0 {
		m.viewport.SetContent(metaStyle.Render("No reviews."))
		return
	}

	var b strings.Builder
	for i, r := range m.reviews {
		icon := reviewStateIcon(r.State)
		b.WriteString(fmt.Sprintf("%s %s  %s\n",
			icon,
			authorStyle.Render(r.Author),
			timeStyle.Render(timeAgo(r.SubmittedAt)),
		))

		if r.Body != "" {
			rendered, err := glamour.Render(r.Body, "auto")
			if err != nil {
				b.WriteString(r.Body + "\n")
			} else {
				b.WriteString(rendered)
			}
		}

		for _, rc := range r.Comments {
			path := metaStyle.Render(rc.Path)
			if rc.Line > 0 {
				path += metaStyle.Render(fmt.Sprintf(":%d", rc.Line))
			}
			b.WriteString(fmt.Sprintf("  %s\n", path))
			rendered, err := glamour.Render(rc.Body, "auto")
			if err != nil {
				b.WriteString("  " + rc.Body + "\n")
			} else {
				b.WriteString(rendered)
			}
		}

		if i < len(m.reviews)-1 {
			b.WriteString(dividerStyle.Render(strings.Repeat("─", m.width)) + "\n")
		}
	}
	m.viewport.SetContent(b.String())
}

func reviewStateIcon(state string) string {
	switch state {
	case "APPROVED":
		return ciPassStyle.Render("✓ Approved")
	case "CHANGES_REQUESTED":
		return ciFailStyle.Render("✗ Changes Requested")
	case "COMMENTED":
		return metaStyle.Render("Commented")
	case "DISMISSED":
		return metaStyle.Render("Dismissed")
	default:
		return metaStyle.Render(state)
	}
}

func (m ReviewsModel) Update(msg tea.Msg) (ReviewsModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m ReviewsModel) View() string {
	if !m.ready {
		return "Loading..."
	}
	header := titleStyle.Render(fmt.Sprintf("Reviews (%d)", len(m.reviews)))
	return header + "\n" + m.viewport.View()
}
