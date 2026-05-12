package ui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/natejsimonsen/gh-tui/internal/debug"
	"github.com/natejsimonsen/gh-tui/internal/github"
)

type Mode int

const (
	ModeList Mode = iota
	ModeDetail
	ModeConversation
	ModeChecks
	ModeFiles
	ModeReview
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
	review       ReviewModel
	help         HelpModel

	client         github.DataSource
	repo           github.Repo
	prDetail       *github.PRDetail
	pageInfo       github.PageInfo
	filter         []string
	author         string
	filterByUser   bool
	filterByReview bool

	width  int
	height int

	loading        bool
	statusMsg      string
	pendingG       bool
	pendingReview  bool
	themeIdx       int
	themeCursor    int
	detailCache    map[int]*github.PRDetail
	prefetchQueue  []int
	prefetchActive int
}

type prsLoadedMsg struct {
	result *github.PRListResult
}

type prDetailLoadedMsg struct {
	detail *github.PRDetail
}

type prPrefetchedMsg struct {
	number int
	detail *github.PRDetail
}

type diffLoadedMsg struct {
	diff string
}

type errMsg struct {
	err error
}

func NewApp(client github.DataSource, repo github.Repo, author string) App {
	return App{
		mode:         ModeList,
		list:         NewListModel(),
		detail:       NewDetailModel(),
		conversation: NewConversationModel(),
		checks:       NewChecksModel(),
		files:        NewDiffModel(),
		review:       NewReviewModel(),
		help:         NewHelpModel(),
		client:       client,
		repo:         repo,
		filter:       []string{"OPEN"},
		author:       author,
		filterByUser: author != "",
		detailCache:  make(map[int]*github.PRDetail),
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
	filterByReview := a.filterByReview
	author := a.author
	return func() tea.Msg {
		debug.Println("fetchPRs: API call start")
		start := time.Now()
		var result *github.PRListResult
		var err error
		if filterByReview && author != "" {
			result, err = client.SearchReviewRequested(context.Background(), repo, author, filter, 50, cursor)
		} else if filterByUser && author != "" {
			result, err = client.SearchPRs(context.Background(), repo, author, filter, 50, cursor)
		} else {
			result, err = client.ListPRs(context.Background(), repo, filter, 50, cursor)
		}
		if err != nil {
			debug.Printf("fetchPRs: API error %dms: %v", time.Since(start).Milliseconds(), err)
			return errMsg{err}
		}
		debug.Printf("fetchPRs: API done %dms (%d PRs)", time.Since(start).Milliseconds(), len(result.PullRequests))
		return prsLoadedMsg{result}
	}
}

func (a App) fetchPRDetail(number int) tea.Cmd {
	client := a.client
	repo := a.repo
	return func() tea.Msg {
		debug.Printf("fetchPRDetail: API call start #%d", number)
		start := time.Now()
		detail, err := client.GetPRDetail(context.Background(), repo, number)
		if err != nil {
			debug.Printf("fetchPRDetail: API error #%d %dms: %v", number, time.Since(start).Milliseconds(), err)
			return errMsg{err}
		}
		debug.Printf("fetchPRDetail: API done #%d %dms (body=%d bytes, comments=%d, reviews=%d)",
			number, time.Since(start).Milliseconds(), len(detail.Body), len(detail.Comments), len(detail.Reviews))
		return prDetailLoadedMsg{detail}
	}
}

func (a App) fetchDiff(number int) tea.Cmd {
	client := a.client
	repo := a.repo
	return func() tea.Msg {
		debug.Printf("fetchDiff: API call start #%d", number)
		start := time.Now()
		diff, err := client.GetPRDiff(context.Background(), repo, number)
		if err != nil {
			debug.Printf("fetchDiff: API error #%d %dms: %v", number, time.Since(start).Milliseconds(), err)
			return errMsg{err}
		}
		debug.Printf("fetchDiff: API done #%d %dms (%d bytes)", number, time.Since(start).Milliseconds(), len(diff))
		return diffLoadedMsg{diff}
	}
}

func (a App) submitReview(number int, input github.ReviewInput) tea.Cmd {
	client := a.client
	repo := a.repo
	return func() tea.Msg {
		debug.Printf("submitReview: start #%d event=%s", number, input.Event)
		start := time.Now()
		err := client.SubmitReview(context.Background(), repo, number, input)
		if err != nil {
			debug.Printf("submitReview: error #%d %dms: %v", number, time.Since(start).Milliseconds(), err)
		} else {
			debug.Printf("submitReview: done #%d %dms", number, time.Since(start).Milliseconds())
		}
		return reviewSubmittedMsg{err: err}
	}
}

const maxPrefetch = 5

func (a App) prefetchPR(number int) tea.Cmd {
	client := a.client
	repo := a.repo
	return func() tea.Msg {
		debug.Printf("prefetch: start #%d", number)
		start := time.Now()
		detail, err := client.GetPRDetail(context.Background(), repo, number)
		if err != nil {
			debug.Printf("prefetch: error #%d %dms: %v", number, time.Since(start).Milliseconds(), err)
			return prPrefetchedMsg{number: number}
		}
		debug.Printf("prefetch: done #%d %dms", number, time.Since(start).Milliseconds())
		return prPrefetchedMsg{number: number, detail: detail}
	}
}

func (a *App) startPrefetch() tea.Cmd {
	var cmds []tea.Cmd
	for a.prefetchActive < maxPrefetch && len(a.prefetchQueue) > 0 {
		number := a.prefetchQueue[0]
		a.prefetchQueue = a.prefetchQueue[1:]
		if _, ok := a.detailCache[number]; ok {
			continue
		}
		a.prefetchActive++
		cmds = append(cmds, a.prefetchPR(number))
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (a *App) enqueuePrefetch(prs []github.PullRequest) tea.Cmd {
	a.prefetchQueue = nil
	a.prefetchActive = 0
	for _, pr := range prs {
		if _, ok := a.detailCache[pr.Number]; !ok {
			a.prefetchQueue = append(a.prefetchQueue, pr.Number)
		}
	}
	debug.Printf("enqueuePrefetch: %d PRs queued", len(a.prefetchQueue))
	return a.startPrefetch()
}

func (a *App) loadDetail(detail *github.PRDetail) {
	if a.prDetail == nil || a.prDetail.Number != detail.Number {
		a.review.Clear()
	}
	a.prDetail = detail
	a.mode = ModeDetail
	a.detail.SetPR(detail)
	a.conversation.SetData(detail.Comments, detail.Reviews)
	a.checks.SetChecks(detail.Checks)
	a.files.Clear()
}

func (a App) isEditing() bool {
	return a.mode == ModeReview && a.review.IsEditing()
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
		a.review.SetSize(msg.Width, contentH)
		a.help.SetSize(msg.Width, contentH)
		return a, nil

	case prsLoadedMsg:
		debug.Printf("prsLoadedMsg: received (%d PRs), existing cache=%d", len(msg.result.PullRequests), len(a.detailCache))
		a.loading = false
		a.pageInfo = msg.result.PageInfo
		a.list.SetPRs(msg.result.PullRequests)
		if a.filterByReview {
			a.statusMsg = fmt.Sprintf("%d reviews (@%s)", msg.result.TotalCount, a.author)
		} else if a.filterByUser {
			a.statusMsg = fmt.Sprintf("%d PRs (@%s)", msg.result.TotalCount, a.author)
		} else {
			a.statusMsg = fmt.Sprintf("%d PRs", msg.result.TotalCount)
		}
		return a, a.enqueuePrefetch(msg.result.PullRequests)

	case prPrefetchedMsg:
		if a.prefetchActive > 0 {
			a.prefetchActive--
		}
		if msg.detail != nil {
			a.detailCache[msg.number] = msg.detail
			debug.Printf("prPrefetchedMsg: cached #%d (total=%d queue=%d active=%d)",
				msg.number, len(a.detailCache), len(a.prefetchQueue), a.prefetchActive)
		}
		return a, a.startPrefetch()

	case prDetailLoadedMsg:
		debug.Printf("prDetailLoadedMsg: #%d received", msg.detail.Number)
		a.loading = false
		a.detailCache[msg.detail.Number] = msg.detail
		a.loadDetail(msg.detail)
		return a, nil

	case diffLoadedMsg:
		debug.Printf("diffLoadedMsg: received (%d bytes)", len(msg.diff))
		a.files.SetDiff(msg.diff)
		a.statusMsg = "Rendering diff..."
		cmd := renderDiffAsync(msg.diff, a.width)

		if a.pendingReview && a.prDetail != nil {
			a.pendingReview = false
			a.review.SetData(a.prDetail.Number, a.prDetail.Title, msg.diff)
			a.mode = ModeReview
			a.loading = false
			a.statusMsg = ""
		}

		return a, cmd

	case diffRenderedMsg:
		debug.Printf("diffRenderedMsg: %d files", msg.fileCount)
		a.loading = false
		a.files.SetRendered(msg)
		if a.mode != ModeReview {
			a.statusMsg = "Diff loaded"
		}
		return a, nil

	case reviewWantsSubmitMsg:
		if a.prDetail == nil {
			return a, nil
		}
		input := github.ReviewInput{
			Event: reviewActionEvents[msg.action],
			Body:  msg.body,
		}
		for _, f := range a.review.files {
			for _, c := range f.comments {
				input.Comments = append(input.Comments, github.ReviewCommentInput{
					Path: f.path,
					Body: c,
				})
			}
		}
		a.loading = true
		a.statusMsg = "Submitting review..."
		return a, a.submitReview(a.prDetail.Number, input)

	case reviewSubmittedMsg:
		a.loading = false
		if msg.err != nil {
			a.statusMsg = "Review error: " + msg.err.Error()
			return a, nil
		}
		a.statusMsg = "Review submitted!"
		a.review.Clear()
		a.mode = ModeDetail
		return a, nil

	case errMsg:
		a.loading = false
		a.statusMsg = "Error: " + msg.err.Error()
		return a, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return a, tea.Quit
		}
		if key.Matches(msg, keys.Quit) && a.mode != ModeThemePicker && !a.isEditing() {
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

	if a.mode == ModeReview {
		if !a.isEditing() {
			if key.Matches(msg, keys.Help) {
				a.prevMode = a.mode
				a.mode = ModeHelp
				return a, nil
			}
		}
		return a.handleReviewKeys(msg)
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
	case ModeDetail, ModeConversation, ModeChecks, ModeFiles:
		return a.handleViewKeys(msg)
	default:
		return a, nil
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
			if cached, ok := a.detailCache[pr.Number]; ok {
				debug.Printf("Enter: CACHE HIT #%d %q", pr.Number, pr.Title)
				a.loadDetail(cached)
				return a, nil
			}
			debug.Printf("Enter: CACHE MISS #%d %q", pr.Number, pr.Title)
			a.loading = true
			a.statusMsg = fmt.Sprintf("Loading #%d...", pr.Number)
			return a, a.fetchPRDetail(pr.Number)
		}
	case key.Matches(msg, keys.Open):
		a.filter = []string{"OPEN"}
		a.pageInfo = github.PageInfo{}
		a.detailCache = make(map[int]*github.PRDetail)
		a.loading = true
		a.statusMsg = "Loading open PRs..."
		return a, a.fetchPRs()
	case key.Matches(msg, keys.Merged):
		a.filter = []string{"MERGED"}
		a.pageInfo = github.PageInfo{}
		a.detailCache = make(map[int]*github.PRDetail)
		a.loading = true
		a.statusMsg = "Loading merged PRs..."
		return a, a.fetchPRs()
	case key.Matches(msg, keys.Closed):
		a.filter = []string{"CLOSED"}
		a.pageInfo = github.PageInfo{}
		a.detailCache = make(map[int]*github.PRDetail)
		a.loading = true
		a.statusMsg = "Loading closed PRs..."
		return a, a.fetchPRs()
	case key.Matches(msg, keys.All):
		a.filter = nil
		a.pageInfo = github.PageInfo{}
		a.detailCache = make(map[int]*github.PRDetail)
		a.loading = true
		a.statusMsg = "Loading all PRs..."
		return a, a.fetchPRs()
	case key.Matches(msg, keys.Mine):
		a.filterByReview = false
		a.filterByUser = !a.filterByUser
		a.pageInfo = github.PageInfo{}
		a.detailCache = make(map[int]*github.PRDetail)
		a.loading = true
		if a.filterByUser {
			a.statusMsg = fmt.Sprintf("Loading @%s PRs...", a.author)
		} else {
			a.statusMsg = "Loading all authors..."
		}
		return a, a.fetchPRs()
	case key.Matches(msg, keys.ReviewRequests):
		a.filterByUser = false
		a.filterByReview = !a.filterByReview
		a.pageInfo = github.PageInfo{}
		a.detailCache = make(map[int]*github.PRDetail)
		a.loading = true
		if a.filterByReview {
			a.statusMsg = fmt.Sprintf("Loading reviews for @%s...", a.author)
		} else {
			a.statusMsg = "Loading all PRs..."
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
			switch a.mode {
			case ModeDetail:
				a.detail.GotoTop()
			case ModeConversation:
				a.conversation.GotoTop()
			case ModeChecks:
				a.checks.GotoTop()
			case ModeFiles:
				a.files.GotoTop()
			}
			return a, nil
		}
	}

	switch {
	case key.Matches(msg, keys.Back):
		switch a.mode {
		case ModeConversation, ModeChecks, ModeFiles:
			a.mode = ModeDetail
		default:
			a.mode = ModeList
		}
		return a, nil

	case key.Matches(msg, keys.Conversation):
		a.mode = ModeConversation
		return a, nil

	case key.Matches(msg, keys.Checks):
		a.mode = ModeChecks
		return a, nil

	case key.Matches(msg, keys.Files):
		if a.prDetail != nil && !a.files.HasContent() && a.files.raw == "" {
			a.loading = true
			a.statusMsg = "Loading diff..."
			a.mode = ModeFiles
			return a, a.fetchDiff(a.prDetail.Number)
		}
		a.mode = ModeFiles
		return a, nil

	case msg.String() == "R":
		return a.enterReviewMode()

	case key.Matches(msg, keys.Browser):
		return a.openInBrowser()

	case msg.String() == "G":
		switch a.mode {
		case ModeDetail:
			a.detail.GotoBottom()
		case ModeConversation:
			a.conversation.GotoBottom()
		case ModeChecks:
			a.checks.GotoBottom()
		case ModeFiles:
			a.files.GotoBottom()
		}
		return a, nil

	case msg.String() == "g":
		a.pendingG = true
		return a, nil
	}

	var cmd tea.Cmd
	switch a.mode {
	case ModeDetail:
		a.detail, cmd = a.detail.Update(msg)
	case ModeConversation:
		a.conversation, cmd = a.conversation.Update(msg)
	case ModeChecks:
		a.checks, cmd = a.checks.Update(msg)
	case ModeFiles:
		a.files, cmd = a.files.Update(msg)
	}
	return a, cmd
}

func (a App) handleReviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEscape && a.review.state == reviewTree {
		a.mode = ModeDetail
		return a, nil
	}

	var cmd tea.Cmd
	a.review, cmd = a.review.Update(msg)
	return a, cmd
}

func (a App) enterReviewMode() (tea.Model, tea.Cmd) {
	if a.prDetail == nil {
		return a, nil
	}
	if a.review.prNumber == a.prDetail.Number && len(a.review.files) > 0 {
		a.mode = ModeReview
		return a, nil
	}
	if a.files.raw != "" {
		a.review.SetData(a.prDetail.Number, a.prDetail.Title, a.files.raw)
		a.mode = ModeReview
		return a, nil
	}
	a.pendingReview = true
	a.loading = true
	a.statusMsg = "Loading diff for review..."
	return a, a.fetchDiff(a.prDetail.Number)
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
	case ModeReview:
		a.review, cmd = a.review.Update(msg)
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
	case ModeReview:
		content = a.review.View()
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
	if a.filterByReview && a.author != "" {
		label += " · reviews:@" + a.author
	} else if a.filterByUser && a.author != "" {
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
		help = "o/m/x/a  u:mine  r:reviews  ?:help  q:quit"
	case ModeDetail:
		help = "c:conv  s:checks  f:files  R:review  Esc:back  ?:help"
	case ModeConversation:
		help = "s:checks  f:files  R:review  Esc:back  ?:help"
	case ModeChecks:
		help = "c:conv  f:files  R:review  Esc:back  ?:help"
	case ModeFiles:
		help = "c:conv  s:checks  R:review  Esc:back  ?:help"
	case ModeReview:
		help = a.review.FooterHelp()
	case ModeHelp:
		help = "Esc:close"
	case ModeThemePicker:
		help = "Enter:apply  Esc:cancel"
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
