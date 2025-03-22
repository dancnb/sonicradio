package components

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/dancnb/sonicradio/ui/styles"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type OptionList struct {
	focused bool
	active  bool
	quick   bool

	prompt string

	options    []OptionValue
	idx        int
	previewIdx int
	jump       JumpInfo

	style *styles.Style

	PartialCallbackFn func(int)
	DoneCallbackFn    func(int)

	Keymap optionsKeymap
}

type OptionValue struct {
	IdxView  int
	NameView string
}

var padOptName = 17

func NewOptionList(prompt string, options []OptionValue, startIdx int, s *styles.Style) OptionList {
	k := optionsKeymap{
		acceptKey: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("space/enter", "accept"),
		),
		closeKey: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		next: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "next"),
		),
		prev: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "prev"),
		),
		digitHelp: key.NewBinding(
			key.WithKeys("#"),
			key.WithHelp("1..", "enter number"),
		),
	}
	for i := 0; i <= 9; i++ {
		x := fmt.Sprintf("%d", i)
		orderkey := key.NewBinding(key.WithKeys(x))
		k.digitKeys = append(k.digitKeys, orderkey)
	}
	o := OptionList{
		prompt:     prompt,
		options:    options,
		idx:        startIdx,
		previewIdx: startIdx,
		style:      s,
		Keymap:     k,
	}
	o.SetActive(false)
	return o
}

func (o *OptionList) Init() tea.Cmd {
	return nil
}

func (o *OptionList) SetIdx(v int) {
	v = max(v, 0)
	v = min(v, len(o.options)-1)
	o.idx = v
	o.previewIdx = v
}

func (o *OptionList) SetActive(v bool) {
	o.active = v
	o.Keymap.setEnable(v)
}

// IsActive: 1.focused + 2.active
func (o *OptionList) IsActive() bool {
	return o.active
}

func (o *OptionList) SetFocused(v bool) {
	o.focused = v
}

// IsFocused: 1.focused
func (o *OptionList) IsFocused() bool {
	return o.focused
}

func (o *OptionList) SetQuick(v bool) {
	o.quick = v
}

type jumpPositionMgs int

func (o *OptionList) jumpPos2Idx(pos int) int {
	for i := 0; i < len(o.options); i++ {
		if o.options[i].IdxView == pos {
			return i
		}
	}
	return -1
}

func (o *OptionList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logTeaMsg(msg, "components.OptionList.Update")

	switch msg := msg.(type) {
	case jumpPositionMgs:
		if msg := int(msg); msg == o.jump.LastPosition() && o.idx == o.previewIdx {
			o.SetActive(false)
			return o, func() tea.Msg {
				return OptionMsg{SelIdx: o.idx, PreviewIdx: o.previewIdx, Done: true, CallbackFn: o.DoneCallbackFn}
			}
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, o.Keymap.digitKeys...):
			digit, _ := strconv.Atoi(msg.String())
			pos := o.jump.NewPosition(digit)
			newIdx := o.jumpPos2Idx(pos)
			if newIdx >= 0 && newIdx < len(o.options) {
				o.idx = newIdx
				o.previewIdx = newIdx
			}
			if o.quick {
				o.SetActive(false)
				return o, func() tea.Msg {
					return OptionMsg{SelIdx: o.idx, PreviewIdx: o.previewIdx, Done: true, CallbackFn: o.DoneCallbackFn}
				}
			}
			return o, tea.Batch(
				func() tea.Msg {
					return OptionMsg{SelIdx: o.idx, PreviewIdx: o.previewIdx, Done: false}
				},
				tea.Tick(o.jump.JumpTimeout(), func(time.Time) tea.Msg {
					return jumpPositionMgs(pos)
				}),
			)

		case key.Matches(msg, o.Keymap.next):
			newIdx := (o.previewIdx + 1) % len(o.options)
			o.previewIdx = newIdx
			return o, func() tea.Msg {
				return OptionMsg{SelIdx: o.idx, PreviewIdx: o.previewIdx, Done: false, CallbackFn: o.PartialCallbackFn}
			}
		case key.Matches(msg, o.Keymap.prev):
			newIdx := o.previewIdx - 1
			if newIdx < 0 {
				newIdx = len(o.options) - 1
			}
			o.previewIdx = newIdx
			return o, func() tea.Msg {
				return OptionMsg{SelIdx: o.idx, PreviewIdx: o.previewIdx, Done: false, CallbackFn: o.PartialCallbackFn}
			}

		case key.Matches(msg, o.Keymap.acceptKey):
			o.idx = o.previewIdx
			o.SetActive(false)
			return o, func() tea.Msg {
				return OptionMsg{SelIdx: o.idx, PreviewIdx: o.previewIdx, Done: true, CallbackFn: o.DoneCallbackFn}
			}
		case key.Matches(msg, o.Keymap.closeKey):
			o.previewIdx = o.idx
			o.SetActive(false)
			return o, func() tea.Msg {
				return OptionMsg{SelIdx: o.idx, PreviewIdx: o.previewIdx, Done: true, CallbackFn: o.DoneCallbackFn}
			}
		}
	}

	return o, nil
}

func (o *OptionList) View() string {
	var b strings.Builder

	optStyle := o.style.SecondaryColorStyle
	previewOptStyle := o.style.ActiveTabInner
	if o.IsActive() {
		optStyle = o.style.PrimaryColorStyle
		previewOptStyle = o.style.SelItemStyle
	}

	for idx := 0; idx < len(o.options); idx++ {
		isSel := idx == o.idx
		isPreview := idx == o.previewIdx

		prefix := ""
		if idx == 0 {
			prefix = o.prompt
		}
		v := styles.MaxFieldLen - styles.IndexStringPadAmt
		padAmt := &v
		if isSel && o.IsFocused() {
			*padAmt -= 1
		}
		b.WriteString(o.style.PromptStyle.Render(styles.PadFieldName(prefix, padAmt)))

		var optS strings.Builder
		optIdx := styles.IndexString(o.options[idx].IdxView)
		optS.WriteString(optStyle.Render(optIdx))
		optName := styles.PadFieldName(o.options[idx].NameView, &padOptName)
		if isPreview {
			optName = previewOptStyle.Render(optName)
		} else {
			optName = optStyle.Render(optName)
		}
		optS.WriteString(optName)
		optName = optS.String()
		if isSel && (o.IsActive() || o.IsFocused()) {
			optName = o.style.SelectedBorderStyle.Render(optName)
		}
		b.WriteString(optName)
		if idx < len(o.options)-1 {
			b.WriteRune('\n')
		}
	}

	return b.String()
}

type optionsKeymap struct {
	acceptKey key.Binding
	closeKey  key.Binding
	next      key.Binding
	prev      key.Binding
	digitKeys []key.Binding
	digitHelp key.Binding
}

func (k *optionsKeymap) ShortHelp() []key.Binding {
	return []key.Binding{k.prev, k.next, k.digitHelp, k.acceptKey, k.closeKey}
}

func (k *optionsKeymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

func (k *optionsKeymap) setEnable(v bool) {
	k.acceptKey.SetEnabled(v)
	k.closeKey.SetEnabled(v)
	k.prev.SetEnabled(v)
	k.next.SetEnabled(v)
	k.digitHelp.SetEnabled(v)
	for i := range k.digitKeys {
		k.digitKeys[i].SetEnabled(v)
	}
}

type OptionMsg struct {
	SelIdx     int
	PreviewIdx int
	Done       bool
	CallbackFn func(int)
}

func logTeaMsg(msg tea.Msg, tag string) {
	log := slog.With("method", tag)
	switch msg.(type) {
	case cursor.BlinkMsg, spinner.TickMsg:
		break
	default:
		log.Info("tea.Msg", "type", fmt.Sprintf("%T", msg), "value", msg, "#", fmt.Sprintf("%#v", msg))
	}
}
