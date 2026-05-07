package ui

import (
	"context"
	"fmt"
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
	ModeComments
	ModeChecks
	ModeReviews
)

type App struct {
	mode     Mode
	list     ListModel
	detail   DetailModel
	comments CommentsModel
	checks   ChecksModel
	reviews  ReviewsModel

	client       *github.Client
	repo         github.Repo
	prDetail     *github.PRDetail
	pageInfo     github.PageInfo
	filter       []string
	author       string
	filterByUser bool

	width  int
	height int

	loading   bool
	statusMsg string
}

type prsLoadedMsg struct {
	result *github.PRListResult
}

type prDetailLoadedMsg struct {
	detail *github.PRDetail
}

type errMsg struct {
	err error
}

func NewApp(client *github.Client, repo github.Repo, author string) App {
	return App{
		mode:         ModeList,
		list:         NewListModel(),
		detail:       NewDetailModel(),
		comments:     NewCommentsModel(),
		checks:       NewChecksModel(),
		reviews:      NewReviewsModel(),
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
	return func() tea.Msg {
		result, err := client.ListPRs(context.Background(), repo, filter, 50, cursor)
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

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		contentH := msg.Height - 2
		a.list.SetSize(msg.Width, contentH)
		a.detail.SetSize(msg.Width, contentH)
		a.comments.SetSize(msg.Width, contentH)
		a.checks.SetSize(msg.Width, contentH)
		a.reviews.SetSize(msg.Width, contentH)
		return a, nil

	case prsLoadedMsg:
		a.loading = false
		a.pageInfo = msg.result.PageInfo
		prs := msg.result.PullRequests
		if a.filterByUser && a.author != "" {
			filtered := prs[:0:0]
			for _, pr := range prs {
				if strings.EqualFold(pr.Author, a.author) {
					filtered = append(filtered, pr)
				}
			}
			prs = filtered
		}
		a.list.SetPRs(prs)
		if a.filterByUser {
			a.statusMsg = fmt.Sprintf("%d/%d PRs (@%s)", len(prs), msg.result.TotalCount, a.author)
		} else {
			a.statusMsg = fmt.Sprintf("%d PRs", msg.result.TotalCount)
		}
		return a, nil

	case prDetailLoadedMsg:
		a.loading = false
		a.prDetail = msg.detail
		a.mode = ModeDetail
		a.detail.SetPR(msg.detail)
		a.comments.SetComments(msg.detail.Comments)
		a.checks.SetChecks(msg.detail.Checks)
		a.reviews.SetReviews(msg.detail.Reviews)
		return a, nil

	case errMsg:
		a.loading = false
		a.statusMsg = "Error: " + msg.err.Error()
		return a, nil

	case tea.KeyMsg:
		if key.Matches(msg, keys.Quit) {
			return a, tea.Quit
		}
		return a.handleKeyForMode(msg)
	}

	return a.updateActiveModel(msg)
}

func (a App) handleKeyForMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch a.mode {
	case ModeList:
		return a.handleListKeys(msg)
	case ModeDetail:
		return a.handleDetailKeys(msg)
	case ModeComments, ModeChecks, ModeReviews:
		return a.handleSubviewKeys(msg)
	}
	return a, nil
}

func (a App) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	}

	var cmd tea.Cmd
	a.list, cmd = a.list.Update(msg)
	return a, cmd
}

func (a App) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back):
		a.mode = ModeList
		return a, nil
	case key.Matches(msg, keys.Comments):
		a.mode = ModeComments
		return a, nil
	case key.Matches(msg, keys.Checks):
		a.mode = ModeChecks
		return a, nil
	case key.Matches(msg, keys.Reviews):
		a.mode = ModeReviews
		return a, nil
	}

	var cmd tea.Cmd
	a.detail, cmd = a.detail.Update(msg)
	return a, cmd
}

func (a App) handleSubviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Back) {
		a.mode = ModeDetail
		return a, nil
	}
	return a.updateActiveModel(msg)
}

func (a App) updateActiveModel(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch a.mode {
	case ModeList:
		a.list, cmd = a.list.Update(msg)
	case ModeDetail:
		a.detail, cmd = a.detail.Update(msg)
	case ModeComments:
		a.comments, cmd = a.comments.Update(msg)
	case ModeChecks:
		a.checks, cmd = a.checks.Update(msg)
	case ModeReviews:
		a.reviews, cmd = a.reviews.Update(msg)
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
	case ModeComments:
		content = a.comments.View()
	case ModeChecks:
		content = a.checks.View()
	case ModeReviews:
		content = a.reviews.View()
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

func (a App) footerView() string {
	var help string
	switch a.mode {
	case ModeList:
		help = "j/k: navigate  Enter: view  o: open  m: merged  x: closed  a: all  u: mine  q: quit"
	case ModeDetail:
		help = "j/k: scroll  Esc: back  c: comments  s: checks  r: reviews  q: quit"
	case ModeComments, ModeChecks, ModeReviews:
		help = "j/k: scroll  Esc: back  q: quit"
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
