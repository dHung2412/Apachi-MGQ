package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"DP_Maintenance/internal/model"
)

type DatasetHandler struct {
	datasets map[string]*model.Dataset
}

func NewDatasetHandler() *DatasetHandler {
	return &DatasetHandler{
		datasets: make(map[string]*model.Dataset),
	}
}

func (h *DatasetHandler) List(c echo.Context) error {
	var result []model.Dataset
	for _, ds := range h.datasets {
		result = append(result, *ds)
	}
	return c.JSON(http.StatusOK, model.SuccessResponse(result, "datasets retrieved"))
}

func (h *DatasetHandler) Get(c echo.Context) error {
	urn := c.Param("urn")
	if ds, exists := h.datasets[urn]; exists {
		return c.JSON(http.StatusOK, model.SuccessResponse(ds, "dataset retrieved"))
	}
	return c.JSON(http.StatusNotFound, model.ErrorResponse("dataset not found"))
}

func (h *DatasetHandler) Create(c echo.Context) error {
	var ds model.Dataset
	if err := c.Bind(&ds); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse("invalid request body"))
	}

	if ds.URN == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse("urn is required"))
	}

	h.datasets[ds.URN] = &ds
	return c.JSON(http.StatusCreated, model.SuccessResponse(ds, "dataset created"))
}

func (h *DatasetHandler) Update(c echo.Context) error {
	urn := c.Param("urn")
	var ds model.Dataset

	if _, exists := h.datasets[urn]; !exists {
		return c.JSON(http.StatusNotFound, model.ErrorResponse("dataset not found"))
	}

	if err := c.Bind(&ds); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse("invalid request body"))
	}

	ds.ID = h.datasets[urn].ID
	h.datasets[urn] = &ds
	return c.JSON(http.StatusOK, model.SuccessResponse(ds, "dataset updated"))
}

func (h *DatasetHandler) Delete(c echo.Context) error {
	urn := c.Param("urn")
	if _, exists := h.datasets[urn]; !exists {
		return c.JSON(http.StatusNotFound, model.ErrorResponse("dataset not found"))
	}

	delete(h.datasets, urn)
	return c.JSON(http.StatusOK, model.SuccessResponse(nil, "dataset deleted"))
}