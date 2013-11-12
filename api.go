package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hoisie/web"
	"strconv"
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

func check_coordinate_sanity(matrix []Coord, country string) (bool, error) {
	bbox := bounding_boxes[country]

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
	debug.Println(buf.String())
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

		mat, err := parse_json_matrix(data)
		if err != nil {
			ctx.Abort(422, err.Error())
			return
		}

		ok, err = check_coordinate_sanity(mat, country)
		if err != nil {
			ctx.Abort(422, err.Error())
			return
		}

		hash, computed = server.CreateMatrixComputation(mat, country, int(speed_profile))

		// launch computation here if the result wasn't proxied.
		if !computed {
			go server.ComputeMatrix(hash)
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
	progress, err := server.GetMatrixComputationProgress(resource)
	if err != nil {
		ctx.Abort(500, err.Error())
		return ""
	}

	if progress == "complete" {
		url := fmt.Sprintf("/spp/%s/result", resource)
		debug.Println("redirect ->", url)
		ctx.Redirect(303, url)
	} else {
		ctx.ContentType("json")
		msg := fmt.Sprintf(`{ "progress": "%s" }`, progress)
		return msg
	}

	return ""
}

func (server *Server) GetMatrixResult(ctx *web.Context, resource string) string {
	if ex, _ := server.client.Exists(resource); !ex {
		ctx.Abort(500, "Result expired. POST again.")
		return ""
	}

	progress, err := server.GetMatrixComputationProgress(resource)
	if progress != "complete" {
		ctx.Abort(403, "Computation is not ready yet.")
		return ""
	}

	if exists, _ := server.client.Hexists(resource, "see"); exists {
		debug.Println("Got proxy result")
		pointer, _ := server.client.Hget(resource, "see")
		resource = string(pointer)
	}

	data, err := server.client.Hget(resource, "result")
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

func (server *Server) GetMatrixStatus(ctx *web.Context) string {
	db, _ := sql.Open("postgres", server.Config.String())
	defer db.Close()

	err := db.Ping()

	if err != nil {
		ctx.Abort(500, "Could not connect to database: "+err.Error())
		return ""
	}

	return "OK"
}

func (server *Server) GetCorrectCoordinate(ctx *web.Context) string {
	s_lat, lat_ok := ctx.Params["lat"]
	s_long, long_ok := ctx.Params["long"]

	// TODO(ane): implement automatic lookup of country
	s_country, country_ok := ctx.Params["country"]

	if !lat_ok || !long_ok || !country_ok {
		ctx.Abort(400, fmt.Sprintf("Missing parameter, need lat, long, country, you gave: %q", ctx.Params))
		return ""
	}

	lat, err := strconv.ParseFloat(s_lat, 32)
	if err != nil {
		ctx.Abort(400, fmt.Sprintf("Latitude %s is invalid, cannot parse!", s_lat))
	}

	long, err := strconv.ParseFloat(s_long, 32)
	if err != nil {
		ctx.Abort(400, fmt.Sprintf("Longitude %s is invalid, cannot parse!", s_long))
	}

	if _, ok := server.Config.AllowedCountries[s_country]; !ok {
		ctx.Abort(500, fmt.Sprintf("Country %s not allowed", s_country))
	}

	coord := Coord{lat, long}

	corr_node, err := CorrectPoint(server.Config, coord, s_country)
	if err != nil {
		ctx.Abort(500, err.Error())
	}

	response, err := json.Marshal(corr_node)
	if err != nil {
		ctx.Abort(500, err.Error())
	}

	ctx.Header().Set("Access-Control-Allow-Origin", "*")
	ctx.ContentType("application/json")
	return string(response)
}

func (server *Server) GetNodesToCoordinates(ctx *web.Context) string {
	nodes, ex := ctx.Params["coords"]

	if !ex {
		ctx.Abort(400, fmt.Sprintf("Missing parameter: coords"))
		return ""
	}

	var parsedNodes []int
	if err := json.Unmarshal([]byte(nodes), &parsedNodes); err != nil {
		ctx.Abort(400, err.Error())
		return ""
	}

	coordinates, err := GetCoordinates(server.Config, "germany", parsedNodes)
	if err != nil {
		ctx.Abort(500, err.Error())
		return ""
	}

	cont, err := json.Marshal(coordinates)
	if err != nil {
		ctx.Abort(500, err.Error())
		return ""
	}
	return string(cont)
}

func (server *Server) GetPath(ctx *web.Context) string {
	p_source, p_source_ok := ctx.Params["source"]
	p_target, p_target_ok := ctx.Params["target"]
	p_country, p_country_ok := ctx.Params["country"]
	p_sp, p_sp_ok := ctx.Params["speed_profile"]

	if !p_source_ok || !p_target_ok || !p_country_ok || !p_sp_ok {
		ctx.Abort(400, fmt.Sprintf("Missing parameter, need country, speed_profile, source, target; you gave: %q", ctx.Params))
		return ""
	}

	if _, ok := server.Config.AllowedCountries[p_country]; !ok {
		ctx.Abort(500, fmt.Sprintf("Country %s not allowed", p_country))
		return ""
	}

	source, _ := strconv.Atoi(p_source)
	target, _ := strconv.Atoi(p_target)
	sp, _ := strconv.Atoi(p_sp)

	path, err := CalculatePath(source, target, p_country, sp)
	if err != nil {
		ctx.Abort(500, err.Error())
		return ""
	}

	data, err := json.Marshal(path)
	if err != nil {
		ctx.Abort(500, "json error: "+err.Error())
		return ""
	}

	return string(data)
}
