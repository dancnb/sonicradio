package ui

import (
	"github.com/dancnb/sonicradio/browser"
)

// tea.Msg
type (
	// used for os signal quit not handled by the list model
	quitMsg struct{}

	// song title
	titleMsg string
	// used for status error message
	statusMsg string
	// view msg instead of list
	viewMsg                 string
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

	searchRespMsg struct {
		viewMsg
		statusMsg
		stations  []browser.Station
		cancelled bool
	}

	toggleFavoriteMsg struct {
		added   bool
		station browser.Station
	}

	playRespMsg struct {
		err string
	}
)
