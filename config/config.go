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
	"text/template"
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

	cfgSubDir       = "sonicRadio"
	cfgFilename     = "config.json"
	historyFilename = "history.json"

	favoritesFilename = "favorites.pls"
	favoritesTmpl     = `[playlist]
NumberOfEntries={{len .List}}
Version=2

{{range $index, $station := .List}}File{{add $index 1}}={{.URL}}
Title{{add $index 1}}={{.Name}}
Length{{add $index 1}}=-1
SR_uuid{{add $index 1}}={{.Stationuuid}}
SR_bitrate{{add $index 1}}={{.Bitrate}}
SR_countrycode{{add $index 1}}={{.Countrycode}}
SR_state{{add $index 1}}={{.State}}
SR_language{{add $index 1}}={{.Language}}
SR_tags{{add $index 1}}={{.Tags}}
SR_homepage{{add $index 1}}={{.Homepage}}
SR_country{{add $index 1}}={{.Country}}
SR_votes{{add $index 1}}={{.Votes}}
SR_codec{{add $index 1}}={{.Codec}}
SR_lastcheckoktime{{add $index 1}}={{.Lastcheckoktime}}
SR_clickcount{{add $index 1}}={{.Clickcount}}
SR_clicktrend{{add $index 1}}={{.Clicktrend}}
SR_geo_lat{{add $index 1}}={{.GeoLat}}
SR_geo_long{{add $index 1}}={{.GeoLong}}

{{end}}`
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
	FavoritesV1 []string    `json:"favorites,omitempty"` // Ordered station UUID's for user favorites
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

	favTmpl *template.Template
}

type Favorites struct {
	RefreshOnStart bool `json:"refreshOnStart"`
	list           []model.Station
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
	return slices.ContainsFunc(v.Favorites.list, func(e model.Station) bool {
		return e.Stationuuid == uuid
	})
}

func (v *Value) HasFavoritesV1() bool {
	return len(v.FavoritesV1) > 0
}

func (v *Value) GetFavoritesV1() []string {
	return v.FavoritesV1
}

func (v *Value) HasFavorites() bool {
	return len(v.Favorites.list) > 0
}

func (v *Value) GetFavorites() []model.Station {
	return v.Favorites.list
}

func (v *Value) SetFavorites(l []model.Station) {
	v.Favorites.list = l
}

func (v *Value) FavoritesCacheEnabled() bool {
	return !v.Favorites.RefreshOnStart
}

// ToggleFavorite return true if station was added, false if it was removed
func (v *Value) ToggleFavorite(s model.Station) bool {
	l1 := len(v.Favorites.list)
	v.Favorites.list = slices.DeleteFunc(
		v.Favorites.list,
		func(el model.Station) bool { return el.Stationuuid == s.Stationuuid },
	)
	l2 := len(v.Favorites.list)
	if l2 == l1 {
		v.Favorites.list = append(v.Favorites.list, s)
		return true
	}
	return false
}

// DeleteFavorite returns true if station was removed, false if not
func (v *Value) DeleteFavorite(s model.Station) bool {
	l1 := len(v.Favorites.list)
	v.Favorites.list = slices.DeleteFunc(
		v.Favorites.list,
		func(el model.Station) bool { return el.Stationuuid == s.Stationuuid },
	)
	l2 := len(v.Favorites.list)
	return l2 != l1
}

func (v *Value) InsertFavorite(s model.Station, idx int) bool {
	if slices.ContainsFunc(v.Favorites.list, func(el model.Station) bool {
		return el.Stationuuid == s.Stationuuid
	}) {
		return false
	}
	if idx >= len(v.Favorites.list) {
		v.Favorites.list = append(v.Favorites.list, s)
		return true
	}
	v.Favorites.list = slices.Insert(v.Favorites.list, idx, s)
	return true
}

func (v *Value) String() string {
	vol := -1
	if v.Volume != nil {
		vol = *v.Volume
	}
	return fmt.Sprintf("{version:%q,   favorites=%d, volume=%d, history=%d, historySaveMax=%v}",
		v.Version, len(v.Favorites.list), vol, len(v.History), v.HistorySaveMax)
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
	cfg.Favorites.list, err = parsePlsFile(filepath.Join(cfgDirPath, favoritesFilename))
	if err != nil {
		return
	}
	cfg.favTmpl = template.Must(
		template.New("favorites").
			Funcs(template.FuncMap{
				"add": func(i, j int) int { return i + j },
			}).
			Parse(favoritesTmpl))

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

	if err := v.saveFavorites(filepath.Join(cfgDirPath, favoritesFilename)); err != nil {
		return err
	}
	v.FavoritesV1 = nil

	entries, err := v.saveCfgFile(cfgDirPath)
	if err != nil {
		return err
	}

	return v.saveHistoryFile(cfgDirPath, entries)
}

func (v *Value) saveFavorites(filename string) (err error) {
	favFile, err := os.Create(filename)
	if err != nil {
		return
	}
	defer func() {
		closeErr := favFile.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	list := struct{ List []model.Station }{
		List: v.Favorites.list,
	}
	err = v.favTmpl.Execute(favFile, list)
	if err != nil {
		return
	}
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
