package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"

	"github.com/hoisie/web"
	"github.com/nfleet/via/geotypes"
)

var allowedSpeeds = []int{40, 60, 80, 100, 120}

func contains(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Starts a computation, validates the matrix in POST.
// If matrix data is missing, returns 400 Bad Request.
// If on the other hand matrix is data is not missing,
// but makes no sense, it returns 422 Unprocessable Entity.
func (server *Server) PostMatrix(ctx *web.Context) {
	defer runtime.GC()
	var hash string
	var computed bool

	// Parse params
	var paramBlob struct {
		Matrix       []int   `json:"matrix"`
		Country      string  `json:"country"`
		SpeedProfile float64 `json:"speed_profile"`
		RatiosHash   string  `json:"ratios"`
	}
	if err := json.NewDecoder(ctx.Request.Body).Decode(&paramBlob); err != nil {
		ctx.Abort(400, err.Error())
		return
	}

	data := paramBlob.Matrix
	country := strings.ToLower(paramBlob.Country)
	sp := int(paramBlob.SpeedProfile)

	ok := len(data) > 0 && country != "" && sp > 0
	if ok {
		// Sanitize speed profile.
		if !contains(sp, allowedSpeeds) {
			msg := fmt.Sprintf("speed profile '%d' makes no sense, must be one of %s", sp, fmt.Sprint(allowedSpeeds))
			ctx.Abort(422, msg)
			return
		}
		// Sanitize country.
		if _, ok := server.AllowedCountries[country]; !ok {
			countries := ""
			for k := range server.AllowedCountries {
				countries += k + " "
			}
			ctx.Abort(422, "country "+country+" not allowed, must be one of: "+countries)
			return
		}

		hash, computed = server.Via.CreateMatrixComputation(data, paramBlob.RatiosHash, country, sp)
		if !computed {
			go server.Via.ComputeMatrix(hash)
		}

	} else {
		body, _ := ioutil.ReadAll(ctx.Request.Body)
		ctx.Abort(400, "Missing or invalid matrix data, speed profile, or country. You sent: "+string(body))
		return
	}

	loc := fmt.Sprintf("/matrix/%s", hash)
	ctx.Redirect(202, loc)
}

type Result struct {
	Progress     string           `json:"progress"`
	Matrix       map[string][]int `json:"matrix"`
	SpeedProfile int              `json:"speed_profile"`
	RatiosHash   string           `json:"ratios_hash"`
}

// Returns a computation from the server as identified by the resource parameter
// in GET.
func (server *Server) GetMatrix(ctx *web.Context, resource string) string {
	defer runtime.GC()
	progress, err := server.Via.GetMatrixComputationProgress(resource)
	if err != nil {
		ctx.Abort(410, err.Error())
		return ""
	}

	if progress == "complete" {
		url := fmt.Sprintf("/matrix/%s/result", resource)
		server.Via.Debug.Println("redirect ->", url)
		ctx.Redirect(303, url)
		return ""
	}
	ctx.ContentType("json")
	ctx.WriteHeader(202)
	json.NewEncoder(ctx.ResponseWriter).Encode(struct {
		Progress string `json:"progress"`
	}{progress})
	return ""
}

func (server *Server) GetMatrixResult(ctx *web.Context, resource string) string {
	defer runtime.GC()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	server.Via.Debug.Printf("getting matrix result, memory used %d MB.\n", memStats.Alloc/1e6)

	if ex, _ := server.Via.Client.Exists(resource); !ex {
		ctx.WriteHeader(410)
		return fmt.Sprintf("Matrix for hash %s has expired. POST again.", resource)
	}

	progress, err := server.Via.GetMatrixComputationProgress(resource)
	if err != nil {
		ctx.WriteHeader(500)
		return fmt.Sprintf("Failed to retrieve computation progress from redis: %s", err.Error())
	}

	if progress != "complete" {
		ctx.WriteHeader(403)
		return "Computation is not ready yet."
	}

	orig := resource
	if exists, _ := server.Via.Client.Hexists(resource, "see"); exists {
		server.Via.Debug.Println("Got proxy result")
		pointer, err := server.Via.Client.Hget(resource, "see")
		if err != nil {
			server.Via.Debug.Println("Proxy result invalid.")
		}
		resource = string(pointer)
	}

	data, err := server.Via.Client.Hget(resource, "result")
	if err != nil {
		ctx.WriteHeader(500)
		return "Redis error: " + err.Error()
	}

	ratios, err := server.Via.Client.Hget(resource, "ratiosHash")
	if err != nil {
		ctx.WriteHeader(500)
		return "Redis error: " + err.Error()
	}

	sp, err := server.Via.Client.Hget(resource, "speed_profile")
	if err != nil {
		ctx.WriteHeader(500)
		return "Redis error: " + err.Error()
	}

	_, err = server.Via.Client.Expire(resource, int64(server.Via.Expiry))
	if err != nil {
		ctx.WriteHeader(500)
		return fmt.Sprintf("Redis error: failed to reset expiry on key %s: %s", resource, err.Error())
	}

	if orig != resource {
		_, err = server.Via.Client.Expire(orig, int64(server.Via.Expiry))
		if err != nil {
			ctx.WriteHeader(500)
			return fmt.Sprintf("Redis error: failed to reset expiry on key %s: %s", resource, err.Error())
		}
	}

	speedProfile, _ := binary.Varint(sp)

	if data != nil {
		runtime.ReadMemStats(&memStats)
		server.Via.Debug.Printf("before decoding: %d mb", memStats.Alloc/1e6)
		var mat map[string][]int
		if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&mat); err != nil {
			ctx.WriteHeader(500)
			return "Failed to parse matrix from Redis: " + err.Error()
		}
		data = []byte{}

		result := Result{
			Progress:     "complete",
			Matrix:       mat,
			RatiosHash:   string(ratios),
			SpeedProfile: int(speedProfile),
		}

		runtime.ReadMemStats(&memStats)
		server.Via.Debug.Printf("before writing results: %d mb", memStats.Alloc/1e6)

		ctx.ContentType("json")
		if err := json.NewEncoder(ctx.ResponseWriter).Encode(result); err != nil {
			fmt.Println("aaguh")
		}
		result = Result{}

		runtime.ReadMemStats(&memStats)
		server.Via.Debug.Printf("before nil assignments: %d mb", memStats.Alloc/1e6)

	}

	runtime.GC()
	runtime.ReadMemStats(&memStats)
	server.Via.Debug.Printf("exiting matrix result, memory used %d MB.\n", memStats.Alloc/1e6)
	return ""
}

func (server *Server) GetServerStatus(ctx *web.Context) string {
	if _, err := server.Via.Client.Ping(); err != nil {
		return fmt.Sprintf("ERROR: Redis responded with %s", err.Error())
	}

	return "OK"
}

func (server *Server) PostPaths(ctx *web.Context) string {
	var input struct {
		Paths        []geotypes.NodeEdge
		Country      string
		SpeedProfile int
	}

	var (
		computed []geotypes.Path
	)

	if err := json.NewDecoder(ctx.Request.Body).Decode(&input); err != nil {
		content, _ := ioutil.ReadAll(ctx.Request.Body)
		ctx.Abort(400, "Couldn't parse JSON: "+err.Error()+" in '"+string(content)+"'")
		return ""
	} else {
		var err error
		computed, err = server.Via.CalculatePaths(input.Paths, input.Country, input.SpeedProfile)
		if err != nil {
			ctx.Abort(422, "Couldn't resolve addresses: "+err.Error())
			return ""
		}
	}

	res, err := json.Marshal(computed)
	if err != nil {
		ctx.Abort(500, "Couldn't serialize paths: "+err.Error())
		return ""
	}

	ctx.Header().Set("Access-Control-Allow-Origin", "*")
	ctx.ContentType("application/json")
	return string(res)
}
