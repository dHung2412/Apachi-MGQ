package handlers

import (
	"DP_Maintenance/internal/model"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type IngestHandler struct {
	eventLog []model.IngestEvent
}

func NewIngestHandler() *IngestHandler {
	return &IngestHandler{
		eventLog: make([]model.IngestEvent, 0),
	}
}

func (h *IngestHandler) Ingest(c echo.Context) error {
	var event model.IngestEvent
	if err := c.Bind(&event); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse("invalid request body"))
	}

	if event.EventType == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse("event_type is required"))
	}

	if event.Payload.Dataset != nil {
		dataset := event.Payload.Dataset
		log.Printf("[INGEST] Dataset: %s (URN: %s)", dataset.Name, dataset.URN)
		log.Printf("[INGEST] Platform: %s, Database: %s, Schema: %s",
			dataset.Platform, dataset.Database, dataset.Schema)

		if len(dataset.Columns) > 0 {
			log.Printf("[INGEST] Columns:")
			for _, col := range dataset.Columns {
				log.Printf("  - %s (%s)", col.Name, col.DataType)
			}
		}
	}

	if event.Payload.Lineage != nil {
		lineage := event.Payload.Lineage
		log.Printf("[INGEST] Lineage: %s -> %s (transform: %s)",
			lineage.SourceDataset, lineage.TargetDataset, lineage.TransformType)
	}

	if event.Payload.Job != nil {
		job := event.Payload.Job
		log.Printf("[INGEST] Job: %s, Dag: %s, Status: %s",
			job.JobID, job.DagID, job.Status)
	}

	h.eventLog = append(h.eventLog, event)

	fmt.Fprintf(c.Response(), "Event accepted")
	return c.JSON(http.StatusAccepted, model.SuccessResponse(
		map[string]interface{}{
			"event_type": event.EventType,
			"timestamp":  event.Timestamp,
		},
		"event accepted",
	))
}

func (h *IngestHandler) ListEvents(c echo.Context) error {
	return c.JSON(http.StatusOK, model.SuccessResponse(
		h.eventLog,
		"events listed successfully",
	))
}
