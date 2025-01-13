package components

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/dancnb/sonicradio/ui/styles"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const enterPrompt = "Enter #       "

type OptionList struct {
	active bool

	prompt  string
	options []string
	idx     int
	jump    JumpInfo

	promptStyle *lipgloss.Style
	selStyle    *lipgloss.Style
	unselStyle  *lipgloss.Style

	Keymap optionsKeymap
}

func NewOptionList(prompt string, options []string, idx int, s *styles.Style) OptionList {
	k := optionsKeymap{
		closeKey: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		// next: key.NewBinding(
		// 	key.WithKeys("down", "j"),
		// 	key.WithHelp("↓/j", "next"),
		// ),
		// prev: key.NewBinding(
		// 	key.WithKeys("up", "k"),
		// 	key.WithHelp("↑/k", "prev"),
		// ),
	}
	for i := 0; i <= 9; i++ {
		x := fmt.Sprintf("%d", i)
		orderkey := key.NewBinding(key.WithKeys(x))
		k.digitKeys = append(k.digitKeys, orderkey)
	}
	o := OptionList{
		prompt:      prompt,
		options:     options,
		idx:         idx,
		promptStyle: &s.SearchPromptStyle,
		selStyle:    &s.PrimaryColorStyle,
		unselStyle:  &s.SecondaryColorStyle,
		Keymap:      k,
	}
	o.SetActive(false)
	return o
}

func (o *OptionList) Init() tea.Cmd {
	return nil
}

func (o *OptionList) SetIdx(v int) {
	o.idx = v
}

func (o *OptionList) SetActive(v bool) {
	o.active = v
	o.Keymap.setEnable(v)
}

func (o *OptionList) IsActive() bool {
	return o.active
}

type jumpMgs int

func (o *OptionList) setIdx(pos int) {
	newIdx := pos - 1
	if newIdx >= 0 && newIdx < len(o.options) {
		o.idx = newIdx
	}
}

func (o *OptionList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log := slog.With("method", "components.OptionList.Update")
	log.Debug("tea.Msg", "type", fmt.Sprintf("%T", msg), "value", msg, "#", fmt.Sprintf("%#v", msg))

	switch msg := msg.(type) {
	case jumpMgs:
		if msg := int(msg); msg == o.jump.LastPosition() {
			o.SetActive(false)
			return o, func() tea.Msg {
				return OptionMsg{o.idx, true}
			}
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, o.Keymap.digitKeys...):
			digit, _ := strconv.Atoi(msg.String())
			pos := o.jump.NewPosition(digit)
			o.setIdx(pos)
			return o, tea.Batch(
				func() tea.Msg {
					return OptionMsg{o.idx, false}
				},
				tea.Tick(o.jump.JumpTimeout(), func(time.Time) tea.Msg {
					return jumpMgs(pos)
				}),
			)

		case key.Matches(msg, o.Keymap.closeKey):
			o.SetActive(false)
			return o, func() tea.Msg {
				return OptionMsg{o.idx, true}
			}
		}
	}

	return o, nil
}

func (o *OptionList) View() string {
	var b strings.Builder
	prompt := o.prompt
	if o.IsActive() {
		prompt = enterPrompt
	}
	for idx := 0; idx < len(o.options); idx++ {
		prefix := ""
		if idx == 0 {
			prefix = prompt
		}
		b.WriteString(o.promptStyle.Render(styles.PadFieldName(prefix)))

		opt := o.options[idx]
		optStyle := o.unselStyle
		if idx == o.idx {
			optStyle = o.selStyle
		}
		optS := fmt.Sprintf("%d. %s", idx+1, opt)
		b.WriteString(optStyle.Render(optS))
		b.WriteRune('\n')
	}
	return b.String()
}

type optionsKeymap struct {
	closeKey key.Binding
	// next      key.Binding
	// prev      key.Binding
	digitKeys []key.Binding
}

func (k *optionsKeymap) ShortHelp() []key.Binding {
	// return []key.Binding{k.prev, k.next, k.closeKey}
	return []key.Binding{k.closeKey}
}

func (k *optionsKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

func (k *optionsKeymap) setEnable(v bool) {
	k.closeKey.SetEnabled(v)
	// k.prev.SetEnabled(v)
	// k.next.SetEnabled(v)
	for i := range k.digitKeys {
		k.digitKeys[i].SetEnabled(v)
	}
}

type OptionMsg struct {
	Val  int
	Done bool
}
