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
		nextTab: key.NewBinding(
			key.WithKeys("right", "l", "tab"),
			key.WithHelp("→/l/tab", "go to next tab"),
		),
		prevTab: key.NewBinding(
			key.WithKeys("left", "h", "shift+tab"),
			key.WithHelp("←/h/shift+tab", "go to prev tab"),
		),
		digits: []key.Binding{
			key.NewBinding(key.WithKeys("1")),
			key.NewBinding(key.WithKeys("2")),
			key.NewBinding(key.WithKeys("3")),
			key.NewBinding(key.WithKeys("4")),
			key.NewBinding(key.WithKeys("5")),
			key.NewBinding(key.WithKeys("6")),
			key.NewBinding(key.WithKeys("7")),
			key.NewBinding(key.WithKeys("8")),
			key.NewBinding(key.WithKeys("9")),
			key.NewBinding(key.WithKeys("0")),
		},
		digitHelp: key.NewBinding(
			key.WithKeys("#"),
			key.WithHelp("1..", "go to number #"),
		),
	}
	return m
}

type listKeymap struct {
	search       key.Binding
	toNowPlaying key.Binding
	nextTab      key.Binding
	prevTab      key.Binding
	digits       []key.Binding
	digitHelp    key.Binding
}

func (k *listKeymap) setEnabled(v bool) {
	k.search.SetEnabled(v)
	k.toNowPlaying.SetEnabled(v)
	k.nextTab.SetEnabled(v)
	k.prevTab.SetEnabled(v)
	for i := range k.digits {
		k.digits[i].SetEnabled(v)
	}
}
