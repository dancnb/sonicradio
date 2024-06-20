package browser

type Country struct {
	Name         string `json:"name"`
	ISO3166_1    string `json:"iso_3166_1"`
	Stationcount int    `json:"stationcount"`
}

type Language struct {
	Name         string `json:"name"`
	ISO639       string `json:"iso_639"`
	Stationcount string `json:"stationcount"`
}

type StationTag struct {
	Name         string `json:"name"`
	Stationcount string `json:"stationcount"`
}
type ClickCounterResponse struct {
	Ok          string `json:"ok"`
	Message     string `json:"message"`
	Stationuuid string `json:"stationuuid"`
	Name        string `json:"name"`
	URL         string `json:"url"`
}

type ServerMirror struct {
	IP   string `json:"ip"`
	Name string `json:"name"`
}
