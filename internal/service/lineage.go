package service

import (
	"fmt"
	"log"

	"DP_Maintenance/internal/model"
	"DP_Maintenance/internal/repository"
)

// LineageService handles data lineage business logic:
// recording column-level mappings, tracing upstream/downstream paths.
type LineageService struct {
	lineageRepo repository.LineageRepository
}

// NewLineageService creates a LineageService with an injected repository.
// Pass nil for stub mode (Phase 1).
func NewLineageService(lineageRepo repository.LineageRepository) *LineageService {
	return &LineageService{
		lineageRepo: lineageRepo,
	}
}

// RecordLineage stores lineage information from an ingest event.
// Creates dataset nodes, column nodes, and DERIVED_FROM edges in Memgraph.
func (s *LineageService) RecordLineage(lineage *model.LineagePayload) error {
	if s.lineageRepo == nil {
		log.Printf("[LINEAGE] Stub: RecordLineage(%s -> %s)",
			lineage.SourceDataset, lineage.TargetDataset)
		for _, m := range lineage.Mappings {
			log.Printf("[LINEAGE]   %s.%s -> %s.%s (%s)",
				lineage.SourceDataset, m.SourceColumn,
				lineage.TargetDataset, m.TargetColumn,
				m.Transform)
		}
		return nil
	}

	// Create lineage edges for each column mapping
	for _, mapping := range lineage.Mappings {
		if err := s.lineageRepo.RecordColumnMapping(
			lineage.SourceDataset,
			lineage.TargetDataset,
			&mapping,
			lineage.JobID,
		); err != nil {
			return fmt.Errorf("failed to record mapping %s->%s: %w",
				mapping.SourceColumn, mapping.TargetColumn, err)
		}
	}

	return nil
}

// TraceUpstream finds all upstream sources for a column in a dataset.
func (s *LineageService) TraceUpstream(datasetURN, columnName string, depth int) ([]model.LineageEdge, error) {
	if s.lineageRepo == nil {
		log.Printf("[LINEAGE] Stub: TraceUpstream(%s.%s, depth=%d)", datasetURN, columnName, depth)
		return []model.LineageEdge{}, nil
	}
	if depth <= 0 {
		depth = 10
	}
	return s.lineageRepo.TraceUpstream(datasetURN, columnName, depth)
}

// TraceDownstream finds all downstream consumers impacted by a column change.
func (s *LineageService) TraceDownstream(datasetURN, columnName string, depth int) ([]model.LineageEdge, error) {
	if s.lineageRepo == nil {
		log.Printf("[LINEAGE] Stub: TraceDownstream(%s.%s, depth=%d)", datasetURN, columnName, depth)
		return []model.LineageEdge{}, nil
	}
	if depth <= 0 {
		depth = 10
	}
	return s.lineageRepo.TraceDownstream(datasetURN, columnName, depth)
}

// GetDatasetLineage returns direct upstream and downstream for a dataset.
func (s *LineageService) GetDatasetLineage(datasetURN string) (upstream, downstream []model.LineageEdge, err error) {
	if s.lineageRepo == nil {
		log.Printf("[LINEAGE] Stub: GetDatasetLineage(%s)", datasetURN)
		return []model.LineageEdge{}, []model.LineageEdge{}, nil
	}
	return s.lineageRepo.GetDatasetLineage(datasetURN)
}
