package geotypes

import "fmt"

type Config struct {
	Host             string
	Port             int
	SslMode          string
	DbUser           string
	DbName           string
	DbHost           string
	DbPort           int
	DbPass           string
	DataDir          string
	RedisAddr        string
	RedisPass        string
	AllowedCountries map[string]bool
}

type CHNode struct {
	Id int
	Coord
}

type Coord []float64

type Debugging bool

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
	Resolution      int
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
	SameRoad bool    `json:"-""`
}

type Matrix struct {
	Nodes []int `json:"sources"`
}

type Ratio struct {
	Id1, Id2, Dist40, Dist60, Dist80, Dist100, Dist120 int
	Ratio, Distance                                    float64
	Oneway                                             bool
	Coord                                              Coord
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
	s := fmt.Sprintf("user=%s dbname=%s host=%s port=%d password=%s sslmode=%s",
		config.DbUser, config.DbName, config.DbHost, config.DbPort, config.DbPass, config.SslMode)
	return s
}
