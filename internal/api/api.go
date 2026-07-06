package api

import (
	"github.com/rs/zerolog/log"
	"github.com/svgstat/svgstat/internal/analytics"
	"github.com/svgstat/svgstat/internal/auth"
	"github.com/svgstat/svgstat/internal/cache"
	"github.com/svgstat/svgstat/internal/config"
	"github.com/svgstat/svgstat/internal/counter"
	"github.com/svgstat/svgstat/internal/database"
	"github.com/svgstat/svgstat/internal/geoip"
	"github.com/svgstat/svgstat/internal/project"
	"github.com/svgstat/svgstat/internal/renderer"
)

type App struct {
	config      *config.Config
	db          *database.Database
	cache       *cache.Cache
	auth        *auth.Manager
	analytics   *analytics.Analytics
	counter     *counter.Counter
	renderer    *renderer.Renderer
	projectRepo project.Repository
	geoIP       *geoip.GeoIP
}

func NewApp(cfg *config.Config) (*App, error) {
	db, err := database.New(cfg)
	if err != nil {
		return nil, err
	}

	c, err := cache.New(cfg)
	if err != nil {
		return nil, err
	}

	var g *geoip.GeoIP
	if cfg.GeoIP.DBPath != "" {
		g, err = geoip.New(cfg.GeoIP.DBPath)
		if err != nil {
			log.Warn().Err(err).Str("path", cfg.GeoIP.DBPath).Msg("Failed to load GeoIP database, geolocation will be disabled")
		}
	}

	projectRepo := project.NewPostgresRepository(db.Pool)

	app := &App{
		config:      cfg,
		db:          db,
		cache:       c,
		auth:        auth.NewManager(db.Pool),
		analytics:   analytics.New(c, projectRepo, g),
		counter:     counter.New(c, projectRepo),
		renderer:    renderer.New(),
		projectRepo: projectRepo,
		geoIP:       g,
	}

	return app, nil
}

func (a *App) Close() {
	a.db.Close()
	_ = a.cache.Close()
	if a.geoIP != nil {
		_ = a.geoIP.Close()
	}
}
