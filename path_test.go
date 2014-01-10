package main

import (
	"encoding/json"
	"fmt"
	"testing"
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
	srcJson := `{"Coordinate":{"Latitude":62.24027,"Longitude":25.74444},
				 "Address": {"Country": "Finland"}}`
	trgJson := `{"Coordinate":{"Latitude":60.45138,"Longitude":22.26666},
				 "Address": {"Country": "Finland"}}`

	var source, target Location
	if err := json.Unmarshal([]byte(srcJson), &source); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal([]byte(trgJson), &target); err != nil {
		t.Fatal(err)
	}

	fmt.Printf("s: %#v t: %#v\n", source, target)

	result, err := CalculateCoordinatePathFromAddresses(server.Config, source, target, 60)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(result.Coords)
}
