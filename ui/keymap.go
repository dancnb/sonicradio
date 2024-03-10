package ui

import "github.com/charmbracelet/bubbles/key"

func newKeymap() keymap {
	m := keymap{
		search: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "search"),
		),
		toNowPlaying: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "now playing"),
		),
	}
	return m
}

type keymap struct {
	search       key.Binding
	toNowPlaying key.Binding
}
