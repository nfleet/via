package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/nfleet/via/ch"
)

type Path struct {
	Length int   `json:"length"`
	Nodes  []int `json:"nodes"`
}

type Edge struct {
	Source Location
	Target Location
}

type NodeEdge struct {
	Source int `json:"source"`
	Target int `json:"target"`
}

type PathsInput struct {
	SpeedProfile int
	Edges        []Edge
}

type CoordinatePath struct {
	Distance int     `json:"distance"`
	Time     int     `json:"time"`
	Coords   []Coord `json:"coords"`
}

func IsMissingCoordinate(loc Location) bool {
	if loc.Coordinate.Latitude == 0.0 && loc.Coordinate.Longitude == 0.0 {
		return true
	}
	return false
}

func calculate_distance(config Config, nodes []int, country string) (int, error) {
	if len(nodes) < 2 {
		return 0, nil
	}

	db, _ := sql.Open("postgres", config.String())
	defer db.Close()

	var edgePairs []string

	for i := 0; i < len(nodes)-1; i++ {
		edgeStart, edgeEnd := nodes[i], nodes[i+1]
		s := fmt.Sprintf("(%d,%d)", edgeStart, edgeEnd)
		edgePairs = append(edgePairs, s)
	}

	edges := strings.Join(edgePairs, ",")

	q := `select sum(dist) from (values%s) as t left join %s_speed_edges on column1=id1 and column2=id2`

	q = fmt.Sprintf(q, edges, country)

	var sum float64
	err := db.QueryRow(q).Scan(&sum)
	switch {
	case err == sql.ErrNoRows:
		return 0, errors.New("No distance found. Check points exist.")
	case err != nil:
		return 0, err
	default:
		return int(sum), nil
	}

}

func correct_coordinates(config Config, source, target Location) (CHNode, CHNode, error) {
	// step 1: coordinate -> node
	srcLat, srcLong := source.Coordinate.Latitude, source.Coordinate.Longitude
	trgLat, trgLong := target.Coordinate.Latitude, target.Coordinate.Longitude

	country := source.Address.Country

	srcNode, err := CorrectPoint(config, Coord{srcLat, srcLong}, strings.ToLower(country))
	if err != nil {
		return CHNode{}, CHNode{}, err
	}

	trgNode, err := CorrectPoint(config, Coord{trgLat, trgLong}, strings.ToLower(country))
	if err != nil {
		return CHNode{}, CHNode{}, err
	}

	return srcNode, trgNode, nil
}

func CalculatePaths(nodeEdges []NodeEdge, country string, speed_profile int) ([]Path, error) {
	input_data, err := json.Marshal(nodeEdges)
	if err != nil {
		return []Path{}, err
	}

	country = strings.ToLower(country)

	// WHY THE HELL IS THIS NECESSARY?
	country += "\x00"

	res := ch.Calc_paths(string(input_data), string(country), speed_profile)
	res = clean_json_cpp_message(res)

	if !strings.HasSuffix(res, "]}]}") {
		res += "]}"
	}

	var edges struct {
		Edges []Path `json:"edges"`
	}

	if err := json.Unmarshal([]byte(res), &edges); err != nil {
		return []Path{}, err
	}

	return edges.Edges, nil
}

func CalculateCoordinatePaths(config Config, input PathsInput) ([]CoordinatePath, error) {
	var edges []NodeEdge

	for _, edge := range input.Edges {
		var source, target Location

		if IsMissingCoordinate(edge.Source) {
			var err error
			source, err = ResolveLocation(config, edge.Source)
			source.Address.Country = edge.Source.Address.Country
			if err != nil {
				return []CoordinatePath{}, err
			}
		} else {
			source = edge.Source
		}

		if IsMissingCoordinate(edge.Target) {
			var err error
			target, err = ResolveLocation(config, edge.Target)
			target.Address.Country = edge.Target.Address.Country
			if err != nil {
				return []CoordinatePath{}, err
			}
		} else {
			target = edge.Target
		}

		srcNode, trgNode, err := correct_coordinates(config, source, target)
		if err != nil {
			return []CoordinatePath{}, err
		}

		edges = append(edges, NodeEdge{srcNode.Id, trgNode.Id})
	}

	country := strings.ToLower(input.Edges[0].Source.Address.Country)

	nodePaths, err := CalculatePaths(edges, country, input.SpeedProfile)
	if err != nil {
		return []CoordinatePath{}, err
	}

	var paths []CoordinatePath

	for _, nodePath := range nodePaths {
		// step 3: get coordinates
		coordinateList, err := GetCoordinates(config, country, nodePath.Nodes)
		if err != nil {
			return []CoordinatePath{}, err
		}

		// step 4: get distance
		distance, err := calculate_distance(config, nodePath.Nodes, country)
		if err != nil {
			return []CoordinatePath{}, err
		}

		paths = append(paths, CoordinatePath{Distance: distance, Time: nodePath.Length, Coords: coordinateList})
	}

	return paths, nil
}
