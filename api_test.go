package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/nfleet/via/geo"
	"github.com/nfleet/via/geotypes"
)

type P struct {
	matrix  []geotypes.Coord
	country string
}

var (
	testConfig, _   = geo.LoadConfig("development.json")
	noCoordsFinland = []geotypes.Edge{
		{Source: geotypes.Location{Address: geotypes.Address{Street: "Erottaja", City: "Helsinki", Country: "Finland"}}, Target: geotypes.Location{Address: geotypes.Address{Street: "Esplanadi", City: "Helsinki", Country: "Finland"}}},
		{Source: geotypes.Location{Address: geotypes.Address{Street: "Taitoniekantie", City: "Jyväskylä", Country: "Finland"}}, Target: geotypes.Location{Address: geotypes.Address{Street: "Vuolteenkatu", City: "Tampere", Country: "Finland"}}},
		{Source: geotypes.Location{Address: geotypes.Address{Street: "Vuolteenkatu", City: "Tampere", Country: "Finland"}}, Target: geotypes.Location{Address: geotypes.Address{Street: "Hikiäntie", City: "Riihimäki", Country: "Finland"}}},
	}
	GermanyNodes = []int{7008268, 5954074, 9120618, 5747785, 6367494, 4381896, 165343, 4695953, 301566, 6781958, 6194766, 2358012, 4692093, 497151, 4355495, 700415, 4390319, 8553348, 6542963, 6634588, 6517688, 5392367, 6190253, 4198881, 504621, 9228869, 7186081, 895834, 5606641, 2647250, 6339514, 691811, 2308165, 10001569, 9120051, 759941, 10073304, 6311612, 6343270, 8179683, 3366117, 6061132, 5643689, 863577, 4931326, 6649953, 294592, 4419734, 4802743, 9567772, 5985694, 468718, 7761320, 5053336, 7260626, 4031757, 4574787, 3596809, 9684347, 164709, 195923, 9834961, 8155072, 7889391, 5200362, 8055855, 8304451, 3560874, 6367318, 5940849, 4633454, 1769665, 7338960, 7345082, 6542540, 1699666, 4264625, 7749655, 10239981, 9265680, 9749596, 6070658, 7983593, 1573825, 5277830, 6966871, 4712734, 7511618, 5665254, 8081312, 2438484, 6431216, 6041290, 2647929, 10208160, 143963, 2291010, 4056536, 5649426, 9047055, 6949841, 4713404, 7111360, 1606682, 735773, 2156000, 1852553, 4645613, 10440701, 3375948, 2775921, 4835420, 1425989, 3591738, 6262150, 7189083, 2183250, 1455662, 9080962, 4939489, 5849163, 2481864, 5358757, 3992900, 8099433, 6030163, 9543396, 3100653, 46963, 6764954, 6741542, 1499985, 1518525, 5777389, 7179979, 6021581, 4558442, 4698864, 10384852, 9110113, 404713, 4015965, 839030, 3539652, 8428891, 171848, 3299486, 6890628, 6617824, 7543977, 7611477, 3287977, 6353455, 10167787, 4457068, 904612, 1125329, 6047234, 2655479, 5740142, 8520248, 334569, 1932239, 7974489, 6232975, 8266989, 5910401, 7383070, 7639188, 906419, 1633652, 4206622, 9425661, 643972, 3654958, 9809412, 9690697, 2206079, 8277645, 5117033, 3990216, 8974041, 5227650, 2746089, 3557062, 127803, 3631098, 5863287, 4719210, 4256235, 8015655, 159987, 7102034, 6788209, 1203170, 6874024, 4575612, 8767789, 2668569, 2569985, 3715654, 393198, 9950700, 9756778, 4282083, 3597652, 5530427, 10165444, 2609535, 3923065, 3748038, 238269, 8862259, 1212190, 5192509, 10063862, 5217308, 6709583, 1871292, 7000930, 7729292, 8338947, 2358378, 10124331, 145863, 10526984, 3662053, 3349217, 7926182, 1962950, 4896276, 8337446, 83303, 4710409, 8901360, 827293, 2015800, 8127920, 2283773, 4786904, 5647141, 5537861, 4039566, 6844087, 538542, 5215683, 10380146, 3521540, 3320306, 3600260, 4355790, 10133601, 3836895, 5575268, 5077219, 2835701, 1422817, 4077361, 6624582, 4221540, 2904403, 4568293, 5892519, 184689, 2476077, 6056339, 109807, 1612422, 3878049, 191945, 10115368, 3764898, 735073, 1606269, 2589551, 2055257, 2949524, 8837422, 4773969, 1356119, 5844509, 8743064, 3449998, 6363927, 1963464, 1983618, 2834810, 9946551, 7951028, 9582813, 1855524, 6821228, 7378037, 3085397, 9599672, 5504082, 3643548, 10565967, 10326036, 7670349, 7211086, 3371299, 8666235, 3342981, 1970318, 8036747, 8855710, 6715344}
	FinlandNodes = []int{728652, 110611, 626130, 715935, 731568, 545875, 239476, 111993, 22456, 484102, 207355, 813582, 704462, 800619, 319546, 447197, 433927, 210866, 397484, 191488, 752941, 858119, 225867, 558678, 751816, 873106, 606328, 44716, 545995, 191895, 417374, 581142, 5765, 766736, 441376, 754878, 137704, 110343, 877904, 462092, 703282, 321789, 271075, 74256, 708213, 487186, 864125, 434782, 897834, 668585, 481143, 328200, 755791, 787787, 804898, 517329, 38329, 385557, 52910, 147543, 332693, 508533, 78022, 182641, 755700, 588135, 171704, 792301, 404424, 12331, 463707, 682149, 514431, 660553, 50353, 301574, 111580, 159890, 697437, 433239, 880503, 213494, 821491, 434268, 463123, 185742, 74775, 267689, 234777, 882035, 629121, 124414, 805268, 34744, 238945, 426994, 358980, 467814, 882226, 348105, 49877, 421225, 886172, 660210, 808690, 117593, 552176, 353276, 381605, 479289, 143576, 215157, 459668, 107667, 297354, 34686, 321908, 485800, 230598, 67389, 872273, 221739, 732023, 587348, 753484, 406863, 418034, 24114, 658888, 585920, 594738, 594682, 9963, 399355, 105233, 235726, 546929, 186023, 670348, 609799, 783980, 513459, 773869, 480130, 108777, 92163, 175665, 828870, 722232, 69790, 790840, 803213, 151523, 828589, 376782, 134263, 253693, 790789, 152821, 365474, 232532, 602094, 835697, 211793, 510894, 764759, 295461, 245014, 598198, 155135, 512634, 483572, 279703, 740451, 83830, 29656, 582624, 57196, 294203, 648408, 67383, 628856, 435138, 418317, 558189, 43173, 651859, 412915, 239282, 502303, 522995, 484783, 322710, 498161, 402752, 536297, 20764, 765584, 449088, 276625, 17355, 533171, 834635, 254981, 383510, 209309, 891922, 772809, 477182, 367671, 667245, 367083, 337482, 725056, 902133, 852205, 401564, 140367, 503914, 131911, 212593, 69331, 181539, 168629, 283736, 183020, 295266, 548967, 518341, 707482, 125622, 434149, 398782, 741374, 8138, 419563, 268158, 618699, 503931, 528594, 37995, 490222, 764459, 407982, 687149, 91850, 890969, 668736, 143210, 480549, 698890, 12724, 803503, 277601, 355647, 257762, 261803, 424876, 423287, 452671, 212446, 406612, 433707, 282632, 174221, 675019, 268424, 192268, 768974, 615963, 52, 180939, 774339, 271686, 77732, 37234, 731231, 713928, 109461, 735758, 253337, 197115, 333738, 510371, 569408, 816725, 37155, 807837, 46437, 746214, 121659, 698470, 43366, 269891, 52628, 12462, 401050, 660123, 899435, 755973, 822040, 654775, 295733, 184192, 185815}
)

func TestBoundingBoxes(t *testing.T) {
	var tests = []struct {
		in  P
		out bool
	}{
		{P{[]geotypes.Coord{{60.0, 25.0}}, "finland"}, true},
		{P{[]geotypes.Coord{{50.0, 25.0}}, "finland"}, false},
		{P{[]geotypes.Coord{{61.0, 19.0}}, "finland"}, false},
		{P{[]geotypes.Coord{{50.0, 10.0}}, "germany"}, true},
		{P{[]geotypes.Coord{{45.0, 10.0}}, "germany"}, false},
		{P{[]geotypes.Coord{{50.0, 4.05}}, "germany"}, false},
	}

	for i, test := range tests {
		res, _ := check_coordinate_sanity(test.in.matrix, test.in.country)
		if res != test.out {
			t.Errorf("%d. check_coordinate_sanity(%q, %q) => %q, want %q", i, test.in.matrix, test.in.country, res, test.out)
		}
	}
}

func api_test_resolve(t *testing.T, locations []geotypes.Location) {
	request := fmt.Sprintf("http://localhost:%d/resolve", testConfig.Port)

	jsonLoc, _ := json.Marshal(locations)
	b := strings.NewReader(string(jsonLoc))

	response, err := http.Post(request, "application/json", b)
	if err != nil {
		t.Fatal(response)
	} else {
		defer response.Body.Close()
		cont, err := ioutil.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}

		var resolved []geotypes.Location
		if err := json.Unmarshal(cont, &resolved); err != nil {
			t.Fatal(string(cont), err)
		}
		for i, loc := range resolved {
			t.Logf("Resolved %v to have coords %v", locations[i], loc)
		}
	}
}

func TestAPIResolvationCoordinate(t *testing.T) {
	locations := []geotypes.Location{
		{Address: geotypes.Address{Country: "finland"}, Coordinate: geotypes.Coordinate{Latitude: 62.24, Longitude: 25.74}},
	}
	api_test_resolve(t, locations)
}

func TestAPIResolvationWithoutCoordinate(t *testing.T) {
	locations := []geotypes.Location{
		{Address: geotypes.Address{Street: "Esplanadi", City: "Helsinki", Country: "finland"}},
	}
	api_test_resolve(t, locations)
}

func api_coordinate_query(t *testing.T, edges []geotypes.Edge, speed_profile int) []geotypes.CoordinatePath {
	payload := struct {
		SpeedProfile int
		Edges        []geotypes.Edge
	}{
		speed_profile,
		edges,
	}

	jsonPayload, err := json.Marshal(payload)
	b := strings.NewReader(string(jsonPayload))

	var paths []geotypes.CoordinatePath

	request := fmt.Sprintf("http://localhost:%d/paths", testConfig.Port)
	response, err := http.Post(request, "application/json", b)
	if err != nil {
		t.Fatal(response)
	} else if response.StatusCode != 200 {
		cont, _ := ioutil.ReadAll(response.Body)
		t.Fatal(string(cont))
	} else {
		defer response.Body.Close()
		cont, err := ioutil.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}

		if err := json.Unmarshal(cont, &paths); err != nil {
			t.Logf(err.Error())
		}

		t.Logf("sent %d paths <=> got %d back", len(edges), len(paths))

		for i, p := range paths {
			t.Logf("Calculated path of %d m / %d secs from %s, %s, %s to %s, %s, %s", p.Distance, p.Time,
				edges[i].Source.Address.Street, edges[i].Source.Address.City, edges[i].Source.Address.Country,
				edges[i].Target.Address.Street, edges[i].Target.Address.City, edges[i].Target.Address.Country)
		}

		t.Logf("Calculated %d paths", len(paths))
	}

	return paths
}

func TestAPICalculateCoordinatePathWithoutCoordinates(t *testing.T) {
	api_coordinate_query(t, noCoordsFinland, 100)
}
