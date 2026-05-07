package ui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/natejsimonsen/gh-tui/internal/github"
)

type Mode int

const (
	ModeList Mode = iota
	ModeDetail
	ModeConversation
	ModeChecks
	ModeFiles
	ModeHelp
	ModeThemePicker
)

type App struct {
	mode     Mode
	prevMode Mode

	list         ListModel
	detail       DetailModel
	conversation ConversationModel
	checks       ChecksModel
	files        DiffModel
	help         HelpModel

	client       *github.Client
	repo         github.Repo
	prDetail     *github.PRDetail
	pageInfo     github.PageInfo
	filter       []string
	author       string
	filterByUser bool

	width  int
	height int

	loading    bool
	statusMsg  string
	pendingG   bool
	themeIdx   int
	themeCursor int
}

type prsLoadedMsg struct {
	result *github.PRListResult
}

type prDetailLoadedMsg struct {
	detail *github.PRDetail
}

type diffLoadedMsg struct {
	diff string
}

type errMsg struct {
	err error
}

func NewApp(client *github.Client, repo github.Repo, author string) App {
	return App{
		mode:         ModeList,
		list:         NewListModel(),
		detail:       NewDetailModel(),
		conversation: NewConversationModel(),
		checks:       NewChecksModel(),
		files:        NewDiffModel(),
		help:         NewHelpModel(),
		client:       client,
		repo:         repo,
		filter:       []string{"OPEN"},
		author:       author,
		filterByUser: author != "",
	}
}

func (a App) Init() tea.Cmd {
	return a.fetchPRs()
}

func (a App) fetchPRs() tea.Cmd {
	client := a.client
	repo := a.repo
	filter := a.filter
	cursor := a.pageInfo.EndCursor
	filterByUser := a.filterByUser
	author := a.author
	return func() tea.Msg {
		var result *github.PRListResult
		var err error
		if filterByUser && author != "" {
			result, err = client.SearchPRs(context.Background(), repo, author, filter, 50, cursor)
		} else {
			result, err = client.ListPRs(context.Background(), repo, filter, 50, cursor)
		}
		if err != nil {
			return errMsg{err}
		}
		return prsLoadedMsg{result}
	}
}

func (a App) fetchPRDetail(number int) tea.Cmd {
	client := a.client
	repo := a.repo
	return func() tea.Msg {
		detail, err := client.GetPRDetail(context.Background(), repo, number)
		if err != nil {
			return errMsg{err}
		}
		return prDetailLoadedMsg{detail}
	}
}

func (a App) fetchDiff(number int) tea.Cmd {
	client := a.client
	repo := a.repo
	return func() tea.Msg {
		diff, err := client.GetPRDiff(context.Background(), repo, number)
		if err != nil {
			return errMsg{err}
		}
		return diffLoadedMsg{diff}
	}
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		contentH := msg.Height - 2
		a.list.SetSize(msg.Width, contentH)
		a.detail.SetSize(msg.Width, contentH)
		a.conversation.SetSize(msg.Width, contentH)
		a.checks.SetSize(msg.Width, contentH)
		a.files.SetSize(msg.Width, contentH)
		a.help.SetSize(msg.Width, contentH)
		return a, nil

	case prsLoadedMsg:
		a.loading = false
		a.pageInfo = msg.result.PageInfo
		a.list.SetPRs(msg.result.PullRequests)
		if a.filterByUser {
			a.statusMsg = fmt.Sprintf("%d PRs (@%s)", msg.result.TotalCount, a.author)
		} else {
			a.statusMsg = fmt.Sprintf("%d PRs", msg.result.TotalCount)
		}
		return a, nil

	case prDetailLoadedMsg:
		a.loading = false
		a.prDetail = msg.detail
		a.mode = ModeDetail
		a.detail.SetPR(msg.detail)
		a.conversation.SetData(msg.detail.Comments, msg.detail.Reviews)
		a.checks.SetChecks(msg.detail.Checks)
		a.files.SetDiff("")
		return a, nil

	case diffLoadedMsg:
		a.loading = false
		a.files.SetDiff(msg.diff)
		a.mode = ModeFiles
		a.statusMsg = "Diff loaded"
		return a, nil

	case errMsg:
		a.loading = false
		a.statusMsg = "Error: " + msg.err.Error()
		return a, nil

	case tea.KeyMsg:
		if key.Matches(msg, keys.Quit) && a.mode != ModeThemePicker {
			return a, tea.Quit
		}
		return a.handleKeys(msg)
	}

	return a.updateActiveModel(msg)
}

func (a App) handleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.mode == ModeThemePicker {
		return a.handleThemePickerKeys(msg)
	}

	if key.Matches(msg, keys.Help) {
		if a.mode == ModeHelp {
			a.mode = a.prevMode
		} else {
			a.prevMode = a.mode
			a.mode = ModeHelp
		}
		return a, nil
	}

	if msg.String() == "t" {
		a.themeCursor = a.themeIdx
		a.prevMode = a.mode
		a.mode = ModeThemePicker
		return a, nil
	}

	switch a.mode {
	case ModeList:
		return a.handleListKeys(msg)
	default:
		return a.handleViewKeys(msg)
	}
}

func (a App) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.pendingG {
		a.pendingG = false
		if msg.String() == "g" {
			a.list.GotoTop()
			return a, nil
		}
	}

	switch {
	case key.Matches(msg, keys.Enter):
		if pr := a.list.SelectedPR(); pr != nil {
			a.loading = true
			a.statusMsg = fmt.Sprintf("Loading #%d...", pr.Number)
			return a, a.fetchPRDetail(pr.Number)
		}
	case key.Matches(msg, keys.Open):
		a.filter = []string{"OPEN"}
		a.pageInfo = github.PageInfo{}
		a.loading = true
		a.statusMsg = "Loading open PRs..."
		return a, a.fetchPRs()
	case key.Matches(msg, keys.Merged):
		a.filter = []string{"MERGED"}
		a.pageInfo = github.PageInfo{}
		a.loading = true
		a.statusMsg = "Loading merged PRs..."
		return a, a.fetchPRs()
	case key.Matches(msg, keys.Closed):
		a.filter = []string{"CLOSED"}
		a.pageInfo = github.PageInfo{}
		a.loading = true
		a.statusMsg = "Loading closed PRs..."
		return a, a.fetchPRs()
	case key.Matches(msg, keys.All):
		a.filter = nil
		a.pageInfo = github.PageInfo{}
		a.loading = true
		a.statusMsg = "Loading all PRs..."
		return a, a.fetchPRs()
	case key.Matches(msg, keys.Mine):
		a.filterByUser = !a.filterByUser
		a.pageInfo = github.PageInfo{}
		a.loading = true
		if a.filterByUser {
			a.statusMsg = fmt.Sprintf("Loading @%s PRs...", a.author)
		} else {
			a.statusMsg = "Loading all authors..."
		}
		return a, a.fetchPRs()
	case key.Matches(msg, keys.Browser):
		return a.openInBrowser()
	case msg.String() == "G":
		a.list.GotoBottom()
		return a, nil
	case msg.String() == "g":
		a.pendingG = true
		return a, nil
	}

	var cmd tea.Cmd
	a.list, cmd = a.list.Update(msg)
	return a, cmd
}

func (a App) handleViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.pendingG {
		a.pendingG = false
		if msg.String() == "g" {
			a.scrollTop()
			return a, nil
		}
	}

	switch {
	case key.Matches(msg, keys.Back):
		if a.mode == ModeDetail {
			a.mode = ModeList
		} else {
			a.mode = ModeDetail
		}
		return a, nil
	case key.Matches(msg, keys.Conversation):
		a.mode = ModeConversation
		return a, nil
	case key.Matches(msg, keys.Checks):
		a.mode = ModeChecks
		return a, nil
	case key.Matches(msg, keys.Files):
		if !a.files.HasContent() && a.prDetail != nil {
			a.loading = true
			a.statusMsg = "Loading diff..."
			return a, a.fetchDiff(a.prDetail.Number)
		}
		a.mode = ModeFiles
		return a, nil
	case key.Matches(msg, keys.Browser):
		return a.openInBrowser()
	case msg.String() == "G":
		a.scrollBottom()
		return a, nil
	case msg.String() == "g":
		a.pendingG = true
		return a, nil
	}

	return a.updateActiveModel(msg)
}

func (a App) handleThemePickerKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back), key.Matches(msg, keys.Quit):
		applyTheme(themes[a.themeIdx])
		a.list.ApplyTheme(themes[a.themeIdx])
		a.mode = a.prevMode
		return a, nil
	case key.Matches(msg, keys.Enter):
		a.themeIdx = a.themeCursor
		a.statusMsg = "Theme: " + themes[a.themeIdx].Name
		a.mode = a.prevMode
		return a, nil
	case key.Matches(msg, keys.Down):
		a.themeCursor = (a.themeCursor + 1) % len(themes)
		applyTheme(themes[a.themeCursor])
		a.list.ApplyTheme(themes[a.themeCursor])
		return a, nil
	case key.Matches(msg, keys.Up):
		a.themeCursor = (a.themeCursor - 1 + len(themes)) % len(themes)
		applyTheme(themes[a.themeCursor])
		a.list.ApplyTheme(themes[a.themeCursor])
		return a, nil
	}
	return a, nil
}

func (a *App) scrollTop() {
	switch a.mode {
	case ModeDetail:
		a.detail.GotoTop()
	case ModeConversation:
		a.conversation.GotoTop()
	case ModeChecks:
		a.checks.GotoTop()
	case ModeFiles:
		a.files.GotoTop()
	case ModeHelp:
		a.help.GotoTop()
	}
}

func (a *App) scrollBottom() {
	switch a.mode {
	case ModeDetail:
		a.detail.GotoBottom()
	case ModeConversation:
		a.conversation.GotoBottom()
	case ModeChecks:
		a.checks.GotoBottom()
	case ModeFiles:
		a.files.GotoBottom()
	case ModeHelp:
		a.help.GotoBottom()
	}
}

func (a App) updateActiveModel(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch a.mode {
	case ModeList:
		a.list, cmd = a.list.Update(msg)
	case ModeDetail:
		a.detail, cmd = a.detail.Update(msg)
	case ModeConversation:
		a.conversation, cmd = a.conversation.Update(msg)
	case ModeChecks:
		a.checks, cmd = a.checks.Update(msg)
	case ModeFiles:
		a.files, cmd = a.files.Update(msg)
	case ModeHelp:
		a.help, cmd = a.help.Update(msg)
	}
	return a, cmd
}

func (a App) View() string {
	var content string
	switch a.mode {
	case ModeList:
		content = a.listView()
	case ModeDetail:
		content = a.detail.View()
	case ModeConversation:
		content = a.conversation.View()
	case ModeChecks:
		content = a.checks.View()
	case ModeFiles:
		content = a.files.View()
	case ModeHelp:
		content = a.help.View()
	case ModeThemePicker:
		content = a.themePickerView()
	}

	return content + "\n" + a.footerView()
}

func (a App) listView() string {
	header := headerStyle.Render(fmt.Sprintf("gh-tui · %s/%s", a.repo.Owner, a.repo.Name))
	filterLabel := a.filterLabel()

	gap := a.width - lipgloss.Width(header) - lipgloss.Width(filterLabel)
	if gap < 0 {
		gap = 0
	}

	return header + strings.Repeat(" ", gap) + filterLabel + "\n" + a.list.View()
}

func (a App) filterLabel() string {
	var state string
	if len(a.filter) == 0 {
		state = "All"
	} else {
		switch a.filter[0] {
		case "OPEN":
			state = openStyle.Render("Open")
		case "MERGED":
			state = mergedStyle.Render("Merged")
		case "CLOSED":
			state = closedStyle.Render("Closed")
		}
	}

	label := state + " PRs"
	if a.filterByUser && a.author != "" {
		label += " · @" + a.author
	}
	return headerStyle.Render(label)
}

func (a App) themePickerView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Select Theme") + "\n\n")
	for i, t := range themes {
		cursor := "  "
		if i == a.themeCursor {
			cursor = "> "
		}
		name := t.Name
		if i == a.themeIdx {
			name += " (current)"
		}
		if i == a.themeCursor {
			b.WriteString(authorStyle.Render(cursor + name) + "\n")
		} else {
			b.WriteString(metaStyle.Render(cursor+name) + "\n")
		}
	}
	return b.String()
}

func (a App) footerView() string {
	var help string
	switch a.mode {
	case ModeList:
		help = "j/k: nav  G/gg: end/top  Enter: view  o/m/x/a: filter  u: mine  b: browser  t: theme  ?: help  q: quit"
	case ModeDetail:
		help = "j/k: scroll  d/u: ½pg  G/gg: end/top  c: convo  s: checks  f: files  b: browser  Esc: back  q: quit"
	case ModeConversation, ModeChecks, ModeFiles:
		help = "j/k: scroll  d/u: ½pg  G/gg: end/top  c/s/f: switch  b: browser  Esc: back  q: quit"
	case ModeHelp:
		help = "j/k: scroll  ?: close  Esc: close  q: quit"
	case ModeThemePicker:
		help = "j/k: navigate  Enter: apply  Esc: cancel"
	}

	status := a.statusMsg
	if a.loading {
		status = "⟳ " + status
	}

	helpText := helpStyle.Render(help)
	statusText := statusStyle.Render(status)
	gap := a.width - lipgloss.Width(helpText) - lipgloss.Width(statusText)
	if gap < 0 {
		gap = 0
	}

	return helpText + strings.Repeat(" ", gap) + statusText
}

func (a App) openInBrowser() (tea.Model, tea.Cmd) {
	var number int
	if a.mode == ModeList {
		if pr := a.list.SelectedPR(); pr != nil {
			number = pr.Number
		}
	} else if a.prDetail != nil {
		number = a.prDetail.Number
	}
	if number == 0 {
		return a, nil
	}
	url := fmt.Sprintf("https://github.com/%s/%s/pull/%d", a.repo.Owner, a.repo.Name, number)
	a.statusMsg = fmt.Sprintf("Opening #%d in browser...", number)
	return a, func() tea.Msg {
		openURL(url)
		return nil
	}
}

func openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return
	}
	cmd.Start()
}
