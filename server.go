package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/hoisie/web"
	"github.com/nfleet/via/geo"
	"github.com/nfleet/via/geotypes"
	"github.com/nfleet/via/postgeodb"

	// register this postgres driver with the SQL module
	_ "github.com/bmizerany/pq"
)

type (
	Server struct {
		Geo              *geo.Geo
		AllowedCountries map[string]bool
		Port             int
	}
)

const (
	expiry int = 3600
)

var (
	Debug    bool
	Parallel bool
)

func Splash(ctx *web.Context) {
	ctx.ContentType("image/jpeg")
	http.ServeFile(ctx, ctx.Request, "./splash.jpg")
}

func parse_flags() {
	flag.BoolVar(&Debug, "debug", false, "toggle debugging on/off")
	flag.BoolVar(&Parallel, "par", false, "turn on parallel execution")
	flag.Parse()
}

func test_connection() {
	// redis
	fmt.Printf("loading services")

}

func Options(ctx *web.Context, route string) string {
	ctx.SetHeader("Access-Control-Allow-Origin", "*", false)
	ctx.SetHeader("Access-Control-Allow-Headers", "Authorization, Content-Type, If-None-Match", false)
	ctx.SetHeader("Access-control-Allow-Methods", "GET, PUT, POST, DELETE", false)
	ctx.ContentType("application/json")
	return "{}"
}

func main() {
	parse_flags()

	var config geotypes.Config
	var configFile string

	args := flag.Args()
	if len(args) < 1 {
		configFile = "production.json"
	} else {
		configFile = args[0]
	}

	fmt.Println("loading config from " + configFile)
	config, err := geo.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("configuration loading from %s failed: %s\n", configFile, err.Error())
		return
	}

	geoDB := postgeodb.GeoPostgresDB{Config: config}
	geo := geo.NewGeo(Debug, geoDB, expiry)

	server := Server{Geo: geo, Port: config.Port, AllowedCountries: config.AllowedCountries}

	// Basic
	web.Get("/", Splash)
	web.Get("/status", server.GetServerStatus)

	// Dmatrix
	web.Get("/dm/(.*)/result", server.GetMatrixResult)
	web.Get("/dm/(.*)$", server.GetMatrix)
	web.Post("/dm/", server.PostMatrix)
	web.Get("/dmatrix/(.*)/result", server.GetMatrixResult)
	web.Get("/dmatrix/(.*)$", server.GetMatrix)
	web.Post("/dmatrix/", server.PostMatrix)

	// Path
	web.Post("/paths", server.PostCoordinatePaths)
	web.Post("/cpaths", server.PostCoordinatePaths)

	// Address/Coordinate
	web.Post("/resolve", server.PostResolve)

	web.Match("OPTIONS", "/(.*)", Options)

	web.Run(fmt.Sprintf("127.0.0.1:%d", config.Port))
}
