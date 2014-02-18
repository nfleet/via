package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"testing"
)

func TestResolvationForAddress(t *testing.T) {
	location := Location{
		Address: Address{City: "Helsinki", Street: "Esplanadi", Country: "finland"},
	}

	loc, err := ResolveLocation(server.Config, location)
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

	loc, err := ResolveLocation(server.Config, location)
	if err != nil {
		t.Fatal(err)
	}

	if IsMissingCoordinate(loc) {
		t.Fatalf("Error: %#v is missing coordinates", loc)
	}

	t.Logf("Fixed coordinates %v => %v", location.Coordinate, loc.Coordinate)
}

func api_test_resolve(t *testing.T, locations []Location) {
	request := fmt.Sprintf("http://localhost:%d/resolve", server.Config.Port)

	jsonLoc, _ := json.Marshal(locations)
	b := strings.NewReader(string(jsonLoc))

	response, err := http.Post(request, "application/json", b)
	if err != nil {
		t.Fatal(response)
	} else {
		defer response.Body.Close()
		cont, err := ioutil.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}

		var resolved []Location
		if err := json.Unmarshal(cont, &resolved); err != nil {
			t.Fatal(string(cont), err)
		}
		for i, loc := range resolved {
			t.Logf("Resolved %v to have coords %v", locations[i], loc)
		}
	}
}

func TestAPIResolvationCoordinate(t *testing.T) {
	locations := []Location{
		{Address: Address{Country: "finland"}, Coordinate: Coordinate{Latitude: 62.24, Longitude: 25.74}},
	}
	api_test_resolve(t, locations)
}

func TestAPIResolvationWithoutCoordinate(t *testing.T) {
	locations := []Location{
		{Address: Address{Street: "Esplanadi", Country: "finland"}},
	}
	api_test_resolve(t, locations)
}
