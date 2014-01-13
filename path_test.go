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
	srcJson = `{"Coordinate":{"Latitude":62.24027,"Longitude":25.74444}, "Address": {"Country": "Finland"}}`
	trgJson = `{"Coordinate":{"Latitude":60.45138,"Longitude":22.26666}, "Address": {"Country": "Finland"}}`
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
	if err := json.Unmarshal([]byte(srcJson), &source); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal([]byte(trgJson), &target); err != nil {
		t.Fatal(err)
	}

	if _, err := CalculateCoordinatePathFromAddresses(server.Config, source, target, 60); err != nil {
		t.Fatal(err)
	}
}

func TestAPICalculateCoordinatePath(t *testing.T) {
	query_string := fmt.Sprintf("source=%s&target=%s&speed_profile=%d",
		url.QueryEscape(srcJson), url.QueryEscape(trgJson), 60)
	request := fmt.Sprintf("http://localhost:%d/cpath?%s", server.Config.Port, query_string)
	fmt.Println(request)

	response, err := http.Get(request)
	if err != nil || response.StatusCode != 200 {
		t.Fatal(err)
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
	}
}
