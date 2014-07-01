package main

import (
	"testing"

	"github.com/nfleet/via/geotypes"
)

var (
	srcCoord = geotypes.Location{Coordinate: geotypes.Coordinate{Latitude: 62.24027, Longitude: 25.74444}, Address: geotypes.Address{Country: "Finland"}}
	trgCoord = geotypes.Location{Coordinate: geotypes.Coordinate{Latitude: 60.45138, Longitude: 22.26666}, Address: geotypes.Address{Country: "Finland"}}
	noCoords = []geotypes.Edge{
		{Source: geotypes.Location{Address: geotypes.Address{Street: "Erottaja", City: "Helsinki", Country: "Finland"}}, Target: geotypes.Location{Address: geotypes.Address{Street: "Esplanadi", City: "Helsinki", Country: "Finland"}}},
	}

	noCoordsFinland = []geotypes.Edge{
		{Source: geotypes.Location{Address: geotypes.Address{Street: "Erottaja", City: "Helsinki", Country: "Finland"}}, Target: geotypes.Location{Address: geotypes.Address{Street: "Esplanadi", City: "Helsinki", Country: "Finland"}}},
		{Source: geotypes.Location{Address: geotypes.Address{Street: "Taitoniekantie", City: "Jyväskylä", Country: "Finland"}}, Target: geotypes.Location{Address: geotypes.Address{Street: "Vuolteenkatu", City: "Tampere", Country: "Finland"}}},
		{Source: geotypes.Location{Address: geotypes.Address{Street: "Vuolteenkatu", City: "Tampere", Country: "Finland"}}, Target: geotypes.Location{Address: geotypes.Address{Street: "Hikiäntie", City: "Riihimäki", Country: "Finland"}}},
	}

	coordsGermany = []geotypes.Edge{
		{Source: geotypes.Location{Coordinate: geotypes.Coordinate{Latitude: 49.0811, Longitude: 9.8795}, Address: geotypes.Address{Country: "Germany"}}, Target: geotypes.Location{Coordinate: geotypes.Coordinate{Latitude: 49.3636, Longitude: 10.1472}, Address: geotypes.Address{Country: "Germany"}}},
		{Source: geotypes.Location{Coordinate: geotypes.Coordinate{Latitude: 49.4413, Longitude: 9.7394}, Address: geotypes.Address{Country: "Germany"}}, Target: geotypes.Location{Coordinate: geotypes.Coordinate{Latitude: 48.4356, Longitude: 12.4750}, Address: geotypes.Address{Country: "Germany"}}},
		{Source: geotypes.Location{Coordinate: geotypes.Coordinate{Latitude: 47.9074, Longitude: 10.4459}, Address: geotypes.Address{Country: "Germany"}}, Target: geotypes.Location{Coordinate: geotypes.Coordinate{Latitude: 48.4133, Longitude: 9.2127}, Address: geotypes.Address{Country: "Germany"}}},
		{Source: geotypes.Location{Coordinate: geotypes.Coordinate{Latitude: 50.1690, Longitude: 11.2960}, Address: geotypes.Address{Country: "Germany"}}, Target: geotypes.Location{Coordinate: geotypes.Coordinate{Latitude: 52.4292, Longitude: 13.2286}, Address: geotypes.Address{Country: "Germany"}}},
		{Source: geotypes.Location{Coordinate: geotypes.Coordinate{Latitude: 52.5876, Longitude: 13.2461}, Address: geotypes.Address{Country: "Germany"}}, Target: geotypes.Location{Coordinate: geotypes.Coordinate{Latitude: 53.4989, Longitude: 13.9805}, Address: geotypes.Address{Country: "Germany"}}},
	}
)

func BenchmarkFinlandPathsExtraction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := test_geo.CalculateCoordinatePaths(geotypes.PathsInput{100, noCoordsFinland})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGermanyPathsExtraction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := test_geo.CalculateCoordinatePaths(geotypes.PathsInput{100, coordsGermany})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestCalculatePathsWithoutCoordinates(t *testing.T) {
	_, err := test_geo.CalculateCoordinatePaths(geotypes.PathsInput{100, noCoordsFinland})

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Finland (%d edges) ok.", len(noCoordsFinland))
}

func TestCalculatePathsWithCoordinates(t *testing.T) {
	_, err := test_geo.CalculateCoordinatePaths(geotypes.PathsInput{100, coordsGermany})

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Germany (%d edges) ok.", len(coordsGermany))
}

func BenchmarkFinlandCoordinatelessPathExtraction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := test_geo.CalculateCoordinatePaths(geotypes.PathsInput{100, noCoordsFinland})
		if err != nil {
			b.Fatal(err)
		}
	}
}
