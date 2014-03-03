package geo

import (
	"fmt"

	"github.com/hoisie/redis"
)

type Config struct {
	Port             int
	DbUser           string
	DbName           string
	AllowedCountries map[string]bool
}

type Coord []float64

type debugging bool

type Coordinate struct {
	Latitude  float64
	Longitude float64
	System    string
}

type Location struct {
	Address    Address
	Coordinate Coordinate
}

type Address struct {
	AdditionalInfo  string
	ApartmentLetter string
	ApartmentNumber int
	City            string
	Confidence      float64
	Country         string
	HouseNumber     int
	PostalCode      string
	Region          string
	Street          string
}

type Path struct {
	Length int   `json:"length"`
	Nodes  []int `json:"nodes"`
}

type Edge struct {
	Source Location
	Target Location
}

type NodeEdge struct {
	Source int `json:"source"`
	Target int `json:"target"`
}

type PathsInput struct {
	SpeedProfile int
	Edges        []Edge
}

type CoordinatePath struct {
	Distance int     `json:"distance"`
	Time     int     `json:"time"`
	Coords   []Coord `json:"coords"`
}

type Matrix struct {
	Nodes []int `json:"sources"`
}

type Geo struct {
	Debug  debugging
	client redis.Client
	Config
}

func (config *Config) String() string {
	s := fmt.Sprintf("sslmode=disable user=%s dbname=%s", config.DbUser, config.DbName)
	return s
}
