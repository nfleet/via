package main

import (
	"database/sql"
	"fmt"
)

type Coordinate struct {
	Latitude  float64
	Longitude float64
}

type Location struct {
	Address    Address
	Coordinate Coordinate
}

type Address struct {
	Street      string
	House       string
	HouseNumber string
	City        string
	Country     string
}

func GetFuzzyAddress(config Config, address, country string, count int) ([]Location, error) {
	newconf := Config(config)
	newconf.DbName = "trgm_test"

	db, err := sql.Open("postgres", newconf.String())
	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf("SELECT name, city, coord[0], coord[1] from get_appr('%s') LIMIT %d", address, count)
	rows, err := db.Query(q)

	if err != nil {
		return []Location{}, err
	}

	var locations []Location

	for rows.Next() {
		var (
			street_name, city string
			lat, long         float64
		)

		if err := rows.Scan(&street_name, &city, &lat, &long); err != nil {
			return []Location{}, err
		}
		locations = append(locations,
			Location{Address{Street: street_name, City: city},
				Coordinate{Latitude: lat, Longitude: long}})
	}

	if err := rows.Err(); err != nil {
		return []Location{}, err
	}

	return locations, nil
}
