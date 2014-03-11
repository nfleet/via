package geo

import (
	"encoding/json"
	"io/ioutil"

	"github.com/hoisie/redis"
	_ "github.com/lib/pq"
	"github.com/nfleet/via/geotypes"
)

type debugging bool

type Geo struct {
	Debug  debugging
	Expiry int
	Client redis.Client
	DB     geotypes.GeoDB
}

func LoadConfig(file string) (geotypes.Config, error) {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return geotypes.Config{}, err
	}

	var config geotypes.Config
	if err := json.Unmarshal(contents, &config); err != nil {
		return geotypes.Config{}, err
	}
	return config, nil
}

func NewGeo(debug bool, db geotypes.GeoDB, expiry int) *Geo {
	return &Geo{
		DB:     db,
		Debug:  debugging(debug),
		Expiry: expiry,
	}
}
