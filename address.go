package main

import (
	"fmt"
	"database/sql"
)

type Address struct {
	Street string
	House string
	Number string
	City string
	Country string
	Coordinate Coord
}

func GetFuzzyAddress(config Config, address, country string, count int) ([]Address, error) {
	newconf := Config(config)
	newconf.DbName = "trgm_test"

	db, err := sql.Open("postgres", newconf.String())
	if err != nil {
		return nil, err
	}


	q := fmt.Sprintf("SELECT name, city, coord[0], coord[1] from get_appr('%s') LIMIT %d", address, count)
	rows, err := db.Query(q)

	if err != nil {
		return []Address{}, err
	}

	var addresses []Address

	for rows.Next() {
		var (
			street_name, city string
			lat, long float64
		)

		if err := rows.Scan(&street_name, &city, &lat, &long); err != nil {
			return []Address{}, err
		}
		addresses = append(addresses, Address{Street: street_name, City: city, Coordinate: Coord{lat, long}})
	}

	if err := rows.Err(); err != nil {
		return []Address{}, err
	}

	return addresses, nil
}
