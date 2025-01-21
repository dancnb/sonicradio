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
		Light: ColorProfile{primaryColor: "#1c467d", secondaryColor: "#1c467d", invertedPrimaryColor: "#abc8ed", invertedSecondaryColor: "#6d9edf"},
	},
}
