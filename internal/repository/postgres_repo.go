package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"DP_Maintenance/internal/model"
)

type DBDatasetRepository struct {
	db *gorm.DB
}

func NewDBDatasetRepository(db *gorm.DB) *DBDatasetRepository {
	return &DBDatasetRepository{db: db}
}

func (r *DBDatasetRepository) Create(dataset *model.Dataset) error {
	if dataset.ID == uuid.Nil {
		dataset.ID = uuid.New()
	}
	return r.db.Create(dataset).Error
}

func (r *DBDatasetRepository) GetByURN(urn string) (*model.Dataset, error) {
	var dataset model.Dataset
	err := r.db.Preload("Columns").Preload("Tags").Where("urn = ?", urn).First(&dataset).Error
	if err != nil {
		return nil, err
	}
	return &dataset, nil
}

func (r *DBDatasetRepository) List(offset, limit int) ([]model.Dataset, int64, error) {
	var datasets []model.Dataset
	var total int64

	r.db.Model(&model.Dataset{}).Count(&total)
	err := r.db.Preload("Columns").Preload("Tags").Offset(offset).Limit(limit).Find(&datasets).Error
	return datasets, total, err
}

func (r *DBDatasetRepository) Update(dataset *model.Dataset) error {
	return r.db.Save(dataset).Error
}

func (r *DBDatasetRepository) Delete(urn string) error {
	return r.db.Where("urn = ?", urn).Delete(&model.Dataset{}).Error
}

func (r *DBDatasetRepository) UpdateOwner(urn string, ownerID uuid.UUID) error {
	return r.db.Model(&model.Dataset{}).Where("urn = ?", urn).Update("owner_id", ownerID).Error
}

func (r *DBDatasetRepository) GetByID(id uuid.UUID) (*model.Dataset, error) {
	var dataset model.Dataset
	err := r.db.Preload("Columns").Preload("Tags").Where("id = ?", id).First(&dataset).Error
	if err != nil {
		return nil, err
	}
	return &dataset, nil
}

type DBUserRepository struct {
	db *gorm.DB
}

func NewDBUserRepository(db *gorm.DB) *DBUserRepository {
	return &DBUserRepository{db: db}
}

func (r *DBUserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

func (r *DBUserRepository) GetByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

func (r *DBUserRepository) Create(user *model.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	return r.db.Create(user).Error
}

func (r *DBUserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

func (r *DBUserRepository) List(offset, limit int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	r.db.Model(&model.User{}).Count(&total)
	err := r.db.Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}

type DBSchemaVersionRepository struct {
	db *gorm.DB
}

func NewDBSchemaVersionRepository(db *gorm.DB) *DBSchemaVersionRepository {
	return &DBSchemaVersionRepository{db: db}
}

func (r *DBSchemaVersionRepository) GetLatest(datasetID uuid.UUID) (*model.SchemaVersion, error) {
	var version model.SchemaVersion
	err := r.db.Preload("Columns").Where("dataset_id = ?", datasetID).Order("version DESC").First(&version).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *DBSchemaVersionRepository) Create(version *model.SchemaVersion) error {
	if version.ID == uuid.Nil {
		version.ID = uuid.New()
	}
	return r.db.Create(version).Error
}

type DBTagRepository struct {
	db *gorm.DB
}

func NewDBTagRepository(db *gorm.DB) *DBTagRepository {
	return &DBTagRepository{db: db}
}

func (r *DBTagRepository) GetByKey(key string) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.Where("key = ?", key).First(&tag).Error
	if err != nil {
		return nil, fmt.Errorf("tag not found: %w", err)
	}
	return &tag, nil
}

func (r *DBTagRepository) Create(tag *model.Tag) error {
	if tag.ID == uuid.Nil {
		tag.ID = uuid.New()
	}
	return r.db.Create(tag).Error
}

func (r *DBTagRepository) AssignToDataset(datasetID uuid.UUID, tagID uuid.UUID) error {
	return r.db.Exec("INSERT INTO dataset_tags (dataset_id, tag_id) VALUES (?, ?) ON CONFLICT DO NOTHING", datasetID, tagID).Error
}

type DBJobRepository struct {
	db *gorm.DB
}

func NewDBJobRepository(db *gorm.DB) *DBJobRepository {
	return &DBJobRepository{db: db}
}

func (r *DBJobRepository) GetByJobID(jobID string) (*model.Job, error) {
	var job model.Job
	err := r.db.Where("job_id = ?", jobID).First(&job).Error
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}
	return &job, nil
}

func (r *DBJobRepository) Create(job *model.Job) error {
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}
	return r.db.Create(job).Error
}