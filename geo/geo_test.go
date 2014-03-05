package geo

import (
	"testing"

	"github.com/nfleet/via/postgeodb"
)

var (
	config, _ = LoadConfig("../development.json")
	geoDB     = postgeodb.GeoPostgresDB{config}
	test_geo  = NewGeo(true, geoDB, 3600)
)

func TestGeo(t *testing.T) {
	_, err := LoadConfig("../development.json")
	if err != nil {
		t.Fatal(err)
	}
}
