package postgeodb

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/nfleet/via/geotypes"
)

// Returns the closest point of point in the graph database, returns the node ID and its coordinates.
func (g GeoPostgresDB) QueryClosestPoint(point geotypes.Coord, country string) (geotypes.CHNode, error) {
	db := g.db

	var (
		lat, long float64
		id        int
	)

	q := fmt.Sprintf("SELECT id, coord[0], coord[1] FROM %s ORDER BY coord <-> point ('%.5f, %.5f') LIMIT 1",
		table_names[country], point[0], point[1])
	err := db.QueryRow(q).Scan(&id, &lat, &long)

	switch {
	case err == sql.ErrNoRows:
		return geotypes.CHNode{}, errors.New("No points found")
	case err != nil:
		return geotypes.CHNode{}, err
	default:
		return geotypes.CHNode{id, geotypes.Coord{lat, long}}, nil
	}

}

// Many-to-many inverse of QueryClosestPoint, returns the coordinates corresponding to the graph nodes. Used
// mostly for rendering paths.
func (g GeoPostgresDB) QueryCoordinates(nodes []int, country string) ([]geotypes.Coord, error) {
	db := g.db

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

	var coords []geotypes.Coord
	rows, err := db.Query(q)
	if err != nil {
		return []geotypes.Coord{}, err
	}
	for rows.Next() {
		var lat, long float64
		if err := rows.Scan(&lat, &long); err != nil {
			return []geotypes.Coord{}, err
		}
		coords = append(coords, geotypes.Coord{lat, long})
	}

	if err := rows.Err(); err != nil {
		return []geotypes.Coord{}, err
	}

	return coords, nil
}
