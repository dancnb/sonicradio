package components

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type FormElement struct {
	input        *textinput.Model
	inputLastVal string

	optionList *OptionList

	description string
}

type FormElementOpt func(f *FormElement)

func WithDescription(desc string) FormElementOpt {
	return func(f *FormElement) {
		f.description = desc
	}
}

func WithTextInput(i *textinput.Model) FormElementOpt {
	return func(f *FormElement) {
		f.input = i
	}
}

func WithOptionList(o *OptionList) FormElementOpt {
	return func(f *FormElement) {
		f.optionList = o
	}
}

func NewFormElement(opts ...FormElementOpt) *FormElement {
	f := &FormElement{}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func (e *FormElement) TextInput() *textinput.Model {
	return e.input
}

func (e *FormElement) Update(msg tea.Msg) (*FormElement, tea.Cmd) {
	logTeaMsg(msg, "components.FormElement.Update")

	var cmd tea.Cmd
	switch {
	case e.input != nil:
		var input textinput.Model
		input, cmd = e.input.Update(msg)
		if input.Err != nil && input.Value() != "" {
			input.SetValue(e.inputLastVal)
		} else {
			e.inputLastVal = input.Value()
		}
		e.input = &input
	case e.optionList != nil:
		var opt tea.Model
		opt, cmd = e.optionList.Update(msg)
		e.optionList = opt.(*OptionList)
	}
	return e, cmd
}

// Focus:
// - input:      1.Focus + 2.SetActive(noop)
// - optionList: 1.Focus
func (e *FormElement) Focus() tea.Cmd {
	var cmd tea.Cmd
	switch {
	case e.input != nil:
		cmd = e.input.Focus()
	case e.optionList != nil:
		e.optionList.SetFocused(true)
	}
	return cmd
}

// SetActive:
// - input:      noop
// - optionList: 2.SetActive
func (e *FormElement) SetActive() {
	switch {
	case e.input != nil:
		break
	case e.optionList != nil:
		e.optionList.SetActive(true)
	}
}

func (e *FormElement) IsActive() bool {
	switch {
	case e.input != nil:
		return e.input.Focused()
	case e.optionList != nil:
		return e.optionList.IsActive()
	}
	return false
}

func (e *FormElement) Blur() {
	switch {
	case e.input != nil:
		e.input.Blur()
	case e.optionList != nil:
		e.optionList.SetActive(false)
		e.optionList.SetFocused(false)
	}
}

func (e *FormElement) View() string {
	switch {
	case e.input != nil:
		return e.input.View()
	case e.optionList != nil:
		return e.optionList.View()
	}
	return ""
}

func (e *FormElement) Value() string {
	switch {
	case e.input != nil:
		return e.input.Value()
	case e.optionList != nil:
		// not used
	}
	return ""
}

func (e *FormElement) SetValue(v any) {
	switch {
	case e.input != nil:
		e.input.SetValue(v.(string))
	case e.optionList != nil:
		e.optionList.SetIdx(v.(int))
	}
}

func (e *FormElement) Keymap() help.KeyMap {
	switch {
	case e.input != nil:
		return nil
	case e.optionList != nil:
		return &e.optionList.Keymap
	}
	return nil
}

func (e *FormElement) Description() string {
	return e.description
}
