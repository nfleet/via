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

var (
	testConfig, _   = geo.LoadConfig("development.json")
	noCoordsFinland = []geo.Edge{
		{Source: geo.Location{Address: geo.Address{Street: "Erottaja", City: "Helsinki", Country: "Finland"}}, Target: geo.Location{Address: geo.Address{Street: "Esplanadi", City: "Helsinki", Country: "Finland"}}},
		{Source: geo.Location{Address: geo.Address{Street: "Taitoniekantie", City: "Jyväskylä", Country: "Finland"}}, Target: geo.Location{Address: geo.Address{Street: "Vuolteenkatu", City: "Tampere", Country: "Finland"}}},
		{Source: geo.Location{Address: geo.Address{Street: "Vuolteenkatu", City: "Tampere", Country: "Finland"}}, Target: geo.Location{Address: geo.Address{Street: "Hikiäntie", City: "Riihimäki", Country: "Finland"}}},
	}
)

func TestBoundingBoxes(t *testing.T) {
	var tests = []struct {
		in  P
		out bool
	}{
		{P{[]geo.Coord{{60.0, 25.0}}, "finland"}, true},
		{P{[]geo.Coord{{50.0, 25.0}}, "finland"}, false},
		{P{[]geo.Coord{{61.0, 19.0}}, "finland"}, false},
		{P{[]geo.Coord{{50.0, 10.0}}, "germany"}, true},
		{P{[]geo.Coord{{45.0, 10.0}}, "germany"}, false},
		{P{[]geo.Coord{{50.0, 4.05}}, "germany"}, false},
	}

	for i, test := range tests {
		res, _ := check_coordinate_sanity(test.in.matrix, test.in.country)
		if res != test.out {
			t.Errorf("%d. check_coordinate_sanity(%q, %q) => %q, want %q", i, test.in.matrix, test.in.country, res, test.out)
		}
	}
}

func api_test_resolve(t *testing.T, locations []geo.Location) {
	request := fmt.Sprintf("http://localhost:%d/resolve", testConfig.Port)

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

func api_coordinate_query(t *testing.T, edges []geo.Edge, speed_profile int) []geo.CoordinatePath {
	payload := struct {
		SpeedProfile int
		Edges        []geo.Edge
	}{
		speed_profile,
		edges,
	}

	jsonPayload, err := json.Marshal(payload)
	b := strings.NewReader(string(jsonPayload))

	var paths []geo.CoordinatePath

	request := fmt.Sprintf("http://localhost:%d/paths", testConfig.Port)
	response, err := http.Post(request, "application/json", b)
	if err != nil {
		t.Fatal(response)
	} else if response.StatusCode != 200 {
		cont, _ := ioutil.ReadAll(response.Body)
		t.Fatal(string(cont))
	} else {
		defer response.Body.Close()
		cont, err := ioutil.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}

		if err := json.Unmarshal(cont, &paths); err != nil {
			t.Logf(err.Error())
		}

		t.Logf("sent %d paths <=> got %d back", len(edges), len(paths))

		for i, p := range paths {
			t.Logf("Calculated path of %d m / %d secs from %s, %s, %s to %s, %s, %s", p.Distance, p.Time,
				edges[i].Source.Address.Street, edges[i].Source.Address.City, edges[i].Source.Address.Country,
				edges[i].Target.Address.Street, edges[i].Target.Address.City, edges[i].Target.Address.Country)
		}

		t.Logf("Calculated %d paths", len(paths))
	}

	return paths
}

func BenchmarkAPIPointCorrection(b *testing.B) {
	for i := 0; i < b.N; i++ {
		query_string := fmt.Sprintf("lat=%f&long=%f&country=%s", 62.24, 25.74, "finland")
		request := fmt.Sprintf("http://localhost:%d/point?%s", test_geo.Config.Port, query_string)
		if _, err := http.Get(request); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAPIFinlandPointsToCoordinates(b *testing.B) {
	for i := 0; i < b.N; i++ {
		values := strings.Replace(fmt.Sprintf("%v", FinlandNodes), " ", ",", -1)
		query_string := fmt.Sprintf("nodes=%s&country=%s", values, "finland")
		request := fmt.Sprintf("http://localhost:%d/points?%s", test_geo.Config.Port, query_string)
		if res, err := http.Get(request); err != nil || res.StatusCode != 200 {
			b.Fatal(err)
		}
	}
}

func BenchmarkAPIGermanyPointsToCoordinates(b *testing.B) {
	for i := 0; i < b.N; i++ {
		values := strings.Replace(fmt.Sprintf("%v", GermanyNodes), " ", ",", -1)
		query_string := fmt.Sprintf("nodes=%s&country=%s", values, "germany")
		request := fmt.Sprintf("http://localhost:%d/points?%s", test_geo.Config.Port, query_string)
		if res, err := http.Get(request); err != nil || res.StatusCode != 200 {
			b.Fatal(err)
		}
	}
}
func TestAPICalculateCoordinatePathWithoutCoordinates(t *testing.T) {
	api_coordinate_query(t, noCoordsFinland, 100)
}
