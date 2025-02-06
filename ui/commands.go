package ui

import (
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancnb/sonicradio/browser"
	"github.com/dancnb/sonicradio/config"
)

func (m *Model) favoritesReqCmd() tea.Msg {
	if len(m.cfg.Favorites) == 0 {
		return favoritesStationRespMsg{
			viewMsg: noFavoritesAddedMsg,
		}
	}

	stations, err := m.browser.GetStations(m.cfg.Favorites)
	res := favoritesStationRespMsg{stations: stations}
	if err != nil {
		res.statusMsg = statusMsg(err.Error())
	} else if len(stations) == 0 {
		res.viewMsg = noStationsFound
	}
	return res
}

func (m *Model) topStationsCmd() tea.Msg {
	stations, err := m.browser.TopStations()
	res := topStationsRespMsg{stations: stations}
	if err != nil {
		res.statusMsg = statusMsg(err.Error())
	} else if len(stations) == 0 {
		res.viewMsg = noStationsFound
	}
	return res
}

func (m *Model) volumeCmd(up bool) tea.Cmd {
	return func() tea.Msg {
		currVol := m.cfg.GetVolume()
		newVol := currVol + config.VolumeStep
		if !up {
			newVol = currVol - config.VolumeStep
		}
		setVol, err := m.player.SetVolume(newVol)
		if err != nil {
			return volumeMsg{err}
		}
		m.cfg.SetVolume(setVol)
		return volumeMsg{}
	}
}

func (m *Model) seekCmd(amtSec int) tea.Cmd {
	return func() tea.Msg {
		m.delegate.playingMtx.RLock()
		defer m.delegate.playingMtx.RUnlock()

		log := slog.With("method", "model.seekCmd")
		var s *browser.Station
		if m.delegate.currPlaying != nil {
			s = m.delegate.currPlaying
		} else if m.delegate.prevPlaying != nil {
			s = m.delegate.prevPlaying
		} else {
			return nil
		}
		metadata := m.player.Seek(amtSec)
		if metadata == nil {
			return nil
		} else if metadata.Err != nil {
			log.Error("seek", "error", metadata.Err)
			return nil
		}
		msg := getMetadataMsg(*s, *metadata)
		return msg
	}
}

func (m *Model) playStationCmd(selStation browser.Station) tea.Cmd {
	m.songTitle = ""
	m.playbackTime = 0
	m.updateStatus(fmt.Sprintf("Connecting to %s...", selStation.Name))
	cmds := []tea.Cmd{m.initSpinner(), m.delegate.playCmd(selStation)}
	return tea.Batch(cmds...)
}

func (m *Model) playUuidCmd(uuid string) tea.Cmd {
	return func() tea.Msg {
		stations, err := m.browser.GetStations([]string{uuid})
		res := playUuidRespMsg{stations: stations}
		if err != nil {
			res.statusMsg = statusMsg(err.Error())
		} else if len(stations) == 0 {
			res.viewMsg = noStationsFound
		}
		return res
	}
}
