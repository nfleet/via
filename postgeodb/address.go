package postgeodb

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/nfleet/via/geotypes"
)

// GeoPoint maps against Postgis geographical point
type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (p *GeoPoint) String() string {
	return fmt.Sprintf("POINT(%v %v)", p.Lat, p.Lng)
}

// Scan implements the Scanner interface and will scan the Postgis POINT(x y) into the GeoPoint struct
func (p *GeoPoint) Scan(val interface{}) error {
	b := string(val.([]uint8))

	var lat, lng float64
	fmt.Sscanf(b, "POINT(%f %f)", &lng, &lat)

	p.Lat = lat
	p.Lng = lng

	return nil
}

func upperFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

// Returns the fixed Location DTO for the Address, i.e. geocodes the Address and returns count results.
func (g GeoPostgresDB) QueryFuzzyAddress(address geotypes.Address, count int) ([]geotypes.Location, error) {
	db := g.DB
	country_funcs := map[string]string{
		"finland": "get_street_city",
		"germany": "get_appr_germany",
	}

	street := strings.ToLower(address.Street)
	city := strings.ToLower(address.City)
	country := strings.ToLower(address.Country)

	if _, ok := country_funcs[country]; !ok {
		return []geotypes.Location{}, errors.New("Country " + country + " not recognized")
	}

	q := fmt.Sprintf("SELECT id, house_numbers, coord, city, name, sml FROM %s($1, $2) LIMIT $3", country_funcs[country])
	rows, err := db.Query(q, street, city, count)

	if err != nil {
		return []geotypes.Location{}, err
	}

	var locations []geotypes.Location

	for rows.Next() {
		var (
			coord, street_name, city string
			real_coord               GeoPoint
			house_num, id            int
			conf                     float64
		)

		if country == "finland" {
			if err := rows.Scan(&id, &house_num, &real_coord, &city, &street_name, &conf); err != nil {
				return []geotypes.Location{}, err
			}
		} else {
			if err := rows.Scan(&id, &coord, &city, &street_name, &conf); err != nil {
				return []geotypes.Location{}, err
			}
		}

		// parse stupid coordinates
		// this workaround is fucking horrible
		var coordinate [2]float64
		if country == "finland" {
			coordinate = [2]float64{real_coord.Lat, real_coord.Lng}

			if house_num == 2 && address.HouseNumber != 0 {
				var c GeoPoint
				q = "SELECT get_house_number FROM get_house_number($1, $2)"
				row := db.QueryRow(q, id, address.HouseNumber)
				if err := row.Scan(&c); err != nil {
					return []geotypes.Location{}, err
				}
				coordinate = [2]float64{c.Lat, c.Lng}
			}
		} else if country == "germany" {
			var lat, long float64
			fmt.Sscanf(coord, "(%f,%f)", &lat, &long)
			coordinate[0], coordinate[1] = long, lat
		}

		var h int
		if house_num == 2 {
			h = address.HouseNumber
		}

		locations = append(locations, geotypes.Location{geotypes.Address{HouseNumber: h, Street: street_name, Country: upperFirst(country), City: city, Confidence: conf * 100}, geotypes.Coordinate{Latitude: coordinate[0], Longitude: coordinate[1], System: "WGS84"}})
	}

	if err := rows.Err(); err != nil {
		return []geotypes.Location{}, err
	}

	return locations, nil
}
