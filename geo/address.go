package geo

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func (g *Geo) GetFuzzyAddress(address Address, count int) ([]Location, error) {
	newconf := Config(g.Config)
	newconf.DbName = "trgm_test"

	db, err := sql.Open("postgres", newconf.String())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	country_funcs := map[string]string{
		"finland": "get_appr2",
		"germany": "get_appr_germany",
	}

	country := strings.ToLower(address.Country)

	if _, ok := country_funcs[country]; !ok {
		return []Location{}, errors.New("Country " + country + " not recognized")
	}

	q := fmt.Sprintf("SELECT id, coord, city, name, sml from %s('%s') WHERE city LIKE '%%%s%%' ORDER BY sml DESC LIMIT %d", country_funcs[country], address.Street, address.City, count)

	rows, err := db.Query(q)

	if err != nil {
		return []Location{}, err
	}

	var locations []Location

	for rows.Next() {
		var (
			coord, street_name, city string
			id                       int
			conf                     float64
		)

		if err := rows.Scan(&id, &coord, &city, &street_name, &conf); err != nil {
			return []Location{}, err
		}

		// parse stupid coordinates
		// this workaround is fucking horrible
		var coordinate [2]float64
		if country == "finland" {
			var coordinateResult [][2]float64

			coord = strings.Replace(coord, "(", "[", -1)
			coord = strings.Replace(coord, ")", "]", -1)

			if err := json.Unmarshal([]byte(coord), &coordinateResult); err != nil {
				return []Location{}, err
			}

			coordinate = coordinateResult[0]
		} else if country == "germany" {
			var lat, long float64
			fmt.Sscanf(coord, "(%f,%f)", &lat, &long)
			coordinate[0], coordinate[1] = long, lat
		}

		locations = append(locations, Location{Address{Street: street_name, City: city, Confidence: conf * 100}, Coordinate{Latitude: coordinate[1], Longitude: coordinate[0], System: "WGS84"}})
	}

	if err := rows.Err(); err != nil {
		return []Location{}, err
	}

	return locations, nil
}

// Resolves a location from the database.
// Returns 20 when everything fails (i.e. database problem), 30 when
// an address could not be found or when the street wasn't supplied.
func (g *Geo) ResolveLocation(location Location) (Location, error) {
	if IsMissingCoordinate(location) {
		if location.Address.Street != "" {
			locs, err := g.GetFuzzyAddress(location.Address, 1)
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
		correctCoord, err := g.CorrectPoint(Coord{coord.Latitude, coord.Longitude}, location.Address.Country)

		if err != nil {
			return location, err
		}

		location.Coordinate.Latitude = correctCoord.Coord[0]
		location.Coordinate.Longitude = correctCoord.Coord[1]

		return location, nil
	}
}
