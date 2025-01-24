package styles

type ColorProfile struct {
	primaryColor           string
	secondaryColor         string
	invertedPrimaryColor   string
	invertedSecondaryColor string
}

type Theme struct {
	Name  string
	Dark  ColorProfile
	Light ColorProfile // 5 4 2 1
}

var Themes = []Theme{
	{
		Name:  "MonoYellow",
		Dark:  ColorProfile{primaryColor: "#ffb641", secondaryColor: "#bd862d", invertedPrimaryColor: "#12100d", invertedSecondaryColor: "#4a4133"},
		Light: ColorProfile{primaryColor: "#342609", secondaryColor: "#9C6902", invertedPrimaryColor: "#FDCD6D", invertedSecondaryColor: "#FEF3DC"},
	},
	{
		Name:  "MonoGreen",
		Dark:  ColorProfile{primaryColor: "#98c379", secondaryColor: "#6b9e47", invertedPrimaryColor: "#243518", invertedSecondaryColor: "#3c5828"},
		Light: ColorProfile{primaryColor: "#1B3409", secondaryColor: "#375F1B", invertedPrimaryColor: "#9BD770", invertedSecondaryColor: "#EBF7E3"},
	},
	{
		Name:  "MonoBlue",
		Dark:  ColorProfile{primaryColor: "#abc8ed", secondaryColor: "#6d9edf", invertedPrimaryColor: "#1c467d", invertedSecondaryColor: "#2969bc"},
		Light: ColorProfile{primaryColor: "#091D34", secondaryColor: "#133863", invertedPrimaryColor: "#abc8ed", invertedSecondaryColor: "#E1ECF9"},
	},
	{
		Name:  "Analog 1",
		Dark:  ColorProfile{primaryColor: "#E1E15B", secondaryColor: "#D58610", invertedPrimaryColor: "#2D2D0B", invertedSecondaryColor: "#827545"},
		Light: ColorProfile{primaryColor: "#2D2D0B", secondaryColor: "#091D34", invertedPrimaryColor: "#D58610", invertedSecondaryColor: "#EDD8B5"},
	},
}
