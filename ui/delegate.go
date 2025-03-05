package ui

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/dancnb/sonicradio/ui/styles"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dancnb/sonicradio/browser"
	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/player"
)

const startWaitMillis = 500 * 3

func newStationDelegate(cfg *config.Value, s *styles.Style, p *player.Player, b *browser.Api) *stationDelegate {
	keymap := newDelegateKeyMap()

	d := list.NewDefaultDelegate()

	st := &stationDelegate{
		player:          p,
		b:               b,
		cfg:             cfg,
		style:           s,
		keymap:          keymap,
		defaultDelegate: d,
	}
	st.setStationView(cfg.StationView)
	return st
}

type stationDelegate struct {
	player *player.Player
	b      *browser.Api
	cfg    *config.Value
	style  *styles.Style

	playingMtx  sync.RWMutex
	prevPlaying *browser.Station
	currPlaying *browser.Station

	deleted *browser.Station

	keymap *delegateKeyMap

	defaultDelegate list.DefaultDelegate
}

func (d *stationDelegate) setStationView(v config.StationView) {
	switch v {
	case config.DefaultView:
		d.defaultDelegate.SetHeight(2)
		d.defaultDelegate.SetSpacing(1)
	case config.CompactView:
		d.defaultDelegate.SetHeight(1)
		d.defaultDelegate.SetSpacing(1)
	case config.MinimalView:
		d.defaultDelegate.SetHeight(1)
		d.defaultDelegate.SetSpacing(0)
	}
}

func (d *stationDelegate) Height() int {
	return d.defaultDelegate.Height()
}

func (d *stationDelegate) Spacing() int {
	return d.defaultDelegate.Spacing()
}

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
		case key.Matches(msg, d.keymap.toggleAutoplay):
			if !isSel {
				break
			}
			if d.cfg.AutoplayFavorite == selStation.Stationuuid {
				d.cfg.AutoplayFavorite = ""
			} else {
				d.cfg.AutoplayFavorite = selStation.Stationuuid
			}

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

func (d *stationDelegate) pauseCmd() tea.Cmd {
	return func() tea.Msg {
		log := slog.With("method", "ui.stationDelegate.pauseCmd")
		log.Debug("begin")
		defer log.Debug("end")

		d.playingMtx.Lock()
		defer d.playingMtx.Unlock()

		if d.currPlaying == nil {
			return nil
		}
		err := d.player.Pause(true)
		if err != nil {
			log.Error(fmt.Sprintf("player pause: %v", err))
			return pauseRespMsg{fmt.Sprintf("Could not pause station %s (%s)!", d.currPlaying.Name, d.currPlaying.URL)}
		}
		d.prevPlaying = d.currPlaying
		d.currPlaying = nil
		return pauseRespMsg{}
	}
}

func (d *stationDelegate) resumeCmd() tea.Cmd {
	return func() tea.Msg {
		log := slog.With("method", "ui.stationDelegate.resumeCmd")
		log.Debug("begin")
		defer log.Debug("end")

		d.playingMtx.Lock()
		defer d.playingMtx.Unlock()

		if d.prevPlaying == nil {
			return nil
		}
		err := d.player.Pause(false)
		if err != nil {
			log.Error(fmt.Sprintf("player resume: %v", err))
			return playRespMsg{fmt.Sprintf("Could not resume playback for station %s (%s)!", d.currPlaying.Name, d.currPlaying.URL)}
		}
		d.currPlaying = d.prevPlaying
		d.prevPlaying = nil
		return playRespMsg{}
	}
}

func (d *stationDelegate) playCmd(s browser.Station) tea.Cmd {
	return func() tea.Msg {
		log := slog.With("method", "ui.stationDelegate.playCmd")
		log.Debug("begin")
		defer log.Debug("end")

		d.playingMtx.Lock()
		defer d.playingMtx.Unlock()

		log.Debug("playing", "id", s.Stationuuid)
		go d.increaseCounter(s)

		err := d.player.Play(s.URL)
		if err != nil {
			errMsg := fmt.Sprintf("error playing station %s: %s", s.Name, err.Error())
			log.Error(errMsg)
			return playRespMsg{fmt.Sprintf("Could not start playback for %s (%s)!", s.Name, s.URL)}
		}
		d.prevPlaying = d.currPlaying
		d.currPlaying = &s
		return playRespMsg{}
	}
}

func (d *stationDelegate) increaseCounter(station browser.Station) {
	d.b.StationCounter(station.Stationuuid)
}

func (d *stationDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	s, ok := listItem.(browser.Station)
	if !ok {
		return
	}
	name := s.Name
	if d.cfg.IsFavorite(s.Stationuuid) {
		name += styles.FavChar
	}
	if d.cfg.AutoplayFavorite == s.Stationuuid {
		name += d.style.BaseBold.Render(styles.AutoplayChar)
	}

	isSel := index == m.Index()

	d.playingMtx.RLock()
	defer d.playingMtx.RUnlock()

	isCurr := d.currPlaying != nil && d.currPlaying.Stationuuid == s.Stationuuid
	isPrev := d.currPlaying == nil && d.prevPlaying != nil && d.prevPlaying.Stationuuid == s.Stationuuid
	var str string

	prefix := styles.IndexString(index + 1)

	listWidth := m.Width()
	if isCurr || isPrev {
		itStyle := d.style.PrimaryColorStyle
		descStyle := d.style.SecondaryColorStyle
		if isSel {
			itStyle = d.style.SelNowPlayingStyle
			descStyle = d.style.SelNowPlayingDescStyle
		}
		prefixStyle := d.style.NowPlayingPrefixStyle
		widthOffset := 1

		str = d.renderStationView(prefix, name, s.Description(), listWidth, widthOffset, prefixStyle, itStyle, descStyle)

		str = d.style.SelectedBorderStyle.Render(str)
	} else {
		itStyle := d.style.PrimaryColorStyle
		descStyle := d.style.SecondaryColorStyle
		if isSel {
			itStyle = d.style.SelItemStyle
			descStyle = d.style.SelDescStyle
		}
		prefixStyle := d.style.PrefixStyle
		widthOffset := 0

		str = d.renderStationView(prefix, name, s.Description(), listWidth, widthOffset, prefixStyle, itStyle, descStyle)
	}

	fmt.Fprint(w, str)
}

func (d *stationDelegate) renderStationView(
	prefix string,
	name string,
	desc string,
	listWidth int,
	widthOffset int,
	prefixStyle lipgloss.Style,
	itStyle lipgloss.Style,
	descStyle lipgloss.Style,
) string {
	switch d.cfg.StationView {
	case config.DefaultView:
		return d.renderDefaultView(prefix, name, desc, listWidth, widthOffset, prefixStyle, itStyle, descStyle)
	case config.CompactView:
		return d.renderCompactView(prefix, name, desc, listWidth, widthOffset, prefixStyle, itStyle, descStyle)
	case config.MinimalView:
		return d.renderMinimalView(prefix, name, desc, listWidth, widthOffset, prefixStyle, itStyle, descStyle)
	default:
		return d.renderDefaultView(prefix, name, desc, listWidth, widthOffset, prefixStyle, itStyle, descStyle)
	}
}

func (d *stationDelegate) renderDefaultView(
	prefix string,
	name string,
	desc string,
	listWidth int,
	widthOffset int,
	prefixStyle lipgloss.Style,
	itStyle lipgloss.Style,
	descStyle lipgloss.Style,
) string {
	var res strings.Builder
	prefixRender := prefixStyle.Render(prefix)
	res.WriteString(prefixRender)
	maxWidth := max(listWidth-lipgloss.Width(prefixRender)-styles.HeaderPadDist-widthOffset, 0)

	for lipgloss.Width(itStyle.Render(name)) > maxWidth-widthOffset && len(name) > 0 {
		name = name[:len(name)-1]
	}
	nameRender := itStyle.Render(name)
	res.WriteString(nameRender)
	hFill := max(listWidth-lipgloss.Width(prefixRender)-lipgloss.Width(nameRender)-styles.HeaderPadDist-widthOffset, 0)
	res.WriteString(itStyle.Render(strings.Repeat(" ", hFill)))
	res.WriteString("\n")

	res.WriteString(prefixStyle.Render(strings.Repeat(" ", utf8.RuneCountInString(prefix))))
	for lipgloss.Width(descStyle.Render(desc)) > maxWidth-widthOffset && len(desc) > 0 {
		desc = desc[:len(desc)-1]
	}
	descRender := descStyle.Render(desc)
	res.WriteString(descRender)
	hFill = max(listWidth-lipgloss.Width(prefixRender)-lipgloss.Width(descRender)-styles.HeaderPadDist-widthOffset, 0)
	res.WriteString(descStyle.Render(strings.Repeat(" ", hFill)))

	return res.String()
}

func (d *stationDelegate) renderCompactView(
	prefix string,
	name string,
	desc string,
	listWidth int,
	widthOffset int,
	prefixStyle lipgloss.Style,
	itStyle lipgloss.Style,
	descStyle lipgloss.Style,
) string {
	var res strings.Builder
	prefixRender := prefixStyle.Render(prefix)
	res.WriteString(prefixRender)
	maxWidth := max(listWidth-lipgloss.Width(prefixRender)-styles.HeaderPadDist-widthOffset, 0)
	width1 := 45
	width2 := maxWidth - width1

	for lipgloss.Width(itStyle.Render(name)) > width1 && len(name) > 0 {
		name = name[:len(name)-1]
	}
	nameRender := itStyle.Render(name)
	res.WriteString(nameRender)
	hFill := max(width1-lipgloss.Width(nameRender), 0)
	res.WriteString(itStyle.Render(strings.Repeat(" ", hFill)))

	for lipgloss.Width(descStyle.Render(desc)) > width2 && len(desc) > 0 {
		desc = desc[:len(desc)-1]
	}
	descRender := descStyle.Render(desc)
	res.WriteString(descRender)
	hFill = max(width2-lipgloss.Width(descRender), 0)
	res.WriteString(descStyle.Render(strings.Repeat(" ", hFill)))

	return res.String()
}

func (d *stationDelegate) renderMinimalView(
	prefix string,
	name string,
	desc string,
	listWidth int,
	widthOffset int,
	prefixStyle lipgloss.Style,
	itStyle lipgloss.Style,
	descStyle lipgloss.Style,
) string {
	var res strings.Builder
	prefixRender := prefixStyle.Render(prefix)
	res.WriteString(prefixRender)
	maxWidth := max(listWidth-lipgloss.Width(prefixRender)-styles.HeaderPadDist, 0)

	for lipgloss.Width(itStyle.Render(name)) > maxWidth-widthOffset && len(name) > 0 {
		name = name[:len(name)-1]
	}
	nameRender := itStyle.Render(name)
	res.WriteString(nameRender)
	hFill := max(listWidth-lipgloss.Width(prefixRender)-lipgloss.Width(nameRender)-styles.HeaderPadDist-widthOffset, 0)
	res.WriteString(itStyle.Render(strings.Repeat(" ", hFill)))

	return res.String()
}

func (d *stationDelegate) ShortHelp() []key.Binding {
	return []key.Binding{
		d.keymap.playSelected, d.keymap.pause, d.keymap.toggleFavorite, d.keymap.toggleAutoplay,
	}
}

func (d *stationDelegate) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.keymap.playSelected,
			d.keymap.pause,
			d.keymap.volumeDown,
			d.keymap.volumeUp,
			d.keymap.seekBack,
			d.keymap.seekFw,
			d.keymap.info,
			d.keymap.toggleFavorite,
			d.keymap.toggleAutoplay,
			d.keymap.delete,
			d.keymap.pasteAfter,
			d.keymap.pasteBefore,
		},
	}
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		pause: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "resume"),
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
		toggleAutoplay: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "autoplay station"),
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
		volumeUp: key.NewBinding(
			key.WithKeys("+", "="),
			key.WithHelp("+", "volume +"),
		),
		volumeDown: key.NewBinding(
			key.WithKeys("-", "_"),
			key.WithHelp("-", "volume -"),
		),
		seekBack: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "seek backwards"),
		),
		seekFw: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "seek forward"),
		),
	}
}

type delegateKeyMap struct {
	pause          key.Binding
	playSelected   key.Binding
	info           key.Binding
	toggleFavorite key.Binding
	toggleAutoplay key.Binding
	delete         key.Binding
	pasteAfter     key.Binding
	pasteBefore    key.Binding
	volumeDown     key.Binding
	volumeUp       key.Binding
	seekBack       key.Binding
	seekFw         key.Binding
}
