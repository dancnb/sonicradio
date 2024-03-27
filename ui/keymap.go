package ui

import "github.com/charmbracelet/bubbles/key"

func newListKeymap() listKeymap {
	m := listKeymap{
		search: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "search"),
		),
		toNowPlaying: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "go to now playing"),
		),
		toFavorites: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "go to favorites"),
		),
		toBrowser: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "go to browser"),
		),
		nextTab: key.NewBinding(
			key.WithKeys("right", "l", "tab"),
			key.WithHelp("→/l/tab", "go to next tab"),
		),
		prevTab: key.NewBinding(
			key.WithKeys("left", "h", "shift+tab"),
			key.WithHelp("←/h/shift+tab", "go to prev tab"),
		),
	}
	return m
}

type listKeymap struct {
	search       key.Binding
	toNowPlaying key.Binding
	toFavorites  key.Binding
	toBrowser    key.Binding
	nextTab      key.Binding
	prevTab      key.Binding
}
