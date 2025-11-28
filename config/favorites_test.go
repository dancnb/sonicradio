package config

import (
	"testing"

	"github.com/dancnb/sonicradio/model"
	"github.com/stretchr/testify/assert"
)

func Test_parsePlsFile(t *testing.T) {
	f := "favorites.pls"
	want := []model.Station{
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
	}
	res, err := parsePlsFile(f)
	assert.Nil(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, res, want)
}
