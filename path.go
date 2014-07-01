package main

import (
	"encoding/json"
	"strings"

	"github.com/nfleet/via/ch"
	"github.com/nfleet/via/geotypes"
)

func (v *Via) CalculatePaths(nodeEdges []geotypes.NodeEdge, country string, speed_profile int) ([]geotypes.Path, error) {
	input_data, err := json.Marshal(nodeEdges)
	if err != nil {
		return []geotypes.Path{}, err
	}

	country = strings.ToLower(country)

	res := ch.Calc_paths(string(input_data), country, speed_profile, v.DataDir)
	var edges struct {
		Edges []geotypes.Path `json:"edges"`
	}

	if err := json.Unmarshal([]byte(res), &edges); err != nil {
		return []geotypes.Path{}, err
	}

	return edges.Edges, nil
}
