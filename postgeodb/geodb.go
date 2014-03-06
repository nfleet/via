// postgeodb implements a PostgresSQL version of the GeoDB
// interface.
package postgeodb

import (
	"database/sql"

	_ "github.com/bmizerany/pq"
	"github.com/nfleet/via/geotypes"
)

var table_names = map[string]string{
	"finland": "finland_nodes",
	"germany": "germany_nodes",
}

type GeoPostgresDB struct {
	db     *sql.DB
	Config geotypes.Config
}

func NewGeoPostgresDB(config geotypes.Config) (*GeoPostgresDB, error) {
	geodb, err := sql.Open("postgres", config.String())

	if err != nil {
		return nil, err
	}

	g := GeoPostgresDB{geodb, config}

	return &g, nil
}

// Returns the status of the server, tests the connection using
// the Ping method.
func (g GeoPostgresDB) QueryStatus() error {
	err := g.db.Ping()

	if err != nil {
		return err
	}

	return nil
}
