package config

import (
	"fmt"
	"log/slog"
	"slices"
	"time"
)

const (
	recentlyPlayed     = 2 * time.Minute
	tsFormat           = "15:04 02.01.2006"
	histTitleSeparator = "|"
)

type HistoryEntry struct {
	Uuid      string    `json:"uuid"`
	Station   string    `json:"station"`
	Song      string    `json:"song"`
	Timestamp time.Time `json:"timestamp"`
}

func (e HistoryEntry) FilterValue() string {
	return e.Station + e.Song
}

func (e HistoryEntry) Title() string {
	ts := e.Timestamp.Format(tsFormat)
	return fmt.Sprintf("%s %s %s", ts, histTitleSeparator, e.Station)
}

func (e HistoryEntry) Description() string { return e.Song }

func (v *Value) DeleteHistoryEntry(delEntry HistoryEntry) {
	v.historyMtx.Lock()
	defer v.historyMtx.Unlock()

	v.History = slices.DeleteFunc(v.History, func(e HistoryEntry) bool {
		return e.Timestamp.Equal(delEntry.Timestamp)
	})
	v.updateHistoryEntries()
}

func (v *Value) ClearHistory() {
	v.historyMtx.Lock()
	defer v.historyMtx.Unlock()

	v.History = v.History[:0]
	v.updateHistoryEntries()
}

func (v *Value) updateHistoryEntries() []HistoryEntry {
	log := slog.With("method", "config.Value.updateHistoryEntries")
	startIx := max(0, len(v.History)-*v.HistorySaveMax)
	entries := v.History[startIx:len(v.History)]
	v.History = entries
	log.Info("updated entries", "len", len(entries), "startIdx", startIx)
	return entries
}

func (v *Value) AddHistoryEntry(timestamp time.Time, uuid string, station string, song string) {
	v.historyMtx.Lock()
	defer v.historyMtx.Unlock()

	log := slog.With("method", "config.Value.AddHistory")
	log.Info("", "uuid", uuid, "stationName", station, "song", song)

	if ok := v.upsertHistory(timestamp, uuid, station, song); ok {
		entries := v.updateHistoryEntries()
		v.HistoryChan <- entries
	}
}

func (v *Value) upsertHistory(timestamp time.Time, uuid string, station string, song string) bool {
	newEntry := HistoryEntry{
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

func (v *Value) equalEntries(a, b HistoryEntry) bool {
	x := a.Uuid == b.Uuid && a.Song == b.Song
	return x
}
