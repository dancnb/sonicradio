package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dancnb/sonicradio/browser"
)

type infoModel struct {
	enabled bool

	style *Style

	b       *browser.Api
	station browser.Station

	keymap infoKeymap
	help   help.Model
	width  int
	height int
}

func newInfoModel(b *browser.Api, s *Style) *infoModel {
	k := newInfoKeymap()

	h := help.New()
	h.ShowAll = false
	h.ShortSeparator = "   "
	h.Styles = s.HelpStyles()

	return &infoModel{
		b:      b,
		style:  s,
		keymap: k,
		help:   h,
	}
}

func (i *infoModel) Init(s browser.Station) tea.Cmd {
	i.station = s
	i.setEnabled(true)
	return nil
}

func (s *infoModel) setSize(width, height int) {
	h, v := s.style.DocStyle.GetFrameSize()
	s.width = width - h
	s.height = height - v
	s.help.Width = s.width
}

func (s *infoModel) isEnabled() bool {
	return s.enabled
}

func (s *infoModel) setEnabled(v bool) {
	s.enabled = v
	s.keymap.setEnable(v)
	s.help.ShowAll = false
}

func (i *infoModel) Update(msg tea.Msg) (*infoModel, tea.Cmd) {
	logTeaMsg(msg, "ui.infoModel.Update")
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		i.setSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, i.keymap.vote):
			return i, func() tea.Msg {
				err := i.b.StationVote(i.station.Stationuuid)
				if err != nil {
					return statusMsg(err.Error())
				}
				return statusMsg(voteSuccesful)
			}
		case key.Matches(msg, i.keymap.cancel):
			return i, func() tea.Msg {
				i.setEnabled(false)
				return toggleInfoMsg{enable: false}
			}
		}
	}

	return i, tea.Batch(cmds...)
}

func (i *infoModel) View() string {
	var b strings.Builder
	i.renderInfoField(&b, "Name          ", i.station.Name)
	i.renderInfoField(&b, "Homepage      ", i.station.Homepage)
	i.renderInfoField(&b, "Stream URL    ", i.station.URL)
	i.renderInfoField(&b, "Tags          ", i.station.Tags)
	i.renderInfoField(&b, "Votes         ", fmt.Sprintf("%d", i.station.Votes))
	i.renderInfoField(&b, "Clicks        ", fmt.Sprintf("%d", i.station.Clickcount))
	trend := fmt.Sprintf("%d", i.station.Clicktrend)
	if i.station.Clicktrend > 0 {
		trend = "+" + trend
	}
	i.renderInfoField(&b, "Trending      ", trend)
	i.renderInfoField(&b, "Codec         ", i.station.Codec)
	br := ""
	if i.station.Bitrate != 0 {
		br = fmt.Sprintf("%d", i.station.Bitrate)
	}
	i.renderInfoField(&b, "Bitrate       ", br)
	country := i.station.Country
	cc := strings.TrimSpace(i.station.Countrycode)
	if cc != "" {
		country += fmt.Sprintf(" [%s]", cc)
	}
	i.renderInfoField(&b, "Country       ", country)
	i.renderInfoField(&b, "State         ", i.station.State)
	i.renderInfoField(&b, "Language      ", i.station.Language)
	i.renderInfoField(&b, "Last ok check ", i.station.Lastcheckoktime)
	lat := ""
	if i.station.GeoLat != nil {
		lat = fmt.Sprintf("%v", i.station.GeoLat)
	}
	i.renderInfoField(&b, "Geo latidude  ", lat)
	long := ""
	if i.station.GeoLong != nil {
		long = fmt.Sprintf("%v", i.station.GeoLong)
	}
	i.renderInfoField(&b, "Geo longitude ", long)

	availHeight := i.height
	help := i.style.HelpStyle.Render(i.help.View(&i.keymap))
	availHeight -= lipgloss.Height(help)

	content := b.String()
	inputsHeight := lipgloss.Height(content)
	for i := 0; i < availHeight-inputsHeight; i++ {
		b.WriteString("\n")
	}
	return b.String() + help
}

func (i *infoModel) renderInfoField(b *strings.Builder, fieldName, fieldValue string) {
	fnRender := i.style.InfoFieldNameStyle.Render(PadFieldName(fieldName, nil))
	b.WriteString(fnRender)
	fnw := lipgloss.Width(fnRender)
	fv := strings.TrimSpace(fieldValue)
	for fnw+lipgloss.Width(i.style.SecondaryColorStyle.Render(fv)) > i.width && len(fv) > 0 {
		fv = fv[:len(fv)-1]
	}
	b.WriteString(i.style.SecondaryColorStyle.Render(fv))
	b.WriteString("\n")
}

type infoKeymap struct {
	cancel key.Binding
	vote   key.Binding
}

func newInfoKeymap() infoKeymap {
	k := infoKeymap{
		cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		vote: key.NewBinding(
			key.WithKeys("ctrl+v"),
			key.WithHelp("ctrl+v", "vote station"),
		),
	}
	return k
}

func (k *infoKeymap) ShortHelp() []key.Binding {
	return []key.Binding{k.vote, k.cancel}
}

func (k *infoKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.vote, k.cancel},
	}
}

func (k *infoKeymap) setEnable(v bool) {
	k.cancel.SetEnabled(v)
	k.vote.SetEnabled(v)
}
