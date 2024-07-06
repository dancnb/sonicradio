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
	"github.com/charmbracelet/lipgloss"
	"github.com/dancnb/sonicradio/browser"
	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player"
)

const startWaitMillis = 500 * 3

func newStationDelegate(cfg *config.Value, p player.Player, b *browser.Api) *stationDelegate {
	keymap := newDelegateKeyMap()

	d := list.NewDefaultDelegate()

	return &stationDelegate{
		player:          p,
		b:               b,
		cfg:             cfg,
		keymap:          keymap,
		defaultDelegate: d,
	}
}

type stationDelegate struct {
	player      player.Player
	b           *browser.Api
	cfg         *config.Value
	prevPlaying *browser.Station
	currPlaying *browser.Station
	deleted     *browser.Station
	keymap      *delegateKeyMap

	defaultDelegate list.DefaultDelegate
}

func (d *stationDelegate) Height() int { return d.defaultDelegate.Height() }

func (d *stationDelegate) Spacing() int { return d.defaultDelegate.Spacing() }

func (d *stationDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	logTeaMsg(msg, "ui.stationDelegate.Update")
	selStation, isSel := m.SelectedItem().(browser.Station)

	switch msg := msg.(type) {
	case toggleInfoMsg:
		if !msg.enable {
			d.keymap.info.SetEnabled(true)
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.keymap.info):
			if !isSel {
				break
			}
			d.keymap.info.SetEnabled(false)
			return func() tea.Msg { return toggleInfoMsg{enable: true, station: selStation} }

		case key.Matches(msg, d.keymap.toggleFavorite):
			if !isSel {
				break
			}
			added := d.cfg.ToggleFavorite(selStation.Stationuuid)
			return func() tea.Msg { return toggleFavoriteMsg{added, selStation} }

		case key.Matches(msg, d.keymap.delete):
			if !isSel {
				break
			}
			idx := m.Index()
			m.RemoveItem(idx)
			d.deleted = &selStation

		case key.Matches(msg, d.keymap.pasteAfter):
			if !d.shouldPaste(m) {
				break
			}
			idx := m.Index()
			if len(m.Items()) > 0 {
				idx++
				m.Select(idx)
			}
			cmd := m.InsertItem(idx, *d.deleted)
			d.deleted = nil
			return cmd

		case key.Matches(msg, d.keymap.pasteBefore):
			if !d.shouldPaste(m) {
				break
			}
			idx := m.Index()
			cmd := m.InsertItem(idx, *d.deleted)
			d.deleted = nil
			return cmd
		}
	}

	return nil
}

func (d *stationDelegate) shouldPaste(m *list.Model) bool {
	if d.deleted == nil {
		return false
	}
	its := m.Items()
	dupl := false
	for ii := range its {
		if d.deleted.Stationuuid == its[ii].(browser.Station).Stationuuid {
			dupl = true
			break
		}
	}
	return !dupl
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

func (d *stationDelegate) increaseCounter(station browser.Station) {
	d.b.StationCounter(station.Stationuuid)
}

func (d *stationDelegate) playStation(station browser.Station) error {
	log := slog.With("method", "ui.stationDelegate.playStation")

	go d.increaseCounter(station)

	log.Debug("playing", "id", station.Stationuuid)
	err := d.player.Play(station.URL)
	if err != nil {
		errMsg := fmt.Sprintf("error playing station %s: %s", station.Name, err.Error())
		log.Error(errMsg)
		return errors.New(errMsg)
	}
	d.prevPlaying = d.currPlaying
	d.currPlaying = &station
	return nil
}

func (d *stationDelegate) stopStation(station browser.Station) (wasPlaying bool, err error) {
	log := slog.With("method", "ui.stationDelegate.stopStation")
	if d.currPlaying != nil && d.currPlaying.Stationuuid == station.Stationuuid {
		log.Debug("stopping", "id", d.currPlaying.Stationuuid)
		d.prevPlaying = &station
		d.currPlaying = nil
		err := d.player.Stop()
		if err != nil {
			errMsg := fmt.Sprintf("error stopping station %s: %s", station.Name, err.Error())
			log.Error(errMsg)
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

	listWidth := m.Width()
	if isCurr || isPrev {
		prefixRender := nowPlayingPrefixStyle.Render(prefix)
		res.WriteString(prefixRender)
		maxWidth := max(listWidth-lipgloss.Width(prefixRender)-padDist, 0)

		itStyle := nowPlayingStyle
		descStyle := nowPlayingDescStyle
		if isSel {
			itStyle = selNowPlayingStyle
			descStyle = selNowPlayingDescStyle
		}

		for lipgloss.Width(itStyle.Render(name)) > maxWidth-1 && len(name) > 0 {
			name = name[:len(name)-1]
		}
		nameRender := itStyle.Render(name)
		res.WriteString(nameRender)
		hFill := max(listWidth-lipgloss.Width(prefixRender)-lipgloss.Width(nameRender)-padDist-1, 0)
		res.WriteString(itStyle.Render(strings.Repeat(" ", hFill)))
		res.WriteString("\n")

		res.WriteString(nowPlayingPrefixStyle.Render(strings.Repeat(" ", utf8.RuneCountInString(prefix))))
		desc := s.Description()
		for lipgloss.Width(descStyle.Render(desc)) > maxWidth-1 && len(desc) > 0 {
			desc = desc[:len(desc)-1]
		}
		descRender := descStyle.Render(desc)
		res.WriteString(descRender)
		hFill = max(listWidth-lipgloss.Width(prefixRender)-lipgloss.Width(descRender)-padDist-1, 0)
		res.WriteString(descStyle.Render(strings.Repeat(" ", hFill)))

		str = res.String()
		str = selectedBorderStyle.Render(str)
	} else {
		prefixRender := prefixStyle.Render(prefix)
		res.WriteString(prefixRender)
		maxWidth := max(listWidth-lipgloss.Width(prefixRender)-padDist, 0)

		itStyle := itemStyle
		descStyle := descStyle
		if isSel {
			itStyle = selItemStyle
			descStyle = selDescStyle
		}

		for lipgloss.Width(itStyle.Render(name)) > maxWidth && len(name) > 0 {
			name = name[:len(name)-1]
		}
		nameRender := itStyle.Render(name)
		res.WriteString(nameRender)
		hFill := max(listWidth-lipgloss.Width(prefixRender)-lipgloss.Width(nameRender)-padDist, 0)
		res.WriteString(itStyle.Render(strings.Repeat(" ", hFill)))
		res.WriteString("\n")

		res.WriteString(prefixStyle.Render(strings.Repeat(" ", utf8.RuneCountInString(prefix))))
		desc := s.Description()
		for lipgloss.Width(descStyle.Render(desc)) > maxWidth && len(desc) > 0 {
			desc = desc[:len(desc)-1]
		}
		descRender := descStyle.Render(desc)
		res.WriteString(descRender)
		hFill = max(listWidth-lipgloss.Width(prefixRender)-lipgloss.Width(descRender)-padDist, 0)
		res.WriteString(descStyle.Render(strings.Repeat(" ", hFill)))

		str = res.String()
	}

	fmt.Fprint(w, str)
}

func (d *stationDelegate) ShortHelp() []key.Binding {
	return []key.Binding{
		d.keymap.playSelected, d.keymap.pause, d.keymap.toggleFavorite,
	}
}

func (d *stationDelegate) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.keymap.playSelected, d.keymap.pause, d.keymap.info, d.keymap.toggleFavorite, d.keymap.delete, d.keymap.pasteAfter, d.keymap.pasteBefore,
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
			key.WithHelp("f", "favorite station"),
		),
		delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		pasteAfter: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "paste after"),
		),
		pasteBefore: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("shift+p", "paste at"),
		),
	}
}

type delegateKeyMap struct {
	pause          key.Binding
	playSelected   key.Binding
	info           key.Binding
	toggleFavorite key.Binding
	delete         key.Binding
	pasteAfter     key.Binding
	pasteBefore    key.Binding
}
