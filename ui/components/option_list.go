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
)

const (
	enterPrompt = "  >>>"
	exitPrompt  = "  <<<"
)

type OptionList struct {
	active bool

	prompt string

	options []string

	idx        int
	previewIdx int
	jump       JumpInfo

	style *styles.Style

	Keymap optionsKeymap
}

func NewOptionList(prompt string, options []string, idx int, s *styles.Style) OptionList {
	k := optionsKeymap{
		acceptKey: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "accept"),
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
		idx:        idx,
		previewIdx: idx,
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

func (o *OptionList) IsActive() bool {
	return o.active
}

type jumpPositionMgs int

func (o *OptionList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log := slog.With("method", "components.OptionList.Update")
	log.Debug("tea.Msg", "type", fmt.Sprintf("%T", msg), "value", msg, "#", fmt.Sprintf("%#v", msg))

	switch msg := msg.(type) {
	case jumpPositionMgs:
		if msg := int(msg); msg == o.jump.LastPosition() && o.idx == o.previewIdx {
			o.SetActive(false)
			return o, func() tea.Msg {
				return OptionMsg{SelIdx: o.idx, PreviewIdx: o.previewIdx, Done: true}
			}
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, o.Keymap.digitKeys...):
			digit, _ := strconv.Atoi(msg.String())
			pos := o.jump.NewPosition(digit)
			newIdx := pos - 1
			if newIdx >= 0 && newIdx < len(o.options) {
				o.idx = newIdx
				o.previewIdx = newIdx
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
				return OptionMsg{SelIdx: o.idx, PreviewIdx: o.previewIdx, Done: false}
			}
		case key.Matches(msg, o.Keymap.prev):
			newIdx := o.previewIdx - 1
			if newIdx < 0 {
				newIdx = len(o.options) - 1
			}
			o.previewIdx = newIdx
			return o, func() tea.Msg {
				return OptionMsg{SelIdx: o.idx, PreviewIdx: o.previewIdx, Done: false}
			}

		case key.Matches(msg, o.Keymap.acceptKey):
			o.idx = o.previewIdx
			o.SetActive(false)
			return o, func() tea.Msg {
				return OptionMsg{SelIdx: o.idx, PreviewIdx: o.previewIdx, Done: true}
			}
		case key.Matches(msg, o.Keymap.closeKey):
			o.previewIdx = o.idx
			o.SetActive(false)
			return o, func() tea.Msg {
				return OptionMsg{SelIdx: o.idx, PreviewIdx: o.previewIdx, Done: true}
			}
		}
	}

	return o, nil
}

func (o *OptionList) View() string {
	var b strings.Builder

	prompt := o.prompt + enterPrompt
	optStyle := o.style.SecondaryColorStyle
	previewOptStyle := o.style.ActiveTabInner
	if o.IsActive() {
		prompt = o.prompt + exitPrompt
		optStyle = o.style.PrimaryColorStyle
		previewOptStyle = o.style.SelItemStyle
	}

	for idx := 0; idx < len(o.options); idx++ {
		isSel := idx == o.idx
		isPreview := idx == o.previewIdx

		prefix := ""
		if idx == 0 {
			prefix = prompt
		}
		v := styles.MaxFieldLen - styles.IndexStringPadAmt
		padAmt := &v
		if isSel {
			*padAmt -= 1
		}
		b.WriteString(o.style.SearchPromptStyle.Render(styles.PadFieldName(prefix, padAmt)))

		var opts strings.Builder
		optIdx := styles.IndexString(idx)
		opts.WriteString(optStyle.Render(optIdx))
		optName := o.options[idx]
		if isPreview {
			optName = previewOptStyle.Render(optName)
		} else {
			optName = optStyle.Render(optName)
		}
		opts.WriteString(optName)
		optName = opts.String()
		if isSel {
			optName = o.style.SelectedBorderStyle.Render(optName)
		}
		b.WriteString(optName)
		b.WriteRune('\n')
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
}
