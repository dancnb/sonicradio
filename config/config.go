package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

var debug = flag.Bool("debug", false, "use -debug arg to log to a file")

const (
	ReqTimeout = 10 * time.Second

	defVersion        = "0.4.1"
	cfgSubDir         = "sonicRadio"
	cfgFilename       = "config.json"
	defHistorySaveMax = 100
)

var defVolume = 100

type Value struct {
	Version        string              `json:"-"`
	Debug          bool                `json:"-"`
	LogPath        string              `json:"-"`
	Favorites      []string            `json:"favorites,omitempty"` // Ordered station UUID's for user favorites
	Volume         *int                `json:"volume,omitempty"`
	historyMtx     *sync.Mutex         `json:"-"`
	History        []HistoryEntry      `json:"history,omitempty"`
	HistorySaveMax int                 `json:"historySaveMax"`
	HistoryChan    chan []HistoryEntry `json:"-"`
}

func (v *Value) GetVolume() int {
	if v.Volume != nil {
		return *v.Volume
	}
	return defVolume
}

func (v *Value) SetVolume(value int) {
	v.Volume = &value
}

func (v *Value) IsFavorite(uuid string) bool {
	return slices.Contains(v.Favorites, uuid)
}

func (v *Value) ToggleFavorite(uuid string) bool {
	l1 := len(v.Favorites)
	v.Favorites = slices.DeleteFunc(v.Favorites, func(el string) bool { return el == uuid })
	l2 := len(v.Favorites)
	if l2 == l1 {
		v.Favorites = append(v.Favorites, uuid)
		return true
	}
	return false
}

func (v *Value) DeleteFavorite(uuid string) bool {
	l1 := len(v.Favorites)
	v.Favorites = slices.DeleteFunc(v.Favorites, func(el string) bool { return el == uuid })
	l2 := len(v.Favorites)
	return l2 != l1
}

func (v *Value) InsertFavorite(uuid string, idx int) bool {
	if slices.Contains(v.Favorites, uuid) {
		return false
	}
	if idx >= len(v.Favorites) {
		v.Favorites = append(v.Favorites, uuid)
		return true
	}
	v.Favorites = slices.Insert(v.Favorites, idx, uuid)
	return true
}

func (v *Value) String() string {
	vol := -1
	if v.Volume != nil {
		vol = *v.Volume
	}
	return fmt.Sprintf("{version:%q, debug: %v, logPath=%q, favorites=%d, volume=%d, history=%d, historySaveMax=%d}",
		v.Version, v.Debug, v.LogPath, len(v.Favorites), vol, len(v.History), v.HistorySaveMax)
}

func Load() (Value, error) {
	flag.Parse()
	versionVal := os.Getenv("SONIC_VERSION")
	if versionVal == "" {
		versionVal = defVersion
	}

	cfg := Value{
		Version:        versionVal,
		Debug:          *debug,
		LogPath:        os.TempDir(),
		Volume:         &defVolume,
		historyMtx:     &sync.Mutex{},
		HistorySaveMax: defHistorySaveMax,
		HistoryChan:    make(chan []HistoryEntry),
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return cfg, err
	}
	fp := filepath.Join(dir, cfgSubDir, cfgFilename)
	f, err := os.Open(fp)
	if err != nil {
		return cfg, err
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return cfg, err
	}
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return cfg, err
	}
	err = f.Close()
	if err != nil {
		return cfg, err
	}

	if cfg.Volume == nil {
		cfg.Volume = &defVolume
	}
	if cfg.HistorySaveMax == 0 {
		cfg.HistorySaveMax = defHistorySaveMax
	}
	return cfg, nil
}

func Save(cfg Value) error {
	dir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	fp := filepath.Join(dir, cfgSubDir)
	_, err = os.Stat(fp)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = os.MkdirAll(fp, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	fp = filepath.Join(fp, cfgFilename)
	f, err := os.Create(fp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("  ", "  ")
	err = enc.Encode(cfg)
	if err != nil {
		return err
	}
	err = f.Close()
	return err
}
