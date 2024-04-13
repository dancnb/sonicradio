package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancnb/sonicradio/browser"
)

// tea.Msg
type (
	// used for os signal quit not handled by the list model
	quitMsg struct{}

	viewMsg   string // used for view message
	statusMsg string // used for status error message

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

// tea.Cmd
func errorMsg(err error) error { return err }

func respMsgCmd(msg string) tea.Cmd {
	return func() tea.Msg {
		return statusMsg(msg)
	}
}
