package geo

var (
	config, _ = LoadConfig("../development.json")
	test_geo  = NewGeo(true, config.DbUser, config.DbName, config.Port, config.AllowedCountries)
)
