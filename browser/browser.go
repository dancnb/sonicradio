package browser

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dancnb/sonicradio/config"
)

const (
	HOST              = "all.api.radio-browser.info"
	backup_server     = "https://de1.api.radio-browser.info/json/servers"
	serverMaxRetry    = 5
	serverRetryMillis = 200
	voteTimeout       = 10 * time.Minute
)

var serverErrMsg = errors.New("Server response not available.")

func NewApi(cfg config.Value) *Api {
	api := Api{
		cfg:           cfg,
		stationsCache: make(map[string][]Station),
		stationVotes:  make(map[string]time.Time),
	}
	res, err := api.getServersDNSLookup(HOST)
	if err != nil {
		res, err = api.getServerMirrors()
		if err != nil {
			msg := fmt.Errorf("could not retrieve %s servers: %w", HOST, err)
			slog.Error(msg.Error())
		}
	}
	slog.Debug("browser servers " + strings.Join(res, "; "))
	api.servers = append(api.servers, res...)

	return &api
}

type Api struct {
	cfg       config.Value
	servers   []string
	countries []Country
	langs     []Language

	stationsMtx   sync.Mutex
	stationsCache map[string][]Station

	stationVotes map[string]time.Time
}

func (a *Api) GetLanguages() ([]Language, error) {
	if len(a.langs) > 0 {
		return a.langs, nil
	}
	log := slog.With("method", "Api.GetLanguages")
	for i := 0; i < serverMaxRetry; i++ {
		res, err := a.doServerRequest(http.MethodGet, urlLangs, nil)
		if err != nil {
			log.Error("", "request error", err)
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		var languages []Language
		err = json.Unmarshal(res, &languages)
		if err != nil {
			log.Error("", "unmarshal error", err)
			log.Error("", "response", string(res))
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		log.Debug("", "length", len(languages))
		a.langs = languages
		return languages, nil
	}
	log.Warn("exceeded max retries")
	return nil, serverErrMsg
}

func (a *Api) GetCountries() ([]Country, error) {
	if len(a.countries) > 0 {
		return a.countries, nil
	}
	log := slog.With("method", "Api.GetCountries")
	for i := 0; i < serverMaxRetry; i++ {
		res, err := a.doServerRequest(http.MethodGet, urlCountries, nil)
		if err != nil {
			log.Error("", "request error", err)
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		var countries []Country
		err = json.Unmarshal(res, &countries)
		if err != nil {
			log.Error("", "unmarshal error", err)
			log.Error("", "response", string(res))
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		log.Debug("", "length", len(countries))
		a.countries = countries
		return countries, nil
	}
	log.Warn("exceeded max retries")
	return nil, serverErrMsg
}

func (a *Api) Search(s SearchParams) ([]Station, error) {
	return a.stationSearch(s)
}

func (a *Api) TopStations() ([]Station, error) {
	s := DefaultSearchParams()
	return a.stationSearch(s)
}

func (a *Api) stationSearch(s SearchParams) ([]Station, error) {
	body := s.toFormData()
	log := slog.With("method", "Api.stationSearch")
	log.Debug("", "request", body)

	a.stationsMtx.Lock()
	if v, ok := a.stationsCache[body]; ok && len(v) > 0 {
		a.stationsMtx.Unlock()
		log.Debug("stations cache hit", "len", len(v))
		return v, nil
	}
	a.stationsMtx.Unlock()
	log.Debug("stations cache miss")

	var err error
	for i := 0; i < serverMaxRetry; i++ {
		var res []byte
		res, err = a.doServerRequest(http.MethodPost, urlStations, []byte(body))
		if err != nil {
			log.Error("", "request error", err)
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		var stations []Station
		err = json.Unmarshal(res, &stations)
		if err != nil {
			log.Error("", "unmarshal error", err)
			log.Error("", "response", string(res))
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		log.Debug("", "length", len(stations))
		if len(stations) > 0 {
			a.stationsMtx.Lock()
			a.stationsCache[body] = stations
			a.stationsMtx.Unlock()
			log.Debug("stations cache set")
		}
		return stations, nil
	}
	log.Warn("exceeded max retries")
	return nil, serverErrMsg
}

func (a *Api) GetStations(uuids []string) ([]Station, error) {
	if len(uuids) == 0 {
		return nil, nil
	}
	log := slog.With("method", "Api.GetStations")
	var reqBody strings.Builder
	reqBody.WriteString(`uuids=`)
	for i, uuid := range uuids {
		reqBody.WriteString(uuid)
		if i < len(uuids)-1 {
			reqBody.WriteString(`,`)
		}
	}
	x := reqBody.String()
	for i := 0; i < serverMaxRetry; i++ {
		res, err := a.doServerRequest(http.MethodPost, urlStationsByUUID, []byte(x))
		if err != nil {
			log.Error("", "request error", err)
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		var stations []Station
		err = json.Unmarshal(res, &stations)
		if err != nil {
			log.Error("", "unmarshal error", err)
			log.Error("", "response", string(res))
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		log.Debug("", "length", len(stations))
		return stations, nil
	}

	log.Warn("exceeded max retries")
	return nil, serverErrMsg
}

func (a *Api) StationCounter(uuid string) error {
	log := slog.With("method", "Api.StationCounter")
	url := urlClickCount + uuid
	res, err := a.doServerRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Error("", "request error", err)
		return err
	}
	log.Debug(string(res))
	return nil
}

var errVoteTimeout = errors.New("Station was voted recently")

func (a *Api) StationVote(uuid string) error {
	log := slog.With("method", "Api.StationVote")

	if voteTime, ok := a.stationVotes[uuid]; ok && time.Now().Before(voteTime.Add(voteTimeout)) {
		log.Debug(fmt.Sprintf("already voted %s at %v", uuid, voteTime))
		return errVoteTimeout
	}
	a.stationVotes[uuid] = time.Now()

	url := urlVote + uuid
	res, err := a.doServerRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Error("", "request error", err)
		return err
	}
	log.Debug(string(res))
	return nil
}

func (a *Api) doServerRequest(method string, path string, body []byte) ([]byte, error) {
	ix := rand.IntN(len(a.servers))
	ip := a.servers[ix]
	url := fmt.Sprintf("http://%s%s", ip, path)
	return a.doRequest(method, url, body)
}

func (a *Api) getServersDNSLookup(host string) ([]string, error) {
	ips, err := net.DefaultResolver.LookupIP(context.Background(), "ip4", host)
	if err != nil {
		return nil, err
	}
	var res []string
	for _, v := range ips {
		res = append(res, v.String())
	}
	return res, nil
}

func (a *Api) getServerMirrors() ([]string, error) {
	res, err := a.doRequest(http.MethodGet, backup_server, nil)
	if err != nil {
		return nil, err
	}
	var srv []ServerMirror
	err = json.Unmarshal(res, &srv)
	if err != nil {
		return nil, err
	}
	var ips []string
	for _, server := range srv {
		ipVal := net.ParseIP(server.IP)
		if ipVal != nil && ipVal.To4() != nil {
			ips = append(ips, server.IP)
		}
	}

	return ips, err
}

func (a *Api) doRequest(method string, url string, body []byte) ([]byte, error) {
	log := slog.With("method", "Api.doRequest")

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		log.Error("create browser request", slog.String("error", err.Error()))
		return nil, err
	}
	ua := fmt.Sprintf("sonicradio/%s", a.cfg.Version)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", ua)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("do browser request", slog.String("error", err.Error()))
		return nil, err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error("read browser response", slog.String("error", err.Error()))
		return nil, err
	}
	return b, nil
}
