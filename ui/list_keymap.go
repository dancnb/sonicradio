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
			key.WithKeys("tab"),
			key.WithHelp("tab", "go to next tab"),
		),
		prevTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "go to prev tab"),
		),
		favoritesTab: key.NewBinding(
			key.WithKeys("F"),
			key.WithHelp("F", "go to favorites tab"),
		),
		browseTab: key.NewBinding(
			key.WithKeys("B"),
			key.WithHelp("B", "go to browse tab"),
		),
		historyTab: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "go to history tab"),
		),
		settingsTab: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "go to settings tab"),
		),
		stationView: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "change view"),
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
	favoritesTab key.Binding
	browseTab    key.Binding
	historyTab   key.Binding
	settingsTab  key.Binding
	stationView  key.Binding
	digits       []key.Binding
	digitHelp    key.Binding
}

func (k *listKeymap) setEnabled(v bool) {
	k.search.SetEnabled(v)
	k.toNowPlaying.SetEnabled(v)
	k.nextTab.SetEnabled(v)
	k.prevTab.SetEnabled(v)
	k.favoritesTab.SetEnabled(v)
	k.browseTab.SetEnabled(v)
	k.historyTab.SetEnabled(v)
	k.settingsTab.SetEnabled(v)
	k.stationView.SetEnabled(v)
	for i := range k.digits {
		k.digits[i].SetEnabled(v)
	}
}
