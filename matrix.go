package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/nfleet/via/ch"
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
func (v *Via) CreateMatrixComputation(matrix []int, ratiosHash, country string, speed_profile int) (string, bool) {
	c := v.Client
	matrixHash := CreateMatrixHash(fmt.Sprint(matrix), country, speed_profile)
	// check if computation exists
	if exists, _ := c.Exists(matrixHash); exists {
		// make this one just point to that node, perturb the hash for uniqueness
		newHash := CreateMatrixHash(fmt.Sprint(matrixHash)+string(time.Now().UnixNano()), country, speed_profile)
		c.Hset(newHash, "see", []byte(matrixHash))
		ttl, _ := c.Ttl(matrixHash)
		c.Expire(newHash, ttl)
		v.Debug.Printf("Created proxy resource %s (expires in %d sec)", newHash, ttl)
		return newHash, true
	}

	c.Hset(matrixHash, "progress", []byte("initializing"))
	c.Hset(matrixHash, "result", []byte("0"))
	c.Hset(matrixHash, "country", []byte(country))
	c.Hset(matrixHash, "ratiosHash", []byte(ratiosHash))

	bif := make([]byte, 8)
	n := binary.PutVarint(bif, int64(speed_profile))
	c.Hset(matrixHash, "speed_profile", bif[:n])

	// convert data into binary things
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(matrix); err != nil {
		v.Debug.Println("encode error", err)
	}

	c.Hset(matrixHash, "data", buf.Bytes())
	c.Expire(matrixHash, int64(v.Expiry))
	v.Debug.Printf("Created computation resource %s (ttl: %d sec)", matrixHash, v.Expiry)

	return matrixHash, false
}

// Returns computation progress for the matrix identified by matrixHash.
func (v *Via) GetMatrixComputationProgress(matrixHash string) (string, error) {
	c := v.Client
	workingHash := matrixHash
	if exists, _ := c.Exists(workingHash); !exists {
		return "", fmt.Errorf("No matrix found for hash %s. It is likely the resource has expired. I recommend posting a new matrix.")
	}
	if exists, _ := c.Hexists(workingHash, "see"); exists {
		v.Debug.Println("Resolving proxy")
		pointer, _ := c.Hget(workingHash, "see")
		workingHash = string(pointer)
	}
	progress, err := c.Hget(workingHash, "progress")
	if err != nil {
		return "", fmt.Errorf("The requested %s resource is not a matrix resource.", matrixHash)
	}
	return string(progress), nil
}

// Computes a matrix hash. This should be launched in a goroutine, not in the main thread.
func (v *Via) ComputeMatrix(matrixHash string) {
	var nodes []int
	rc := v.Client
	t0 := time.Now()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	v.Debug.Printf("entering ComputeMatrix for hash %s. memory used: %d mb.", matrixHash, memStats.Alloc/1e6)

	set_status := func(status string) {
		v.Debug.Println(status)
		rc.Hset(matrixHash, "progress", []byte(status))
	}

	// error handling
	defer func() {
		if r := recover(); r != nil {
			v.Debug.Println("Recovered in Compute.", r)
			set_status("error")
		}
	}()

	data, err := rc.Hget(matrixHash, "data")
	if err != nil {
		panic("redis problem: " + err.Error())
	}

	// unpack from binary data
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(&nodes); err != nil {
		panic("decode error: " + err.Error())
	}

	buf.Reset()

	country, _ := rc.Hget(matrixHash, "country")

	matrixData := struct {
		Sources []int `json:"sources"`
	}{nodes}
	json_data, err := json.Marshal(matrixData)
	if err != nil {
		panic("nodes to json error:" + err.Error())
	}

	speed_profile_raw, _ := rc.Hget(matrixHash, "speed_profile")
	speed_profile, _ := binary.Varint(speed_profile_raw)
	v.Debug.Println("got country", string(country), "with profile", speed_profile)

	set_status("computing")
	// todo: use nodes
	var res string

	res = ch.Calc_dm(string(json_data), string(country), int(speed_profile), v.DataDir)

	var matrix map[string][]int
	if err := json.NewDecoder(strings.NewReader(res)).Decode(&matrix); err != nil {
		set_status("error")
		v.Debug.Println("failed to parse CH results")
	}
	res = ""

	buf.Reset()
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(matrix); err != nil {
		set_status("error")
		v.Debug.Println("failed to encode CH matrix into a byte form")
	}
	rc.Hset(matrixHash, "result", buf.Bytes())
	buf.Reset()

	t1 := time.Since(t0)
	v.Debug.Println("calculated matrix in", t1)

	set_status("complete")

	runtime.GC()
	runtime.ReadMemStats(&memStats)
	v.Debug.Printf("Computation completed with memory usage still at %d mb.\n", memStats.Alloc/1e6)
}
