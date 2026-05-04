package memgraph

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"DP_Maintenance/internal/model"
)

type LineageRepository struct {
	conn *Connection
}

func NewLineageRepository(conn *Connection) *LineageRepository {
	return &LineageRepository{conn: conn}
}

func (r *LineageRepository) UpsertDataset(dataset *model.Dataset) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := neo4j.ExecuteQuery(ctx, r.conn.Driver(), QueryUpsertDataset, map[string]any{
		"id":          dataset.ID.String(),
		"name":        dataset.Name,
		"database":    dataset.Database,
		"schema":     dataset.Schema,
		"description": dataset.Description,
	}, neo4j.EagerResultTransformer)
	return err
}

func (r *LineageRepository) UpsertColumn(column *model.ColumnDef, datasetID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if column.ID == uuid.Nil {
		column.ID = uuid.New()
	}

	_, err := neo4j.ExecuteQuery(ctx, r.conn.Driver(), QueryUpsertColumn, map[string]any{
		"id":               column.ID.String(),
		"name":             column.Name,
		"data_type":        column.DataType,
		"nullable":        column.IsNullable,
		"dataset_id":       datasetID.String(),
		"ordinal_position": column.Position,
	}, neo4j.EagerResultTransformer)
	return err
}

func (r *LineageRepository) LinkDatasetColumn(datasetID, columnID uuid.UUID, position int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := neo4j.ExecuteQuery(ctx, r.conn.Driver(), QueryLinkDatasetColumn, map[string]any{
		"dataset_id":      datasetID.String(),
		"column_id":      columnID.String(),
		"ordinal_position": position,
	}, neo4j.EagerResultTransformer)
	return err
}

func (r *LineageRepository) RecordColumnMapping(sourceDataset string, targetDataset string, mapping *model.ColumnMapping, jobID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sourceID := fmt.Sprintf("%s.%s", sourceDataset, mapping.SourceColumn)
	targetID := fmt.Sprintf("%s.%s", targetDataset, mapping.TargetColumn)

	_, err := neo4j.ExecuteQuery(ctx, r.conn.Driver(), QueryRecordLineage, map[string]any{
		"source_id":      sourceID,
		"target_id":      targetID,
		"transformation": mapping.TransformRule,
		"job_id":        jobID,
	}, neo4j.EagerResultTransformer)
	return err
}

func (r *LineageRepository) RecordEdge(edge *model.LineageEdge) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	relType := "DERIVED_FROM"
	if edge.TargetType == "dataset" {
		relType = "TRANSFORMS_TO"
	}

	query := fmt.Sprintf(`
		MATCH (src:Node {id: $source_id})
		MATCH (tgt:Node {id: $target_id})
		MERGE (tgt)-[r:%s]->(src)
		SET r.transformation = $transformation,
		    r.timestamp = timestamp()
		RETURN r
	`, relType)

	_, err := neo4j.ExecuteQuery(ctx, r.conn.Driver(), query, map[string]any{
		"source_id":      edge.SourceID,
		"target_id":      edge.TargetID,
		"transformation": edge.Transform,
	}, neo4j.EagerResultTransformer)
	return err
}

func (r *LineageRepository) TraceUpstream(datasetURN, columnName string, depth int) ([]model.LineageEdge, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	columnID := fmt.Sprintf("%s.%s", datasetURN, columnName)
	query := fmt.Sprintf(QueryTraceUpstream, depth)

	result, err := neo4j.ExecuteQuery(ctx, r.conn.Driver(), query, map[string]any{
		"column_id": columnID,
	}, neo4j.EagerResultTransformer)
	if err != nil {
		return nil, err
	}

	var edges []model.LineageEdge
	for _, record := range result.Records {
		edge := model.LineageEdge{
			SourceType: "column",
			TargetType: "column",
		}
		if v, ok := record.Get("transformation"); ok {
			edge.Transform = v.(string)
		}
		if v, ok := record.Get("job_id"); ok {
			edge.SourceID = v.(string)
		}
		edges = append(edges, edge)
	}
	return edges, nil
}

func (r *LineageRepository) TraceDownstream(datasetURN, columnName string, depth int) ([]model.LineageEdge, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	columnID := fmt.Sprintf("%s.%s", datasetURN, columnName)
	query := fmt.Sprintf(QueryTraceDownstream, depth)

	result, err := neo4j.ExecuteQuery(ctx, r.conn.Driver(), query, map[string]any{
		"column_id": columnID,
	}, neo4j.EagerResultTransformer)
	if err != nil {
		return nil, err
	}

	var edges []model.LineageEdge
	for _, record := range result.Records {
		edge := model.LineageEdge{
			SourceType: "column",
			TargetType: "column",
		}
		if v, ok := record.Get("transformation"); ok {
			edge.Transform = v.(string)
		}
		if v, ok := record.Get("job_id"); ok {
			edge.TargetID = v.(string)
		}
		edges = append(edges, edge)
	}
	return edges, nil
}

func (r *LineageRepository) GetDatasetLineage(datasetURN string) ([]model.LineageEdge, []model.LineageEdge, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := neo4j.ExecuteQuery(ctx, r.conn.Driver(), QueryGetDatasetLineage, map[string]any{
		"dataset_id": datasetURN,
	}, neo4j.EagerResultTransformer)
	if err != nil {
		return nil, nil, err
	}

	var upstream, downstream []model.LineageEdge
	for _, record := range result.Records {
		sourceID, _ := record.Get("source_id")
		targetID, _ := record.Get("target_id")
		transform, _ := record.Get("transformation")
		jobID, _ := record.Get("job_id")

		edge := model.LineageEdge{
			SourceID:   sourceID.(string),
			TargetID:   targetID.(string),
			Transform:  transform.(string),
		}
		if jobID != nil {
			edge.SourceID = jobID.(string)
		}

		if sourceID == datasetURN {
			downstream = append(downstream, edge)
		} else {
			upstream = append(upstream, edge)
		}
	}
	return upstream, downstream, nil
}

func (r *LineageRepository) GetLineageGraph(datasetURN string) (*model.LineageGraph, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := neo4j.ExecuteQuery(ctx, r.conn.Driver(), QueryGetLineageGraph, map[string]any{
		"dataset_id": datasetURN,
	}, neo4j.EagerResultTransformer)
	if err != nil {
		return nil, err
	}

	nodesMap := make(map[string]model.LineageNode)
	var edges []model.LineageEdge

	for _, record := range result.Records {
		nodeID, _ := record.Get("node_id")
		nodeType, _ := record.Get("node_type")
		nodeName, _ := record.Get("node_name")
		targetID, _ := record.Get("target_id")
		transform, _ := record.Get("transformation")

		srcID := nodeID.(string)
		nodesMap[srcID] = model.LineageNode{
			ID:   srcID,
			Type: nodeType.(string),
			Name: nodeName.(string),
		}

		tgtID := targetID.(string)
		targetType, _ := record.Get("target_type")
		targetName, _ := record.Get("target_name")
		nodesMap[tgtID] = model.LineageNode{
			ID:   tgtID,
			Type: targetType.(string),
			Name: targetName.(string),
		}

		edges = append(edges, model.LineageEdge{
			SourceID:   srcID,
			SourceType: nodeType.(string),
			TargetID:   tgtID,
			TargetType: targetType.(string),
			Transform:  transform.(string),
		})
	}

	var nodes []model.LineageNode
	for _, node := range nodesMap {
		nodes = append(nodes, node)
	}

	return &model.LineageGraph{Nodes: nodes, Edges: edges}, nil
}