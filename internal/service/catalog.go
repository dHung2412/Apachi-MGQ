package service

import (
	"DP_Maintenance/internal/model"
	"DP_Maintenance/internal/repository"
	"fmt"
	"log"

	"github.com/google/uuid"
)

// CatalogService handles metadata business logic: dataset CRUD,
// schema versioning, ownership, and tag management.
type CatalogService struct {
	datasetRepo repository.DatasetRepository
	schemaRepo  repository.SchemaVersionRepository
	userRepo    repository.UserRepository
	tagRepo     repository.TagRepository
}

// NewCatalogService creates a CatalogService with injected repositories.
// Pass nil for repos that aren't available yet (stub mode).
func NewCatalogService(
	datasetRepo repository.DatasetRepository,
	schemaRepo repository.SchemaVersionRepository,
	userRepo repository.UserRepository,
	tagRepo repository.TagRepository,
) *CatalogService {
	return &CatalogService{
		datasetRepo: datasetRepo,
		schemaRepo:  schemaRepo,
		userRepo:    userRepo,
		tagRepo:     tagRepo,
	}
}

// --- Dataset Operations ---

// CreateDataset creates a new dataset entry with its columns.
func (s *CatalogService) CreateDataset(dataset *model.Dataset) error {
	if s.datasetRepo == nil {
		log.Printf("[CATALOG] Stub: CreateDataset(%s)", dataset.URN)
		return nil
	}
	return s.datasetRepo.Create(dataset)
}

// GetDataset retrieves a dataset by URN.
func (s *CatalogService) GetDataset(urn string) (*model.Dataset, error) {
	if s.datasetRepo == nil {
		log.Printf("[CATALOG] Stub: GetDataset(%s)", urn)
		return nil, fmt.Errorf("dataset %s not found (stub mode)", urn)
	}
	return s.datasetRepo.GetByURN(urn)
}

// ListDatasets returns a paginated list of datasets.
func (s *CatalogService) ListDatasets(page, pageSize int) ([]model.Dataset, int64, error) {
	if s.datasetRepo == nil {
		log.Printf("[CATALOG] Stub: ListDatasets(page=%d, size=%d)", page, pageSize)
		return []model.Dataset{}, 0, nil
	}
	offset := (page - 1) * pageSize
	return s.datasetRepo.List(offset, pageSize)
}

// UpdateDataset updates an existing dataset's metadata.
func (s *CatalogService) UpdateDataset(dataset *model.Dataset) error {
	if s.datasetRepo == nil {
		log.Printf("[CATALOG] Stub: UpdateDataset(%s)", dataset.URN)
		return nil
	}
	return s.datasetRepo.Update(dataset)
}

// DeleteDataset removes a dataset by URN.
func (s *CatalogService) DeleteDataset(urn string) error {
	if s.datasetRepo == nil {
		log.Printf("[CATALOG] Stub: DeleteDataset(%s)", urn)
		return nil
	}
	return s.datasetRepo.Delete(urn)
}

// --- Ownership ---

// UpdateOwner changes the owner of a dataset.
func (s *CatalogService) UpdateOwner(urn string, ownerID uuid.UUID) error {
	if s.datasetRepo == nil {
		log.Printf("[CATALOG] Stub: UpdateOwner(%s, %s)", urn, ownerID)
		return nil
	}
	return s.datasetRepo.UpdateOwner(urn, ownerID)
}

// --- Schema Versioning ---

// RecordSchemaVersion saves a new schema version and computes the diff.
func (s *CatalogService) RecordSchemaVersion(datasetID uuid.UUID, columns []model.ColumnDef, changedBy uuid.UUID) error {
	if s.schemaRepo == nil {
		log.Printf("[CATALOG] Stub: RecordSchemaVersion(dataset=%s, columns=%d)", datasetID, len(columns))
		return nil
	}

	// Get latest version number
	latest, _ := s.schemaRepo.GetLatest(datasetID)
	newVersion := 1
	if latest != nil {
		newVersion = latest.Version + 1
	}

	version := &model.SchemaVersion{
		DatasetID: datasetID,
		Version:   newVersion,
		CreatedBy: changedBy,
		ChangeLog: fmt.Sprintf("Schema version %d recorded", newVersion),
	}

	return s.schemaRepo.Create(version)
}

// --- Tags ---

// AssignTag assigns a tag to a dataset.
func (s *CatalogService) AssignTag(datasetID uuid.UUID, key, value string) error {
	if s.tagRepo == nil {
		log.Printf("[CATALOG] Stub: AssignTag(dataset=%s, %s=%s)", datasetID, key, value)
		return nil
	}

	tag, err := s.tagRepo.GetByKey(key)
	if err != nil {
		// Create tag if it doesn't exist
		tag = &model.Tag{Key: key, Value: value}
		if err := s.tagRepo.Create(tag); err != nil {
			return err
		}
	}

	return s.tagRepo.AssignToDataset(datasetID, tag.ID)
}
