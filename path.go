package main

import (
	"encoding/json"
	"github.com/co-sky-developers/via/dmatrix"
)

type Path struct {
	Length int   `json:"length"`
	Nodes  []int `json:"nodes"`
}

func CalculatePath(source, target int, country string, speed_profile int) (Path, error) {
	var input = struct {
		Source int `json:"source"`
		Target int `json:"target"`
	}{
		source,
		target,
	}

	input_data, err := json.Marshal(input)
	if err != nil {
		return Path{}, err
	}

	// WHY THE HELL IS THIS NECESSARY?
	country += "\x00"

	res := dmatrix.Calc_path(string(input_data), string(country), speed_profile)
	res = clean_json_cpp_message(res)

	var path Path
	if err := json.Unmarshal([]byte(res), &path); err != nil {
		return Path{}, err
	}

	return path, nil
}
