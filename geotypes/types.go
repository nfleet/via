package geotypes

import "fmt"

type Config struct {
	Port             int
	DbUser           string
	DbName           string
	DbHost           string
	DbPass           string
	AllowedCountries map[string]bool
}

type CHNode struct {
	Id int
	Coord
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

// GeoDB implements a database abstraction layer that hides all
// functionality related to databases.
type GeoDB interface {
	// Returns the point closest to the coordinates
	// in the road network, based on geoindexed data.
	QueryClosestPoint(point Coord, country string) (CHNode, error)

	// Transforms the nodes from the graph nodes,
	// which are indexed by integers, and returns the corresponding
	// coordinates.
	QueryCoordinates(nodes []int, country string) ([]Coord, error)

	// Gets an address using fuzzy search. Uses the street and city at
	// the moment.
	QueryFuzzyAddress(address Address, count int) ([]Location, error)

	// Calculates the distance between edges. If the array nodes is indexed
	// with the variable i, then edge pairs are formed like thus (i, i+1), (i+1, i+2)
	// and so on.
	QueryDistance(nodes []int, country string) (int, error)

	// Returns the server status, nil if everything works fine.
	QueryStatus() error
}

func (config *Config) String() string {
	s := fmt.Sprintf("sslmode=disable user=%s dbname=%s host=%s password=%s",
		config.DbUser, config.DbName, config.DbHost, config.DbPass)
	return s
}
