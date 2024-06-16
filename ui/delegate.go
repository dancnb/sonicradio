package ui

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancnb/sonicradio/browser"
	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player"
)

const startWaitMillis = 500 * 3

func newStationDelegate(cfg *config.Value, p player.Player) *stationDelegate {
	keymap := newDelegateKeyMap()

	d := list.NewDefaultDelegate()

	return &stationDelegate{
		player:          p,
		cfg:             cfg,
		keymap:          keymap,
		defaultDelegate: d,
	}
}

type stationDelegate struct {
	player      player.Player
	cfg         *config.Value
	prevPlaying *browser.Station
	currPlaying *browser.Station
	keymap      *delegateKeyMap

	defaultDelegate list.DefaultDelegate
}

func (d *stationDelegate) Height() int { return d.defaultDelegate.Height() }

func (d *stationDelegate) Spacing() int { return d.defaultDelegate.Spacing() }

func (d *stationDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	logTeaMsg(msg, "update stationDelegate")
	selStation, ok := m.SelectedItem().(browser.Station)
	if !ok {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.keymap.toggleFavorite):
			added := d.cfg.ToggleFavorite(selStation.Stationuuid)
			return func() tea.Msg { return toggleFavoriteMsg{added, selStation} }
		}
	}

	return nil
}

func (d *stationDelegate) playCmd(s *browser.Station) tea.Cmd {
	return func() tea.Msg {
		err := d.playStation(*s)
		if err != nil {
			return playRespMsg{fmt.Sprintf("Could not start playback for %s (%s)!", s.Name, s.URL)}
		}

		return playRespMsg{}
	}
}

func (d *stationDelegate) playStation(station browser.Station) error {
	slog.Debug("playing", "id", station.Stationuuid)
	err := d.player.Play(station.URL)
	if err != nil {
		errMsg := fmt.Sprintf("error playing station %s: %s", station.Name, err.Error())
		slog.Error(errMsg)
		return errors.New(errMsg)
	}
	d.prevPlaying = d.currPlaying
	d.currPlaying = &station
	return nil
}

func (d *stationDelegate) stopStation(station browser.Station) (wasPlaying bool, err error) {
	if d.currPlaying != nil && d.currPlaying.Stationuuid == station.Stationuuid {
		slog.Debug("stopping", "id", d.currPlaying.Stationuuid)
		d.prevPlaying = &station
		d.currPlaying = nil
		err := d.player.Stop()
		if err != nil {
			errMsg := fmt.Sprintf("error stopping station %s: %s", station.Name, err.Error())
			slog.Error(errMsg)
			return true, errors.New(errMsg)
		}
		return true, nil
	}
	return false, nil
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

	isSel := index == m.Index()
	isCurr := d.currPlaying != nil && d.currPlaying.Stationuuid == s.Stationuuid
	isPrev := d.currPlaying == nil && d.prevPlaying != nil && d.prevPlaying.Stationuuid == s.Stationuuid
	var res strings.Builder
	var str string

	prefix := fmt.Sprintf("%d. ", index+1)
	if index+1 < 10 {
		prefix = fmt.Sprintf("   %s", prefix)
	} else if index+1 < 100 {
		prefix = fmt.Sprintf("  %s", prefix)
	} else if index+1 < 1000 {
		prefix = fmt.Sprintf(" %s", prefix)
	}

	if isCurr || isPrev {
		res.WriteString(nowPlayingPrefixStyle.Render(prefix))
		itStyle := nowPlayingStyle
		descStyle := nowPlayingDescStyle
		if isSel {
			itStyle = selNowPlayingStyle
			descStyle = selNowPlayingDescStyle
		}

		res.WriteString(itStyle.Render(name))
		w := m.Width()
		hFill := max(w-utf8.RuneCountInString(prefix)-utf8.RuneCountInString(name), 0)
		res.WriteString(itStyle.Render(strings.Repeat(" ", hFill)))
		res.WriteString("\n")
		res.WriteString(nowPlayingPrefixStyle.Render(strings.Repeat(" ", utf8.RuneCountInString(prefix))))
		res.WriteString(descStyle.Render(s.Description()))
		hFill = max(w-utf8.RuneCountInString(prefix)-utf8.RuneCountInString(s.Description()), 0)
		res.WriteString(descStyle.Render(strings.Repeat(" ", hFill)))

		str = res.String()
		str = selectedBorderStyle.Render(str)
	} else {
		res.WriteString(prefixStyle.Render(prefix))
		itStyle := itemStyle
		descStyle := descStyle
		if isSel {
			itStyle = selItemStyle
			descStyle = selDescStyle
		}
		res.WriteString(itStyle.Render(name))

		w := m.Width()
		hFill := max(w-utf8.RuneCountInString(prefix)-utf8.RuneCountInString(name), 0)
		res.WriteString(itStyle.Render(strings.Repeat(" ", hFill)))
		res.WriteString("\n")
		res.WriteString(prefixStyle.Render(strings.Repeat(" ", utf8.RuneCountInString(prefix))))
		res.WriteString(descStyle.Render(s.Description()))
		hFill = max(w-utf8.RuneCountInString(prefix)-utf8.RuneCountInString(s.Description()), 0)
		res.WriteString(descStyle.Render(strings.Repeat(" ", hFill)))
		str = res.String()
	}

	fmt.Fprint(w, str)
}

// Additional short help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d *stationDelegate) ShortHelp() []key.Binding {
	return []key.Binding{
		d.keymap.playSelected, d.keymap.pause, d.keymap.toggleFavorite,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d *stationDelegate) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.keymap.playSelected, d.keymap.pause, d.keymap.toggleFavorite, d.keymap.info,
		},
	}
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		pause: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "play/pause"),
		),
		playSelected: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "play"),
		),
		info: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "station info"),
		),
		toggleFavorite: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "set favorite"),
		),
	}
}

type delegateKeyMap struct {
	pause          key.Binding
	playSelected   key.Binding
	info           key.Binding
	toggleFavorite key.Binding
}
