package config

import (
	"testing"
)

func Test_load(t *testing.T) {
	testLoadConfig(t)
}

func testLoadConfig(t *testing.T) (*Value, error) {
	cfg, err := Load()
	if cfg == nil {
		t.Error("config load: expected a non-nil config")
	}
	if err != nil {
		t.Log(err)
	}
	return cfg, err
}

func Test_save(t *testing.T) {
	cfg, _ := testLoadConfig(t)
	err := cfg.Save()
	if err != nil {
		t.Error(err)
	}
}
