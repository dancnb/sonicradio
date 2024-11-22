package config

import "time"

type historyEntry struct {
	Uuid      string    `json:"uuid"`
	Station   string    `json:"station"`
	Song      string    `json:"song"`
	Timestamp time.Time `json:"timestamp"`
}
