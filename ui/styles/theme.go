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
	Light ColorProfile
}

var Themes = []Theme{
	{
		Name:  "Duo Yellow",
		Dark:  ColorProfile{primaryColor: "#D4DAF7", secondaryColor: "#D58610", invertedPrimaryColor: "#2D2D0B", invertedSecondaryColor: "#827545"},
		Light: ColorProfile{primaryColor: "#2D2D0B", secondaryColor: "#827545", invertedPrimaryColor: "#D4DAF7", invertedSecondaryColor: "#D58610"},
	},
	{
		Name:  "Duo Green",
		Dark:  ColorProfile{primaryColor: "#F7D4D6", secondaryColor: "#6b9e47", invertedPrimaryColor: "#243518", invertedSecondaryColor: "#3c5828"},
		Light: ColorProfile{primaryColor: "#243518", secondaryColor: "#3c5828", invertedPrimaryColor: "#F7D4D6", invertedSecondaryColor: "#6b9e47"},
	},
	{
		Name:  "Duo Blue",
		Dark:  ColorProfile{primaryColor: "#F7EDD4", secondaryColor: "#6d9edf", invertedPrimaryColor: "#1c467d", invertedSecondaryColor: "#2969bc"},
		Light: ColorProfile{primaryColor: "#1c467d", secondaryColor: "#2969bc", invertedPrimaryColor: "#F7EDD4", invertedSecondaryColor: "#6d9edf"},
	},
	{
		Name:  "Duo Red",
		Dark:  ColorProfile{primaryColor: "#E3F7D4", secondaryColor: "#DE5145", invertedPrimaryColor: "#351D10", invertedSecondaryColor: "#8C4D2B"},
		Light: ColorProfile{primaryColor: "#351D10", secondaryColor: "#8C4D2B", invertedPrimaryColor: "#E3F7D4", invertedSecondaryColor: "#DE5145"},
	},
	{
		Name:  "Mono Yellow",
		Dark:  ColorProfile{primaryColor: "#ffb641", secondaryColor: "#bd862d", invertedPrimaryColor: "#12100d", invertedSecondaryColor: "#4a4133"},
		Light: ColorProfile{primaryColor: "#12100d", secondaryColor: "#4a4133", invertedPrimaryColor: "#ffb641", invertedSecondaryColor: "#bd862d"},
	},
	{
		Name:  "Mono Green",
		Dark:  ColorProfile{primaryColor: "#98c379", secondaryColor: "#6b9e47", invertedPrimaryColor: "#243518", invertedSecondaryColor: "#3c5828"},
		Light: ColorProfile{primaryColor: "#243518", secondaryColor: "#3c5828", invertedPrimaryColor: "#98c379", invertedSecondaryColor: "#6b9e47"},
	},
	{
		Name:  "Mono Blue",
		Dark:  ColorProfile{primaryColor: "#abc8ed", secondaryColor: "#6d9edf", invertedPrimaryColor: "#1c467d", invertedSecondaryColor: "#2969bc"},
		Light: ColorProfile{primaryColor: "#1c467d", secondaryColor: "#2969bc", invertedPrimaryColor: "#abc8ed", invertedSecondaryColor: "#6d9edf"},
	},
	{
		Name:  "Mono Red",
		Dark:  ColorProfile{primaryColor: "#e48189", secondaryColor: "#d7424e", invertedPrimaryColor: "#69161d", invertedSecondaryColor: "#931f29"},
		Light: ColorProfile{primaryColor: "#69161d", secondaryColor: "#931f29", invertedPrimaryColor: "#e48189", invertedSecondaryColor: "#d7424e"},
	},
	{
		Name:  "Grayscale",
		Dark:  ColorProfile{primaryColor: "#e5e5e5ff", secondaryColor: "#bdbdbdff", invertedPrimaryColor: "#2e2e2eff", invertedSecondaryColor: "#818181ff"},
		Light: ColorProfile{primaryColor: "#2e2e2eff", secondaryColor: "#818181ff", invertedPrimaryColor: "#e5e5e5ff", invertedSecondaryColor: "#bdbdbdff"},
	},
}
