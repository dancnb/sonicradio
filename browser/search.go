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
	Name        string
	Country     string
	CountryCode string
	State       string
	Language    string
	TagList     []string
	Order       SearchOrder
	Reverse     bool
	Offset      int
	Limit       int
	HideBroken  bool
}

// `
// curl 'https://de1.api.radio-browser.info/json/stations/search'
// -X POST
// -H 'User-Agent: Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:123.0) Gecko/20100101 Firefox/123.0'
// -H 'Accept: */*'
// -H 'Accept-Language: en-US,en;q=0.5'
// -H 'Accept-Encoding: gzip, deflate, br'
// -H 'Referer: https://radiolise.gitlab.io/'
// -H 'content-type: application/x-www-form-urlencoded'
// -H 'Origin: https://radiolise.gitlab.io'
// -H 'Connection: keep-alive'
// -H 'Sec-Fetch-Dest: empty'
// -H 'Sec-Fetch-Mode: cors'
// -H 'Sec-Fetch-Site: cross-site'
// --data-raw 'name=zen+relax&tagList=&country=&state=&language=&tagExact=true&countryExact=false&stateExact=false&languageExact=false&offset=0&limit=20&order=clickcount&bitrateMin=0&bitrateMax=&reverse=true&hidebroken=true'
//
// `
func (p SearchParams) toFormData() string {
	var tags strings.Builder
	for _, v := range p.TagList {
		tags.WriteString(strings.TrimSpace(v))
	}
	fname := strings.Join(strings.Fields(p.Name), "+")
	return fmt.Sprintf("name=%s&tagList=%s&country=%s&state=%s&language=%s&tagExact=true&offset=%d&limit=%d&order=%s&bitrateMin=0&bitrateMax=&reverse=true&hidebroken=true",
		fname, tags.String(), p.Country, p.State, p.Language, p.Offset, p.Limit, p.Order)
}
