package main

import (
	"fmt"
	"net/http"
	"testing"
)

func BenchmarkPointCorrection(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CorrectPoint(server.Config, Coord{62.24, 25.74}, "finland")
	}
}

func BenchmarkAPIPointCorrection(b *testing.B) {
	for i := 0; i < b.N; i++ {
		query_string := fmt.Sprintf("lat=%f&long=%f&country=%s", 62.24, 25.74, "finland")
		request := fmt.Sprintf("http://localhost:%d/point?%s", server.Config.Port, query_string)
		_, err := http.Get(request)
		if err != nil {
			b.Fatal(err)
		}
	}
}
