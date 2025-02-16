package browser

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dancnb/sonicradio/config"
)

func Test_NewApi(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := NewApi(ctx, &config.Value{Version: ""})
	if err != nil {
		t.Error(err)
	}
}

func Test_topStations(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a, err := NewApi(ctx, &config.Value{Version: ""})
	if err != nil {
		t.Fatal(err)
	}
	res, err := a.TopStations()
	if err != nil {
		t.Error(err)
	}
	if len(res) == 0 {
		t.Error("missing stations response")
	}
	t.Log(res)
}

func Test_getStation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a, err := NewApi(ctx, &config.Value{Version: ""})
	if err != nil {
		t.Fatal(err)
	}
	uuid := []string{
		"748d830c-d934-41e8-bd14-870add931e1d",
		"a06ed3d2-ba59-4969-825d-4e9b3f336b93",
		"96133c49-0601-11e8-ae97-52543be04c81",
	}
	res, err := a.GetStations(uuid)
	if err != nil {
		t.Error(err)
	}
	if res == nil {
		t.Error("missing station response")
	}
	t.Log(res)
	fmt.Printf("%#v", res)
}

func Test_searchStations(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a, err := NewApi(ctx, &config.Value{Version: ""})
	if err != nil {
		t.Fatal(err)
	}

	params := DefaultSearchParams()
	params.Name = strings.TrimSpace("deea")
	params.TagList = strings.TrimSpace("")
	params.Country = strings.TrimSpace("")
	params.State = strings.TrimSpace("")
	params.Language = strings.TrimSpace("")
	res, err := a.Search(params)
	if err != nil {
		t.Error(err)
	}
	if res == nil {
		t.Error("missing station response")
	}
	t.Log(res)
	fmt.Printf("%#v", res)
}

func Test_getCountries(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a, err := NewApi(ctx, &config.Value{Version: ""})
	if err != nil {
		t.Fatal(err)
	}
	res, err := a.GetCountries()
	if err != nil {
		t.Error(err)
	}
	t.Log(res)
}

func TestApi_StationCounter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a, err := NewApi(ctx, &config.Value{Version: ""})
	if err != nil {
		t.Fatal(err)
	}
	err = a.StationCounter("748d830c-d934-41e8-bd14-870add931e1d")
	if err != nil {
		t.Error(err)
	}
}

func TestApi_StationVote(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a, err := NewApi(ctx, &config.Value{Version: ""})
	if err != nil {
		t.Fatal(err)
	}
	err = a.StationVote("748d830c-d934-41e8-bd14-870add931e1d")
	if err != nil {
		t.Error(err)
	}
	time.Sleep(300 * time.Millisecond)
	err = a.StationVote("748d830c-d934-41e8-bd14-870add931e1d")
	if err != errVoteTimeout {
		t.Error(err)
	}
}
