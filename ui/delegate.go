package ui

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancnb/sonicradio/browser"
	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player"
)

func newStationDelegate(cfg *config.Value, p player.Player) *stationDelegate {
	keymap := newDelegateKeyMap()

	d := list.NewDefaultDelegate()
	// d.ShortHelpFunc = func() []key.Binding {
	// 	return []key.Binding{keymap.play}
	// }
	// d.FullHelpFunc = func() [][]key.Binding {
	// 	return [][]key.Binding{{keymap.play, keymap.info, keymap.toggleFavorite}}
	// }

	return &stationDelegate{
		p:               p,
		cfg:             cfg,
		keymap:          keymap,
		defaultDelegate: d,
	}
}

type stationDelegate struct {
	p          player.Player
	cfg        *config.Value
	nowPlaying *browser.Station
	keymap     *delegateKeyMap

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
		}
	}

	return nil
}

func (d *stationDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	s, ok := listItem.(browser.Station)
	if !ok {
		return
	}
	name := s.Name
	if d.cfg.IsFavorite(s.Stationuuid) {
		name += favChar
	}

	itStyle := itemStyle
	descStyle := descStyle
	if index == m.Index() {
		itStyle = selItemStyle
		descStyle = selDescStyle
	}
	var res strings.Builder
	prefix := fmt.Sprintf("%d. ", index+1)
	if index+1 < 10 {
		prefix = fmt.Sprintf("   %s", prefix)
	} else if index+1 < 100 {
		prefix = fmt.Sprintf("  %s", prefix)
	} else if index+1 < 1000 {
		prefix = fmt.Sprintf(" %s", prefix)
	}

	if d.nowPlaying != nil && d.nowPlaying.Stationuuid == s.Stationuuid {
		res.WriteString(itStyle.Render(prefix))

		npItStyle := nowPlayingStyle
		npDescStyle := nowPlayingDescStyle
		if index == m.Index() {
			npItStyle = selNowPlayingStyle
			npDescStyle = selNowPlayingDescStyle
		}

		res.WriteString(npItStyle.Render(name))
		w := m.Width()
		hFill := max(w-len(prefix)-len(name), 0)
		res.WriteString(npItStyle.Render(strings.Repeat(" ", hFill)))
		res.WriteString("\n")
		res.WriteString(descStyle.Render(strings.Repeat(" ", len(prefix))))
		res.WriteString(npDescStyle.Render(s.Description()))
		hFill = max(w-len(prefix)-len(s.Description()), 0)
		res.WriteString(npItStyle.Render(strings.Repeat(" ", hFill)))
	} else {
		res.WriteString(itStyle.Render(prefix + name))
		res.WriteString("\n")
		res.WriteString(descStyle.Render(strings.Repeat(" ", len(prefix)) + s.Description()))
	}

	str := res.String()
	if index == m.Index() {
		str = selectedBorderStyle.Render(str)
	}
	fmt.Fprint(w, str)
}

// Additional short help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d *stationDelegate) ShortHelp() []key.Binding {
	return []key.Binding{
		d.keymap.play,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d *stationDelegate) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.keymap.play, d.keymap.info, d.keymap.toggleFavorite,
		},
	}
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		play: key.NewBinding(
			key.WithKeys(" ", "enter"),
			key.WithHelp("space/enter", "play/pause"),
		),
		info: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "info"),
		),
		toggleFavorite: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "toggle favorite"),
		),
	}
}

type delegateKeyMap struct {
	play           key.Binding
	info           key.Binding
	toggleFavorite key.Binding
}
