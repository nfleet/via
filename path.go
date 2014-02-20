package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nfleet/via/dmatrix"
	"strings"
)

type Path struct {
	Length int   `json:"length"`
	Nodes  []int `json:"nodes"`
}

type CoordinatePath struct {
	Distance int     `json:"distance"`
	Time     int     `json:"time"`
	Coords   []Coord `json:"coords"`
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

func CalculateCoordinatePathFromAddresses(config Config, source, target Location, speed_profile int) (CoordinatePath, error) {
	srcCountry := source.Address.Country
	trgCountry := target.Address.Country
	if IsMissingCoordinate(source) {
		// resolve i
		var err error
		source, err = ResolveLocation(config, source)
		if err != nil {
			return CoordinatePath{}, err
		}
	}

	if IsMissingCoordinate(target) {
		// resolve it
		var err error
		target, err = ResolveLocation(config, target)
		if err != nil {
			return CoordinatePath{}, err
		}
	}


	// step 1: coordinate -> node
	srcLat, srcLong := source.Coordinate.Latitude, source.Coordinate.Longitude
	trgLat, trgLong := target.Coordinate.Latitude, target.Coordinate.Longitude

	srcNode, err := CorrectPoint(config, Coord{srcLat, srcLong}, strings.ToLower(srcCountry))
	if err != nil {
		return CoordinatePath{}, err
	}

	trgNode, err := CorrectPoint(config, Coord{trgLat, trgLong}, strings.ToLower(trgCountry))
	if err != nil {
		return CoordinatePath{}, err
	}

	// step 2: calculate path
	path, err := CalculatePath(srcNode.Id, trgNode.Id, strings.ToLower(srcCountry), speed_profile)
	if err != nil {
		return CoordinatePath{}, err
	}

	// step 3: get coordinates
	coordinateList, err := GetCoordinates(config, strings.ToLower(srcCountry), path.Nodes)
	if err != nil {
		return CoordinatePath{}, err
	}

	// step 4: get distance
	distance, err := calculate_distance(config, path.Nodes, strings.ToLower(srcCountry))
	if err != nil {
		return CoordinatePath{}, err
	}

	return CoordinatePath{Distance: distance, Time: path.Length, Coords: coordinateList}, nil
}
