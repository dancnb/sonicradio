package styles

type colorProfile struct {
	primaryColor           string
	secondaryColor         string
	invertedPrimaryColor   string
	invertedSecondaryColor string
}

type theme struct {
	dark  colorProfile
	light colorProfile
}

var themes = []theme{
	{
		dark:  colorProfile{primaryColor: "#ffb641", secondaryColor: "#bd862d", invertedPrimaryColor: "#12100d", invertedSecondaryColor: "#4a4133"},
		light: colorProfile{primaryColor: "#12100d", secondaryColor: "#4a4133", invertedPrimaryColor: "#ffb641", invertedSecondaryColor: "#bd862d"},
	},
	{
		dark:  colorProfile{primaryColor: "#98c379", secondaryColor: "#6b9e47", invertedPrimaryColor: "#243518", invertedSecondaryColor: "#3c5828"},
		light: colorProfile{primaryColor: "#243518", secondaryColor: "#3c5828", invertedPrimaryColor: "#98c379", invertedSecondaryColor: "#6b9e47"},
	},
	{
		dark:  colorProfile{primaryColor: "#abc8ed", secondaryColor: "#6d9edf", invertedPrimaryColor: "#1c467d", invertedSecondaryColor: "#2969bc"},
		light: colorProfile{primaryColor: "#1c467d", secondaryColor: "#1c467d", invertedPrimaryColor: "#abc8ed", invertedSecondaryColor: "#6d9edf"},
	},
}
