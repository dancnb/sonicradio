package ui

import "github.com/dancnb/sonicradio/browser"

// tea.Msg
type (
	// used for os signal quit not handled by the list model
	quitMsg struct{}

	respMsg struct {
		viewMsg string // used for view message
		errMsg  error  // used for status error message
	}

	favoritesStationRespMsg struct {
		respMsg
		stations []browser.Station
	}

	topStationsRespMsg struct {
		respMsg
		stations []browser.Station
	}

	toggleFavoriteMsg struct {
		added   bool
		station browser.Station
	}
)

// tea.Cmd
func errorMsg(err error) error { return err }
