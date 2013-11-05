package main

import (
	"database/sql"
	"errors"
	"fmt"
)

type CHNode struct {
	Id int
	Coord
}

func ParseCoord(value string) (Coord, error) {
	var lat, long float64
	n_parsed, err := fmt.Sscanf(value, "(%f,%f)", &lat, &long)
	if n_parsed != 2 || err != nil {
		return Coord{}, err
	}

	return Coord{lat, long}, nil
}

func CorrectPoint(config Config, point Coord, country string) (CHNode, error) {
	db, err := sql.Open("postgres", config.String())
	if err != nil {
		debug.Println("ARGh")
		return CHNode{}, err
	}
	defer db.Close()

	var coord []byte
	var id int

	q := fmt.Sprintf("SELECT id, coord FROM %s_nodes ORDER BY coord <-> point ('%.5f, %.5f') LIMIT 1",
		country, point[0], point[1])
	err = db.QueryRow(q).Scan(&id, &coord)

	switch {
	case err == sql.ErrNoRows:
		return CHNode{}, errors.New("No points found")
	case err != nil:
		return CHNode{}, err
	}

	c, err := ParseCoord(string(coord))
	if err != nil {
		return CHNode{}, err
	}

	return CHNode{id, c}, nil
}
