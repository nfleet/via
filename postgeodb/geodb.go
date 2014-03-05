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
	Config geotypes.Config
}

// Returns the status of the server, tests the connection using
// the Ping method.
func (g GeoPostgresDB) QueryStatus() error {
	db, _ := sql.Open("postgres", g.Config.String())
	defer db.Close()

	err := db.Ping()

	if err != nil {
		return err
	}

	return nil
}
