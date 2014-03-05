package geo

import "github.com/nfleet/via/postgeodb"

var (
	config, _ = LoadConfig("../development.json")
	geoDB     = postgeodb.GeoPostgresDB{config}
	test_geo  = NewGeo(true, geoDB, 3600)
)
