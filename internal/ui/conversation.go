package ui

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/natejsimonsen/gh-tui/internal/debug"
	"github.com/natejsimonsen/gh-tui/internal/github"

	glamour "charm.land/glamour/v2"
)

type timelineEntry struct {
	timestamp time.Time
	content   string
}

type ConversationModel struct {
	viewport viewport.Model
	count    int
	width    int
	ready    bool
}

func NewConversationModel() ConversationModel {
	return ConversationModel{}
}

func newRenderer(width int) *glamour.TermRenderer {
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil
	}
	return r
}

func renderMarkdown(r *glamour.TermRenderer, body string) string {
	if r == nil {
		return body + "\n"
	}
	out, err := r.Render(body)
	if err != nil {
		return body + "\n"
	}
	return out
}

func (m *ConversationModel) SetData(comments []github.Comment, reviews []github.Review) {
	start := time.Now()
	total := len(comments) + len(reviews)
	entries := make([]timelineEntry, total)

	var wg sync.WaitGroup
	wg.Add(total)

	for i, c := range comments {
		go func(idx int, c github.Comment) {
			defer wg.Done()
			r := newRenderer(m.width)
			var b strings.Builder
			b.WriteString(fmt.Sprintf("%s  %s\n",
				authorStyle.Render(c.Author),
				timeStyle.Render(timeAgo(c.CreatedAt)),
			))
			b.WriteString(renderMarkdown(r, c.Body))
			entries[idx] = timelineEntry{timestamp: c.CreatedAt, content: b.String()}
		}(i, c)
	}

	for i, r := range reviews {
		go func(idx int, r github.Review) {
			defer wg.Done()
			renderer := newRenderer(m.width)
			var b strings.Builder
			icon := reviewStateIcon(r.State)
			b.WriteString(fmt.Sprintf("%s %s  %s\n",
				icon,
				authorStyle.Render(r.Author),
				timeStyle.Render(timeAgo(r.SubmittedAt)),
			))
			if r.Body != "" {
				b.WriteString(renderMarkdown(renderer, r.Body))
			}
			for _, rc := range r.Comments {
				path := metaStyle.Render(rc.Path)
				if rc.Line > 0 {
					path += metaStyle.Render(fmt.Sprintf(":%d", rc.Line))
				}
				b.WriteString(fmt.Sprintf("  %s\n", path))
				b.WriteString(renderMarkdown(renderer, rc.Body))
			}
			entries[idx] = timelineEntry{timestamp: r.SubmittedAt, content: b.String()}
		}(len(comments)+i, r)
	}

	wg.Wait()

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].timestamp.Before(entries[j].timestamp)
	})

	var full strings.Builder
	for i, e := range entries {
		full.WriteString(e.content)
		if i < len(entries)-1 {
			full.WriteString(dividerStyle.Render(strings.Repeat("─", m.width)) + "\n")
		}
	}

	m.count = len(entries)
	if m.ready {
		m.viewport.SetContent(full.String())
	}
	debug.Printf("SetData: %dms (%d entries)", time.Since(start).Milliseconds(), total)
}

func (m *ConversationModel) SetSize(w, h int) {
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
}

func (m *ConversationModel) GotoTop()    { m.viewport.GotoTop() }
func (m *ConversationModel) GotoBottom() { m.viewport.GotoBottom() }

func (m ConversationModel) Update(msg tea.Msg) (ConversationModel, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m ConversationModel) View() string {
	if !m.ready {
		return "Loading..."
	}
	header := titleStyle.Render(fmt.Sprintf("Conversation (%d)", m.count))
	return header + "\n" + m.viewport.View()
}

func reviewStateIcon(state string) string {
	switch state {
	case "APPROVED":
		return ciPassStyle.Render("✓ Approved")
	case "CHANGES_REQUESTED":
		return ciFailStyle.Render("✗ Changes Requested")
	case "COMMENTED":
		return metaStyle.Render("💬 Commented")
	case "DISMISSED":
		return metaStyle.Render("Dismissed")
	default:
		return metaStyle.Render(state)
	}
}
