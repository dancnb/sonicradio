package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dancnb/sonicradio/model"
)

var debug = flag.Bool("debug", false, "use -debug arg to log to a file")

const (
	APIReqTimeout     = 10 * time.Second
	MpvIpcConnTimeout = 10 * time.Second
	MpdConnTimeout    = 10 * time.Second
	VlcConnTimeout    = 10 * time.Millisecond

	VolumeStep  = 5
	SeekStepSec = 10

	cfgSubDir         = "sonicRadio"
	cfgFilename       = "config.json"
	historyFilename   = "history.json"
	favoritesFilename = "favorites.json"
)

const (
	DefVolume                  = 100
	DefHistorySaveMax          = 100
	DefFavoritesRefreshOnStart = false

	DefMpdHost = ""
	DefMpdPort = 6600

	DefInternalBufferSeconds = 0
)

type Value struct {
	Version     string      `json:"-"`
	Favorites   Favorites   `json:"favoritesv2"`
	Volume      *int        `json:"volume,omitempty"`
	Theme       int         `json:"theme"`
	StationView StationView `json:"stationView"`

	Player         PlayerType `json:"playerType"`
	MpdHost        string     `json:"mpdHost,omitempty"`
	MpdPort        int        `json:"mpdPort,omitempty"`
	MpdPassword    *string    `json:"mpdPassword,omitempty"`
	mpdEnvPassword *string    `json:"-"`

	Internal InternalPlayer `json:"internal"`

	historyMtx     sync.Mutex          `json:"-"`
	History        []HistoryEntry      `json:"history,omitempty"`
	HistorySaveMax *int                `json:"historySaveMax,omitempty"`
	HistoryChan    chan []HistoryEntry `json:"-"`

	AutoplayFavorite string `json:"autoplayFavorite"`
}

type Favorites struct {
	RefreshOnStart bool            `json:"refreshOnStart"`
	List           []model.Station `json:"-"`
}

type InternalPlayer struct {
	BufferSeconds int `json:"bufferSeconds"`
}

type PlayerType uint8

const (
	Mpv PlayerType = iota
	FFPlay
	Vlc
	MPlayer
	MPD
	Internal
)

var Players = [6]PlayerType{Mpv, FFPlay, Vlc, MPlayer, MPD, Internal}

var playerNames = map[PlayerType]string{
	Mpv:      "Mpv",
	FFPlay:   "FFplay",
	Vlc:      "VLC",
	MPlayer:  "MPlayer",
	MPD:      "MPD",
	Internal: "Internal (experimental)",
}

func (p PlayerType) String() string {
	return playerNames[p]
}

func (v *Value) GetVolume() int {
	if v.Volume != nil {
		return *v.Volume
	}
	return DefVolume
}

func (v *Value) SetVolume(value int) {
	v.Volume = &value
}

func (v *Value) IsFavorite(uuid string) bool {
	return slices.ContainsFunc(v.Favorites.List, func(e model.Station) bool {
		return e.Stationuuid == uuid
	})
}

func (v *Value) HasFavorites() bool {
	return len(v.Favorites.List) > 0
}

func (v *Value) GetFavorites() []model.Station {
	return v.Favorites.List
}

func (v *Value) SetFavorites(l []model.Station) {
	v.Favorites.List = l
}

func (v *Value) AddFavorite(s model.Station) {
	v.Favorites.List = append(v.Favorites.List, s)
}

func (v *Value) FavoritesCacheEnabled() bool {
	return !v.Favorites.RefreshOnStart
}

// ToggleFavorite return true if station was added, false if it was removed
func (v *Value) ToggleFavorite(s model.Station) bool {
	l1 := len(v.Favorites.List)
	v.Favorites.List = slices.DeleteFunc(
		v.Favorites.List,
		func(el model.Station) bool { return el.Stationuuid == s.Stationuuid },
	)
	l2 := len(v.Favorites.List)
	if l2 == l1 {
		v.Favorites.List = append(v.Favorites.List, s)
		return true
	}
	return false
}

// DeleteFavorite returns true if station was removed, false if not
func (v *Value) DeleteFavorite(s model.Station) bool {
	l1 := len(v.Favorites.List)
	v.Favorites.List = slices.DeleteFunc(
		v.Favorites.List,
		func(el model.Station) bool { return el.Stationuuid == s.Stationuuid },
	)
	l2 := len(v.Favorites.List)
	return l2 != l1
}

func (v *Value) InsertFavorite(s model.Station, idx int) bool {
	if slices.ContainsFunc(v.Favorites.List, func(el model.Station) bool {
		return el.Stationuuid == s.Stationuuid
	}) {
		return false
	}
	if idx >= len(v.Favorites.List) {
		v.Favorites.List = append(v.Favorites.List, s)
		return true
	}
	v.Favorites.List = slices.Insert(v.Favorites.List, idx, s)
	return true
}

func (v *Value) String() string {
	vol := -1
	if v.Volume != nil {
		vol = *v.Volume
	}
	return fmt.Sprintf("{version:%q,   favorites=%d, volume=%d, history=%d, historySaveMax=%v}",
		v.Version, len(v.Favorites.List), vol, len(v.History), v.HistorySaveMax)
}

// Load must return a non-nil config Value and an error specifying why it could not read the config file:
//
// - either a default value if no previously saved config is found in the file system
//
// - either the found config Value
func Load(version string) (cfg *Value, err error) {
	defVolume := DefVolume
	defHistorySaveMax := DefHistorySaveMax
	cfg = &Value{
		Version:        version,
		Volume:         &defVolume,
		HistorySaveMax: &defHistorySaveMax,
		HistoryChan:    make(chan []HistoryEntry),
	}

	cfgDirPath, err := getOrCreateConfigDir()
	if err != nil {
		return
	}
	f, err := os.Open(filepath.Join(cfgDirPath, cfgFilename))
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

	// history
	if cfg.HistorySaveMax == nil {
		cfg.HistorySaveMax = &defHistorySaveMax
	}
	err = cfg.loadHistory(filepath.Join(cfgDirPath, historyFilename))
	if err != nil {
		return
	}
	if len(cfg.History) > *cfg.HistorySaveMax {
		cfg.History = cfg.History[len(cfg.History)-*cfg.HistorySaveMax:]
	}

	// favorites
	err = cfg.loadFavorites(filepath.Join(cfgDirPath, favoritesFilename))
	if err != nil {
		return
	}

	// environment variables overwrite config file
	if v, ok := os.LookupEnv("MPD_HOST"); ok && v != "" {
		parts := strings.Split(v, "@")
		if len(parts) == 1 {
			cfg.MpdHost = parts[0]
		} else {
			cfg.mpdEnvPassword = &parts[0]
			cfg.MpdHost = parts[1]
		}
	}

	if v, ok := os.LookupEnv("MPD_PORT"); ok && v != "" {
		intV, _ := strconv.Atoi(v)
		cfg.MpdPort = intV
	} else if cfg.MpdPort == 0 {
		cfg.MpdPort = DefMpdPort
	}

	return
}

func (v *Value) loadHistory(historyFilePath string) (err error) {
	// if history found in config file (older version)
	if len(v.History) > 0 {
		return
	}

	_, err = os.Stat(historyFilePath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	hf, err := os.Open(historyFilePath)
	if err != nil {
		return
	}
	defer func() {
		closeErr := hf.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	b, err := io.ReadAll(hf)
	if err != nil {
		return
	}
	var entries []HistoryEntry
	err = json.Unmarshal(b, &entries)
	if err != nil {
		return
	}
	v.History = entries

	return
}

func (v *Value) loadFavorites(favoritesFilePath string) (err error) {
	// if favorites found in config file (older version)
	if len(v.Favorites.List) > 0 {
		return
	}

	_, err = os.Stat(favoritesFilePath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	hf, err := os.Open(favoritesFilePath)
	if err != nil {
		return
	}
	defer func() {
		closeErr := hf.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	b, err := io.ReadAll(hf)
	if err != nil {
		return
	}
	var entries []model.Station
	err = json.Unmarshal(b, &entries)
	if err != nil {
		return
	}
	v.Favorites.List = entries

	return
}

func (v *Value) GetMpdPassword() *string {
	if v.mpdEnvPassword != nil {
		return v.mpdEnvPassword
	}
	return v.MpdPassword
}

func (v *Value) Save() error {
	cfgDirPath, err := getOrCreateConfigDir()
	if err != nil {
		return err
	}

	if err := v.saveFavorites(cfgDirPath, v.Favorites.List); err != nil {
		return err
	}

	entries, err := v.saveCfgFile(cfgDirPath)
	if err != nil {
		return err
	}

	return v.saveHistoryFile(cfgDirPath, entries)
}

func (v *Value) saveFavorites(cfgDirPath string, entries []model.Station) (err error) {
	favoritesFile, err := os.Create(filepath.Join(cfgDirPath, favoritesFilename))
	if err != nil {
		return err
	}
	defer func() {
		closeErr := favoritesFile.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	enc := json.NewEncoder(favoritesFile)
	enc.SetIndent("  ", "  ")
	err = enc.Encode(entries)

	return
}

func (v *Value) saveCfgFile(cfgDirPath string) (entries []HistoryEntry, err error) {
	cfgFile, err := os.Create(filepath.Join(cfgDirPath, cfgFilename))
	if err != nil {
		return
	}
	defer func() {
		closeErr := cfgFile.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	enc := json.NewEncoder(cfgFile)
	enc.SetIndent("  ", "  ")
	entries = slices.Clone(v.History)
	v.History = nil
	err = enc.Encode(v)
	if err != nil {
		return
	}

	return
}

func (*Value) saveHistoryFile(cfgDirPath string, entries []HistoryEntry) (err error) {
	historyFile, err := os.Create(filepath.Join(cfgDirPath, historyFilename))
	if err != nil {
		return
	}
	defer func() {
		closeErr := historyFile.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	enc := json.NewEncoder(historyFile)
	enc.SetIndent("  ", "  ")
	err = enc.Encode(entries)

	return
}

func getOrCreateConfigDir() (string, error) {
	logger := slog.With("method", "getOrCreateConfigDir")

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("get user config dir: %v", err)
	}

	fp := filepath.Join(dir, cfgSubDir)
	_, err = os.Stat(fp)
	if err == nil {
		logger.Info(fmt.Sprintf("found config dir at path %s", fp))
		return fp, nil
	}

	if !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("checking config dir at path %s", fp)
	}

	logger.Info(fmt.Sprintf("creating config dir at path %s", fp))
	if err = os.MkdirAll(fp, os.ModePerm); err != nil {
		return "", fmt.Errorf("creating config dir at path %s: %v", fp, err)
	}

	return fp, nil
}

func Debug() bool {
	return *debug
}
