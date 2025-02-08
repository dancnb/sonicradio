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
	ApiReqTimeout     = 10 * time.Second
	MpvIpcConnTimeout = 10 * time.Second

	VolumeStep  = 5
	SeekStepSec = 10

	defVersion  = "0.5.8"
	cfgSubDir   = "sonicRadio"
	cfgFilename = "config.json"
)

var (
	defVolume         = 100
	defHistorySaveMax = 100
)

type Value struct {
	Version   string   `json:"-"`
	Debug     bool     `json:"-"`
	LogPath   string   `json:"-"`
	Favorites []string `json:"favorites,omitempty"` // Ordered station UUID's for user favorites
	Volume    *int     `json:"volume,omitempty"`
	Theme     int      `json:"theme"`

	Player PlayerType `json:"playerType"`

	historyMtx     sync.Mutex          `json:"-"`
	History        []HistoryEntry      `json:"history,omitempty"`
	HistorySaveMax *int                `json:"historySaveMax,omitempty"`
	HistoryChan    chan []HistoryEntry `json:"-"`

	AutoplayFavorite string `json:"autoplayFavorite"`

	IsRunning bool `json:"isRunning"`

	saveMtx sync.Mutex
}

type PlayerType uint8

const (
	Mpv PlayerType = iota
	FFPlay
)

var Players = [2]PlayerType{Mpv, FFPlay}

var playerNames = map[PlayerType]string{
	Mpv:    "Mpv",
	FFPlay: "FFplay",
}

func (p PlayerType) String() string {
	return playerNames[p]
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

// ToggleFavorite return true if uuid was added, false if it was removed
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

// DeleteFavorite returns true if uuid was removed, false if not
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
	return fmt.Sprintf("{version:%q, debug: %v, logPath=%q, favorites=%d, volume=%d, history=%d, historySaveMax=%v}",
		v.Version, v.Debug, v.LogPath, len(v.Favorites), vol, len(v.History), v.HistorySaveMax)
}

// Load must return a non-nil config Value and an error specifying why it could not read the config file:
//
// - either a default value if no previously saved config is found in the file system
//
// - either the found config Value
func Load() (cfg *Value, err error) {
	flag.Parse()
	versionVal := os.Getenv("SONIC_VERSION")
	if versionVal == "" {
		versionVal = defVersion
	}

	cfg = &Value{
		Version:        versionVal,
		Debug:          *debug,
		LogPath:        os.TempDir(),
		Volume:         &defVolume,
		HistorySaveMax: &defHistorySaveMax,
		HistoryChan:    make(chan []HistoryEntry),
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return
	}
	fp := filepath.Join(dir, cfgSubDir, cfgFilename)
	f, err := os.Open(fp)
	if err != nil {
		return
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return
	}
	err = f.Close()
	if err != nil {
		return
	}

	if cfg.Volume == nil {
		cfg.Volume = &defVolume
	}
	if cfg.HistorySaveMax == nil {
		cfg.HistorySaveMax = &defHistorySaveMax
	}
	if len(cfg.History) > *cfg.HistorySaveMax {
		cfg.History = cfg.History[len(cfg.History)-*cfg.HistorySaveMax:]
	}
	return
}

func (v *Value) Save() error {
	v.saveMtx.Lock()
	defer v.saveMtx.Unlock()

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
	err = enc.Encode(v)
	if err != nil {
		return err
	}
	err = f.Close()
	return err
}
