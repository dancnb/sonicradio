package ui

import (
	"fmt"
	"time"

	smodel "github.com/dancnb/sonicradio/model"
	"github.com/dancnb/sonicradio/player/model"
)

// tea.Msg
type (
	// used for os signal quit not handled by the list model
	quitMsg struct{}

	// song title
	metadataMsg struct {
		stationUUID  string
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
		stations []smodel.Station
	}

	topStationsRespMsg struct {
		viewMsg
		statusMsg
		stations []smodel.Station
	}

	searchRespMsg struct {
		viewMsg
		statusMsg
		stations  []smodel.Station
		cancelled bool
	}

	toggleFavoriteMsg struct {
		added   bool
		station smodel.Station
	}

	toggleInfoMsg struct {
		enable  bool
		station smodel.Station
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

	playUUIDRespMsg struct {
		viewMsg
		statusMsg
		stations []smodel.Station
	}
)

func getMetadataMsg(s smodel.Station, m model.Metadata) metadataMsg {
	msg := metadataMsg{
		stationUUID: s.Stationuuid,
		stationName: s.Name,
		songTitle:   m.Title,
	}
	if m.PlaybackTimeSec != nil {
		t := time.Second * (time.Duration(*m.PlaybackTimeSec))
		msg.playbackTime = &t
	}
	return msg
}

func (m metadataMsg) String() string {
	var pt time.Duration
	if m.playbackTime != nil {
		pt = *m.playbackTime
	}
	return fmt.Sprintf("{uuid=%s, name=%s, title=%s, playbackTime=%d}",
		m.stationUUID, m.stationName, m.songTitle, int(pt.Seconds()))
}
