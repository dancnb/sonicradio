package ui

import (
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

	playRespMsg struct {
		err string
	}
)
