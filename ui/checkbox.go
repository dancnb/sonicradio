package ui

import (
	"log/slog"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Checkbox struct {
	value   bool
	label   string
	focused bool

	style *Style
}

func NewCheckbox(label string, defValue bool, s *Style) *Checkbox {
	return &Checkbox{
		value: defValue,
		label: label,
		style: s,
	}
}

func (c *Checkbox) SetValue(v bool) {
	c.value = v
}

func (c *Checkbox) Value() bool {
	return c.value
}

func (c *Checkbox) Toggle() {
	c.value = !c.value
}

func (c *Checkbox) Label() string {
	return c.label
}

func (c *Checkbox) SetFocused(v bool) {
	c.focused = v
}

func (c *Checkbox) IsFocused() bool {
	return c.focused
}

func (c *Checkbox) Init() tea.Cmd {
	return nil
}

func (c *Checkbox) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logTeaMsg(msg, "components.Checkbox.Update")
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeySpace:
			c.Toggle()
		}
	}
	return c, nil
}

func (c *Checkbox) View() string {
	var b strings.Builder

	v := MaxFieldLen - IndexStringPadAmt
	padAmt := &v
	if c.IsFocused() {
		*padAmt -= 1
	}
	labelS := c.style.PromptStyle.Render(PadFieldName(c.label, padAmt))
	slog.Info("--->", "label", len(labelS))
	b.WriteString(labelS)

	optStyle := c.style.SecondaryColorStyle
	previewOptStyle := c.style.ActiveTabInner
	// on := "[x]"
	// on := "[✓]"
	on := "[✔]"
	off := "[ ]"
	val := off
	ss := optStyle
	if c.value {
		val = on
		ss = previewOptStyle
	}
	optName := "   " + ss.Render(string(val))
	if c.focused {
		optName = c.style.SelectedBorderStyle.Render(optName)
	}
	b.WriteString(optName)

	return b.String()
}
