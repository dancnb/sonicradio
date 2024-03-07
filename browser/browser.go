package browser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net"
	"net/http"

	"github.com/dancnb/sonicradio/config"
)

const (
	HOST          = "all.api.radio-browser.info"
	backup_server = "https://de1.api.radio-browser.info/json/servers"
)

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

	api.servers = append(api.servers, res...)

	return &api
}

type Api struct {
	cfg     config.Value
	servers []string
}

func (a *Api) TopStations() []Station {
	s := SearchParams{
		Offset: 0,
		Limit:  100,
		Order:  Votes,
	}
	body := s.toFormData()
	res, err := a.doServerRequest(http.MethodPost, urlStations, []byte(body))
	if err != nil {
		return nil
	}
	var stations []Station
	err = json.Unmarshal(res, &stations)
	if err != nil {
		return nil
	}
	return stations
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

// ONLY USE THIS if your client is not able to do DNS look-ups
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
		return nil, err
	}
	ua := fmt.Sprintf("sonicradio/%s", a.cfg.Version)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", ua)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}
