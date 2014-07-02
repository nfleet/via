package geo

import (
	"testing"

	"github.com/nfleet/via/geotypes"
)

var locations = []geotypes.Location{
	{Address: geotypes.Address{City: "Helsinki", HouseNumber: 28, Street: "Mechelininkatu", Country: "finland"}},
	{Address: geotypes.Address{City: "Jyväskylä", HouseNumber: 9, Street: "Taitoniekantie", Country: "finland"}},
	{Address: geotypes.Address{City: "Jyväskylä", HouseNumber: 0, Street: "Taitoniekantie", Country: "finland"}},
}

var germany = []geotypes.Location{
	{Address: geotypes.Address{City: "Stuttgart", Street: "Calwer Straße", Country: "germany"}},
}

func BenchmarkAddressFinlandResolvation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := test_geo.ResolveLocation(locations[0], 1)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAddressGermanyResolvation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := test_geo.ResolveLocation(locations[1], 1)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func geocode(loc geotypes.Location, t *testing.T) {
	resolvedLocs, err := test_geo.ResolveLocation(loc, 1)
	if err != nil {
		t.Fatal(err)
	}

	resolvedLoc := resolvedLocs[0]

	if IsMissingCoordinate(resolvedLoc) {
		t.Fatalf("Geocoding failed: %#v is missing coordinates", resolvedLoc)
	}

	t.Logf("Geocoded %v => %v", loc, resolvedLoc)
}

func TestResolvationForFinnishAddress(t *testing.T) {
	for _, loc := range locations {
		geocode(loc, t)
	}
}

func TestResolvationForGermanAddress(t *testing.T) {
	for _, loc := range germany {
		geocode(loc, t)
	}
}

func TestResolvationCoordinateFixing(t *testing.T) {
	location := geotypes.Location{
		Address:    geotypes.Address{Country: "finland"},
		Coordinate: geotypes.Coordinate{Latitude: 62.24, Longitude: 25.74},
	}

	locs, err := test_geo.ResolveLocation(location, 1)
	if err != nil {
		t.Fatal(err)
	}

	if IsMissingCoordinate(locs[0]) {
		t.Fatalf("Error: %#v is missing coordinates", locs)
	}

	t.Logf("Fixed coordinates %v => %v", location.Coordinate, locs[0].Coordinate)
}
