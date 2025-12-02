package ui

import (
	"fmt"
	"log/slog"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dancnb/sonicradio/config"
	"github.com/dancnb/sonicradio/model"
)

func (m *Model) favoritesReqCmd() tea.Msg {

	var reqList []string
	favorites := m.cfg.GetFavorites()
	slog.Info(fmt.Sprintf("cached favorites len: %#v", len(favorites)))
	for _, s := range favorites {
		if m.cfg.FavoritesCacheEnabled() || s.IsCustom {
			continue
		}
		reqList = append(reqList, s.Stationuuid)
	}
	slices.Sort(reqList)
	reqList = slices.Compact(reqList)
	slog.Info(fmt.Sprintf("favorites request list: %#v", reqList))
	if len(reqList) == 0 {
		return favoritesStationRespMsg{stations: favorites}
	}

	newStations, err := m.browser.GetStations(reqList)
	for i := range newStations {
		found := false
		for j := range favorites {
			if favorites[j].Stationuuid != newStations[i].Stationuuid {
				continue
			}
			found = true
			favorites[j] = newStations[i]
			break
		}
		if !found {
			favorites = append(favorites, newStations[i])
		}
	}
	m.cfg.SetFavorites(favorites)
	slog.Info(fmt.Sprintf("updated favorites len: %#v", len(favorites)))

	res := favoritesStationRespMsg{stations: favorites}
	if err != nil {
		res.statusMsg = statusMsg(err.Error())
	} else if len(favorites) == 0 {
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
		log := slog.With("method", "ui.Model.seekCmd")
		log.Info("begin")
		defer log.Info("end")

		m.delegate.playingMtx.RLock()
		defer m.delegate.playingMtx.RUnlock()

		var s *model.Station
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

func (m *Model) playStationCmd(selStation model.Station) tea.Cmd {
	m.songTitle = ""
	m.playbackTime = 0
	m.updateStatus(fmt.Sprintf("Connecting to %s...", selStation.Name))
	cmds := []tea.Cmd{m.initSpinner(), m.delegate.playCmd(selStation)}
	return tea.Batch(cmds...)
}

func (m *Model) playUUIDCmd(uuid string) tea.Cmd {
	return func() tea.Msg {
		stations, err := m.browser.GetStations([]string{uuid})
		res := playUUIDRespMsg{stations: stations}
		if err != nil {
			res.statusMsg = statusMsg(err.Error())
		} else if len(stations) == 0 {
			res.viewMsg = noStationsFound
		}
		return res
	}
}
