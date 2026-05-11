package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up              key.Binding
	Down            key.Binding
	Enter           key.Binding
	Back            key.Binding
	Quit            key.Binding
	Help            key.Binding
	Open            key.Binding
	Merged          key.Binding
	Closed          key.Binding
	All             key.Binding
	Mine            key.Binding
	ReviewRequests  key.Binding
	Conversation    key.Binding
	Checks          key.Binding
	Files           key.Binding
	Browser         key.Binding
}

var keys = keyMap{
	Up:             key.NewBinding(key.WithKeys("k", "up")),
	Down:           key.NewBinding(key.WithKeys("j", "down")),
	Enter:          key.NewBinding(key.WithKeys("enter")),
	Back:           key.NewBinding(key.WithKeys("esc")),
	Quit:           key.NewBinding(key.WithKeys("q", "ctrl+c")),
	Help:           key.NewBinding(key.WithKeys("?")),
	Open:           key.NewBinding(key.WithKeys("o")),
	Merged:         key.NewBinding(key.WithKeys("m")),
	Closed:         key.NewBinding(key.WithKeys("x")),
	All:            key.NewBinding(key.WithKeys("a")),
	Mine:           key.NewBinding(key.WithKeys("u")),
	ReviewRequests: key.NewBinding(key.WithKeys("r")),
	Conversation:   key.NewBinding(key.WithKeys("c")),
	Checks:         key.NewBinding(key.WithKeys("s")),
	Files:          key.NewBinding(key.WithKeys("f")),
	Browser:        key.NewBinding(key.WithKeys("b")),
}
