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
	"time"

	"github.com/dancnb/sonicradio/config"
)

const (
	HOST              = "all.api.radio-browser.info"
	backup_server     = "https://de1.api.radio-browser.info/json/servers"
	serverMaxRetry    = 5
	serverRetryMillis = 200
)

var serverErrMsg = errors.New("Server response not available.")

func NewApi(cfg config.Value) *Api {
	api := Api{
		cfg: cfg,
	}
	res, err := api.getServersDNSLookup(HOST)
	if err != nil {
		res, err = api.getServerMirrors()
		if err != nil {
			msg := fmt.Errorf("could not retrieve %s servers: %w", HOST, err)
			slog.Info(msg.Error())
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
}

func (a *Api) GetCountries() ([]Country, error) {
	if len(a.countries) > 0 {
		return a.countries, nil
	}
	for i := 0; i < serverMaxRetry; i++ {
		res, err := a.doServerRequest(http.MethodGet, urlCountries, nil)
		if err != nil {
			slog.Error("get countries", "request error", err)
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		var countries []Country
		err = json.Unmarshal(res, &countries)
		if err != nil {
			slog.Error("get countries", "unmarshal error", err)
			slog.Error("get countries", "response", string(res))
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		slog.Info("get countries", "length", len(countries))
		a.countries = countries
		return countries, nil
	}
	slog.Warn("get countries", "", "exceeded max retries")
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
	slog.Debug("stationSearch", "request", body)
	var err error
	for i := 0; i < serverMaxRetry; i++ {
		var res []byte
		res, err = a.doServerRequest(http.MethodPost, urlStations, []byte(body))
		if err != nil {
			slog.Error("stationSearch", "request error", err)
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		var stations []Station
		err = json.Unmarshal(res, &stations)
		if err != nil {
			slog.Error("stationSearch", "unmarshal error", err)
			slog.Error("stationSearch", "response", string(res))
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		slog.Info("stationSearch", "length", len(stations))
		return stations, nil
	}
	slog.Warn("stationSearch", "", "exceeded max retries")
	return nil, serverErrMsg
}

func (a *Api) GetStations(uuids []string) ([]Station, error) {
	if len(uuids) == 0 {
		return nil, nil
	}
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
			slog.Error("get stations", "request error", err)
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		var stations []Station
		err = json.Unmarshal(res, &stations)
		if err != nil {
			slog.Error("get stations", "unmarshal error", err)
			slog.Error("get stations", "response", string(res))
			time.Sleep(serverRetryMillis * time.Millisecond)
			continue
		}
		slog.Info("get stations", "length", len(stations))
		return stations, nil
	}

	slog.Warn("get station exceeded max retries")
	return nil, serverErrMsg
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
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		slog.Error("create browser request", slog.String("error", err.Error()))
		return nil, err
	}
	ua := fmt.Sprintf("sonicradio/%s", a.cfg.Version)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", ua)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("do browser request", slog.String("error", err.Error()))
		return nil, err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("read browser response", slog.String("error", err.Error()))
		return nil, err
	}
	return b, nil
}
