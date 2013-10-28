package main

import (
	"github.com/hoisie/redis"
	"testing"
)

var (
	client    redis.Client
	config, _ = loadConfig("development.json")
	server    = Server{client, config}
)

// The matrix sample we will be using.
var testPayload = struct {
	matrix        []Coord
	country       string
	speed_profile int
}{
	[]Coord{{21.5, 62.2}, {25.0, 64}},
	"finland",
	120,
}

func erase_computation(hash string, t *testing.T) {
	ok, err := server.client.Del(hash)
	if err != nil {
		t.Fatalf("deleting %s failed: %s", hash, err.Error())
	}

	if ok, _ = server.client.Exists(hash); ok {
		t.Fatalf("%s should be deleted", hash)
	}
}

func TestHashUniqueness(t *testing.T) {
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

func TestComputationCreation(t *testing.T) {
	hash, res := server.CreateComputation(testPayload.matrix, testPayload.country, testPayload.speed_profile)
	defer erase_computation(hash, t)

	if res != false {
		t.Errorf("CreateComputation(%q) => %q, wanted false, must create new", testPayload, res)
	}
}

func TestProxyCreation(t *testing.T) {
	// res ignored, tested by above method.
	hash, _ := server.CreateComputation(testPayload.matrix, testPayload.country, testPayload.speed_profile)
	defer erase_computation(hash, t)

	// create again
	hash2, res2 := server.CreateComputation(testPayload.matrix, testPayload.country, testPayload.speed_profile)
	defer erase_computation(hash2, t)

	if hash == hash2 {
		t.Errorf("Proxy creation failed: %q == %q, wanted different", hash, hash2)
	}

	if res2 != true {
		t.Error("Proxy creation didn't succeed.")
	}

	ttl, _ := client.Ttl(hash)
	ttl2, _ := client.Ttl(hash2)

	if ttl != ttl2 {
		t.Errorf("TTL different: %q != %q, TTL times should be equal for proxy resources.", ttl, ttl2)
	}

}
