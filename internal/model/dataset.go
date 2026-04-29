package model

import "time"

type Dataset struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	URN         string    `json:"urn" gorm:"uniqueIndex;not null"`
	Description string    `json:"description"`
	Platform    string    `json:"platform"`
	Database    string    `json:"database"`
	Schema      string    `json:"schema"`
	TableType   string    `json:"table_type"`
	OwnerID     uint      `json:"owner_id"`
	Tags        []Tag     `json:"tags" gorm:"many2many:dataset_tags;"`
	Columns     []ColumnDef `json:"columns" gorm:"foreignKey:DatasetID"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ColumnDef struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	DatasetID   uint   `json:"dataset_id" gorm:"index"`
	Name        string `json:"name" gorm:"not null"`
	DataType    string `json:"data_type"`
	Description string `json:"description"`
	IsNullable  bool   `json:"is_nullable"`
	IsPrimary   bool   `json:"is_primary"`
	IsForeign   bool   `json:"is_foreign"`
	Position    int    `json:"position"`
}

type SchemaVersion struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	DatasetID uint      `json:"dataset_id" gorm:"index;not null"`
	Version   int       `json:"version"`
	Columns   []ColumnDef `json:"columns" gorm:"foreignKey:DatasetID"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
	ChangeLog string    `json:"change_log"`
}