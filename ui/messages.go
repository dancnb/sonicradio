package ui

import (
	"time"

	"github.com/dancnb/sonicradio/browser"
	"github.com/dancnb/sonicradio/player"
)

// tea.Msg
type (
	// used for os signal quit not handled by the list model
	quitMsg struct{}

	// song title
	metadataMsg struct {
		stationUuid  string
		stationName  string
		songTitle    string
		playbackTime *time.Duration
	}

	volumeMsg struct {
		err error
	}

	// used for status info/error message
	statusMsg string

	// view msg instead of list
	viewMsg string

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

	toggleInfoMsg struct {
		enable  bool
		station browser.Station
	}

	playRespMsg struct {
		err string
	}

	pauseRespMsg struct {
		err string
	}

	playHistoryEntryMsg struct {
		uuid string
	}

	playUuidRespMsg struct {
		viewMsg
		statusMsg
		stations []browser.Station
	}
)

func getMetadataMsg(s browser.Station, m player.Metadata) metadataMsg {
	msg := metadataMsg{
		stationUuid: s.Stationuuid,
		stationName: s.Name,
		songTitle:   m.Title,
	}
	if m.PlaybackTimeSec != nil {
		t := time.Second * (time.Duration(*m.PlaybackTimeSec))
		msg.playbackTime = &t
	}
	return msg
}
