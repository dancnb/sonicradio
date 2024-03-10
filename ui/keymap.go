package ui

import "github.com/charmbracelet/bubbles/key"

func newKeymap() keymap {
	m := keymap{
		search: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "search station"),
		),
	}
	return m
}

type keymap struct {
	search key.Binding
}
