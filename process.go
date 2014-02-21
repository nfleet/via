package main

import (
	"encoding/json"
	"errors"
	"strings"
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

func clean_json_cpp_message(msg string) string {
	res := msg
	if strings.LastIndex(res, "}") == -1 {
		res = res + "}"
		debug.Println("brace missing")
	} else if strings.LastIndex(res, "}") != len(res)-1 {
		braceIndex := strings.LastIndex(res, "}")
		junk := res[braceIndex+1:]
		res = res[:braceIndex+1]
		debug.Printf("Stripped extra data: %s", junk)
	}
	return res
}
