package geo

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nfleet/via/ch"
	"github.com/nfleet/via/geotypes"
)

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
// Returns the hash to be used with Redis and true whether a proxy resource was created,
// false if the resource is new.
func (g *Geo) CreateMatrixComputation(matrix []geotypes.Coord, country string, speed_profile int) (string, bool) {
	c := g.Client
	matrixHash := CreateMatrixHash(fmt.Sprint(matrix), country, speed_profile)
	// check if computation exists
	if exists, _ := c.Exists(matrixHash); exists {
		// make this one just point to that node, perturb the hash for uniqueness
		newHash := CreateMatrixHash(fmt.Sprint(matrixHash)+string(time.Now().UnixNano()), country, speed_profile)
		c.Hset(newHash, "see", []byte(matrixHash))
		ttl, _ := c.Ttl(matrixHash)
		c.Expire(newHash, ttl)
		g.Debug.Printf("Created proxy resource %s (expires in %d sec)", newHash, ttl)
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
	if err := enc.Encode(matrix); err != nil {
		g.Debug.Println("encode error", err)
	}

	c.Hset(matrixHash, "data", buf.Bytes())
	c.Expire(matrixHash, int64(g.Expiry))
	g.Debug.Printf("Created computation resource %s (ttl: %d sec)", matrixHash, g.Expiry)

	return matrixHash, false
}

// Returns computation progress for the matrix identified by matrixHash.
func (g *Geo) GetMatrixComputationProgress(matrixHash string) (string, error) {
	c := g.Client
	workingHash := matrixHash
	if exists, _ := c.Exists(workingHash); !exists {
		return "", errors.New("no matrix found for hash " + workingHash)
	}
	if exists, _ := c.Hexists(workingHash, "see"); exists {
		g.Debug.Println("Resolving proxy")
		pointer, _ := c.Hget(workingHash, "see")
		workingHash = string(pointer)
	}
	progress, _ := c.Hget(workingHash, "progress")
	return string(progress), nil
}

// Computes a matrix hash. This should be launched in a goroutine, not in the main thread.
func (g *Geo) ComputeMatrix(matrixHash string) {
	var coords []geotypes.Coord
	rc := g.Client
	t0 := time.Now()

	set_status := func(status string) {
		g.Debug.Println(status)
		rc.Hset(matrixHash, "progress", []byte(status))
	}

	// error handling
	defer func() {
		if r := recover(); r != nil {
			g.Debug.Println("Recovered in Compute.", r)
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

	if err := dec.Decode(&coords); err != nil {
		panic("decode error: " + err.Error())
	}

	country, _ := rc.Hget(matrixHash, "country")

	nodes, err := g.RunGeoindexer(string(country), coords, 4)
	if err != nil {
		panic("Geoindex error:" + err.Error())
	}

	set_status("computing")
	json_data, err := MatrixToJson(nodes)
	if err != nil {
		panic("nodes to json error:" + err.Error())
	}

	speed_profile_raw, _ := rc.Hget(matrixHash, "speed_profile")
	speed_profile, _ := binary.Varint(speed_profile_raw)
	g.Debug.Println("got country", string(country), "with profile", speed_profile)

	// todo: use nodes
	var res string
	res = ch.Calc_dm(string(json_data), string(country), int(speed_profile))
	// something weird might happen with rapidjson serialization - fix this
	// case 1: missing }
	if strings.Index(res, "}") == -1 {
		res = res + "}"
		g.Debug.Println("brace missing")
	} else if strings.Index(res, "}") != len(res)-1 {
		braceIndex := strings.Index(res, "}")
		junk := res[braceIndex+1:]
		res = res[:braceIndex+1]
		g.Debug.Printf("Stripped extra data: %s", junk)
	}

	rc.Hset(matrixHash, "result", []byte(res))
	res = ""

	t1 := time.Since(t0)
	g.Debug.Println("calculated matrix in", t1)

	set_status("complete")
}
