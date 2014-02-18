package main

import (
	"database/sql"
	"strings"
	"encoding/json"
	"errors"
	"fmt"
)

type Coordinate struct {
	Latitude  float64
	Longitude float64
	System string
}

type Location struct {
	Address    Address
	Coordinate Coordinate
}

type Address struct {
	Street     string
	HouseNumber int
	City       string
	Country    string
	Region     string
	PostalCode string
	Confidence float64
	ApartmentLetter string
	ApartmentNumber int
	AdditionalInfo string
}

func GetFuzzyAddress(config Config, address Address, count int) ([]Location, error) {
	newconf := Config(config)
	newconf.DbName = "trgm_test"

	db, err := sql.Open("postgres", newconf.String())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	q := fmt.Sprintf("SELECT id, coord, city, name, sml from get_appr2('%s') WHERE city LIKE '%%%s%%' ORDER BY sml DESC", address.Street, address.City)

	rows, err := db.Query(q)

	if err != nil {
		return []Location{}, err
	}

	var locations []Location

	for rows.Next() {
		var (
			coord, street_name, city string
			id int
			conf float64
		)

		if err := rows.Scan(&id, &coord, &city, &street_name, &conf); err != nil {
			return []Location{}, err
		}

		// parse stupid coordinates
		var KORDIKYRPÄ [][]float64

		coord = strings.Replace(coord, "(", "[", -1)
		coord = strings.Replace(coord, ")", "]", -1)

		if err := json.Unmarshal([]byte(coord), &KORDIKYRPÄ); err != nil {
			fmt.Println(coord)
			return []Location{}, err
		}

		coordinate := KORDIKYRPÄ[0]

		locations = append(locations, Location{Address{Street: street_name, City: city, Confidence: conf*100}, Coordinate{Latitude: coordinate[1], Longitude: coordinate[0], System: "WGS84"}})
	}

	if err := rows.Err(); err != nil {
		return []Location{}, err
	}

	return locations, nil
}

func ResolveLocation(config Config, location Location) (Location, error) {
	if IsMissingCoordinate(location) {
		if location.Address.Street != "" {
			locs, err := GetFuzzyAddress(config, location.Address, 1)
			if err != nil {
				return Location{}, err
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
