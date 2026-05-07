# gh-tui

A minimal terminal UI for GitHub pull requests.

Browse, review, and track pull requests without leaving the terminal. Built with Go using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework and the GitHub API.

## Features (planned)

- List pull requests for any repository
- View PR details: description, comments, review status, CI checks
- Navigate between PRs with keyboard shortcuts
- Minimal, distraction-free interface

## Requirements

- Go 1.23+
- A GitHub personal access token (or `gh` CLI auth)

## Usage

```
gh-tui owner/repo
```

## License

MIT
