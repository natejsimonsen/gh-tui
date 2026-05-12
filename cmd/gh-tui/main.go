package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/natejsimonsen/gh-tui/internal/debug"
	"github.com/natejsimonsen/gh-tui/internal/github"
	"github.com/natejsimonsen/gh-tui/internal/ui"
)

func main() {
	mockData := flag.String("mock-data", "", "path to mock data directory (bypasses GitHub API)")
	debugMode := flag.Bool("debug", false, "enable debug logging to stderr")
	flag.Usage = printUsage
	flag.Parse()

	if *debugMode {
		debug.Enable()
	}

	var client github.DataSource
	if *mockData != "" {
		mc, err := github.NewMockClient(*mockData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		client = mc
	} else {
		c, err := github.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		client = c
	}

	var repo github.Repo
	var err error
	if flag.NArg() > 0 {
		repo, err = github.ParseRepo(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
	} else if *mockData != "" {
		repo = github.Repo{Owner: "mock-org", Name: "mock-repo"}
	} else {
		repo, err = github.DetectRepo()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\nUsage: gh-tui [owner/repo]\n", err)
			os.Exit(1)
		}
	}

	author := "mockuser"
	if *mockData == "" {
		author = github.DetectUser()
	}

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
  gh-tui [flags] [owner/repo]

Flags:
  --mock-data <dir>  Use mock JSON data from directory (bypasses GitHub API)

If no repo is specified, detects from git remote origin.

Keys:
  j/k         Navigate up/down
  d/u         Half page down/up
  G/gg        Jump to bottom/top
  Enter       View PR detail
  Esc         Go back
  o/m/x/a     Filter: open/merged/closed/all
  u           Toggle my PRs
  r           Toggle review requests
  c/s/f       Conversation/checks/files
  b           Open in browser
  t           Theme picker
  ?           Help
  q           Quit`)
}
