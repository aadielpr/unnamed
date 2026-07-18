package server

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/aadielpr/unnamed/internal/config"
	"github.com/aadielpr/unnamed/internal/db"
	"github.com/aadielpr/unnamed/internal/handlers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Server wires the Echo router to the application dependencies.
type Server struct {
	e *echo.Echo
	cfg config.Config
}

// New creates an Echo server with routes configured.
func New(cfg config.Config, database *db.DB) *Server {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	healthDeps := handlers.HealthDeps{DB: database}
	e.GET("/api/health", handlers.Health(healthDeps))

	// Serve the SPA static assets. If a file is missing, fall back to index.html
	// so the client-side router can handle the path.
	e.GET("/*", spaHandler(cfg.StaticDir))

	return &Server{
		e:   e,
		cfg: cfg,
	}
}

// Start runs the HTTP server.
func (s *Server) Start() error {
	return s.e.Start(":" + s.cfg.Port)
}

// ServeHTTP implements http.Handler for use in tests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.e.ServeHTTP(w, r)
}

// spaHandler serves static files from root and falls back to index.html.
func spaHandler(root string) echo.HandlerFunc {
	return func(c echo.Context) error {
		path := c.Request().URL.Path
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
