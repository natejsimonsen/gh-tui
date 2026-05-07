package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Back     key.Binding
	Quit     key.Binding
	Help     key.Binding
	Open     key.Binding
	Merged   key.Binding
	Closed   key.Binding
	All      key.Binding
	Comments key.Binding
	Checks   key.Binding
	Reviews  key.Binding
}

var keys = keyMap{
	Up:       key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "up")),
	Down:     key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "down")),
	Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Back:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Open:     key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open")),
	Merged:   key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "merged")),
	Closed:   key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "closed")),
	All:      key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "all")),
	Comments: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "comments")),
	Checks:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "checks")),
	Reviews:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reviews")),
}
