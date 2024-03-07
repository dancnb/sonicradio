package browser

import (
	"net/http"
	"testing"

	"github.com/dancnb/sonicradio/config"
)

func Test_getServers(t *testing.T) {
	a := NewApi(config.Value{Version: "", Debug: true})
	res, err := a.getServersDNSLookup(HOST)
	if err != nil {
		t.Error(err)
	}
	t.Log(res)

	// ONLY USE THIS if your client is not able to do DNS look-ups
	// res, err = getServerMirrors()
	// if err != nil {
	// 	t.Error(err)
	// }
	t.Log(res)
}

func Test_doRrequest(t *testing.T) {
	a := NewApi(config.Value{Version: "", Debug: true})
	res, err := a.doServerRequest(http.MethodGet, "/json/servers", nil)
	if err != nil {
		t.Error(err)
	}
	t.Log(res)
}

func Test_topStations(t *testing.T) {
	a := NewApi(config.Value{Version: "", Debug: true})
	res := a.TopStations()
	if len(res) == 0 {
		t.Error("missing stations response")
	}
	t.Log(res)
}
