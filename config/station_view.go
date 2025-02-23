package config

type StationView uint8

func (s StationView) String() string {
	switch s {
	case DefaultView:
		return "DefaultView"
	case CompactView:
		return "CompactView"
	case MinimalView:
		return "MinimalView"
	}
	return "unknown StationView"
}

const (
	DefaultView StationView = iota
	CompactView
	MinimalView
)
