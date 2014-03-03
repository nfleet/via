package geo

import (
	"encoding/json"
	"io/ioutil"
)

func LoadConfig(file string) (Config, error) {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := json.Unmarshal(contents, &config); err != nil {
		return Config{}, err
	}
	return config, nil
}

func NewGeo(debug bool, dbUser, dbName string, port int, allowedCountries map[string]bool) *Geo {
	g := new(Geo)
	g.Debug = debug
	g.Config.AllowedCountries = allowedCountries
	g.Config.DbName = dbName
	g.Config.DbUser = dbUser
	g.Config.Port = port
	return g
}
