package ui

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dancnb/sonicradio/browser"
	"github.com/dancnb/sonicradio/player"
)

func newStationDelegate(keymap *delegateKeyMap, p player.Player) *stationDelegate {
	help := []key.Binding{keymap.play}

	d := list.NewDefaultDelegate()
	d.ShortHelpFunc = func() []key.Binding {
		return help
	}
	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	descStyle := d.Styles.NormalDesc.Copy().PaddingLeft(4)

	return &stationDelegate{
		p:               p,
		keymap:          keymap,
		defaultDelegate: d,
		descStyle:       descStyle,
	}
}

type stationDelegate struct {
	p          player.Player
	nowPlaying *browser.Station
	keymap     *delegateKeyMap
	descStyle  lipgloss.Style

	defaultDelegate list.DefaultDelegate
}

func (d *stationDelegate) Height() int { return d.defaultDelegate.Height() }

func (d *stationDelegate) Spacing() int { return d.defaultDelegate.Spacing() }

func (d *stationDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	var title string
	station, ok := m.SelectedItem().(browser.Station)
	if ok {
		title = station.Name
	} else {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.keymap.play):
			if d.nowPlaying != nil && d.nowPlaying.Stationuuid == station.Stationuuid {
				slog.Debug("stopping", "id", d.nowPlaying.Stationuuid)
				d.nowPlaying = nil
				err := d.p.Stop()
				if err != nil {
					errMsg := fmt.Sprintf("error stopping station %s: %s", station.Name, err.Error())
					slog.Error(errMsg)
					return m.NewStatusMessage(statusErrMessageStyle("Error stopping " + title))
				}
				return m.NewStatusMessage(statusMessageStyle("Stopped playing " + title))
			}

			slog.Debug("playing", "id", station.Stationuuid)
			err := d.p.Play(station.URL)
			if err != nil {
				errMsg := fmt.Sprintf("error playing station %s: %s", station.Name, err.Error())
				slog.Error(errMsg)
				return m.NewStatusMessage(statusErrMessageStyle("Error playing " + title))
			}

			d.nowPlaying = &station
			return m.NewStatusMessage(statusMessageStyle("Playing " + title))

			// case key.Matches(msg, m.KeyMap.Quit)
		}
	}

	return nil
}

func (d *stationDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	s, ok := listItem.(browser.Station)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, s.Name)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}
	str = fn(str) + "\n"
	str += d.descStyle.Render(s.Description())

	fmt.Fprint(w, str)

	// d.DefaultDelegate.Render(w, m, index, listItem)
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		play: key.NewBinding(
			key.WithKeys(" ", "enter"),
			key.WithHelp("space/enter", "play/pause"),
		),
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
