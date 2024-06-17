package browser

import (
	"fmt"
	"strings"
)

type SearchOrder string

const (
	Name            SearchOrder = "name"
	Url             SearchOrder = "url"
	Homepage        SearchOrder = "homepage"
	Favicon         SearchOrder = "favicon"
	Tags            SearchOrder = "tags"
	CountryOrder    SearchOrder = "country"
	State           SearchOrder = "state"
	LanguageOrder   SearchOrder = "language"
	Votes           SearchOrder = "votes"
	Codec           SearchOrder = "codec"
	Bitrate         SearchOrder = "bitrate"
	Lastcheckok     SearchOrder = "lastcheckok"
	Lastchecktime   SearchOrder = "lastchecktime"
	Clicktimestamp  SearchOrder = "clicktimestamp"
	Clickcount      SearchOrder = "clickcount"
	Clicktrend      SearchOrder = "clicktrend"
	Changetimestamp SearchOrder = "changetimestamp"
	Random          SearchOrder = "random"
)

type SearchParams struct {
	Name     string
	TagList  string
	Country  string
	State    string
	Language string
	Limit    int
	Order    SearchOrder
	Reverse  bool

	Offset int
	// CountryCode string
	// TagExact    string //always "true"
	// HideBroken  string //always "true"
}

func DefaultSearchParams() SearchParams {
	return SearchParams{
		Order:   Votes,
		Reverse: true,
		Offset:  0,
		Limit:   30,
	}
}

func (p SearchParams) toFormData() string {
	fname := strings.Join(strings.Fields(p.Name), "+")
	fTags := strings.Join(strings.Fields(p.TagList), "+")

	return fmt.Sprintf("name=%s&tagList=%s&country=%s&countryExact=false&state=%s&language=%s&tagExact=true&offset=%d&limit=%d&order=%s&bitrateMin=0&bitrateMax=&reverse=%s&hidebroken=true",
		fname, fTags, p.Country, p.State, p.Language, p.Offset, p.Limit, p.Order, boolString(p.Reverse))
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
