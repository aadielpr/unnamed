package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aadielpr/unnamed/internal/config"
	"github.com/aadielpr/unnamed/internal/db"
	"github.com/aadielpr/unnamed/internal/handlers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	e   *echo.Echo
	cfg config.Config
}

func New(cfg config.Config, database *db.DB) *Server {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	healthDeps := handlers.HealthDeps{DB: database}
	e.GET("/api/health", handlers.Health(healthDeps))

	e.GET("/*", spaHandler(cfg.StaticDir))

	return &Server{
		e:   e,
		cfg: cfg,
	}
}

func (s *Server) Start() error {
	return s.e.Start(":" + s.cfg.Port)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.e.ServeHTTP(w, r)
}

// spaHandler serves static files from root and falls back to index.html so
// the client-side router can take unknown paths. Requests under /api/ that no
// API route matched are a 404, not the SPA: per decision #4, /api is the only
// API prefix and must not masquerade as the SPA.
func spaHandler(root string) echo.HandlerFunc {
	return func(c echo.Context) error {
		path := c.Request().URL.Path
		if strings.HasPrefix(path, "/api/") {
			return c.NoContent(http.StatusNotFound)
		}
		cleanPath := filepath.Clean(path)
		if cleanPath == "/" {
			cleanPath = "/index.html"
		}

		fullPath := filepath.Join(root, cleanPath)
		info, err := os.Stat(fullPath)
		if err == nil && !info.IsDir() {
			return c.File(fullPath)
		}

		return c.File(filepath.Join(root, "index.html"))
	}
}
