package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/dancnb/sonicradio/model"
)

const (
	plsSep = "="

	prefixFile            = "file"
	prefixTitle           = "title"
	prefixUUID            = "sr_uuid"
	prefixBitrate         = "sr_bitrate"
	prefixCountrycode     = "sr_countrycode"
	prefixState           = "sr_state"
	prefixLanguage        = "sr_language"
	prefixTags            = "sr_tags"
	prefixHomepage        = "sr_homepage"
	prefixCountry         = "sr_country"
	prefixVotes           = "sr_votes"
	prefixCodec           = "sr_codec"
	prefixLastcheckoktime = "sr_lastcheckoktime"
	prefixClickcount      = "sr_clickcount"
	prefixClicktrend      = "sr_clicktrend"
	prefixGeolat          = "sr_geo_lat"
	prefixGeolong         = "sr_geo_long"
)

func getStringValue(line string) string {
	p := strings.Split(line, plsSep)
	if len(p) == 2 && len(p[1]) > 0 {
		return strings.TrimSpace(p[1])
	}
	return ""
}

func parsePlsFile(filename string) ([]model.Station, error) {
	if _, err := os.Stat(filename); errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	b, err := os.ReadFile(filename)
	if err != nil {
		slog.Error(fmt.Sprintf("open pls file %s: %v", filename, err))
		return nil, fmt.Errorf("could not read file: %s", filename)
	}

	var res []model.Station
	var elem *model.Station

	s := bufio.NewScanner(bytes.NewReader(b))
	for s.Scan() {
		l := s.Text()
		ll := strings.ToLower(l)

		if strings.HasPrefix(ll, prefixFile) {
			if elem != nil && len(elem.Stationuuid) > 0 {
				res = append(res, *elem)
				elem = nil
			}
			if v := getStringValue(l); v != "" {
				elem = &model.Station{
					URL: v,
				}
			}
		} else if elem == nil {
			continue
		}

		switch {
		case strings.HasPrefix(ll, prefixTitle):
			if v := getStringValue(l); v != "" {
				elem.Name = v
			}
		case strings.HasPrefix(ll, prefixUUID):
			if v := getStringValue(l); v != "" {
				elem.Stationuuid = v
			}
		case strings.HasPrefix(ll, prefixBitrate):
			if v := getStringValue(l); v != "" {
				bitR, err := strconv.Atoi(v)
				if err != nil {
					slog.Error(fmt.Sprintf("invalid bitrate value: %v", v))
					continue
				}
				elem.Bitrate = int64(bitR)
			}
		case strings.HasPrefix(ll, prefixCountrycode):
			if v := getStringValue(l); v != "" {
				elem.Countrycode = v
			}
		case strings.HasPrefix(ll, prefixState):
			if v := getStringValue(l); v != "" {
				elem.State = v
			}
		case strings.HasPrefix(ll, prefixLanguage):
			if v := getStringValue(l); v != "" {
				elem.Language = v
			}
		case strings.HasPrefix(ll, prefixTags):
			if v := getStringValue(l); v != "" {
				elem.Tags = v
			}
		//
		case strings.HasPrefix(ll, prefixHomepage):
			if v := getStringValue(l); v != "" {
				elem.Homepage = v
			}
		case strings.HasPrefix(ll, prefixCountry):
			if v := getStringValue(l); v != "" {
				elem.Country = v
			}
		case strings.HasPrefix(ll, prefixVotes):
			if v := getStringValue(l); v != "" {
				nr, err := strconv.Atoi(v)
				if err != nil {
					slog.Error(fmt.Sprintf("invalid votes value: %v", v))
					continue
				}
				elem.Votes = int64(nr)
			}
		case strings.HasPrefix(ll, prefixCodec):
			if v := getStringValue(l); v != "" {
				elem.Codec = v
			}
		case strings.HasPrefix(ll, prefixLastcheckoktime):
			if v := getStringValue(l); v != "" {
				elem.Lastcheckoktime = v
			}
		case strings.HasPrefix(ll, prefixClickcount):
			if v := getStringValue(l); v != "" {
				nr, err := strconv.Atoi(v)
				if err != nil {
					slog.Error(fmt.Sprintf("invalid clickcount value: %v", v))
					continue
				}
				elem.Clickcount = int64(nr)
			}
		case strings.HasPrefix(ll, prefixClicktrend):
			if v := getStringValue(l); v != "" {
				nr, err := strconv.Atoi(v)
				if err != nil {
					slog.Error(fmt.Sprintf("invalid clicktrend value: %v", v))
					continue
				}
				elem.Clicktrend = int64(nr)
			}
		case strings.HasPrefix(ll, prefixGeolat):
			if v := getStringValue(l); v != "" {
				elem.GeoLat = v
			}
		case strings.HasPrefix(ll, prefixGeolong):
			if v := getStringValue(l); v != "" {
				elem.GeoLong = v
			}
		}

	}

	if err := s.Err(); err != nil {
		slog.Error(fmt.Sprintf("%s scan error: %v", filename, err))
	}

	if elem != nil && len(elem.Stationuuid) > 0 {
		res = append(res, *elem)
		elem = nil
	}

	return res, nil
}
