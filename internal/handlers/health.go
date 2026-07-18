package handlers

import (
	"net/http"

	"github.com/aadielpr/unnamed/internal/db"
	"github.com/labstack/echo/v4"
)

// HealthDeps contains the dependencies required by health handlers.
type HealthDeps struct {
	DB *db.DB
}

// HealthResponse is the JSON body returned by GET /api/health.
type HealthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

// Health returns a handler that pings the database and reports status.
func Health(deps HealthDeps) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := deps.DB.Ping(); err != nil {
			return c.JSON(http.StatusServiceUnavailable, HealthResponse{
				Status: "degraded",
				DB:     "unreachable",
			})
		}

		return c.JSON(http.StatusOK, HealthResponse{
			Status: "ok",
			DB:     "reachable",
		})
	}
}
