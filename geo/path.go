package geo

import (
	"encoding/json"
	"strings"

	"github.com/nfleet/via/ch"
	"github.com/nfleet/via/geotypes"
)

func IsMissingCoordinate(loc geotypes.Location) bool {
	if loc.Coordinate.Latitude == 0.0 && loc.Coordinate.Longitude == 0.0 {
		return true
	}
	return false
}

func (g *Geo) CorrectCoordinates(source, target geotypes.Location) (geotypes.CHNode, geotypes.CHNode, error) {
	// step 1: coordinate -> node
	srcLat, srcLong := source.Coordinate.Latitude, source.Coordinate.Longitude
	trgLat, trgLong := target.Coordinate.Latitude, target.Coordinate.Longitude

	country := source.Address.Country

	srcNode, err := g.DB.QueryClosestPoint(geotypes.Coord{srcLat, srcLong}, strings.ToLower(country))
	if err != nil {
		return geotypes.CHNode{}, geotypes.CHNode{}, err
	}

	trgNode, err := g.DB.QueryClosestPoint(geotypes.Coord{trgLat, trgLong}, strings.ToLower(country))
	if err != nil {
		return geotypes.CHNode{}, geotypes.CHNode{}, err
	}

	return srcNode, trgNode, nil
}

func (g *Geo) CalculatePaths(nodeEdges []geotypes.NodeEdge, country string, speed_profile int) ([]geotypes.Path, error) {
	input_data, err := json.Marshal(nodeEdges)
	if err != nil {
		return []geotypes.Path{}, err
	}

	country = strings.ToLower(country)

	res := ch.Calc_paths(string(input_data), country, speed_profile, g.DataDir)
	res = g.CleanCppMessage(res)

	if !strings.HasSuffix(res, "]}]}") {
		res += "]}"
	}

	var edges struct {
		Edges []geotypes.Path `json:"edges"`
	}

	if err := json.Unmarshal([]byte(res), &edges); err != nil {
		return []geotypes.Path{}, err
	}

	return edges.Edges, nil
}

func (g *Geo) CalculateCoordinatePaths(input geotypes.PathsInput) ([]geotypes.CoordinatePath, error) {
	var edges []geotypes.NodeEdge

	for _, edge := range input.Edges {
		var source, target geotypes.Location

		if IsMissingCoordinate(edge.Source) {
			var err error
			locs, err := g.ResolveLocation(edge.Source, 1)
			if err != nil {
				return []geotypes.CoordinatePath{}, err
			}
			source = locs[0]
			source.Address.Country = edge.Source.Address.Country
		} else {
			source = edge.Source
		}

		if IsMissingCoordinate(edge.Target) {
			var err error
			locs, err := g.ResolveLocation(edge.Target, 1)
			if err != nil {
				return []geotypes.CoordinatePath{}, err
			}
			target = locs[0]
			target.Address.Country = edge.Target.Address.Country
		} else {
			target = edge.Target
		}

		srcNode, trgNode, err := g.CorrectCoordinates(source, target)
		if err != nil {
			return []geotypes.CoordinatePath{}, err
		}

		edges = append(edges, geotypes.NodeEdge{srcNode.Id, trgNode.Id})
	}

	country := strings.ToLower(input.Edges[0].Source.Address.Country)

	nodePaths, err := g.CalculatePaths(edges, country, input.SpeedProfile)
	if err != nil {
		return []geotypes.CoordinatePath{}, err
	}

	var paths []geotypes.CoordinatePath

	for _, nodePath := range nodePaths {
		// step 3: get coordinates
		coordinateList, err := g.DB.QueryCoordinates(nodePath.Nodes, country)
		if err != nil {
			return []geotypes.CoordinatePath{}, err
		}

		// step 4: get distance
		distance, err := g.DB.QueryDistance(nodePath.Nodes, country)
		if err != nil {
			return []geotypes.CoordinatePath{}, err
		}

		paths = append(paths, geotypes.CoordinatePath{Distance: distance, Time: nodePath.Length, Coords: coordinateList})
	}

	return paths, nil
}
