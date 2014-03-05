package postgeodb

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/nfleet/via/geotypes"
)

func (g GeoPostgresDB) QueryFuzzyAddress(address geotypes.Address, count int) ([]geotypes.Location, error) {
	newconf := geotypes.Config(g.Config)
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
		return []geotypes.Location{}, errors.New("Country " + country +
			" not recognized")
	}

	q := fmt.Sprintf("SELECT id, coord, city, name, sml "+
		"from %s('%s') WHERE city LIKE '%%%s%%'"+
		"ORDER BY sml DESC LIMIT %d",
		country_funcs[country],
		address.Street, address.City, count)

	rows, err := db.Query(q)

	if err != nil {
		return []geotypes.Location{}, err
	}

	var locations []geotypes.Location

	for rows.Next() {
		var (
			coord, street_name, city string
			id                       int
			conf                     float64
		)

		if err := rows.Scan(&id, &coord,
			&city, &street_name, &conf); err != nil {
			return []geotypes.Location{}, err
		}

		// parse stupid coordinates
		// this workaround is fucking horrible
		var coordinate [2]float64
		if country == "finland" {
			var coordinateResult [][2]float64

			coord = strings.Replace(coord, "(", "[", -1)
			coord = strings.Replace(coord, ")", "]", -1)

			if err := json.Unmarshal([]byte(coord), &coordinateResult); err != nil {
				return []geotypes.Location{}, err
			}

			coordinate = coordinateResult[0]
		} else if country == "germany" {
			var lat, long float64
			fmt.Sscanf(coord, "(%f,%f)", &lat, &long)
			coordinate[0], coordinate[1] = long, lat
		}

		locations = append(locations, geotypes.Location{geotypes.Address{Street: street_name, City: city, Confidence: conf * 100}, geotypes.Coordinate{Latitude: coordinate[1], Longitude: coordinate[0], System: "WGS84"}})
	}

	if err := rows.Err(); err != nil {
		return []geotypes.Location{}, err
	}

	return locations, nil
}
