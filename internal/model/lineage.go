package model

import (
	"time"

	"github.com/google/uuid"
)

type LineageEdge struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	SourceID    string  `json:"source_id" gorm:"index"`
	SourceType string  `json:"source_type"`
	TargetID   string  `json:"target_id" gorm:"index"`
	TargetType string  `json:"target_type"`
	Transform  string  `json:"transform"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ColumnMapping struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	SourceDataset  string    `json:"source_dataset" gorm:"index"`
	SourceColumn  string    `json:"source_column"`
	TargetDataset string    `json:"target_dataset" gorm:"index"`
	TargetColumn  string    `json:"target_column"`
	TransformRule string    `json:"transform_rule"`
	JobID          string    `json:"job_id" gorm:"index"`
	CreatedAt      time.Time `json:"created_at"`
}

type LineageGraph struct {
	Nodes []LineageNode `json:"nodes"`
	Edges []LineageEdge `json:"edges"`
}

type LineageNode struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}