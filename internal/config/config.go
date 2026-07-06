package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	HTTP     HTTPConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	Logging  LoggingConfig
	Cache    CacheConfig
	GeoIP    GeoIPConfig
}

type HTTPConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
	PoolMax  int
	PoolMin  int
}

type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
}

type LoggingConfig struct {
	Level string
}

type CacheConfig struct {
	TTL time.Duration
}

type GeoIPConfig struct {
	DBPath string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		HTTP: HTTPConfig{
			Addr:         getEnv("HTTP_ADDR", ":8080"),
			ReadTimeout:  getDurationEnv("HTTP_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("HTTP_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDurationEnv("HTTP_IDLE_TIMEOUT", 60*time.Second),
		},
		Postgres: PostgresConfig{
			Host:     getEnv("PG_HOST", "localhost"),
			Port:     getEnv("PG_PORT", "5432"),
			User:     getEnv("PG_USER", "svgstat"),
			Password: getEnv("PG_PASSWORD", "svgstat"),
			Database: getEnv("PG_DATABASE", "svgstat"),
			SSLMode:  getEnv("PG_SSL_MODE", "disable"),
			PoolMax:  getIntEnv("PG_POOL_MAX", 20),
			PoolMin:  getIntEnv("PG_POOL_MIN", 5),
		},
		Redis: RedisConfig{
			Addr:         getEnv("REDIS_ADDR", "localhost:6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getIntEnv("REDIS_DB", 0),
			PoolSize:     getIntEnv("REDIS_POOL_SIZE", 20),
			MinIdleConns: getIntEnv("REDIS_MIN_IDLE_CONNS", 5),
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		Cache: CacheConfig{
			TTL: getDurationEnv("CACHE_TTL", 5*time.Minute),
		},
		GeoIP: GeoIPConfig{
			DBPath: getEnv("GEOIP_DB_PATH", "resource/GeoLite2-City.mmdb"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		var result int
		_, err := time.ParseDuration(value)
		if err == nil {
			d, _ := time.ParseDuration(value)
			return int(d.Seconds())
		}
		_, err = fmt.Sscanf(value, "%d", &result)
		if err == nil {
			return result
		}
		log.Warn().Str("key", key).Msg("Failed to parse int env, using default")
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		d, err := time.ParseDuration(value)
		if err == nil {
			return d
		}
		log.Warn().Str("key", key).Msg("Failed to parse duration env, using default")
	}
	return defaultValue
}
