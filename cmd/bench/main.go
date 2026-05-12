package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/natejsimonsen/gh-tui/internal/github"
	"github.com/natejsimonsen/gh-tui/internal/ui"
)

func main() {
	mockDir := flag.String("mock-data", "testdata/mock", "path to mock data directory")
	width := flag.Int("width", 120, "terminal width")
	height := flag.Int("height", 40, "terminal height")
	iterations := flag.Int("n", 5, "number of PRs to cycle through")
	flag.Parse()

	client, err := github.NewMockClient(*mockDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	repo := github.Repo{Owner: "mock-org", Name: "mock-repo"}
	app := ui.NewApp(client, repo, "mockuser")

	total := time.Now()

	t := time.Now()
	model, _ := app.Update(tea.WindowSizeMsg{Width: *width, Height: *height})
	app = model.(ui.App)
	fmt.Printf("%-40s %8s\n", "WindowSizeMsg", time.Since(t))

	t = time.Now()
	cmd := app.Init()
	msg := cmd()
	fmt.Printf("%-40s %8s\n", "Init (fetchPRs from mock)", time.Since(t))

	t = time.Now()
	model, cmd = app.Update(msg)
	app = model.(ui.App)
	fmt.Printf("%-40s %8s\n", "prsLoadedMsg → SetPRs", time.Since(t))

	if cmd != nil {
		drainCmds(cmd, &app)
	}

	t = time.Now()
	view := app.View()
	fmt.Printf("%-40s %8s  (%d bytes)\n", "List View()", time.Since(t), len(view))
	fmt.Println()

	prs, _ := client.ListPRs(context.Background(), repo, []string{"OPEN"}, 50, "")
	n := *iterations
	if n > len(prs.PullRequests) {
		n = len(prs.PullRequests)
	}

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	escKey := tea.KeyMsg{Type: tea.KeyEscape}
	runeKey := func(r rune) tea.KeyMsg {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
	}

	for i := 0; i < n; i++ {
		pr := prs.PullRequests[i]
		fmt.Printf("--- PR #%d: %s ---\n", pr.Number, pr.Title)

		t = time.Now()
		model, cmd = app.Update(enterKey)
		app = model.(ui.App)
		dur := time.Since(t)
		fmt.Printf("  %-38s %8s\n", "Enter (detail+conversation sync)", dur)

		if cmd != nil {
			t = time.Now()
			drainCmds(cmd, &app)
			dur = time.Since(t)
			fmt.Printf("  %-38s %8s\n", "async cmds after Enter", dur)
		}

		t = time.Now()
		view = app.View()
		dur = time.Since(t)
		fmt.Printf("  %-38s %8s  (%d bytes)\n", "Detail View()", dur, len(view))

		t = time.Now()
		model, cmd = app.Update(runeKey('c'))
		app = model.(ui.App)
		dur = time.Since(t)
		fmt.Printf("  %-38s %8s\n", "Press 'c' (conversation mode)", dur)

		t = time.Now()
		view = app.View()
		dur = time.Since(t)
		fmt.Printf("  %-38s %8s  (%d bytes)\n", "Conversation View()", dur, len(view))

		t = time.Now()
		model, cmd = app.Update(runeKey('s'))
		app = model.(ui.App)
		dur = time.Since(t)
		fmt.Printf("  %-38s %8s\n", "Press 's' (checks mode)", dur)

		t = time.Now()
		view = app.View()
		dur = time.Since(t)
		fmt.Printf("  %-38s %8s  (%d bytes)\n", "Checks View()", dur, len(view))

		t = time.Now()
		model, cmd = app.Update(runeKey('f'))
		app = model.(ui.App)
		dur = time.Since(t)
		fmt.Printf("  %-38s %8s\n", "Press 'f' (files mode + fetch)", dur)

		if cmd != nil {
			t = time.Now()
			drainCmds(cmd, &app)
			dur = time.Since(t)
			fmt.Printf("  %-38s %8s\n", "async (diff fetch+render)", dur)
		}

		t = time.Now()
		view = app.View()
		dur = time.Since(t)
		fmt.Printf("  %-38s %8s  (%d bytes)\n", "Files View()", dur, len(view))

		t = time.Now()
		shiftR := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}}
		model, cmd = app.Update(shiftR)
		app = model.(ui.App)
		dur = time.Since(t)
		fmt.Printf("  %-38s %8s\n", "Press 'R' (review mode)", dur)

		if cmd != nil {
			t = time.Now()
			drainCmds(cmd, &app)
			dur = time.Since(t)
			fmt.Printf("  %-38s %8s\n", "async (review diff load)", dur)
		}

		t = time.Now()
		view = app.View()
		dur = time.Since(t)
		fmt.Printf("  %-38s %8s  (%d bytes)\n", "Review View()", dur, len(view))

		model, _ = app.Update(escKey)
		app = model.(ui.App)
		model, _ = app.Update(escKey)
		app = model.(ui.App)

		model, _ = app.Update(runeKey('j'))
		app = model.(ui.App)

		fmt.Println()
	}

	fmt.Printf("Total elapsed: %s\n", time.Since(total))
}

func drainCmds(cmd tea.Cmd, app *ui.App) {
	if cmd == nil {
		return
	}
	msg := cmd()
	if msg == nil {
		return
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			drainCmds(c, app)
		}
		return
	}
	model, nextCmd := app.Update(msg)
	*app = model.(ui.App)
	if nextCmd != nil {
		drainCmds(nextCmd, app)
	}
}
