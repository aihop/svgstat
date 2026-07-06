package geoip

import (
	"fmt"
	"net"
	"sync"

	"github.com/oschwald/maxminddb-golang"
	"github.com/rs/zerolog/log"
)

type Location struct {
	CountryISO string  // ISO 3166-1 alpha-2, e.g., "US", "CN"
	Country    string  // Country name, e.g., "United States", "China"
	Region     string  // Region/state
	City       string  // City name
	Timezone   string  // Timezone
	Latitude   float64 // Latitude
	Longitude  float64 // Longitude
}

type GeoIP struct {
	db   *maxminddb.Reader
	once sync.Once
}

func New(dbPath string) (*GeoIP, error) {
	db, err := maxminddb.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open geoip database: %w", err)
	}

	log.Info().Str("path", dbPath).Msg("GeoIP database loaded successfully")
	return &GeoIP{db: db}, nil
}

func (g *GeoIP) Lookup(ipStr string) (*Location, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	var result geoIPRecord
	err := g.db.Lookup(ip, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup IP %s: %w", ipStr, err)
	}

	location := &Location{
		CountryISO: result.Country.ISOCode,
		Country:    result.Country.Names["en"],
		Timezone:   result.Location.Timezone,
		Latitude:   result.Location.Latitude,
		Longitude:  result.Location.Longitude,
	}

	if len(result.Subdivisions) > 0 {
		location.Region = result.Subdivisions[0].Names["en"]
	}

	if result.City.Names["en"] != "" {
		location.City = result.City.Names["en"]
	}

	return location, nil
}

func (g *GeoIP) LookupCountryISO(ipStr string) string {
	location, err := g.Lookup(ipStr)
	if err != nil {
		log.Debug().Err(err).Str("ip", ipStr).Msg("Failed to lookup geolocation")
		return ""
	}
	return location.CountryISO
}

func (g *GeoIP) Close() error {
	return g.db.Close()
}

// Internal structs for maxminddb
type geoIPRecord struct {
	Country struct {
		ISOCode string            `maxminddb:"iso_code"`
		Names   map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Location struct {
		Timezone string  `maxminddb:"time_zone"`
		Latitude float64 `maxminddb:"latitude"`
		Longitude float64 `maxminddb:"longitude"`
	} `maxminddb:"location"`
	Subdivisions []struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"subdivisions"`
}
