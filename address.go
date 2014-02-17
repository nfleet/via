package main

import (
	"database/sql"
	"errors"
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
	Street     string
	City       string
	Country    string
	Region     string
	PostalCode string

	ApartmentLetter string
	ApartmentNumber int
	Similarity      float64
}

func GetFuzzyAddress(config Config, address string, count int) ([]Location, error) {
	newconf := Config(config)
	newconf.DbName = "trgm_test"

	db, err := sql.Open("postgres", newconf.String())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	q := fmt.Sprintf("SELECT name, city, coord[0], coord[1], sml from get_appr('%s') LIMIT %d", address, count)
	rows, err := db.Query(q)

	if err != nil {
		return []Location{}, err
	}

	var locations []Location

	for rows.Next() {
		var (
			street_name, city string
			lat, long, sml    float64
		)

		if err := rows.Scan(&street_name, &city, &lat, &long, &sml); err != nil {
			return []Location{}, err
		}
		locations = append(locations,
			Location{Address{Street: street_name, City: city, Similarity: sml},
				Coordinate{Latitude: lat, Longitude: long}})
	}

	if err := rows.Err(); err != nil {
		return []Location{}, err
	}

	return locations, nil
}

func ResolveLocation(config Config, location Location) (Location, error) {
	if IsMissingCoordinate(location) {
		street := location.Address.Street
		if street != "" {
			locs, err := GetFuzzyAddress(config, street, 1)
			if err != nil {
				return Location{}, err
			}

			if len(locs) == 0 {
				e := errors.New("Couldn't find any address for street " + street)
				return Location{}, e
			}

			location.Coordinate = locs[0].Coordinate
			return location, nil
		} else {
			return Location{}, errors.New("Street empty, cannot search.")
		}
	} else {
		if location.Address.Country == "" {
			return Location{}, errors.New("Must provide country in Location.Address!")
		}

		coord := location.Coordinate
		correctCoord, err := CorrectPoint(config,
			Coord{coord.Latitude, coord.Longitude}, location.Address.Country)

		if err != nil {
			return Location{}, err
		}

		location.Coordinate.Latitude = correctCoord.Coord[0]
		location.Coordinate.Longitude = correctCoord.Coord[1]

		return location, nil
	}
}
