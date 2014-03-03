package geo

import "testing"

func TestResolvationForAddress(t *testing.T) {
	location := Location{
		Address: Address{City: "Helsinki", Street: "Esplanadi", Country: "finland"},
	}

	loc, err := test_geo.ResolveLocation(location)
	if err != nil {
		t.Fatal(err)
	}

	if IsMissingCoordinate(loc) {
		t.Fatalf("Geocoding failed: %#v is missing coordinates", loc)
	}

	t.Logf("Geocoded %v => %v", location, loc)
}

func TestResolvationCoordinateFixing(t *testing.T) {
	location := Location{
		Address:    Address{Country: "finland"},
		Coordinate: Coordinate{Latitude: 62.24, Longitude: 25.74},
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
