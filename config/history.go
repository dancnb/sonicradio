package config

import (
	"log/slog"
	"time"
)

const (
	recentlyPlayed = 2 * time.Minute
)

type historyEntry struct {
	Uuid      string    `json:"uuid"`
	Station   string    `json:"station"`
	Song      string    `json:"song"`
	Timestamp time.Time `json:"timestamp"`
}

func (v *Value) AddHistory(timestamp time.Time, uuid string, station string, song string) {
	v.historyMtx.Lock()
	defer v.historyMtx.Unlock()

	log := slog.With("method", "config.Value.AddHistory")
	log.Debug("", "uuid", uuid, "stationName", station, "song", song)
	if ok := v.updateHistory(timestamp, uuid, station, song); ok {
		startIx := max(0, len(v.History)-v.HistorySaveMax)
		err := Save(Value{
			Favorites:      v.Favorites,
			Volume:         v.Volume,
			History:        v.History[startIx:len(v.History)],
			HistorySaveMax: v.HistorySaveMax,
		})
		if err != nil {
			log.Error("save config", "err", err)
		}
	}
}

func (v *Value) updateHistory(timestamp time.Time, uuid string, station string, song string) bool {
	newEntry := historyEntry{
		Uuid:      uuid,
		Station:   station,
		Song:      song,
		Timestamp: timestamp,
	}
	minTs := timestamp.Add(-recentlyPlayed)
	found := false
	for i := len(v.History) - 1; i >= 0; i-- {
		entry := v.History[i]

		if i == len(v.History)-1 && entry.Uuid == newEntry.Uuid {
			if v.equalEntries(entry, newEntry) {
				return false
			} else if entry.Song == "" {
				v.History[i] = newEntry
				return true
			} else if newEntry.Song == "" {
				return false
			}
		}

		if entry.Timestamp.Before(minTs) {
			break
		} else if v.equalEntries(entry, newEntry) {
			found = true
			break
		}
	}
	if !found {
		v.History = append(v.History, newEntry)
		return true
	}
	return false
}

func (v *Value) equalEntries(a, b historyEntry) bool {
	x := a.Uuid == b.Uuid && a.Song == b.Song
	return x
}
