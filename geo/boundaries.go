package geo

var BoundingBoxes = map[string]map[string]float64{
	"finland": map[string]float64{
		"long_min": 20.54,     // Border between Swe-Nor-Fin
		"long_max": 31.5867,   // Somewhere in Ilomantsi
		"lat_min":  59.807983, // Hanko
		"lat_max":  70.092283, // Nuorgam
	},
	"germany": map[string]float64{
		"long_min": 5.8666667, // Isenbruch, Nordrhein-Westfalen
		"long_max": 15.033333, // Deschake, Neißeaue, Saxony
		"lat_min":  47.270108, // Haldenwanger Eck, Oberstdorf, Bavaria
		"lat_max":  54.9,      // Aventoft, Schleswig-Holstein
	},
}
