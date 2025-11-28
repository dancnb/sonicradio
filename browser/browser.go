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
	"github.com/dancnb/sonicradio/model"
)

const (
	HOST              = "radio-browser.info"
	backupServer      = "https://de1.api.radio-browser.info/json/servers"
	serverMaxRetry    = 5
	serverRetryMillis = 200
	voteTimeout       = 10 * time.Minute
)

var ErrServerMsg = errors.New("server response not available")

func NewAPI(ctx context.Context, cfg *config.Value) (*API, error) {
	api := API{
		cfg:           cfg,
		stationsCache: make(map[string][]model.Station),
		stationVotes:  make(map[string]time.Time),
	}
	res, err := api.getServers(ctx, HOST)
	if err != nil {
		msg := fmt.Errorf("could not perform DNS lookup for %q: %w", HOST, err)
		slog.Error(msg.Error())
		res, err = api.getServerMirrors()
		if err != nil {
			msg := fmt.Errorf("could not retrieve %s servers: %w", HOST, err)
			slog.Error(msg.Error())
		}
	}
	slog.Info("browser servers: " + strings.Join(res, "; "))
	api.servers = append(api.servers, res...)

	if len(api.servers) == 0 {
		return nil, ErrServerMsg
	}
	return &api, nil
}

type API struct {
	cfg       *config.Value
	servers   []string
	countries []model.Country
	langs     []model.Language

	stationsMtx   sync.Mutex
	stationsCache map[string][]model.Station

	stationVotes map[string]time.Time
}

func (a *API) GetLanguages() ([]model.Language, error) {
	if len(a.langs) > 0 {
		return a.langs, nil
	}
	log := slog.With("method", "Api.GetLanguages")
	for range serverMaxRetry {
		res, err := a.doServerRequest(http.MethodGet, urlLangs, nil)
		if err != nil {
			log.Error("", "request error", err)
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		var languages []model.Language
		err = json.Unmarshal(res, &languages)
		if err != nil {
			log.Error("", "unmarshal error", err)
			log.Error("", "response", string(res))
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		log.Info("", "length", len(languages))
		a.langs = languages
		return languages, nil
	}
	log.Warn("exceeded max retries")
	return nil, fmt.Errorf("Get languages: %w", ErrServerMsg)
}

func (a *API) GetCountries() ([]model.Country, error) {
	if len(a.countries) > 0 {
		return a.countries, nil
	}
	log := slog.With("method", "Api.GetCountries")
	for range serverMaxRetry {
		res, err := a.doServerRequest(http.MethodGet, urlCountries, nil)
		if err != nil {
			log.Error("", "request error", err)
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		var countries []model.Country
		err = json.Unmarshal(res, &countries)
		if err != nil {
			log.Error("", "unmarshal error", err)
			log.Error("", "response", string(res))
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		log.Info("", "length", len(countries))
		a.countries = countries
		return countries, nil
	}
	log.Warn("exceeded max retries")
	return nil, fmt.Errorf("Get countries: %w", ErrServerMsg)
}

func (a *API) Search(s SearchParams) ([]model.Station, error) {
	return a.stationSearch(s)
}

func (a *API) TopStations() ([]model.Station, error) {
	s := DefaultSearchParams()
	return a.stationSearch(s)
}

func (a *API) stationSearch(s SearchParams) ([]model.Station, error) {
	body := s.toFormData()
	log := slog.With("method", "Api.stationSearch")
	log.Info("", "request", body)

	a.stationsMtx.Lock()
	if v, ok := a.stationsCache[body]; ok && len(v) > 0 {
		a.stationsMtx.Unlock()
		log.Info("stations cache hit", "len", len(v))
		return v, nil
	}
	a.stationsMtx.Unlock()
	log.Info("stations cache miss")

	var err error
	for range serverMaxRetry {
		var res []byte
		res, err = a.doServerRequest(http.MethodPost, urlStations, []byte(body))
		if err != nil {
			log.Error("", "request error", err)
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		var stations []model.Station
		err = json.Unmarshal(res, &stations)
		if err != nil {
			log.Error("", "unmarshal error", err)
			log.Error("", "response", string(res))
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		log.Info("", "length", len(stations))
		if len(stations) > 0 {
			a.stationsMtx.Lock()
			a.stationsCache[body] = stations
			a.stationsMtx.Unlock()
			log.Info("stations cache set")
		}
		return stations, nil
	}
	log.Warn("exceeded max retries")
	return nil, fmt.Errorf("Get stations: %w", ErrServerMsg)
}

func (a *API) GetStations(uuids []string) ([]model.Station, error) {
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
	for range serverMaxRetry {
		res, err := a.doServerRequest(http.MethodPost, urlStationsByUUID, []byte(x))
		if err != nil {
			log.Error("", "request error", err)
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		var stations []model.Station
		err = json.Unmarshal(res, &stations)
		if err != nil {
			log.Error("", "unmarshal error", err)
			log.Error("", "response", string(res))
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		log.Info("", "length", len(stations))
		return stations, nil
	}

	log.Warn("exceeded max retries")
	return nil, fmt.Errorf("Get stations: %w", ErrServerMsg)
}

func (a *API) StationCounter(uuid string) error {
	log := slog.With("method", "Api.StationCounter")
	url := urlClickCount + uuid
	res, err := a.doServerRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Error("", "request error", err)
		return err
	}
	log.Info(string(res))
	return nil
}

var (
	errVoteTimeout = errors.New("Station was voted recently")
	errVoteReq     = errors.New("Vote request error")
	errVoteOften   = errors.New("You are voting for the same station too often")
)

func (a *API) StationVote(uuid string) error {
	log := slog.With("method", "Api.StationVote")

	if voteTime, ok := a.stationVotes[uuid]; ok && time.Now().Before(voteTime.Add(voteTimeout)) {
		log.Info(fmt.Sprintf("already voted %s at %v", uuid, voteTime))
		return errVoteTimeout
	}
	a.stationVotes[uuid] = time.Now()

	url := urlVote + uuid
	res, err := a.doServerRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Error("", "request error", err)
		return errVoteReq
	}
	log.Info(string(res))
	var voteRes struct {
		Ok      bool
		Message string
	}
	err = json.Unmarshal(res, &voteRes)
	if err != nil {
		return errVoteReq
	} else if strings.Contains(voteRes.Message, "you are voting for the same station too often") {
		return errVoteOften
	}
	return nil
}

func (a *API) doServerRequest(method string, path string, body []byte) ([]byte, error) {
	ix := rand.IntN(len(a.servers))
	host := a.servers[ix]
	url := fmt.Sprintf("http://%s%s", host, path)
	return a.doRequest(method, url, body)
}

func (a *API) getServers(ctx context.Context, name string) ([]string, error) {
	_, hosts, err := net.DefaultResolver.LookupSRV(ctx, "api", "tcp", name)
	if err != nil {
		return nil, err
	}
	var res []string
	for _, v := range hosts {
		res = append(res, v.Target)
	}
	return res, nil
}

func (a *API) getServerMirrors() ([]string, error) {
	res, err := a.doRequest(http.MethodGet, backupServer, nil)
	if err != nil {
		return nil, err
	}
	var srv []model.ServerMirror
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

func (a *API) doRequest(method string, url string, body []byte) ([]byte, error) {
	log := slog.With("method", "Api.doRequest")

	ctx, cancel := context.WithTimeout(context.Background(), config.APIReqTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
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
	defer func() { _ = res.Body.Close() }()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error("read browser response", slog.String("error", err.Error()))
		return nil, err
	}
	return b, nil
}
