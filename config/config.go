package config

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

var debug = flag.Bool("debug", false, "use -debug arg to log to a file")

const (
	defVersion  = "0.1.0"
	cfgSubDir   = "sonicRadio"
	cfgFilename = "config.json"
)

type Value struct {
	Version   string   `json:"-"`
	Debug     bool     `json:"-"`
	LogPath   string   `json:"logPath"`             //`json:"logPath"`
	Favorites []string `json:"favorites,omitempty"` // Ordered station UUID's for user favorites
	History   []string `json:"history,omitempty"`   // Ordered station UUID's for user listening history
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
