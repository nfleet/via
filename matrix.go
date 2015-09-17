package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/nfleet/via/ch"
)

// Computes a matrix hash. This should be launched in a goroutine, not in the main thread.
func (v *Via) ComputeMatrix(nodes []int, country string, speedProfile int) (map[string][]int, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	v.Debug.Printf("entering ComputeMatrix for hash, memory used: %d mb.", memStats.Alloc/1e6)
	empty := map[string][]int{}
	t0 := time.Now()

	matrixData := struct {
		Sources []int `json:"sources"`
	}{nodes}
	jsonData, err := json.Marshal(matrixData)
	if err != nil {
		panic("nodes to json error:" + err.Error())
	}

	v.Debug.Println("got country", string(country), "with profile", speedProfile)

	res := ch.Calc_dm(string(jsonData), string(country), int(speedProfile), v.DataDir)

	var matrix map[string][]int
	if err := json.NewDecoder(strings.NewReader(res)).Decode(&matrix); err != nil {
		v.Debug.Println("failed to parse CH results")
		return empty, fmt.Errorf("Failed to parse CH results: %s", err.Error())
	}

	t1 := time.Since(t0)
	v.Debug.Println("calculated matrix in", t1)

	runtime.GC()
	runtime.ReadMemStats(&memStats)
	v.Debug.Printf("Computation completed with memory usage still at %d mb.\n", memStats.Alloc/1e6)

	return matrix, nil
}
