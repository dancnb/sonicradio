package ui

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dancnb/sonicradio/config"
)

const (
	emptyHistoryMsg = "\n No playback history available. \n"
)

type historyTab struct {
	cfg     *config.Value
	viewMsg string
	list    list.Model
	keymap  historyKeymap
}

func newHistoryTab(ctx context.Context, cfg *config.Value) *historyTab {
	t := &historyTab{
		cfg: cfg,
		keymap: historyKeymap{
			nextTab: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "go to next tab"),
			),
			prevTab: key.NewBinding(
				key.WithKeys("shift+tab"),
				key.WithHelp("shift+tab", "go to prev tab"),
			),
			search: key.NewBinding(
				key.WithKeys("s"),
				key.WithHelp("s", "search"),
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
	if len(entries) == 0 {
		return nil
	}
	items := make([]list.Item, len(entries))
	for i := len(entries) - 1; i >= 0; i-- {
		items[len(entries)-i-1] = entries[i]
	}
	cmd := t.list.SetItems(items)
	if len(entries) > 0 {
		t.viewMsg = ""
	}
	t.list.Select(0)
	slog.Debug("setEntries", "len", len(t.list.Items()), "index", t.list.Index())
	return cmd
}

func (t *historyTab) createList(width int, height int) {
	delegate := historyEntryDelegate{list.NewDefaultDelegate()}
	l := list.New([]list.Item{}, &delegate, 0, 0)
	l.InfiniteScrolling = true
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetShowFilter(true)
	l.Filter = list.UnsortedFilter
	l.SetStatusBarItemName("entry", "entries")
	l.Styles.NoItems = noItemsStyle
	l.FilterInput.ShowSuggestions = true
	l.KeyMap.Quit.SetKeys("q")
	l.KeyMap.PrevPage.SetKeys("pgup", "ctrl+b")
	l.KeyMap.PrevPage.SetHelp("ctrl+b/pgup", "prev page")
	l.KeyMap.NextPage.SetKeys("pgdown", "ctrl+f")
	l.KeyMap.NextPage.SetHelp("ctrl+f/pgdn", "next page")
	h, v := docStyle.GetFrameSize()
	l.SetSize(width-h, height-v)

	l.Help.ShortSeparator = "   "
	l.Help.Styles = helpStyles()
	l.Styles.HelpStyle = helpStyle
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{t.keymap.search}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			t.keymap.search,
			t.keymap.prevTab,
			t.keymap.nextTab,
		}
	}

	textInputSyle(&l.FilterInput, "Filter:       ", "station name or song")

	t.list = l
}

func (t *historyTab) Update(m *Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	logTeaMsg(msg, "ui.historyTab.Update")

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		t.list.SetSize(msg.Width-h, msg.Height-m.headerHeight-v)

	case tea.KeyMsg:
		if t.IsFiltering() {
			break
		}

		switch {
		case key.Matches(msg, t.list.KeyMap.Quit, t.list.KeyMap.ForceQuit):
			return m, tea.Quit

		case key.Matches(msg, t.keymap.search):
			m.toBrowseTab()
			return m.tabs[browseTabIx].Update(m, msg)

		case key.Matches(msg, t.keymap.nextTab):
			m.toFavoritesTab()

		case key.Matches(msg, t.keymap.prevTab):
			m.toBrowseTab()
		}
	}

	newListModel, cmd := t.list.Update(msg)
	t.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
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
		viewSection := viewStyle.Height(availHeight).Render(t.viewMsg)
		sections = append(sections, viewSection)
		sections = append(sections, help)
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}
	return t.list.View()
}

type historyEntryDelegate struct {
	defaultDelegate list.DefaultDelegate
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

	prefixRender := prefixStyle.Render(prefix)
	res.WriteString(prefixRender)
	maxWidth := max(listWidth-lipgloss.Width(prefixRender)-headerPadDist, 0)

	itStyle := historyItemStyle
	descStyle := historyDescStyle
	if isSel {
		itStyle = historySelItemStyle
		descStyle = historySelDescStyle
	}

	for lipgloss.Width(itStyle.Render(station)) > maxWidth && len(station) > 0 {
		station = station[:len(station)-1]
	}
	nameRender := itStyle.Render(station)
	res.WriteString(nameRender)
	hFill := max(listWidth-lipgloss.Width(prefixRender)-lipgloss.Width(nameRender)-headerPadDist, 0)
	res.WriteString(itStyle.Render(strings.Repeat(" ", hFill)))
	res.WriteString("\n")

	res.WriteString(prefixStyle.Render(strings.Repeat(" ", utf8.RuneCountInString(prefix))))
	desc := entry.Description()
	for lipgloss.Width(descStyle.Render(desc)) > maxWidth && len(desc) > 0 {
		desc = desc[:len(desc)-1]
	}
	descRender := descStyle.Render(desc)
	res.WriteString(descRender)
	hFill = max(listWidth-lipgloss.Width(prefixRender)-lipgloss.Width(descRender)-headerPadDist, 0)
	res.WriteString(descStyle.Render(strings.Repeat(" ", hFill)))

	str := res.String()
	fmt.Fprint(w, str)
}

type historyKeymap struct {
	nextTab key.Binding
	prevTab key.Binding
	search  key.Binding
}
