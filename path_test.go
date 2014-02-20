package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

type TestPayload struct {
	Source Location
	Target Location
}

var (
	srcCoord = Location{Coordinate: Coordinate{Latitude: 62.24027, Longitude: 25.74444}, Address: Address{Country: "Finland"}}
	trgCoord = Location{Coordinate: Coordinate{Latitude: 60.45138, Longitude: 22.26666}, Address: Address{Country: "Finland"}}
	noCoords = []TestPayload{
		{Source: Location{Address: Address{Street: "Erottaja", City: "Helsinki", Country: "Finland"}}, Target: Location{Address: Address{Street: "Esplanadi", City: "Helsinki", Country: "Finland"}}},
	}
)

func BenchmarkFinlandPathExtraction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := CalculatePath(253299, 762749, "finland", 60)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGermanyPathExtraction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := CalculatePath(54184, 3165075, "germany", 60)
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

func api_coordinate_query(t *testing.T, edges []TestPayload, speed_profile int) []CoordinatePath {
	payload := struct {
		SpeedProfile int `json:'speed_profile'`
		Edges        []TestPayload
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

		for i, p := range paths {
			t.Logf("Calculated path of %d km / %d secs from %s, %s, %s to %s, %s, %s", p.Distance, p.Time, 
				edges[i].Source.Address.Street, edges[i].Source.Address.City, edges[i].Source.Address.Country,
				edges[i].Target.Address.Street, edges[i].Target.Address.City, edges[i].Target.Address.Country)
		}

		t.Logf("Calculated %d paths", len(paths))
	}

	return paths
}

func TestAPICalculateCoordinatePathWithExplicitCoordinates(t *testing.T) {
	api_coordinate_query(t, []TestPayload{{srcCoord, trgCoord}}, 100)
}

func TestAPICalculateCoordinatePathWithoutCoordinates(t *testing.T) {
	api_coordinate_query(t, noCoords, 100)
}
