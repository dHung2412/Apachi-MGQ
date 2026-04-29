package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"DP_Maintenance/internal/model"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, model.SuccessResponse(
		map[string]string{
			"status":  "healthy",
			"service": "dp_maintenance",
		},
		"service is running",
	))
}