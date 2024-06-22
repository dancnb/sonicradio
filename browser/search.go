package browser

import (
	"fmt"
	"strings"
)

type OrderBy string

const (
	Name            OrderBy = "name"
	Url             OrderBy = "url"
	Homepage        OrderBy = "homepage"
	Favicon         OrderBy = "favicon"
	Tags            OrderBy = "tags"
	CountryOrder    OrderBy = "country"
	State           OrderBy = "state"
	LanguageOrder   OrderBy = "language"
	Votes           OrderBy = "votes"
	Codec           OrderBy = "codec"
	Bitrate         OrderBy = "bitrate"
	Lastcheckok     OrderBy = "lastcheckok"
	Lastchecktime   OrderBy = "lastchecktime"
	Clicktimestamp  OrderBy = "clicktimestamp"
	Clickcount      OrderBy = "clickcount"
	Clicktrend      OrderBy = "clicktrend"
	Changetimestamp OrderBy = "changetimestamp"
	Random          OrderBy = "random"
)

const DefLimit = 30

type SearchParams struct {
	Name     string
	TagList  string
	Country  string
	State    string
	Language string
	Limit    int
	Order    OrderBy
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
		Limit:   DefLimit,
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
