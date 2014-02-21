package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

var (
	srcCoord = Location{Coordinate: Coordinate{Latitude: 62.24027, Longitude: 25.74444}, Address: Address{Country: "Finland"}}
	trgCoord = Location{Coordinate: Coordinate{Latitude: 60.45138, Longitude: 22.26666}, Address: Address{Country: "Finland"}}
	noCoords = []Edge{
		{Source: Location{Address: Address{Street: "Erottaja", City: "Helsinki", Country: "Finland"}}, Target: Location{Address: Address{Street: "Esplanadi", City: "Helsinki", Country: "Finland"}}},
	}

	noCoordsFinland = []Edge{
		{Source: Location{Address: Address{Street: "Erottaja", City: "Helsinki", Country: "Finland"}}, Target: Location{Address: Address{Street: "Esplanadi", City: "Helsinki", Country: "Finland"}}},
		{Source: Location{Address: Address{Street: "Taitoniekantie", City: "Jyväskylä", Country: "Finland"}}, Target: Location{Address: Address{Street: "Vuolteenkatu", City: "Tampere", Country: "Finland"}}},
		{Source: Location{Address: Address{Street: "Vuolteenkatu", City: "Tampere", Country: "Finland"}}, Target: Location{Address: Address{Street: "Hikiäntie", City: "Riihimäki", Country: "Finland"}}},
	}

	coordsGermany = []Edge{
		{Source: Location{Coordinate: Coordinate{Latitude: 49.0811, Longitude: 9.8795}, Address: Address{Country: "Germany"}}, Target: Location{Coordinate: Coordinate{Latitude: 49.3636, Longitude: 10.1472}, Address: Address{Country: "Germany"}}},
		{Source: Location{Coordinate: Coordinate{Latitude: 49.4413, Longitude: 9.7394}, Address: Address{Country: "Germany"}}, Target: Location{Coordinate: Coordinate{Latitude: 48.4356, Longitude: 12.4750}, Address: Address{Country: "Germany"}}},
		{Source: Location{Coordinate: Coordinate{Latitude: 47.9074, Longitude: 10.4459}, Address: Address{Country: "Germany"}}, Target: Location{Coordinate: Coordinate{Latitude: 48.4133, Longitude: 9.2127}, Address: Address{Country: "Germany"}}},
		{Source: Location{Coordinate: Coordinate{Latitude: 50.1690, Longitude: 11.2960}, Address: Address{Country: "Germany"}}, Target: Location{Coordinate: Coordinate{Latitude: 52.4292, Longitude: 13.2286}, Address: Address{Country: "Germany"}}},
		{Source: Location{Coordinate: Coordinate{Latitude: 52.5876, Longitude: 13.2461}, Address: Address{Country: "Germany"}}, Target: Location{Coordinate: Coordinate{Latitude: 53.4989, Longitude: 13.9805}, Address: Address{Country: "Germany"}}},
	}
)

func BenchmarkFinlandPathsExtraction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := CalculateCoordinatePaths(server.Config, PathsInput{100, noCoordsFinland})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGermanyPathsExtraction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := CalculateCoordinatePaths(server.Config, PathsInput{100, coordsGermany})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestCalculateCoordinatePathWithCoordinates(t *testing.T) {
	if _, err := CalculateCoordinatePathFromAddresses(server.Config, srcCoord, trgCoord, 100); err != nil {
		t.Fatal(err)
	}
}

func TestCalculatePathsWithoutCoordinates(t *testing.T) {
	_, err := CalculateCoordinatePaths(server.Config, PathsInput{100, noCoordsFinland})

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Finland (%d edges) ok.", len(noCoordsFinland))

	_, err = CalculateCoordinatePaths(server.Config, PathsInput{100, coordsGermany})

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Germany (%d edges) ok.", len(coordsGermany))
}

func BenchmarkFinlandCoordinatelessPathExtraction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := CalculateCoordinatePaths(server.Config, PathsInput{100, noCoordsFinland})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func api_coordinate_query(t *testing.T, edges []Edge, speed_profile int) []CoordinatePath {
	payload := struct {
		SpeedProfile int
		Edges        []Edge
	}{
		speed_profile,
		edges,
	}

	jsonPayload, err := json.Marshal(payload)
	b := strings.NewReader(string(jsonPayload))

	var paths []CoordinatePath

	request := fmt.Sprintf("http://localhost:%d/cpaths", server.Config.Port)
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

func TestAPICalculateCoordinatePathWithExplicitCoordinates(t *testing.T) {
	api_coordinate_query(t, []Edge{{srcCoord, trgCoord}}, 100)
}

func TestAPICalculateCoordinatePathWithoutCoordinates(t *testing.T) {
	api_coordinate_query(t, noCoordsFinland, 100)
}
