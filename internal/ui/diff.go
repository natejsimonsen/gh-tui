package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DiffModel struct {
	viewport  viewport.Model
	raw       string
	fileCount int
	additions int
	deletions int
	width     int
	ready     bool
}

func NewDiffModel() DiffModel {
	return DiffModel{}
}

func (m *DiffModel) SetDiff(content string) {
	m.raw = content
	m.fileCount = strings.Count(content, "diff --git")
	m.additions = 0
	m.deletions = 0
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			m.additions++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			m.deletions++
		}
	}
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
	if m.width == 0 || !m.ready {
		return
	}
	if m.raw == "" {
		m.viewport.SetContent(metaStyle.Render("No changes."))
		return
	}

	const gutterW = 5
	contentW := m.width - gutterW - 1

	lines := strings.Split(m.raw, "\n")

	type fileSection struct {
		name      string
		codeTexts []string
	}
	var files []fileSection

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			files = append(files, fileSection{name: extractFileName(line)})
		} else if len(files) > 0 {
			switch {
			case strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
				files[len(files)-1].codeTexts = append(files[len(files)-1].codeTexts, line[1:])
			case strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---"):
				files[len(files)-1].codeTexts = append(files[len(files)-1].codeTexts, line[1:])
			case len(line) > 0 && line[0] == ' ':
				files[len(files)-1].codeTexts = append(files[len(files)-1].codeTexts, line[1:])
			}
		}
	}

	fileTokens := make([][][]chroma.Token, len(files))
	for i, f := range files {
		fileTokens[i] = tokenizeLines(getLexer(f.name), f.codeTexts)
	}

	var b strings.Builder
	var oldNum, newNum int
	fileIdx := -1
	codeIdx := 0

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "diff --git"):
			fileIdx++
			codeIdx = 0
			if fileIdx > 0 {
				b.WriteByte('\n')
			}
			banner := diffFileBannerStyle.Width(m.width).Render(" " + files[fileIdx].name)
			b.WriteString(banner + "\n")

		case strings.HasPrefix(line, "index "),
			strings.HasPrefix(line, "--- "),
			strings.HasPrefix(line, "+++ "),
			strings.HasPrefix(line, "new file"),
			strings.HasPrefix(line, "deleted file"),
			strings.HasPrefix(line, "old mode"),
			strings.HasPrefix(line, "new mode"),
			strings.HasPrefix(line, "similarity"),
			strings.HasPrefix(line, "rename from"),
			strings.HasPrefix(line, "rename to"),
			strings.HasPrefix(line, "Binary"):
			continue

		case strings.HasPrefix(line, "@@"):
			oldNum, newNum = parseHunkHeader(line)
			hunk := diffHunkStyle.Render(extractHunkFunc(line))
			b.WriteString(hunk + "\n")

		case strings.HasPrefix(line, "+"):
			num := diffGutterStyle.Render(padLeft(strconv.Itoa(newNum), gutterW))
			text := line[1:]
			var hl string
			if fileIdx >= 0 && fileTokens[fileIdx] != nil && codeIdx < len(fileTokens[fileIdx]) {
				hl = renderTokens(fileTokens[fileIdx][codeIdx], activeTheme.DiffAddBg)
			} else {
				hl = renderPlainWithBg(text, activeTheme.DiffAddBg)
			}
			codeIdx++
			pad := contentW - lipgloss.Width(text)
			if pad > 0 {
				hl += lipgloss.NewStyle().Background(activeTheme.DiffAddBg).Render(strings.Repeat(" ", pad))
			}
			b.WriteString(num + " " + hl + "\n")
			newNum++

		case strings.HasPrefix(line, "-"):
			num := diffGutterStyle.Render(padLeft(strconv.Itoa(oldNum), gutterW))
			text := line[1:]
			var hl string
			if fileIdx >= 0 && fileTokens[fileIdx] != nil && codeIdx < len(fileTokens[fileIdx]) {
				hl = renderTokens(fileTokens[fileIdx][codeIdx], activeTheme.DiffDelBg)
			} else {
				hl = renderPlainWithBg(text, activeTheme.DiffDelBg)
			}
			codeIdx++
			pad := contentW - lipgloss.Width(text)
			if pad > 0 {
				hl += lipgloss.NewStyle().Background(activeTheme.DiffDelBg).Render(strings.Repeat(" ", pad))
			}
			b.WriteString(num + " " + hl + "\n")
			oldNum++

		case len(line) > 0 && line[0] == ' ':
			num := diffGutterStyle.Render(padLeft(strconv.Itoa(newNum), gutterW))
			var hl string
			if fileIdx >= 0 && fileTokens[fileIdx] != nil && codeIdx < len(fileTokens[fileIdx]) {
				hl = renderTokens(fileTokens[fileIdx][codeIdx], "")
			} else {
				hl = line[1:]
			}
			codeIdx++
			b.WriteString(num + " " + hl + "\n")
			oldNum++
			newNum++
		}
	}

	m.viewport.SetContent(b.String())
}

func extractFileName(diffLine string) string {
	parts := strings.SplitN(diffLine, " b/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return diffLine
}

func extractHunkFunc(line string) string {
	// @@ -10,5 +10,7 @@ func something()
	idx := strings.Index(line, "@@")
	if idx < 0 {
		return line
	}
	rest := line[idx+2:]
	idx2 := strings.Index(rest, "@@")
	if idx2 < 0 {
		return line
	}
	after := strings.TrimSpace(rest[idx2+2:])
	range_ := strings.TrimSpace(rest[:idx2])
	if after != "" {
		return "  " + range_ + "  " + after
	}
	return "  " + range_
}

func parseHunkHeader(line string) (oldStart, newStart int) {
	line = strings.TrimPrefix(line, "@@ ")
	parts := strings.Fields(line)
	if len(parts) >= 2 {
		if old := strings.TrimPrefix(parts[0], "-"); old != parts[0] {
			if n, _, ok := splitRange(old); ok {
				oldStart = n
			}
		}
		if new_ := strings.TrimPrefix(parts[1], "+"); new_ != parts[1] {
			if n, _, ok := splitRange(new_); ok {
				newStart = n
			}
		}
	}
	return
}

func splitRange(s string) (start, length int, ok bool) {
	parts := strings.SplitN(s, ",", 2)
	n, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, false
	}
	l := 1
	if len(parts) == 2 {
		l, _ = strconv.Atoi(parts[1])
	}
	return n, l, true
}

func padLeft(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return strings.Repeat(" ", w-len(s)) + s
}

func padRight(s string, w int) string {
	vis := lipgloss.Width(s)
	if vis >= w {
		return s
	}
	return s + strings.Repeat(" ", w-vis)
}

func (m *DiffModel) HasContent() bool { return m.raw != "" }

func (m *DiffModel) Clear() {
	m.raw = ""
	m.fileCount = 0
	m.additions = 0
	m.deletions = 0
	if m.ready {
		m.viewport.SetContent("")
	}
}

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
	header := titleStyle.Render(fmt.Sprintf("Files Changed (%d)", m.fileCount)) +
		"  " + ciPassStyle.Render(fmt.Sprintf("+%d", m.additions)) +
		"  " + ciFailStyle.Render(fmt.Sprintf("-%d", m.deletions))
	return header + "\n" + m.viewport.View()
}
