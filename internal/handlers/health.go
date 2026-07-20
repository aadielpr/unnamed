package handlers

import (
	"net/http"

	"github.com/aadielpr/unnamed/internal/db"
	"github.com/labstack/echo/v4"
)

type HealthDeps struct {
	DB *db.DB
}

type HealthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

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
