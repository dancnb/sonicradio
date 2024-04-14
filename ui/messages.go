package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancnb/sonicradio/browser"
)

// tea.Msg
type (
	// used for os signal quit not handled by the list model
	quitMsg struct{}

	statusMsg string // used for status error message
	titleMsg  string

	viewMsg                 string // used for view message
	favoritesStationRespMsg struct {
		viewMsg
		statusMsg
		stations []browser.Station
	}

	topStationsRespMsg struct {
		viewMsg
		statusMsg
		stations []browser.Station
	}

	toggleFavoriteMsg struct {
		added   bool
		station browser.Station
	}
)

type stringMsg struct {
	statusMsg *string // used for status error message
	titleMsg  *string
}

func (s stringMsg) String() string {
	var res strings.Builder
	if s.statusMsg != nil {
		res.WriteString("statusMsg=" + *s.statusMsg)
	}
	if s.titleMsg != nil {
		res.WriteString("; titleMsg=" + *s.titleMsg)
	}
	return res.String()
}

// tea.Cmd
func errorMsg(err error) error { return err }

func statusMsgCmd(msg string) tea.Cmd {
	return func() tea.Msg {
		return statusMsg(msg)
	}
}

func titleMsgCmd(msg string) tea.Cmd {
	return func() tea.Msg {
		return titleMsg(msg)
	}
}

func stringMsgCmd(msg stringMsg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
