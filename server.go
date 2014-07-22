package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/ane/redis"
	"github.com/hoisie/web"

	// register this postgres driver with the SQL module
	_ "github.com/lib/pq"
	_ "net/http/pprof"
)

type (
	Server struct {
		Via              *Via
		AllowedCountries map[string]bool
		Host             string
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

	var config ViaConfig
	var configFile string

	args := flag.Args()
	if len(args) < 1 {
		configFile = "production.json"
	} else {
		configFile = args[0]
	}

	log.Print("loading config from " + configFile + "... ")
	config, err := LoadConfig(configFile)
	if err != nil {
		log.Printf("failed: %s\n", configFile, err.Error())
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
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGABRT)
	go func() {
		for sig := range c {
			log.Printf("received %v, exiting...", sig)
			os.Exit(1)
		}
	}()

	procs := runtime.NumCPU()
	runtime.GOMAXPROCS(procs)

	log.Printf("starting server, running on %d cores...", procs)

	via := NewVia(Debug, client, expiry, config.DataDir)
	server := Server{Via: via, Host: config.Host, Port: config.Port, AllowedCountries: config.AllowedCountries}

	// Basic
	web.Get("/", Splash)
	web.Get("/status", server.GetServerStatus)

	// Dmatrix
	web.Get("/matrix/(.*)/result", server.GetMatrixResult)
	web.Get("/matrix/(.*)$", server.GetMatrix)
	web.Post("/matrix/", server.PostMatrix)

	// Path
	web.Post("/paths", server.PostPaths)

	web.Match("OPTIONS", "/(.*)", Options)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	web.Run(fmt.Sprintf("%s:%d", config.Host, config.Port))
}
