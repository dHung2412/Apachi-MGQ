package model

import (
	"time"

	"github.com/google/uuid"
)

type Dataset struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`// `
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	URN         string    `json:"urn" gorm:"uniqueIndex;not null"`
	Description string    `json:"description"`
	Platform    string    `json:"platform"`
	Database    string    `json:"database"`
	Schema      string    `json:"schema"`
	TableType   string    `json:"table_type"`
	OwnerID     uuid.UUID `json:"owner_id" gorm:"type:uuid"`
	Tags        []Tag     `json:"tags" gorm:"many2many:dataset_tags;"`
	Columns     []ColumnDef `json:"columns" gorm:"foreignKey:DatasetID"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ColumnDef struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	DatasetID   uuid.UUID `json:"dataset_id" gorm:"type:uuid;index"`
	Name        string `json:"name" gorm:"not null"`
	DataType    string `json:"data_type"`
	Description string `json:"description"`
	IsNullable  bool   `json:"is_nullable"`
	IsPrimary   bool   `json:"is_primary"`
	IsForeign   bool   `json:"is_foreign"`
	Position    int    `json:"position"`
}

type SchemaVersion struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	DatasetID uuid.UUID `json:"dataset_id" gorm:"type:uuid;index;not null"`
	Version   int      `json:"version"`
	Columns   []ColumnDef `json:"columns" gorm:"foreignKey:DatasetID"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy uuid.UUID `json:"created_by" gorm:"type:uuid"`
	ChangeLog string    `json:"change_log"`
}