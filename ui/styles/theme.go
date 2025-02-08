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
		Name:  "Duo 1",
		Dark:  ColorProfile{primaryColor: "#E1E15B", secondaryColor: "#D58610", invertedPrimaryColor: "#2D2D0B", invertedSecondaryColor: "#827545"},
		Light: ColorProfile{primaryColor: "#2D2D0B", secondaryColor: "#827545", invertedPrimaryColor: "#E1E15B", invertedSecondaryColor: "#D58610"},
	},
	{
		Name:  "Duo 2",
		Dark:  ColorProfile{primaryColor: "#E6E6E6", secondaryColor: "#DE5145", invertedPrimaryColor: "#351D10", invertedSecondaryColor: "#8C4D2B"},
		Light: ColorProfile{primaryColor: "#351D10", secondaryColor: "#8C4D2B", invertedPrimaryColor: "#E6E6E6", invertedSecondaryColor: "#DE5145"},
	},
	{
		Name:  "MonoYellow",
		Dark:  ColorProfile{primaryColor: "#ffb641", secondaryColor: "#bd862d", invertedPrimaryColor: "#12100d", invertedSecondaryColor: "#4a4133"},
		Light: ColorProfile{primaryColor: "#12100d", secondaryColor: "#4a4133", invertedPrimaryColor: "#ffb641", invertedSecondaryColor: "#bd862d"},
	},
	{
		Name:  "MonoGreen",
		Dark:  ColorProfile{primaryColor: "#98c379", secondaryColor: "#6b9e47", invertedPrimaryColor: "#243518", invertedSecondaryColor: "#3c5828"},
		Light: ColorProfile{primaryColor: "#243518", secondaryColor: "#3c5828", invertedPrimaryColor: "#98c379", invertedSecondaryColor: "#6b9e47"},
	},
	{
		Name:  "MonoBlue",
		Dark:  ColorProfile{primaryColor: "#abc8ed", secondaryColor: "#6d9edf", invertedPrimaryColor: "#1c467d", invertedSecondaryColor: "#2969bc"},
		Light: ColorProfile{primaryColor: "#1c467d", secondaryColor: "#2969bc", invertedPrimaryColor: "#abc8ed", invertedSecondaryColor: "#6d9edf"},
	},
}
