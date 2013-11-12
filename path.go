package main

import (
	"encoding/json"
	"github.com/co-sky-developers/via/dmatrix"
)

type Path struct {
	length int
	nodes  []int
}

func CalculatePath(source, target int, country string, speed_profile int) (Path, error) {
	var input = struct {
		source int
		target int
	}{
		source,
		target,
	}

	input_data, err := json.Marshal(input)
	if err != nil {
		return Path{}, err
	}

	res := dmatrix.Calc_path(input_data, country, speed_profile)

	var path Path
	if err := json.Unmarshal([]byte(res), &path); err != nil {
		return Path{}, err
	}

	return path, nil
}
