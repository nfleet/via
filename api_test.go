package main

import (
	"testing"
)

type P struct {
	matrix  []Coord
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
