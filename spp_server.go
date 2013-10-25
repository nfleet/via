package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hoisie/redis"
	"github.com/hoisie/web"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	// register this postgres driver with the SQL module
	_ "github.com/bmizerany/pq"
)

type Config struct {
	Port   int
	DbUser string
	DbName string
}

func (config *Config) String() string {
	s := fmt.Sprintf("sslmode=disable user=%s dbname=%s", config.DbUser, config.DbName)
	return s
}

type Server struct {
	client redis.Client
	Config
}

type Coord []float64

type debugging bool

const debug debugging = true
const expiry int64 = 3600

func CreateMatrixHash(matrixData, country string, speed_profile int) string {
	hash := sha256.New()
	data := fmt.Sprintf("%s%s%d", matrixData, country, speed_profile)
	hash.Write([]byte(data))
	md := hash.Sum(nil)
	return hex.EncodeToString(md)
}

// Creates a computation ID by getting an ID for it in Redis.
// Uses matrix to create a hash for the data, which is the raw string data that
// has not been parsed into JSON. Stores parsedData into Redis in binary format.
func (server *Server) CreateComputation(matrix []Coord, country string, speed_profile int) (string, bool) {
	c := server.client
	matrixHash := CreateMatrixHash(fmt.Sprint(matrix), country, speed_profile)
	// check if computation exists
	if exists, _ := c.Exists(matrixHash); exists {
		// make this one just point to that node, perturb the hash for uniqueness
		newHash := CreateMatrixHash(fmt.Sprint(matrixHash)+string(time.Now().UnixNano()), country, speed_profile)
		c.Hset(newHash, "see", []byte(matrixHash))
		ttl, _ := c.Ttl(matrixHash)
		c.Expire(newHash, ttl)
		debug.Printf("Created proxy resource %s (expires in %d sec)", newHash, ttl)
		return newHash, true
	}

	c.Hset(matrixHash, "progress", []byte("initializing"))
	c.Hset(matrixHash, "result", []byte("0"))
	c.Hset(matrixHash, "country", []byte(country))

	bif := make([]byte, 8)
	n := binary.PutVarint(bif, int64(speed_profile))
	c.Hset(matrixHash, "speed_profile", bif[:n])

	// convert data into binary things
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(matrix)
	if err != nil {
		fmt.Println("encode error", err)
	}

	c.Hset(matrixHash, "data", buf.Bytes())
	c.Expire(matrixHash, 3600)
	debug.Printf("Created computation resource %s (ttl: %d sec)", matrixHash, expiry)

	return matrixHash, false
}

// Returns computation progress for the matrix identified by matrixHash.
func (server *Server) GetComputationProgress(matrixHash string) (string, error) {
	c := server.client
	workingHash := matrixHash
	if exists, _ := c.Exists(workingHash); !exists {
		return "", errors.New("no matrix found for hash " + workingHash)
	}
	if exists, _ := c.Hexists(workingHash, "see"); exists {
		debug.Println("Resolving proxy")
		pointer, _ := c.Hget(workingHash, "see")
		workingHash = string(pointer)
	}
	progress, _ := c.Hget(workingHash, "progress")
	return string(progress), nil
}

// Computes a matrix hash. This should be launched in a goroutine, not in the main thread.
func (server *Server) Compute(matrixHash string) {
	var coords []Coord
	rc := server.client
	t0 := time.Now()

	set_status := func(status string) {
		debug.Println(status)
		rc.Hset(matrixHash, "progress", []byte(status))
	}

	// error handling
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in Compute.", r)
			set_status("error")
		}
	}()

	// start with Geoindexing
	set_status("Geoindexing")

	data, err := rc.Hget(matrixHash, "data")
	if err != nil {
		panic("redis problem: " + err.Error())
	}

	// unpack from binary data
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	err = dec.Decode(&coords)
	if err != nil {
		panic("decode error: " + err.Error())
	}

	country, _ := rc.Hget(matrixHash, "country")

	nodes, err := server.RunGeoindexer(string(country), coords, 4)
	if err != nil {
		panic("Geoindex error:" + err.Error())
	}

	set_status("computing")
	json_data, err := dump_matrix_to_json(nodes)
	if err != nil {
		panic("nodes to json error:" + err.Error())
	}

	speed_profile_raw, _ := rc.Hget(matrixHash, "speed_profile")
	speed_profile, _ := binary.Varint(speed_profile_raw)
	debug.Println("got country", string(country), "with profile", speed_profile)

	// todo: use nodes
	var res string
	res = Calc(string(json_data), string(country), int(speed_profile))
	// something weird might happen with rapidjson serialization - fix this
	// case 1: missing }
	if strings.Index(res, "}") == -1 {
		res = res + "}"
		debug.Println("brace missing")
	} else if strings.Index(res, "}") != len(res)-1 {
		braceIndex := strings.Index(res, "}")
		junk := res[braceIndex+1:]
		res = res[:braceIndex+1]
		debug.Printf("Stripped extra data: %s", junk)
	}

	rc.Hset(matrixHash, "result", []byte(res))
	res = ""

	t1 := time.Since(t0)
	debug.Println("calculated matrix in", t1)

	set_status("complete")
}

func (s *Server) Splash(ctx *web.Context) {
	ctx.ContentType("image/jpeg")
	http.ServeFile(ctx, ctx.Request, "./splash.jpg")
}

func loadConfig(file string) Config {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		panic("config file loading failed: " + err.Error())
	}

	var config Config
	err = json.Unmarshal(contents, &config)
	if err != nil {
		panic("config file loading failed: " + err.Error())
	}
	return config
}

func main() {
	var config Config
	var configFile string
	var redis redis.Client

	args := os.Args
	if len(args) < 2 {
		configFile = "production.json"
	} else {
		configFile = os.Args[1]
	}

	debug.Println("loading config from " + configFile)
	config = loadConfig(configFile)

	server := Server{client: redis, Config: config}

	// Routes.
	web.Get("/", server.Splash)
	web.Get("/status", server.Status)
	web.Get("/spp/(.*)/result", server.GetResult)
	web.Get("/spp/(.*)$", server.Get)

	web.Post("/spp/", server.Begin)

	web.Run(fmt.Sprintf("127.0.0.1:%d", config.Port))
}
