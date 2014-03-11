package geo

import (
	"testing"

	"github.com/nfleet/via/geotypes"
)

var locations = []geotypes.Location{
	{Address: geotypes.Address{City: "Helsinki", Street: "Esplanadi", Country: "finland"}},
	{Address: geotypes.Address{City: "Stuttgart", Street: "Calwer Straße", Country: "germany"}},
}

func BenchmarkAddressFinlandResolvation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := test_geo.ResolveLocation(locations[0])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAddressGermanyResolvation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := test_geo.ResolveLocation(locations[1])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func geocode(loc geotypes.Location, t *testing.T) {
	resolvedLoc, err := test_geo.ResolveLocation(loc)
	if err != nil {
		t.Fatal(err)
	}

	if IsMissingCoordinate(resolvedLoc) {
		t.Fatalf("Geocoding failed: %#v is missing coordinates", resolvedLoc)
	}

	t.Logf("Geocoded %v => %v", loc, resolvedLoc)
}

func TestResolvationForFinnishAddress(t *testing.T) {
	geocode(locations[0], t)
}

func TestResolvationForGermanAddress(t *testing.T) {
	geocode(locations[1], t)
}

func TestResolvationCoordinateFixing(t *testing.T) {
	location := geotypes.Location{
		Address:    geotypes.Address{Country: "finland"},
		Coordinate: geotypes.Coordinate{Latitude: 62.24, Longitude: 25.74},
	}

	loc, err := test_geo.ResolveLocation(location)
	if err != nil {
		t.Fatal(err)
	}

	if IsMissingCoordinate(loc) {
		t.Fatalf("Error: %#v is missing coordinates", loc)
	}

	t.Logf("Fixed coordinates %v => %v", location.Coordinate, loc.Coordinate)
}
