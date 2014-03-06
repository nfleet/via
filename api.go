package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/hoisie/web"
	"github.com/nfleet/via/geo"
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

func check_coordinate_sanity(matrix []geotypes.Coord, country string) (bool, error) {
	bbox := geo.BoundingBoxes[country]

	verify := func(pair []float64) bool {
		lat, long := pair[0], pair[1]
		if long > bbox["long_max"] || long < bbox["long_min"] || lat > bbox["lat_max"] || lat < bbox["lat_min"] {
			return false
		}
		return true
	}

	for i, pair := range matrix {
		res := verify(pair)
		if !res {
			return false, errors.New(fmt.Sprintf("Coordinate (lat: %f, long: %f) at matrix index %d is outside the limits for country \"%s\" which is %vi. Make sure you use [[LAT, LONG]...].", pair[0], pair[1], i, country, bbox))
		}
	}
	return true, nil
}

// Starts a computation, validates the matrix in POST.
// If matrix data is missing, returns 400 Bad Request.
// If on the other hand matrix is data is not missing,
// but makes no sense, it returns 422 Unprocessable Entity.
func (server *Server) PostMatrix(ctx *web.Context) {
	var hash string
	var computed bool
	params := []string{"matrix", "speed_profile", "country"}

	body := ctx.Request.Body

	var buf bytes.Buffer
	buf.ReadFrom(body)
	bodyParams := buf.Bytes()
	server.Geo.Debug.Println(buf.String())
	var paramBlob map[string]interface{}
	// Parse params
	json.Unmarshal(bodyParams, &paramBlob)

	ok := false
	for _, param := range params {
		// ok will be set to false if ctx.Params doesn't contain param
		_, ok = paramBlob[param]
	}

	if ok {
		data := paramBlob["matrix"].(string)
		country := paramBlob["country"].(string)
		sp, jep := paramBlob["speed_profile"].(float64)

		speed_profile := int(sp)
		// Sanitize speed profile.
		if !jep || !contains(speed_profile, allowed_speeds) {
			msg := fmt.Sprintf("speed profile '%d' makes no sense, must be one of %s", speed_profile, fmt.Sprint(allowed_speeds))
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

		mat, err := geo.ParseJsonMatrix(data)
		if err != nil {
			ctx.Abort(422, err.Error())
			return
		}

		ok, err = check_coordinate_sanity(mat, country)
		if err != nil {
			ctx.Abort(422, err.Error())
			return
		}

		hash, computed = server.Geo.CreateMatrixComputation(mat, country, int(speed_profile))

		// launch computation here if the result wasn't proxied.
		if !computed {
			go server.Geo.ComputeMatrix(hash)
		}
	} else {
		ctx.Abort(400, "Missing matrix data or speed profile or country. You sent: "+buf.String())
		return
	}

	loc := fmt.Sprintf("/spp/%s", hash)

	ctx.Redirect(201, loc)
}

// Returns a computation from the server as identified by the resource parameter
// in GET.
func (server *Server) GetMatrix(ctx *web.Context, resource string) string {
	progress, err := server.Geo.GetMatrixComputationProgress(resource)
	if err != nil {
		ctx.Abort(500, err.Error())
		return ""
	}

	if progress == "complete" {
		url := fmt.Sprintf("/spp/%s/result", resource)
		server.Geo.Debug.Println("redirect ->", url)
		ctx.Redirect(303, url)
	} else {
		ctx.ContentType("json")
		msg := fmt.Sprintf(`{ "progress": "%s" }`, progress)
		return msg
	}

	return ""
}

func (server *Server) GetMatrixResult(ctx *web.Context, resource string) string {
	if ex, _ := server.Geo.Client.Exists(resource); !ex {
		ctx.Abort(500, "Result expired. POST again.")
		return ""
	}

	progress, err := server.Geo.GetMatrixComputationProgress(resource)
	if progress != "complete" {
		ctx.Abort(403, "Computation is not ready yet.")
		return ""
	}

	if exists, _ := server.Geo.Client.Hexists(resource, "see"); exists {
		server.Geo.Debug.Println("Got proxy result")
		pointer, _ := server.Geo.Client.Hget(resource, "see")
		resource = string(pointer)
	}

	data, err := server.Geo.Client.Hget(resource, "result")
	if err != nil {
		ctx.Abort(500, "Redis error: "+err.Error())
		return ""
	}

	if data != nil {
		ctx.ContentType("json")
		return fmt.Sprintf("{ \"Matrix\": %s }", string(data))
	}

	return ""
}

func (server *Server) GetServerStatus(ctx *web.Context) string {
	err := server.Geo.DB.QueryStatus()

	if err == nil {
		return "OK"
	}

	return fmt.Sprintf("Could not connect to database: %s\n", err.Error())
}

func (server *Server) PostCoordinatePaths(ctx *web.Context) string {
	content, err := ioutil.ReadAll(ctx.Request.Body)

	var (
		paths    geotypes.PathsInput
		computed []geotypes.CoordinatePath
	)

	if err := json.Unmarshal(content, &paths); err != nil {
		ctx.Abort(400, "Couldn't parse JSON: "+err.Error()+" in '"+string(content)+"'")
		return ""
	} else {
		var err error
		computed, err = server.Geo.CalculateCoordinatePaths(paths)
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

func (server *Server) PostResolve(ctx *web.Context) string {
	ctx.Header().Set("Access-Control-Allow-Origin", "*")
	ctx.ContentType("application/json")
	content, err := ioutil.ReadAll(ctx.Request.Body)
	var locations, resolvedLocations []geotypes.Location

	// Parse params
	if err := json.Unmarshal(content, &locations); err != nil {
		ctx.Abort(400, "Couldn't parse JSON: "+err.Error())
		return ""
	} else {
		for i := 0; i < len(locations); i++ {
			newLoc, err := server.Geo.ResolveLocation(locations[i])
			if err != nil {
				ctx.Abort(422, "Resolvation failure: "+err.Error())
			}
			resolvedLocations = append(resolvedLocations, newLoc)
		}
	}

	res, err := json.Marshal(resolvedLocations)
	if err != nil {
		ctx.Abort(500, err.Error())
	}

	return string(res)
}
