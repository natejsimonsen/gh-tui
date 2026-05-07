# gh-tui

A minimal terminal UI for GitHub pull requests.

Browse and track pull requests without leaving the terminal. Built with Go using [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss), and [Glamour](https://github.com/charmbracelet/glamour).

## Features

- List pull requests with status, CI, labels, and reviewers at a glance
- View PR details with rendered markdown descriptions
- Browse comments, CI checks, and code reviews
- Filter by state: open, merged, closed, or all
- Auto-detects repository from git remote origin
- Vim-style modal navigation

## Install

```
go install github.com/natejsimonsen/gh-tui/cmd/gh-tui@latest
```

## Requirements

- Go 1.23+
- GitHub auth via `gh` CLI (`gh auth login`) or `GITHUB_TOKEN` env var

## Usage

```
gh-tui [owner/repo]
```

Omit the argument to auto-detect from the current git repo's origin remote.

## Keys

| Key | Action |
|-----|--------|
| `j/k` | Navigate / scroll |
| `Enter` | View PR detail |
| `Esc` | Go back |
| `o` | Show open PRs |
| `m` | Show merged PRs |
| `x` | Show closed PRs |
| `a` | Show all PRs |
| `c` | View comments (from detail) |
| `s` | View CI checks (from detail) |
| `r` | View reviews (from detail) |
| `q` | Quit |

## License

MIT
