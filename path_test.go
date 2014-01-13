package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

var (
	srcJsonWithCoord = `{"Coordinate":{"Latitude":62.24027,"Longitude":25.74444}, "Address": {"Country": "Finland"}}`
	trgJsonWithCoord = `{"Coordinate":{"Latitude":60.45138,"Longitude":22.26666}, "Address": {"Country": "Finland"}}`
	srcJson          = `{ "Address": {"Street": "Taitoniekantie", "Country": "Finland"}}`
	trgJson          = `{ "Address": {"Street": "Viitaniementie", "Country": "Finland"}}`
)

func BenchmarkFinlandPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := CalculatePath(253299, 762749, "finland", 60)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGermanyPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := CalculatePath(54184, 3165075, "germany", 60)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestCalculateCoordinatePathWithCoordinates(t *testing.T) {

	var source, target Location
	if err := json.Unmarshal([]byte(srcJsonWithCoord), &source); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal([]byte(trgJsonWithCoord), &target); err != nil {
		t.Fatal(err)
	}

	if _, err := CalculateCoordinatePathFromAddresses(server.Config, source, target, 60); err != nil {
		t.Fatal(err)
	}
}

func api_coordinate_query(t *testing.T, source, target string, speed_profile int) {
	query_string := fmt.Sprintf("source=%s&target=%s&speed_profile=%d",
		url.QueryEscape(source), url.QueryEscape(target), 60)
	request := fmt.Sprintf("http://localhost:%d/cpath?%s", server.Config.Port, query_string)

	response, err := http.Get(request)
	if err != nil || response.StatusCode != 200 {
		t.Fatal(response)
	} else {
		defer response.Body.Close()
		cont, err := ioutil.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}

		var path CoordinatePath
		err = json.Unmarshal(cont, &path)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("Got path length of %d seconds (%d coords)", path.Length, len(path.Coords))
	}
}

func TestAPICalculateCoordinatePathWithExplicitCoordinates(t *testing.T) {
	api_coordinate_query(t, srcJsonWithCoord, trgJsonWithCoord, 60)
}

func TestAPICalculateCoordinatePathWithoutCoordinates(t *testing.T) {
	api_coordinate_query(t, srcJson, trgJson, 60)
}
