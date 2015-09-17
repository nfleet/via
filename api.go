package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"

	"github.com/hoisie/web"
	viaErr "github.com/nfleet/via/error"
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

type Result struct {
	Progress     string           `json:"progress"`
	Matrix       map[string][]int `json:"matrix"`
	SpeedProfile int              `json:"speed_profile"`
}

// Starts a computation, validates the matrix in POST.
// If matrix data is missing, returns 400 Bad Request.
// If on the other hand matrix is data is not missing,
// but makes no sense, it returns 422 Unprocessable Entity.
func (server *Server) PostMatrix(ctx *web.Context) {
	defer runtime.GC()

	// Parse params
	var paramBlob struct {
		Matrix       []int   `json:"matrix"`
		Country      string  `json:"country"`
		SpeedProfile float64 `json:"speed_profile"`
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

		matrix, err := server.Via.ComputeMatrix(data, country, sp)
		if err != nil {
			viaErr.NewError(viaErr.ErrMatrixComputation, err.Error()).WriteTo(ctx.ResponseWriter)
			return
		}

		result := Result{
			Progress:     "complete",
			Matrix:       matrix,
			SpeedProfile: sp,
		}

		ctx.ContentType("json")
		if err := json.NewEncoder(ctx.ResponseWriter).Encode(result); err != nil {
			viaErr.NewError(viaErr.ErrMatrixComputation, "Failed to encode results to response writer.").WriteTo(ctx.ResponseWriter)
			return
		}

	} else {
		body, _ := ioutil.ReadAll(ctx.Request.Body)
		ctx.Abort(400, "Missing or invalid matrix data, speed profile, or country. You sent: "+string(body))
		return
	}
}

func (server *Server) GetServerStatus(ctx *web.Context) string {
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
