package geo

import (
	"testing"

	"github.com/nfleet/via/geotypes"
)

// The matrix sample we will be using.
var testPayload = struct {
	matrix        []geotypes.Coord
	country       string
	speed_profile int
}{
	[]geotypes.Coord{{21.5, 62.2}, {25.0, 64}},
	"finland",
	120,
}

func erase_computation(hash string, t *testing.T) {
	ok, err := test_geo.Client.Del(hash)
	if err != nil {
		t.Fatalf("deleting %s failed: %s", hash, err.Error())
	}

	if ok, _ = test_geo.Client.Exists(hash); ok {
		t.Fatalf("%s should be deleted", hash)
	}
}

func TestMatrixHashUniqueness(t *testing.T) {
	baseMatrix, baseCountry, baseSpeed := "foo", "finland", 40
	h1 := CreateMatrixHash(baseMatrix, baseCountry, baseSpeed)
	h2 := CreateMatrixHash(baseMatrix, "germany", baseSpeed)
	h3 := CreateMatrixHash(baseMatrix, "germany", 60)

	testUniq := func(s1, s2 string) {
		if s1 == s2 {
			t.Fatalf("%s equal to %s, wtf?", h1, h2)
		}
	}

	testUniq(h1, h2)
	testUniq(h2, h3)
}

func TestMatrixComputationCreation(t *testing.T) {
	hash, res := test_geo.CreateMatrixComputation(testPayload.matrix, testPayload.country, testPayload.speed_profile)
	defer erase_computation(hash, t)

	if res != false {
		t.Errorf("CreateComputation(%q) => %q, wanted false, must create new", testPayload, res)
	}
}

func TestMatrixProxyCreation(t *testing.T) {
	// res ignored, tested by above method.
	hash, _ := test_geo.CreateMatrixComputation(testPayload.matrix, testPayload.country, testPayload.speed_profile)
	defer erase_computation(hash, t)

	// create again
	hash2, res2 := test_geo.CreateMatrixComputation(testPayload.matrix, testPayload.country, testPayload.speed_profile)
	defer erase_computation(hash2, t)

	if hash == hash2 {
		t.Errorf("Proxy creation failed: %q == %q, wanted different", hash, hash2)
	}

	if res2 != true {
		t.Error("Proxy creation didn't succeed.")
	}

	ttl, _ := test_geo.Client.Ttl(hash)
	ttl2, _ := test_geo.Client.Ttl(hash2)

	if ttl != ttl2 {
		t.Errorf("TTL different: %q != %q, TTL times should be equal for proxy resources.", ttl, ttl2)
	}

}
