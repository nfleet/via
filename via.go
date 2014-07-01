package main

import (
	"encoding/json"
	"github.com/ane/redis"
	"io/ioutil"
)

type Via struct {
	Debug   Debugging
	Expiry  int
	Client  redis.Client
	DataDir string
}

type ViaConfig struct {
	Host             string
	Port             int
	SslMode          string
	DataDir          string
	RedisAddr        string
	RedisPass        string
	AllowedCountries map[string]bool
}

func LoadConfig(file string) (ViaConfig, error) {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return ViaConfig{}, err
	}

	var config ViaConfig
	if err := json.Unmarshal(contents, &config); err != nil {
		return ViaConfig{}, err
	}
	return config, nil
}

func NewVia(debug bool, client redis.Client, expiry int, dataDir string) *Via {
	return &Via{
		Debug:   Debugging(debug),
		Expiry:  expiry,
		Client:  client,
		DataDir: dataDir,
	}
}
