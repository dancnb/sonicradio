package components

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type FormElement struct {
	input      *textinput.Model
	optionList *OptionList
}

func NewFormElement(input *textinput.Model, options *OptionList) *FormElement {
	return &FormElement{input, options}
}

func (e *FormElement) Update(msg tea.Msg) (*FormElement, tea.Cmd) {
	logTeaMsg(msg, "components.FormElement.Update")

	var cmd tea.Cmd
	switch {
	case e.input != nil:
		var input textinput.Model
		input, cmd = e.input.Update(msg)
		e.input = &input
	case e.optionList != nil:
		var opt tea.Model
		opt, cmd = e.optionList.Update(msg)
		e.optionList = opt.(*OptionList)
	}
	return e, cmd
}

func (e *FormElement) Focus() tea.Cmd {
	var cmd tea.Cmd
	switch {
	case e.input != nil:
		cmd = e.input.Focus()
	case e.optionList != nil:
		if !e.optionList.IsActive() {
			e.optionList.SetActive(true)
		}
	}
	return cmd
}

func (e *FormElement) Blur() {
	switch {
	case e.input != nil:
		e.input.Blur()
	case e.optionList != nil:
		if e.optionList.IsActive() {
			e.optionList.SetActive(false)
		}
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

// func (e *FormElement) checkSelected(isSel bool, s *styles.Style) {
// 	if isSel {
// 		e.input.PromptStyle = s.SelPromptStyle
// 		// e.input.TextStyle = s.SelDescStyle
// 	} else {
// 		e.input.PromptStyle = s.PromptStyle
// 		// e.input.TextStyle = s.PrimaryColorStyle
// 	}
// }

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

func (e *FormElement) IsActive() bool {
	switch {
	case e.input != nil:
		return e.input.Focused()
	case e.optionList != nil:
		return e.optionList.IsActive()
	}
	return false
}
