package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/natejsimonsen/gh-tui/internal/github"
	"github.com/natejsimonsen/gh-tui/internal/ui"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		printUsage()
		os.Exit(0)
	}

	client, err := github.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	var repo github.Repo
	if len(os.Args) > 1 {
		repo, err = github.ParseRepo(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
	} else {
		repo, err = github.DetectRepo()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\nUsage: gh-tui [owner/repo]\n", err)
			os.Exit(1)
		}
	}

	author := github.DetectUser()
	app := ui.NewApp(client, repo, author)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`gh-tui - A minimal terminal UI for GitHub pull requests

Usage:
  gh-tui [owner/repo]

If no repo is specified, detects from git remote origin.

Keys:
  j/k         Navigate up/down
  Enter       View PR detail
  Esc         Go back
  o/m/x/a     Filter: open/merged/closed/all
  c/s/r       Comments/checks/reviews (in detail view)
  q           Quit`)
}
