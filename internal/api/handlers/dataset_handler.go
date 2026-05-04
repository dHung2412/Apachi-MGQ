package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"DP_Maintenance/internal/model"
	"DP_Maintenance/internal/service"
)

type DatasetHandler struct {
	catalogSvc *service.CatalogService
}

func NewDatasetHandler(catalogSvc *service.CatalogService) *DatasetHandler {
	return &DatasetHandler{
		catalogSvc: catalogSvc,
	}
}

func (h *DatasetHandler) List(c echo.Context) error {
	page := 1
	pageSize := 20
	datasets, total, err := h.catalogSvc.ListDatasets(page, pageSize)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to list datasets"))
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    datasets,
		"total":   total,
		"page":    page,
		"size":    pageSize,
	})
}

func (h *DatasetHandler) Get(c echo.Context) error {
	urn := c.Param("urn")
	dataset, err := h.catalogSvc.GetDataset(urn)
	if err != nil {
		return c.JSON(http.StatusNotFound, model.ErrorResponse("dataset not found"))
	}
	return c.JSON(http.StatusOK, model.SuccessResponse(dataset, "dataset retrieved"))
}

func (h *DatasetHandler) Create(c echo.Context) error {
	var ds model.Dataset
	if err := c.Bind(&ds); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse("invalid request body"))
	}

	if ds.URN == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse("urn is required"))
	}

	ds.ID = uuid.New()
	if err := h.catalogSvc.CreateDataset(&ds); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to create dataset"))
	}

	return c.JSON(http.StatusCreated, model.SuccessResponse(ds, "dataset created"))
}

func (h *DatasetHandler) Update(c echo.Context) error {
	urn := c.Param("urn")
	var ds model.Dataset

	if err := c.Bind(&ds); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse("invalid request body"))
	}

	existing, err := h.catalogSvc.GetDataset(urn)
	if err != nil {
		return c.JSON(http.StatusNotFound, model.ErrorResponse("dataset not found"))
	}

	ds.ID = existing.ID
	if err := h.catalogSvc.UpdateDataset(&ds); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse("failed to update dataset"))
	}

	return c.JSON(http.StatusOK, model.SuccessResponse(ds, "dataset updated"))
}

func (h *DatasetHandler) Delete(c echo.Context) error {
	urn := c.Param("urn")
	if err := h.catalogSvc.DeleteDataset(urn); err != nil {
		return c.JSON(http.StatusNotFound, model.ErrorResponse("dataset not found"))
	}

	return c.JSON(http.StatusOK, model.SuccessResponse(nil, "dataset deleted"))
}