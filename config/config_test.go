package config

import (
	"testing"
	"text/template"

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
	tests := []struct {
		name     string // description of this test case
		filename string
		wantErr  bool
	}{
		{
			name:     "1",
			filename: "./__test.pls",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Value{
				Favorites: Favorites{
					list: []model.Station{
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
			v.favTmpl = template.Must(
				template.New("favorites").
					Funcs(template.FuncMap{
						"add": func(i, j int) int { return i + j },
					}).
					Parse(favoritesTmpl))

			gotErr := v.saveFavorites(tt.filename)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("saveFavorites() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("saveFavorites() succeeded unexpectedly")
			}
		})
	}
}
