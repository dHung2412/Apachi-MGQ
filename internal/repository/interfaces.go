package repository

import (
	"github.com/google/uuid"

	"DP_Maintenance/internal/model"
)

type DatasetRepository interface {
	Create(dataset *model.Dataset) error
	GetByURN(urn string) (*model.Dataset, error)
	List(offset, limit int) ([]model.Dataset, int64, error)
	Update(dataset *model.Dataset) error
	Delete(urn string) error
	UpdateOwner(urn string, ownerID uuid.UUID) error
}

type SchemaVersionRepository interface {
	GetLatest(datasetID uuid.UUID) (*model.SchemaVersion, error)
	Create(version *model.SchemaVersion) error
}

type UserRepository interface {
	GetByUsername(username string) (*model.User, error)
	GetByID(id uuid.UUID) (*model.User, error)
	Create(user *model.User) error
}

type TagRepository interface {
	GetByKey(key string) (*model.Tag, error)
	Create(tag *model.Tag) error
	AssignToDataset(datasetID uuid.UUID, tagID uuid.UUID) error
}

type JobRepository interface {
	GetByJobID(jobID string) (*model.Job, error)
	Create(job *model.Job) error
}

type LineageRepository interface {
	RecordColumnMapping(sourceDataset string, targetDataset string, mapping *model.ColumnMapping, jobID string) error
	TraceUpstream(datasetURN, columnName string, depth int) ([]model.LineageEdge, error)
	TraceDownstream(datasetURN, columnName string, depth int) ([]model.LineageEdge, error)
	GetDatasetLineage(datasetURN string) ([]model.LineageEdge, []model.LineageEdge, error)
	RecordEdge(edge *model.LineageEdge) error
	GetLineageGraph(datasetURN string) (*model.LineageGraph, error)
}