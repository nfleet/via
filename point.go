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

func CorrectPoint(config Config, point Coord, country string) (CHNode, error) {
	db, err := sql.Open("postgres", config.String())
	if err != nil {
		return CHNode{}, err
	}
	defer db.Close()

	var (
		lat, long float64
		id        int
	)

	q := fmt.Sprintf("SELECT id, coord[0], coord[1] FROM %s ORDER BY coord <-> point ('%.5f, %.5f') LIMIT 1",
		table_names[country], point[0], point[1])
	err = db.QueryRow(q).Scan(&id, &lat, &long)

	switch {
	case err == sql.ErrNoRows:
		return CHNode{}, errors.New("No points found")
	case err != nil:
		return CHNode{}, err
	default:
		return CHNode{id, Coord{lat, long}}, nil
	}
}
