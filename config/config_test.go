package config

import (
	"os"
	"testing"
)

func Test_load(t *testing.T) {
	defCfg := Value{
		Version: defVersion,
		Debug:   *debug,
		LogPath: os.TempDir(),
		Volume:  &defVolume,
	}
	err := Save(defCfg)
	if err != nil {
		t.Error(err)
	}

	_, err = Load()
	if err != nil {
		t.Error(err)
	}
}

func Test_save(t *testing.T) {
	defCfg := Value{
		Version: defVersion,
		Debug:   *debug,
		LogPath: os.TempDir(),
		Volume:  &defVolume,
	}
	err := Save(defCfg)
	if err != nil {
		t.Error(err)
	}
}
