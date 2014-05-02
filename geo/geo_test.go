package geo

import (
	"testing"

	"github.com/ane/redis"
	"github.com/nfleet/via/postgeodb"
)

var (
	config, _  = LoadConfig("../development.json")
	geoDB, err = postgeodb.NewGeoPostgresDB(config)
	test_geo   = NewGeo(true, geoDB, redis.Client{Addr: config.RedisAddr, Password: config.RedisPass}, 3600, config.DataDir)
)

func TestGeo(t *testing.T) {
	_, err := LoadConfig("../development.json")
	if err != nil {
		t.Fatal(err)
	}
}
