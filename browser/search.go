package browser

import (
	"fmt"
	"strings"
)

type OrderBy string

const (
	Votes         OrderBy = "votes"      // Number of votes for this station. This number is by server and only ever increases. It will never be reset to 0.
	Clickcount    OrderBy = "clickcount" // Clicks within the last 24 hours
	Clicktrend    OrderBy = "clicktrend" // The difference of the clickcounts within the last 2 days. Posivite values mean an increase, negative a decrease of clicks.
	Bitrate       OrderBy = "bitrate"
	Name          OrderBy = "name"
	Tags          OrderBy = "tags"
	CountryOrder  OrderBy = "country"
	LanguageOrder OrderBy = "language"
	Codec         OrderBy = "codec"
	Random        OrderBy = "random"

	// Url             OrderBy = "url"
	// Homepage        OrderBy = "homepage"
	// State           OrderBy = "state"
	// Favicon         OrderBy = "favicon"
	// Lastcheckok     OrderBy = "lastcheckok"
	// Lastchecktime   OrderBy = "lastchecktime"
	// Clicktimestamp  OrderBy = "clicktimestamp"
	// Changetimestamp OrderBy = "changetimestamp"
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
