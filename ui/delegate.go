package ui

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancnb/sonicradio/browser"
)

func newItemDelegate() list.DefaultDelegate {
	keys := newDelegateKeyMap()
	help := []key.Binding{keys.play}

	d := list.NewDefaultDelegate()
	d.UpdateFunc = udpateStation(keys)
	d.ShortHelpFunc = func() []key.Binding {
		return help
	}
	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

func udpateStation(keys *delegateKeyMap) func(tea.Msg, *list.Model) tea.Cmd {
	return func(msg tea.Msg, m *list.Model) tea.Cmd {
		var title string
		sta, ok := m.SelectedItem().(browser.Station)
		if ok {
			title = sta.Title()
		} else {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.play):
				slog.Info("Playing station " + sta.Name)
				return m.NewStatusMessage(statusMessageStyle("Playing " + title))

				// case key.Matches(msg, keys.remove):
				// 	index := m.Index()
				// 	m.RemoveItem(index)
				// 	if len(m.Items()) == 0 {
				// 		keys.remove.SetEnabled(false)
				// 	}
				// 	return m.NewStatusMessage(statusMessageStyle("Deleted " + title))
				// }
			}

		}
		return nil
	}
}

type delegateKeyMap struct {
	play key.Binding
}

// Additional short help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.play,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.play,
		},
	}
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		play: key.NewBinding(
			key.WithKeys(" ", "enter"),
			key.WithHelp("space/enter", "play/pause"),
		),
	}
}
