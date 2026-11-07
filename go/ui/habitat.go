//go:build ui || windows || (darwin && cgo)

package ui

import (
	"image/color"
	"math"
	"strings"
)

// netlogoBase are the NetLogo base colours 5, 15, ..., 135 (gray, red, orange, brown, yellow, green, lime,
// turquoise, cyan, sky, blue, violet, magenta, pink).
var netlogoBase = []color.RGBA{
	{141, 141, 141, 255}, {215, 50, 41, 255}, {241, 106, 21, 255}, {157, 110, 72, 255},
	{237, 237, 49, 255}, {89, 176, 60, 255}, {44, 209, 59, 255}, {29, 159, 120, 255},
	{84, 196, 196, 255}, {45, 141, 190, 255}, {52, 93, 169, 255}, {124, 80, 164, 255},
	{167, 27, 106, 255}, {224, 127, 150, 255},
}

// netlogoColor converts a NetLogo colour number to RGB: shade 0 is black, 5 the base colour, 10 white.
func netlogoColor(c float64) color.RGBA {
	c = math.Mod(c, 140)
	if c < 0 {
		c += 140
	}
	base := netlogoBase[int(c/10)]
	shade := c - 10*math.Floor(c/10)
	mix := func(v uint8) uint8 {
		if shade < 5 {
			return uint8(float64(v) * shade / 5)
		}
		return uint8(float64(v) + (255-float64(v))*(shade-5)/5)
	}
	return color.RGBA{mix(base.R), mix(base.G), mix(base.B), 255}
}

// habitatColours are the NetLogo colour numbers of the habitat types. Landscape files hold no colours, so the UI
// keeps the column patchColour of the NetLogo foodsources files of the example farm and Sussex
// (_SYSTEM_Example_Farm_Foodsources.txt, _SYSTEM_Sussex1_Foodsources.txt), and for the other habitat types the
// column habitatColourID of _SYSTEM_Habitats.csv.
var habitatColours = map[string]float64{
	// foodsources files
	"Crop_Cereals": 75, "Crop_Maize": 55, "Crop_Oilseed_rape": 45, "Grass_improved": 115, "Grassland": 135,
	"Hedgerow": 105, "Scrub": 125, "Woodland": 35,
	// _SYSTEM_Habitats.csv
	"Grass_margin": 12, "Crop_Field_beans": 15, "Grass_multispecies_ley": 18, "Woodland_mixed": 32,
	"GameCover_Plot_57": 42, "Wildflower": 48, "Rush_pasture": 51, "Permanent_pasture_GS2": 62,
	"Permanent_pasture": 65, "Heath": 85, "Legume_3": 92, "Traditional_orchard": 95, "Nectar_plot": 98,
	"Wetland_mosaic": 102, "Improved": 115, "Flower_margin": 122, "BirdBumble_mix": 128,
}

// colUnknownHabitat is the colour of a habitat type that is not in habitatColours.
var colUnknownHabitat = color.RGBA{0xc8, 0xc8, 0xc8, 0xff}

// habitatColour returns the colour of a habitat type.
func habitatColour(patchType string) color.RGBA {
	if c, ok := habitatColours[patchType]; ok {
		return netlogoColor(c)
	}
	return colUnknownHabitat
}

// prettyName turns an input file name like "Crop_Cereals" into "Crop Cereals".
func prettyName(s string) string {
	return strings.ReplaceAll(s, "_", " ")
}
