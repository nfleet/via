package geodb

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func (g GeoPostgresDB) QueryDistance(nodes []int, country string) (int, error) {
	if len(nodes) < 2 {
		return 0, nil
	}

	db, _ := sql.Open("postgres", g.Config.String())
	defer db.Close()

	var edgePairs []string

	for i := 0; i < len(nodes)-1; i++ {
		edgeStart, edgeEnd := nodes[i], nodes[i+1]
		s := fmt.Sprintf("(%d,%d)", edgeStart, edgeEnd)
		edgePairs = append(edgePairs, s)
	}

	edges := strings.Join(edgePairs, ",")

	q := `select sum(dist) from (values%s) as t left join %s_speed_edges on column1=id1 and column2=id2`

	q = fmt.Sprintf(q, edges, country)

	var sum float64
	err := db.QueryRow(q).Scan(&sum)
	switch {
	case err == sql.ErrNoRows:
		return 0, errors.New("No distance found. Check points exist.")
	case err != nil:
		return 0, err
	default:
		return int(sum), nil
	}

}
