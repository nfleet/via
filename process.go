package main

import (
	"encoding/json"
	"errors"
)

type Matrix struct {
	Nodes []int `json:"sources"`
}

// Parses the JSON input coordinates into an array.
func parse_json_matrix(matrix string) ([]Coord, error) {
	var target []Coord

	err := json.Unmarshal([]byte(matrix), &target)
	if err != nil {
		return nil, errors.New("JSON parsing error: " + err.Error())
	}
	return target, nil
}

// Converts the coordinate array back into JSON.
func dump_matrix_to_json(nodes []int) ([]byte, error) {
	cont, err := json.Marshal(Matrix{Nodes: nodes})
	if err != nil {
		return []byte{}, err
	}
	return cont, nil
}
