package ui

import (
	"log/slog"
	"strconv"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dancnb/sonicradio/model"
)

const (
	stationsFilterPrompt      = "Filter:       "
	stationsFilterPlaceholder = "station name"
)

type uiTabIndex uint8

func (t uiTabIndex) String() string {
	switch t {
	case favoriteTabIx:
		return " Favorites "
	case browseTabIx:
		return "  Browse  "
	case historyTabIx:
		return "  History  "
	case settingsTabIx:
		return " Settings "
	}
	return ""
}

const (
	favoriteTabIx uiTabIndex = iota
	browseTabIx
	historyTabIx
	settingsTabIx
)

type uiTab interface {
	Init(m *Model) tea.Cmd
	Update(m *Model, msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
}

type filteringTab interface {
	IsFiltering() bool
}

type stationTab interface {
	uiTab
	filteringTab
	Stations() *stationsTabBase
	IsSearchEnabled() bool
	IsCustomStationEnabled() bool
	IsInfoEnabled() bool
	createList(delegate *stationDelegate, width int, height int) list.Model
}

type stationsTabBase struct {
	uiTab
	style      *Style
	list       list.Model
	viewMsg    string
	listKeymap listKeymap
	jump       JumpInfo
	infoModel  *infoModel
}

func newStationsTab(k listKeymap, infoModel *infoModel, s *Style) stationsTabBase {
	t := stationsTabBase{
		style:      s,
		listKeymap: k,
		infoModel:  infoModel,
	}
	return t
}

func (t *stationsTabBase) doJump(msg tea.KeyMsg) {
	digit, _ := strconv.Atoi(msg.String())
	jumpIdx := t.jump.NewPosition(digit)
	if jumpIdx > 0 && jumpIdx <= len(t.list.Items()) {
		t.list.Select(jumpIdx - 1)
	}
}

func (t *stationsTabBase) View() string {
	if t.viewMsg != "" {
		var sections []string
		availHeight := t.list.Height()
		help := t.list.Styles.HelpStyle.Render(t.list.Help.View(t.list))
		availHeight -= lipgloss.Height(help)
		viewSection := t.style.ViewStyle.Height(availHeight).Render(t.viewMsg)
		sections = append(sections, viewSection)
		sections = append(sections, help)
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}
	return t.list.View()
}

func (t *stationsTabBase) Stations() *stationsTabBase { return t }

func (t *stationsTabBase) IsFiltering() bool {
	return t.list.FilterState() == list.Filtering
}

func (t *stationsTabBase) toNowPlaying(m *Model) {
	log := slog.With("method", "ui.stationsTabBase.toNowPlaying")
	log.Info("begin")
	defer log.Info("end")

	m.delegate.playingMtx.RLock()
	defer m.delegate.playingMtx.RUnlock()

	uuid := ""
	if m.delegate.currPlaying != nil {
		uuid = m.delegate.currPlaying.Stationuuid
	} else if m.delegate.prevPlaying != nil {
		uuid = m.delegate.prevPlaying.Stationuuid
	} else {
		return
	}
	selIndex := -1
	items := t.list.VisibleItems()
	for ix := range items {
		if items[ix].(model.Station).Stationuuid == uuid {
			selIndex = ix
			break
		}
	}
	if selIndex > -1 {
		t.list.Select(selIndex)
	}
}

func (t *stationsTabBase) IsSearchEnabled() bool {
	return false
}

func (t *stationsTabBase) IsCustomStationEnabled() bool {
	return false
}

func (t *stationsTabBase) IsInfoEnabled() bool {
	return t.infoModel != nil && t.infoModel.enabled
}

func (*stationsTabBase) newSizeMsg(sizeMsg tea.WindowSizeMsg, m *Model) tea.WindowSizeMsg {
	availableHeight := sizeMsg.Height - m.headerHeight
	newSizeMsg := tea.WindowSizeMsg{Width: sizeMsg.Width, Height: availableHeight}
	return newSizeMsg
}

func (t *stationsTabBase) initInfoModel(m *Model, msg toggleInfoMsg) tea.Cmd {
	t.listKeymap.setEnabled(false)
	t.infoModel.setSize(m.width, m.totHeight-m.headerHeight)
	return t.infoModel.Init(msg.station)
}

func (t *stationsTabBase) getListStationByUUID(uuid string) (*model.Station, *int) {
	var s *model.Station
	var idx *int
	for i := range t.list.Items() {
		itS, ok := t.list.Items()[i].(model.Station)
		if ok && itS.Stationuuid == uuid {
			idx = &i
			s = &itS
			break
		}
	}
	return s, idx
}

func createList(delegate *stationDelegate, width int, height int) list.Model {
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.InfiniteScrolling = true
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetShowFilter(true)
	l.SetStatusBarItemName("station", "stations")
	l.Styles.NoItems = delegate.style.NoItemsStyle
	l.FilterInput.ShowSuggestions = true
	l.KeyMap.Quit.SetKeys("q")
	l.KeyMap.PrevPage.SetKeys("pgup", "ctrl+b")
	l.KeyMap.PrevPage.SetHelp("ctrl+b/pgup", "prev page")
	l.KeyMap.NextPage.SetKeys("pgdown", "ctrl+f")
	l.KeyMap.NextPage.SetHelp("ctrl+f/pgdn", "next page")
	h, v := delegate.style.DocStyle.GetFrameSize()
	l.SetSize(width-h, height-v)

	l.Help.ShortSeparator = "   "
	l.Help.Styles = delegate.style.HelpStyles()
	l.Styles.HelpStyle = delegate.style.HelpStyle

	delegate.style.TextInputSyle(&l.FilterInput, stationsFilterPrompt, stationsFilterPlaceholder)
	return l
}
