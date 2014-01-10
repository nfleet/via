package main

import (
	"encoding/json"
	"errors"
	"github.com/co-sky-developers/via/dmatrix"
	"strings"
)

type Path struct {
	Length int   `json:"length"`
	Nodes  []int `json:"nodes"`
}

type CoordinatePath struct {
	Length int     `json:"length"`
	Coords []Coord `json:"coords"`
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
	if loc.Coordinate.Latitude == 0.0 || loc.Coordinate.Longitude == 0.0 {
		return true
	}
	return false
}

func CalculateCoordinatePathFromAddresses(config Config, source, target Location, speed_profile int) (CoordinatePath, error) {
	if IsMissingCoordinate(source) {
		// resolve it
		return CoordinatePath{}, errors.New("missing coord from source")
	}
	if IsMissingCoordinate(target) {
		// resolve it
		return CoordinatePath{}, errors.New("missing coord from target")
	}

	// step 1: coordinate -> node
	srcLat, srcLong := source.Coordinate.Latitude, source.Coordinate.Longitude
	trgLat, trgLong := target.Coordinate.Latitude, target.Coordinate.Longitude

	srcNode, err := CorrectPoint(config, Coord{srcLat, srcLong}, strings.ToLower(source.Address.Country))
	if err != nil {
		return CoordinatePath{}, err
	}
	trgNode, err := CorrectPoint(config, Coord{trgLat, trgLong}, strings.ToLower(target.Address.Country))
	if err != nil {
		return CoordinatePath{}, err
	}

	// step 2: calculate path
	path, err := CalculatePath(srcNode.Id, trgNode.Id, strings.ToLower(source.Address.Country), speed_profile)
	if err != nil {
		return CoordinatePath{}, err
	}

	// step 3: get coordinates
	coordinateList, err := GetCoordinates(config, strings.ToLower(source.Address.Country), path.Nodes)
	if err != nil {
		return CoordinatePath{}, err
	}

	return CoordinatePath{Length: path.Length, Coords: coordinateList}, nil
}
