package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"syscall"

	"github.com/hoisie/web"
	"github.com/nfleet/via/geotypes"
)

var allowed_speeds = []int{40, 60, 80, 100, 120}

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
	var hash string
	var computed bool
	stuff, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Abort(400, err.Error())
		return
	}

	// Parse params
	var paramBlob struct {
		Matrix       []int   `json:"matrix"`
		Country      string  `json:"country"`
		SpeedProfile float64 `json:"speed_profile"`
		RatiosHash   string  `json:"ratios"`
	}
	if err := json.Unmarshal(stuff, &paramBlob); err != nil {
		ctx.Abort(400, err.Error())
		return
	}

	data := paramBlob.Matrix
	country := strings.ToLower(paramBlob.Country)
	sp := int(paramBlob.SpeedProfile)

	ok := len(data) > 0 && country != "" && sp > 0
	if ok {
		// Sanitize speed profile.
		if !contains(sp, allowed_speeds) {
			msg := fmt.Sprintf("speed profile '%d' makes no sense, must be one of %s", sp, fmt.Sprint(allowed_speeds))
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
		ctx.Abort(400, "Missing or invalid matrix data, speed profile, or country. You sent: "+string(stuff))
		return
	}

	loc := fmt.Sprintf("/matrix/%s", hash)
	ctx.Redirect(201, loc)
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
	progress, err := server.Via.GetMatrixComputationProgress(resource)
	if err != nil {
		ctx.Abort(500, err.Error())
		return ""
	}

	fmt.Println(progress)

	if progress == "complete" {
		url := fmt.Sprintf("/matrix/%s/result", resource)
		server.Via.Debug.Println("redirect ->", url)
		ctx.Redirect(303, url)
	} else {
		ctx.ContentType("json")
		result := Result{Progress: progress, Matrix: map[string][]int{}}
		msg, _ := json.Marshal(result)
		return string(msg)
	}

	return ""
}

func (server *Server) GetMatrixResult(ctx *web.Context, resource string) string {
	if ex, _ := server.Via.Client.Exists(resource); !ex {
		ctx.Abort(500, "Result expired. POST again.")
		return ""
	}

	progress, err := server.Via.GetMatrixComputationProgress(resource)
	if err != nil {
		ctx.Abort(403, err.Error())
	}

	if progress != "complete" {
		ctx.Abort(403, "Computation is not ready yet.")
		return ""
	}

	if exists, _ := server.Via.Client.Hexists(resource, "see"); exists {
		server.Via.Debug.Println("Got proxy result")
		pointer, _ := server.Via.Client.Hget(resource, "see")
		resource = string(pointer)
	}

	data, err := server.Via.Client.Hget(resource, "result")
	if err != nil {
		ctx.Abort(500, "Redis error: "+err.Error())
		return ""
	}

	ratios, err := server.Via.Client.Hget(resource, "ratiosHash")
	if err != nil {
		ctx.Abort(500, "Redis error: "+err.Error())
		return ""
	}

	sp, err := server.Via.Client.Hget(resource, "speed_profile")
	if err != nil {
		ctx.Abort(500, "Redis error: "+err.Error())
		return ""
	}

	speed_profile, _ := binary.Varint(sp)

	if data != nil {
		ctx.ContentType("json")
		var mat map[string][]int
		if err := json.Unmarshal(data, &mat); err != nil {
			ctx.Abort(500, "Could not parse json matrix from ch: "+err.Error())
			return ""
		}

		result := Result{Progress: "complete", Matrix: mat, RatiosHash: string(ratios), SpeedProfile: int(speed_profile)}
		str, _ := json.Marshal(result)
		return string(str)
	}

	return ""
}

func (server *Server) GetServerStatus(ctx *web.Context) string {
	s := syscall.Statfs_t{}
	syscall.Statfs("./", &s)
	space := (uint64(s.Bsize) * s.Bavail) / 1000000000
	err_disk := space < 2

	if err_disk != false {
		ctx.Abort(500, fmt.Sprintf("Disk space is below 2G!"))
		return ""
	}

	return "OK"
}

func (server *Server) PostPaths(ctx *web.Context) string {
	content, err := ioutil.ReadAll(ctx.Request.Body)

	var input struct {
		Paths        []geotypes.NodeEdge
		Country      string
		SpeedProfile int
	}

	var (
		computed []geotypes.Path
	)

	if err := json.Unmarshal(content, &input); err != nil {
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
