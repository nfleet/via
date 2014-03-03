package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/nfleet/via/geo"
)

type P struct {
	matrix  []geo.Coord
	country string
}

func TestBoundingBoxes(t *testing.T) {
	var tests = []struct {
		in  P
		out bool
	}{
		{P{[]Coord{{60.0, 25.0}}, "finland"}, true},
		{P{[]Coord{{50.0, 25.0}}, "finland"}, false},
		{P{[]Coord{{61.0, 19.0}}, "finland"}, false},
		{P{[]Coord{{50.0, 10.0}}, "germany"}, true},
		{P{[]Coord{{45.0, 10.0}}, "germany"}, false},
		{P{[]Coord{{50.0, 4.05}}, "germany"}, false},
	}

	for i, test := range tests {
		res, _ := check_coordinate_sanity(test.in.matrix, test.in.country)
		if res != test.out {
			t.Errorf("%d. check_coordinate_sanity(%q, %q) => %q, want %q", i, test.in.matrix, test.in.country, res, test.out)
		}
	}
}

func api_test_resolve(t *testing.T, locations []Location) {
	request := fmt.Sprintf("http://localhost:%d/resolve", test_via.Config.Port)

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

		var resolved []geo.Location
		if err := json.Unmarshal(cont, &resolved); err != nil {
			t.Fatal(string(cont), err)
		}
		for i, loc := range resolved {
			t.Logf("Resolved %v to have coords %v", locations[i], loc)
		}
	}
}

func TestAPIResolvationCoordinate(t *testing.T) {
	locations := []geo.Location{
		{Address: geo.Address{Country: "finland"}, Coordinate: geo.Coordinate{Latitude: 62.24, Longitude: 25.74}},
	}
	api_test_resolve(t, locations)
}

func TestAPIResolvationWithoutCoordinate(t *testing.T) {
	locations := []geo.Location{
		{Address: geo.Address{Street: "Esplanadi", City: "Helsinki", Country: "finland"}},
	}
	api_test_resolve(t, locations)
}
