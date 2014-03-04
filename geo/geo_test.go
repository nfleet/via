package geo

import "github.com/nfleet/via/geodb"

var (
	config, _ = LoadConfig("../development.json")
	geoDB     = geodb.GeoPostgresDB{config}
	test_geo  = NewGeo(true, geoDB)
)
