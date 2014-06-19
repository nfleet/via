package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"

	"github.com/ane/redis"
	"github.com/hoisie/web"
	"github.com/nfleet/via/geo"
	"github.com/nfleet/via/geotypes"
	"github.com/nfleet/via/postgeodb"

	// register this postgres driver with the SQL module
	_ "github.com/lib/pq"
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
	var geoDB *postgeodb.GeoPostgresDB

	args := flag.Args()
	if len(args) < 1 {
		configFile = "production.json"
	} else {
		configFile = args[0]
	}

	log.Print("loading config from " + configFile + "... ")
	config, err := geo.LoadConfig(configFile)
	if err != nil {
		log.Printf("failed: %s\n", configFile, err.Error())
		return
	}

	log.Print("establishing database connection... ")
	geoDB, err = postgeodb.NewGeoPostgresDB(config)
	if err != nil {
		log.Println("error: " + err.Error())
		return
	}

	client := redis.Client{Addr: config.RedisAddr, Password: config.RedisPass}
	log.Printf("connecting to redis server at %s... ", client.Addr)
	if _, err := client.Ping(); err != nil {
		log.Println("error: ", err.Error())
		return
	}

	// Handle SIGINT and SIGKILL
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		for sig := range c {
			log.Printf("received %v, closing database connections and exiting...", sig)
			geoDB.DB.Close()
			os.Exit(1)
		}
	}()

	procs := runtime.NumCPU()
	runtime.GOMAXPROCS(procs)

	log.Printf("starting server, running on %d cores...", procs)

	geo := geo.NewGeo(Debug, geoDB, client, expiry, config.DataDir)
	server := Server{Geo: geo, Port: config.Port, AllowedCountries: config.AllowedCountries}

	// Basic
	web.Get("/", Splash)
	web.Get("/status", server.GetServerStatus)

	// Dmatrix
	web.Get("/dm/(.*)/result", server.GetMatrixResult)
	web.Get("/dm/(.*)$", server.GetMatrix)
	web.Post("/dm/", server.PostMatrix)
	web.Get("/spp/(.*)/result", server.GetMatrixResult)
	web.Get("/spp/(.*)$", server.GetMatrix)
	web.Post("/spp/", server.PostMatrix)

	// Path
	web.Post("/paths", server.PostCoordinatePaths)
	web.Post("/cpaths", server.PostCoordinatePaths)

	// Address/Coordinate
	web.Post("/resolve", server.PostResolve)

	web.Match("OPTIONS", "/(.*)", Options)

	web.Run(fmt.Sprintf("%s:%d", config.Host, config.Port))
}
