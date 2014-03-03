package geo

import (
	"encoding/json"
	"errors"
	"strings"
)

// Parses the JSON input coordinates into an array.
func ParseJsonMatrix(matrix string) ([]Coord, error) {
	var target []Coord

	err := json.Unmarshal([]byte(matrix), &target)
	if err != nil {
		return nil, errors.New("JSON parsing error: " + err.Error())
	}
	return target, nil
}

// Converts the coordinate array back into JSON.
func MatrixToJson(nodes []int) ([]byte, error) {
	cont, err := json.Marshal(Matrix{Nodes: nodes})
	if err != nil {
		return []byte{}, err
	}
	return cont, nil
}

func (g *Geo) CleanCppMessage(msg string) string {
	res := msg
	if strings.LastIndex(res, "}") == -1 {
		res = res + "}"
		g.Debug.Println("brace missing")
	} else if strings.LastIndex(res, "}") != len(res)-1 {
		braceIndex := strings.LastIndex(res, "}")
		junk := res[braceIndex+1:]
		res = res[:braceIndex+1]
		g.Debug.Printf("Stripped extra data: %s", junk)
	}
	return res
}
