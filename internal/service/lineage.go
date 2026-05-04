package service

import (
	"fmt"
	"log"
	"time"

	"DP_Maintenance/internal/model"
	"DP_Maintenance/internal/repository"
)

type LineageService struct {
	lineageRepo repository.LineageRepository
}

func NewLineageService(lineageRepo repository.LineageRepository) *LineageService {
	return &LineageService{
		lineageRepo: lineageRepo,
	}
}

func (s *LineageService) RecordLineage(lineage *model.LineagePayload) error {
	if s.lineageRepo == nil {
		log.Printf("[LINEAGE] Stub: RecordLineage(%s -> %s)",
			lineage.SourceDataset, lineage.TargetDataset)
		for _, m := range lineage.Mappings {
			log.Printf("[LINEAGE]   %s.%s -> %s.%s (%s)",
				lineage.SourceDataset, m.SourceColumn,
				lineage.TargetDataset, m.TargetColumn,
				m.TransformRule)
		}
		return nil
	}

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

	sourceEdge := &model.LineageEdge{
		SourceType: "table",
		SourceID:   lineage.SourceDataset,
		TargetType: "table",
		TargetID:   lineage.TargetDataset,
		Transform:  lineage.TransformType,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := s.lineageRepo.RecordEdge(sourceEdge); err != nil {
		log.Printf("[LINEAGE] Warning: failed to record edge: %v", err)
	}

	return nil
}

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

func (s *LineageService) GetDatasetLineage(datasetURN string) (upstream, downstream []model.LineageEdge, err error) {
	if s.lineageRepo == nil {
		log.Printf("[LINEAGE] Stub: GetDatasetLineage(%s)", datasetURN)
		return []model.LineageEdge{}, []model.LineageEdge{}, nil
	}
	return s.lineageRepo.GetDatasetLineage(datasetURN)
}

func (s *LineageService) GetLineageGraph(datasetURN string) (*model.LineageGraph, error) {
	if s.lineageRepo == nil {
		log.Printf("[LINEAGE] Stub: GetLineageGraph(%s)", datasetURN)
		return &model.LineageGraph{Nodes: []model.LineageNode{}, Edges: []model.LineageEdge{}}, nil
	}
	return s.lineageRepo.GetLineageGraph(datasetURN)
}