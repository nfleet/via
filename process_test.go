package main

import (
	"encoding/json"
	"testing"
)

func TestParsing(t *testing.T) {
	dummy := Matrix{Nodes: []int{1, 2, 3, 4, 5}}

	blob, err := dump_matrix_to_json(dummy.Nodes)
	if err != nil {
		t.Error("json marshal failed: " + err.Error())
	}

	var target Matrix

	err = json.Unmarshal(blob, &target)
	if err != nil {
		t.Error("json unmarshal failed: " + err.Error())
	}
}
