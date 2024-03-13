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
		toFavorites: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "favorites"),
		),
		toBrowser: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "browser"),
		),
	}
	return m
}

type keymap struct {
	search       key.Binding
	toNowPlaying key.Binding
	toFavorites  key.Binding
	toBrowser    key.Binding
}
