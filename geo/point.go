package geo

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type CHNode struct {
	Id int
	Coord
}

func (g *Geo) CorrectPoint(point Coord, country string) (CHNode, error) {
	db, _ := sql.Open("postgres", g.Config.String())
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

func (g *Geo) GetCoordinates(country string, nodes []int) ([]Coord, error) {
	db, _ := sql.Open("postgres", g.Config.String())
	if err != nil {
		return CHNode{}, err
	}
	defer db.Close()

	query := `
SELECT n.coord[0], n.coord[1]
FROM (
	SELECT *, arr[rn] AS node_id
	FROM (
		SELECT *, generate_subscripts(arr, 1) AS rn
		FROM (
			SELECT ARRAY%s AS arr
		) x
	) y
) z
JOIN %s_nodes n ON n.id = z.node_id
ORDER BY z.arr, z.rn;`

	values := strings.Replace(fmt.Sprintf("%v", nodes), " ", ",", -1)
	q := fmt.Sprintf(query, values, country)

	var coords []Coord
	rows, err := db.Query(q)
	if err != nil {
		return []Coord{}, err
	}
	for rows.Next() {
		var lat, long float64
		if err := rows.Scan(&lat, &long); err != nil {
			return []Coord{}, err
		}
		coords = append(coords, Coord{lat, long})
	}

	if err := rows.Err(); err != nil {
		return []Coord{}, err
	}

	return coords, nil
}
