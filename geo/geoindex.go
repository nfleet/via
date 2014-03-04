package geo

import (
	"errors"
	"sync"
	"time"

	"github.com/nfleet/via/geotypes"
)

var table_names = map[string]string{
	"finland": "finland_nodes",
	"germany": "germany_nodes",
}

// Geoindexes the coordinates. Modifies mapped with the results since
// modifying a slice modifies the underlying array. Note! Launch this as a goroutine!
func (g *Geo) Geoindex(country string, coordinates []geotypes.Coord, mapped []int) error {

	t := time.Now()
	var coordNodes []int

	for _, coord := range coordinates {
		node, err := g.DB.QueryClosestPoint(geotypes.Coord{coord[0], coord[1]}, country)

		if err != nil {
			return err
		}

		coordNodes = append(coordNodes, node.Id)
	}
	t2 := time.Since(t)
	g.Debug.Printf("goroutine: Geoindexed %d nodes in %s", len(coordinates), t2)
	copy(mapped, coordNodes)

	return nil
}

// Runs the geoindexer for a coordinate set. Launches N connections to the PostgreSQL
// geoindexing database. Currently you should use N = 4,6,8, the benefit tapers off
// after 8 connections.
func (g *Geo) RunGeoindexer(country string, coordinates []geotypes.Coord, connections int) ([]int, error) {
	if connections > 1 && (connections%2) != 0 {
		return []int{}, errors.New("connections must be even when more than 1, you gave: " + string(connections))
	}
	chunkSize := len(coordinates) / connections
	src := make([][]geotypes.Coord, connections)
	dst := make([][]int, connections)

	offset := 0
	jobs := 0
	i := 0
	for ; i < connections; i++ {
		src[i] = coordinates[offset : offset+chunkSize]
		dst[i] = make([]int, chunkSize)
		offset += chunkSize
		jobs++
	}

	// process leftovers, e.g. 10 === 2 (mod 4) -> need a fifth connection in that case
	remnants := len(coordinates) % connections
	if remnants > 0 {
		g.Debug.Printf("Got %d remnants.", remnants)
		src = append(src, coordinates[offset:offset+remnants])
		dst = append(dst, make([]int, remnants))
		jobs++
	}

	var wg sync.WaitGroup
	t := time.Now()
	err_chan := make(chan error)
	for j := 0; j < jobs; j++ {
		wg.Add(1)
		j := j
		go func() {
			err := g.Geoindex(country, src[j], dst[j])
			if err != nil {
				// send the error to err_chan
				err_chan <- err
			}
			wg.Done()
		}()
	}

	// Each wg.Done() decrements the counter by 1, wg.Wait() until this is 0.
	// As a result, wg.Wait() blocks until all the geoindexer goroutines have
	// completed (called wg.Done()). Then close will close the error channel.
	go func() { wg.Wait(); close(err_chan) }()

	// This will cause the function to block until err_chan is closed.
	for err := range err_chan {
		g.Debug.Println("Geoindexing error:", err.Error())
		return nil, err
	}

	g.Debug.Printf("Geoindexing done in %s (%d+%d connections)", time.Since(t), connections, jobs-connections)
	var res []int
	for _, chunk := range dst {
		res = append(res, chunk...)
	}
	return res, nil
}
