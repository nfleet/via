package geo

import (
	"encoding/json"
	"testing"

	"github.com/nfleet/via/geotypes"
)

func TestParsing(t *testing.T) {
	dummy := geotypes.Matrix{Nodes: []int{1, 2, 3, 4, 5}}

	blob, err := MatrixToJson(dummy.Nodes)
	if err != nil {
		t.Error("json marshal failed: " + err.Error())
	}

	var target geotypes.Matrix

	if err := json.Unmarshal(blob, &target); err != nil {
		t.Error("json unmarshal failed: " + err.Error())
	}
}
