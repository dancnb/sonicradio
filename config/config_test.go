package config

import (
	"os"
	"testing"

	"github.com/dancnb/sonicradio/model"
)

func Test_load(t *testing.T) {
	testLoadConfig(t)
}

func testLoadConfig(t *testing.T) (*Value, error) {
	cfg, err := Load("test")
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

func TestValue_saveFavorites(t *testing.T) {
	v := &Value{
		Favorites: Favorites{
			List: []model.Station{
				{
					Stationuuid:     "96133c49-0601-11e8-ae97-52543be04c81",
					Name:            "My Zen Relax",
					URL:             "http://vibration.stream2net.eu:8220/;stream/1",
					Bitrate:         128,
					Countrycode:     "CH",
					State:           "Suisse Romande",
					Homepage:        "http://www.vibration108.ch/player/vibrationzenrelax/",
					Country:         "Switzerland",
					Votes:           111111,
					Codec:           "MP3",
					Lastcheckoktime: "",
					Clickcount:      221,
					Clicktrend:      10,
					GeoLat:          "44.4",
					GeoLong:         "26.14",
				},
				{
					Stationuuid:     "748d830c-d934-41e8-bd14-870add931e1d",
					Name:            "My Radio Deea",
					URL:             "http://radiocdn.nxthost.com/radio-deea",
					Bitrate:         320,
					Countrycode:     "RO",
					State:           "Bucharest",
					Language:        "english,romanian",
					Tags:            "club,dance",
					Homepage:        "https://radiodeea.ro/",
					Country:         "Romania",
					Votes:           111112,
					Codec:           "MP3",
					Lastcheckoktime: "",
					Clickcount:      222,
					Clicktrend:      11,
					GeoLat:          "44.4",
					GeoLong:         "26.14",
				},
			},
		},
	}
	tmpDir, err := os.MkdirTemp("", "sonicradio_*")
	if err != nil {
		t.Fatal("unable to create temporary directory")
	}
	err = v.saveFavorites(tmpDir, v.Favorites.List)
	if err != nil {
		t.Errorf("saveFavorites() failed: %v", err)
		return
	}
	t.Logf("test saved favorites to: %s", tmpDir)
}
