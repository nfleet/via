package geo

import (
	"errors"

	"github.com/nfleet/via/geotypes"
)

// Resolves a location from the database.
// Returns 20 when everything fails (i.e. database problem), 30 when
// an address could not be found or when the street wasn't supplied.
func (g *Geo) ResolveLocation(location geotypes.Location) (geotypes.Location, error) {
	if IsMissingCoordinate(location) {
		if location.Address.Street != "" {
			locs, err := g.DB.QueryFuzzyAddress(location.Address, 1)
			if err != nil {
				location.Address.Confidence = 20.0
				return location, err
			}

			if len(locs) == 0 {
				location.Address.Confidence = 30.0
				return location, nil
			}

			location.Coordinate = locs[0].Coordinate
			location.Address = locs[0].Address
			return location, nil
		} else {
			// skip
			location.Address.Confidence = 30.0
			return location, nil
		}
	} else {
		if location.Address.Country == "" {
			return location, errors.New("Must provide country in Location.Address!")
		}

		coord := location.Coordinate
		correctCoord, err := g.DB.QueryClosestPoint(geotypes.Coord{coord.Latitude, coord.Longitude}, location.Address.Country)

		if err != nil {
			return location, err
		}

		location.Coordinate.Latitude = correctCoord.Coord[0]
		location.Coordinate.Longitude = correctCoord.Coord[1]

		return location, nil
	}
}
