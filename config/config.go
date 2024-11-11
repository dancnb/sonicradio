package config

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

var debug = flag.Bool("debug", false, "use -debug arg to log to a file")

const (
	defVersion  = "0.3.2"
	cfgSubDir   = "sonicRadio"
	cfgFilename = "config.json"
)

var defVolume = 100

type Value struct {
	Version   string   `json:"-"`
	Debug     bool     `json:"-"`
	LogPath   string   `json:"-"`
	Favorites []string `json:"favorites,omitempty"` // Ordered station UUID's for user favorites
	Volume    *int     `json:"volume,omitempty"`
	// History   []string `json:"history,omitempty"`   // Ordered station UUID's for user listening history
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

func Load() (Value, error) {
	flag.Parse()
	versionVal := os.Getenv("SONIC_VERSION")
	if versionVal == "" {
		versionVal = defVersion
	}

	defCfg := Value{
		Version: versionVal,
		Debug:   *debug,
		LogPath: os.TempDir(),
		Volume:  &defVolume,
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return defCfg, err
	}
	fp := filepath.Join(dir, cfgSubDir, cfgFilename)

	f, err := os.Open(fp)
	if err != nil {
		return defCfg, err
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return defCfg, err
	}
	var cfg Value
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return defCfg, err
	}
	err = f.Close()
	if err != nil {
		return defCfg, err
	}

	if cfg.Volume == nil {
		cfg.Volume = &defVolume
	}
	cfg.Debug = *debug
	cfg.Version = versionVal
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
	err = json.NewEncoder(f).Encode(cfg)
	if err != nil {
		return err
	}
	err = f.Close()
	return err
}
