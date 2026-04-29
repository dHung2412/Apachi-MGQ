package model

import "time"

type LineageEdge struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	SourceDataset  string    `json:"source_dataset" gorm:"index;not null"`
	TargetDataset  string    `json:"target_dataset" gorm:"index;not null"`
	SourceColumn   string    `json:"source_column"`
	TargetColumn   string    `json:"target_column"`
	TransformType  string    `json:"transform_type"`
	JobID          string    `json:"job_id" gorm:"index"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ColumnMapping struct {
	SourceColumn string `json:"source_column"`
	TargetColumn string `json:"target_column"`
	Transform    string `json:"transform"`
}