package ui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type reviewWantsSubmitMsg struct {
	action int
	body   string
}

type reviewSubmittedMsg struct {
	err error
}

var reviewActionEvents = []string{"APPROVE", "REQUEST_CHANGES", "COMMENT"}
var reviewActionLabels = []string{"Approve", "Request Changes", "Comment Only"}

type reviewState int

const (
	reviewTree     reviewState = iota
	reviewFileDiff
	reviewComment
	reviewSubmit
)

type reviewFile struct {
	path      string
	additions int
	deletions int
	diffChunk string
	reviewed  bool
	comments  []string
}

type ReviewModel struct {
	state    reviewState
	files    []reviewFile
	cursor   int
	prNumber int
	prTitle  string

	diffVP    viewport.Model
	diffReady bool

	commentTA textarea.Model

	submitAction int
	submitBody   textarea.Model
	submitFocus  int

	width  int
	height int
}

func NewReviewModel() ReviewModel {
	return ReviewModel{}
}

func (m *ReviewModel) SetData(prNumber int, prTitle string, rawDiff string) {
	m.prNumber = prNumber
	m.prTitle = prTitle
	m.files = parseReviewFiles(rawDiff)
	m.cursor = 0
	m.state = reviewTree
}

func (m *ReviewModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	if m.diffReady {
		m.diffVP.Width = w
		m.diffVP.Height = h - 3
	}
}

func (m *ReviewModel) Clear() {
	m.files = nil
	m.cursor = 0
	m.state = reviewTree
	m.prNumber = 0
	m.prTitle = ""
}

func (m ReviewModel) IsEditing() bool {
	return m.state == reviewComment || (m.state == reviewSubmit && m.submitFocus == 1)
}

func (m ReviewModel) FooterHelp() string {
	switch m.state {
	case reviewTree:
		return "m:mark  a:comment  x:del  Enter:diff  S:submit  Esc:back"
	case reviewFileDiff:
		return "m:mark  Esc:back"
	case reviewComment:
		return "Ctrl+S:save  Esc:cancel"
	case reviewSubmit:
		return "Tab:switch  Ctrl+S:submit  Esc:cancel"
	}
	return ""
}

// --- Update ---

func (m ReviewModel) Update(msg tea.Msg) (ReviewModel, tea.Cmd) {
	switch m.state {
	case reviewTree:
		return m.updateTree(msg)
	case reviewFileDiff:
		return m.updateFileDiff(msg)
	case reviewComment:
		return m.updateComment(msg)
	case reviewSubmit:
		return m.updateSubmit(msg)
	}
	return m, nil
}

func (m ReviewModel) updateTree(msg tea.Msg) (ReviewModel, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "j", "down":
		if m.cursor < len(m.files)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "G":
		if len(m.files) > 0 {
			m.cursor = len(m.files) - 1
		}
	case "g":
		m.cursor = 0
	case "m":
		if m.cursor < len(m.files) {
			m.files[m.cursor].reviewed = !m.files[m.cursor].reviewed
		}
	case "a":
		if m.cursor < len(m.files) {
			cmd := m.enterComment()
			return m, cmd
		}
	case "x":
		if m.cursor < len(m.files) {
			if n := len(m.files[m.cursor].comments); n > 0 {
				m.files[m.cursor].comments = m.files[m.cursor].comments[:n-1]
			}
		}
	case "enter":
		if m.cursor < len(m.files) {
			m.enterFileDiff()
		}
	case "S":
		m.enterSubmit()
	}

	return m, nil
}

func (m ReviewModel) updateFileDiff(msg tea.Msg) (ReviewModel, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case keyMsg.Type == tea.KeyEscape:
			m.state = reviewTree
			return m, nil
		case keyMsg.String() == "m":
			if m.cursor < len(m.files) {
				m.files[m.cursor].reviewed = !m.files[m.cursor].reviewed
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.diffVP, cmd = m.diffVP.Update(msg)
	return m, cmd
}

func (m ReviewModel) updateComment(msg tea.Msg) (ReviewModel, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyEscape:
			m.state = reviewTree
			return m, nil
		case tea.KeyCtrlS:
			text := strings.TrimSpace(m.commentTA.Value())
			if text != "" {
				m.files[m.cursor].comments = append(m.files[m.cursor].comments, text)
			}
			m.state = reviewTree
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.commentTA, cmd = m.commentTA.Update(msg)
	return m, cmd
}

func (m ReviewModel) updateSubmit(msg tea.Msg) (ReviewModel, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyEscape:
			m.state = reviewTree
			return m, nil
		case tea.KeyCtrlS:
			action := m.submitAction
			body := m.submitBody.Value()
			return m, func() tea.Msg {
				return reviewWantsSubmitMsg{action: action, body: body}
			}
		case tea.KeyTab:
			m.submitFocus = (m.submitFocus + 1) % 2
			if m.submitFocus == 1 {
				return m, m.submitBody.Focus()
			}
			m.submitBody.Blur()
			return m, nil
		}

		if m.submitFocus == 0 {
			switch keyMsg.String() {
			case "j", "down":
				m.submitAction = (m.submitAction + 1) % len(reviewActionLabels)
				return m, nil
			case "k", "up":
				m.submitAction = (m.submitAction - 1 + len(reviewActionLabels)) % len(reviewActionLabels)
				return m, nil
			}
		}
	}

	if m.submitFocus == 1 {
		var cmd tea.Cmd
		m.submitBody, cmd = m.submitBody.Update(msg)
		return m, cmd
	}

	return m, nil
}

// --- State transitions ---

func (m *ReviewModel) enterFileDiff() {
	f := m.files[m.cursor]
	content := renderDiffContent(f.diffChunk, m.width)
	h := m.height - 3
	if h < 1 {
		h = 1
	}
	if !m.diffReady {
		m.diffVP = viewport.New(m.width, h)
		m.diffVP.KeyMap = viewportKeyMap()
		m.diffReady = true
	} else {
		m.diffVP.Width = m.width
		m.diffVP.Height = h
	}
	m.diffVP.SetContent(content)
	m.diffVP.GotoTop()
	m.state = reviewFileDiff
}

func (m *ReviewModel) enterComment() tea.Cmd {
	m.commentTA = textarea.New()
	m.commentTA.Placeholder = "Type your comment..."
	m.commentTA.SetWidth(m.width - 4)
	h := m.height - 10
	if h < 4 {
		h = 4
	}
	m.commentTA.SetHeight(h)
	m.state = reviewComment
	return m.commentTA.Focus()
}

func (m *ReviewModel) enterSubmit() {
	m.submitBody = textarea.New()
	m.submitBody.Placeholder = "Review body (optional)..."
	m.submitBody.SetWidth(m.width - 4)
	h := m.height - 14
	if h < 4 {
		h = 4
	}
	m.submitBody.SetHeight(h)
	m.submitFocus = 0
	m.submitAction = 0
	m.state = reviewSubmit
}

// --- View ---

func (m ReviewModel) View() string {
	switch m.state {
	case reviewTree:
		return m.treeView()
	case reviewFileDiff:
		return m.fileDiffView()
	case reviewComment:
		return m.commentView()
	case reviewSubmit:
		return m.submitView()
	}
	return ""
}

func (m ReviewModel) treeView() string {
	if len(m.files) == 0 {
		return titleStyle.Render("Review") + "\n" + metaStyle.Render("No files to review.")
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("Review PR #%d", m.prNumber)) + "\n")
	b.WriteString(dividerStyle.Render(strings.Repeat("─", m.width)) + "\n")

	lastDir := "\x00"
	for i, f := range m.files {
		dir := filepath.Dir(f.path)
		if dir == "." {
			dir = ""
		}
		if dir != lastDir {
			lastDir = dir
			if dir != "" {
				b.WriteString("\n  " + metaStyle.Render(dir+"/") + "\n")
			} else {
				b.WriteByte('\n')
			}
		}

		icon := metaStyle.Render("○")
		if f.reviewed {
			icon = ciPassStyle.Render("✓")
		}

		name := filepath.Base(f.path)
		stats := ciPassStyle.Render(fmt.Sprintf("+%d", f.additions)) + " " +
			ciFailStyle.Render(fmt.Sprintf("-%d", f.deletions))

		badge := ""
		if len(f.comments) > 0 {
			badge = metaStyle.Render(fmt.Sprintf(" 💬%d", len(f.comments)))
		}

		if i == m.cursor {
			b.WriteString(authorStyle.Render("> ") + icon + " " + authorStyle.Render(name) + "  " + stats + badge + "\n")
		} else {
			b.WriteString("  " + icon + " " + name + "  " + stats + badge + "\n")
		}
	}

	reviewed := 0
	for _, f := range m.files {
		if f.reviewed {
			reviewed++
		}
	}
	b.WriteString(fmt.Sprintf("\n  %s\n",
		metaStyle.Render(fmt.Sprintf("%d/%d files reviewed", reviewed, len(m.files))),
	))

	return b.String()
}

func (m ReviewModel) fileDiffView() string {
	if m.cursor >= len(m.files) {
		return ""
	}
	f := m.files[m.cursor]

	icon := metaStyle.Render("○")
	if f.reviewed {
		icon = ciPassStyle.Render("✓")
	}

	header := titleStyle.Render(f.path) + "  " + icon + "  " +
		ciPassStyle.Render(fmt.Sprintf("+%d", f.additions)) + " " +
		ciFailStyle.Render(fmt.Sprintf("-%d", f.deletions))

	return header + "\n" + m.diffVP.View()
}

func (m ReviewModel) commentView() string {
	if m.cursor >= len(m.files) {
		return ""
	}
	f := m.files[m.cursor]

	var b strings.Builder
	b.WriteString(titleStyle.Render("Add Comment") + "\n")
	b.WriteString(metaStyle.Render(f.path) + "\n")
	b.WriteString(dividerStyle.Render(strings.Repeat("─", m.width)) + "\n\n")

	for i, c := range f.comments {
		preview := c
		if len(preview) > 60 {
			preview = preview[:57] + "..."
		}
		b.WriteString(metaStyle.Render(fmt.Sprintf("  #%d: %s", i+1, preview)) + "\n")
	}
	if len(f.comments) > 0 {
		b.WriteByte('\n')
	}

	b.WriteString(m.commentTA.View())
	return b.String()
}

func (m ReviewModel) submitView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Submit Review") + "\n")
	b.WriteString(dividerStyle.Render(strings.Repeat("─", m.width)) + "\n\n")

	actionIcons := []string{
		ciPassStyle.Render("✓"),
		ciFailStyle.Render("✗"),
		metaStyle.Render("💬"),
	}

	for i, label := range reviewActionLabels {
		cursor := "  "
		if i == m.submitAction {
			cursor = "> "
		}
		if i == m.submitAction {
			b.WriteString(authorStyle.Render(cursor) + actionIcons[i] + " " + authorStyle.Render(label) + "\n")
		} else {
			b.WriteString(metaStyle.Render(cursor) + actionIcons[i] + " " + metaStyle.Render(label) + "\n")
		}
	}

	b.WriteString("\n" + metaStyle.Render("Review body:") + "\n")
	b.WriteString(m.submitBody.View() + "\n")

	commentCount := 0
	for _, f := range m.files {
		commentCount += len(f.comments)
	}
	if commentCount > 0 {
		b.WriteString(metaStyle.Render(fmt.Sprintf("\n%d file comment(s) will be included", commentCount)) + "\n")
	}

	return b.String()
}

// --- Helpers ---

func parseReviewFiles(raw string) []reviewFile {
	if raw == "" {
		return nil
	}

	var files []reviewFile
	lines := strings.Split(raw, "\n")
	var current *reviewFile
	var chunk strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			if current != nil {
				current.diffChunk = chunk.String()
				files = append(files, *current)
			}
			current = &reviewFile{path: extractFileName(line)}
			chunk.Reset()
			chunk.WriteString(line + "\n")
		} else if current != nil {
			chunk.WriteString(line + "\n")
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				current.additions++
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				current.deletions++
			}
		}
	}
	if current != nil {
		current.diffChunk = chunk.String()
		files = append(files, *current)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].path < files[j].path
	})

	return files
}
