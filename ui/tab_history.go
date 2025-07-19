package ui

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/dancnb/sonicradio/ui/components"
	"github.com/dancnb/sonicradio/ui/styles"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dancnb/sonicradio/config"
)

const (
	historyFilterPlaceholder = "station name or song"
)

type historyTab struct {
	cfg     *config.Value
	style   *styles.Style
	viewMsg string
	jump    components.JumpInfo
	list    list.Model
	keymap  historyKeymap
}

func newHistoryTab(ctx context.Context, cfg *config.Value, s *styles.Style) *historyTab {
	t := &historyTab{
		cfg:   cfg,
		style: s,
		keymap: historyKeymap{
			play: key.NewBinding(
				key.WithKeys("enter", "l"),
				key.WithHelp("enter/l", "play"),
			),
			deleteOne: key.NewBinding(
				key.WithKeys("d"),
				key.WithHelp("d", "delete entry"),
			),
			deleteAll: key.NewBinding(
				key.WithKeys("D"),
				key.WithHelp("D", "clear entries   "),
			),
			nextTab: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "go to next tab"),
			),
			prevTab: key.NewBinding(
				key.WithKeys("shift+tab"),
				key.WithHelp("shift+tab", "go to prev tab"),
			),
			settingsTab: key.NewBinding(
				key.WithKeys("S"),
				key.WithHelp("S", "go to settings tab"),
			),
			favoritesTab: key.NewBinding(
				key.WithKeys("F"),
				key.WithHelp("F", "go to favorites tab"),
			),
			browseTab: key.NewBinding(
				key.WithKeys("B"),
				key.WithHelp("B", "go to browse tab"),
			),
			search: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "search"),
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
		},
	}

	go t.handleHistoryUpdates(ctx)

	return t
}

func (t *historyTab) handleHistoryUpdates(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case entries := <-t.cfg.HistoryChan:
			t.setEntries(entries)
		}
	}
}

func (t *historyTab) Init(m *Model) tea.Cmd {
	t.viewMsg = emptyHistoryMsg
	t.createList(m.width, m.totHeight-m.headerHeight)
	return t.setEntries(t.cfg.History)
}

func (t *historyTab) setEntries(entries []config.HistoryEntry) tea.Cmd {
	items := make([]list.Item, len(entries))
	for i := len(entries) - 1; i >= 0; i-- {
		items[len(entries)-i-1] = entries[i]
	}
	cmd := t.list.SetItems(items)
	if len(entries) > 0 {
		t.viewMsg = ""
	} else {
		t.viewMsg = emptyHistoryMsg
	}
	t.list.Select(0)
	return cmd
}

func (t *historyTab) deleteOneCmd() tea.Cmd {
	return func() tea.Msg {
		if t.list.SelectedItem() == nil {
			return nil
		}
		e, ok := t.list.SelectedItem().(config.HistoryEntry)
		if !ok {
			return nil
		}
		idx := t.list.Index()

		t.cfg.DeleteHistoryEntry(e)

		t.list.RemoveItem(idx)
		if idx >= len(t.list.Items()) {
			t.list.Select(len(t.list.Items()) - 1)
		}
		t.viewMsg = ""
		if len(t.list.Items()) == 0 {
			t.viewMsg = emptyHistoryMsg
		}
		return nil
	}
}

func (t *historyTab) deleteAllCmd() tea.Cmd {
	return func() tea.Msg {
		t.cfg.ClearHistory()
		return t.setEntries([]config.HistoryEntry{})
	}
}

func (t *historyTab) createList(width int, height int) {
	delegate := historyEntryDelegate{
		defaultDelegate: list.NewDefaultDelegate(),
		keymap:          &t.keymap,
		style:           t.style,
	}
	l := list.New([]list.Item{}, &delegate, 0, 0)
	l.InfiniteScrolling = true
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetShowFilter(true)
	l.Filter = list.UnsortedFilter
	l.SetStatusBarItemName("entry", "entries")
	l.Styles.NoItems = t.style.NoItemsStyle
	l.FilterInput.ShowSuggestions = true
	l.KeyMap.Quit.SetKeys("q")
	l.KeyMap.PrevPage.SetKeys("pgup", "ctrl+b")
	l.KeyMap.PrevPage.SetHelp("ctrl+b/pgup", "prev page")
	l.KeyMap.NextPage.SetKeys("pgdown", "ctrl+f")
	l.KeyMap.NextPage.SetHelp("ctrl+f/pgdn", "next page")
	h, v := t.style.DocStyle.GetFrameSize()
	l.SetSize(width-h, height-v)

	l.Help.ShortSeparator = "   "
	l.Help.Styles = t.style.HelpStyles()
	l.Styles.HelpStyle = t.style.HelpStyle
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{t.keymap.search}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			t.keymap.search,
			t.keymap.digitHelp,
			t.keymap.prevTab,
			t.keymap.nextTab,
			t.keymap.favoritesTab,
			t.keymap.browseTab,
			t.keymap.settingsTab,
		}
	}

	t.style.TextInputSyle(&l.FilterInput, stationsFilterPrompt, historyFilterPlaceholder)

	t.list = l
}

func (t *historyTab) Update(m *Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	logTeaMsg(msg, "ui.historyTab.Update")

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := t.style.DocStyle.GetFrameSize()
		t.list.SetSize(msg.Width-h, msg.Height-m.headerHeight-v)

	case tea.KeyMsg:
		if t.IsFiltering() {
			break
		}

		switch {
		case key.Matches(msg, t.list.KeyMap.Quit, t.list.KeyMap.ForceQuit):
			return m, tea.Quit

		case key.Matches(msg, t.keymap.play):
			e, _ := t.list.SelectedItem().(config.HistoryEntry)
			if slices.Contains(m.cfg.Favorites, e.Uuid) {
				m.toFavoritesTab()
				return m.tabs[favoriteTabIx].Update(m, playHistoryEntryMsg{e.Uuid})
			}
			m.toBrowseTab()
			return m.tabs[browseTabIx].Update(m, playHistoryEntryMsg{e.Uuid})

		case key.Matches(msg, t.keymap.deleteOne):
			return m, t.deleteOneCmd()
		case key.Matches(msg, t.keymap.deleteAll):
			return m, t.deleteAllCmd()

		case key.Matches(msg, t.keymap.search):
			m.toBrowseTab()
			return m.tabs[browseTabIx].Update(m, msg)
		case key.Matches(msg, t.keymap.digits...):
			t.doJump(msg)

		case key.Matches(msg, t.keymap.nextTab, t.keymap.settingsTab):
			return m, m.toSettingsTab()

		case key.Matches(msg, t.keymap.favoritesTab):
			m.toFavoritesTab()

		case key.Matches(msg, t.keymap.prevTab, t.keymap.browseTab):
			m.toBrowseTab()
		}
	}

	newListModel, cmd := t.list.Update(msg)
	t.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (t *historyTab) doJump(msg tea.KeyMsg) {
	digit, _ := strconv.Atoi(msg.String())
	jumpIdx := t.jump.NewPosition(digit)
	if jumpIdx > 0 && jumpIdx <= len(t.list.Items()) {
		t.list.Select(jumpIdx - 1)
	}
}

func (t *historyTab) IsFiltering() bool {
	return t.list.FilterState() == list.Filtering
}

func (t *historyTab) View() string {
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

type historyEntryDelegate struct {
	defaultDelegate list.DefaultDelegate
	keymap          *historyKeymap
	style           *styles.Style
}

func (d *historyEntryDelegate) ShortHelp() []key.Binding {
	return []key.Binding{d.keymap.play}
}

func (d *historyEntryDelegate) FullHelp() [][]key.Binding {
	return [][]key.Binding{{d.keymap.play, d.keymap.deleteOne, d.keymap.deleteAll}}
}

func (d *historyEntryDelegate) Height() int { return d.defaultDelegate.Height() }

func (d *historyEntryDelegate) Spacing() int { return d.defaultDelegate.Spacing() }

func (d *historyEntryDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	logTeaMsg(msg, "ui.historyEntryDelegate.Update")
	return nil
}

func (d *historyEntryDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	entry, ok := item.(config.HistoryEntry)
	if !ok {
		return
	}
	isSel := index == m.Index()
	var res strings.Builder

	prefix := fmt.Sprintf("%d. ", index+1)
	if index+1 < 10 {
		prefix = fmt.Sprintf("   %s", prefix)
	} else if index+1 < 100 {
		prefix = fmt.Sprintf("  %s", prefix)
	} else if index+1 < 1000 {
		prefix = fmt.Sprintf(" %s", prefix)
	}
	listWidth := m.Width()
	station := entry.Title()

	prefixRender := d.style.PrefixStyle.Render(prefix)
	res.WriteString(prefixRender)
	maxWidth := max(listWidth-lipgloss.Width(prefixRender)-styles.HeaderPadDist, 0)

	itStyle := d.style.SecondaryColorStyle
	descStyle := d.style.HistoryDescStyle
	if isSel {
		itStyle = d.style.HistorySelItemStyle
		descStyle = d.style.HistorySelDescStyle
	}

	for lipgloss.Width(itStyle.Render(station)) > maxWidth && len(station) > 0 {
		station = station[:len(station)-1]
	}
	nameRender := itStyle.Render(station)
	res.WriteString(nameRender)
	hFill := max(listWidth-lipgloss.Width(prefixRender)-lipgloss.Width(nameRender)-styles.HeaderPadDist, 0)
	res.WriteString(itStyle.Render(strings.Repeat(" ", hFill)))
	res.WriteString("\n")

	res.WriteString(d.style.PrefixStyle.Render(strings.Repeat(" ", utf8.RuneCountInString(prefix))))
	desc := entry.Description()
	for lipgloss.Width(descStyle.Render(desc)) > maxWidth && len(desc) > 0 {
		desc = desc[:len(desc)-1]
	}
	descRender := descStyle.Render(desc)
	res.WriteString(descRender)
	hFill = max(listWidth-lipgloss.Width(prefixRender)-lipgloss.Width(descRender)-styles.HeaderPadDist, 0)
	res.WriteString(descStyle.Render(strings.Repeat(" ", hFill)))

	str := res.String()
	fmt.Fprint(w, str)
}

type historyKeymap struct {
	play         key.Binding
	deleteOne    key.Binding
	deleteAll    key.Binding
	nextTab      key.Binding
	prevTab      key.Binding
	favoritesTab key.Binding
	settingsTab  key.Binding
	browseTab    key.Binding
	search       key.Binding
	digits       []key.Binding
	digitHelp    key.Binding
}
