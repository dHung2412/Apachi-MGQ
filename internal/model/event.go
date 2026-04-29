package model

import "time"

type IngestEvent struct {
	EventType   string         `json:"event_type" validate:"required"`
	Timestamp   time.Time      `json:"timestamp"`
	Payload     IngestPayload  `json:"payload" validate:"required"`
	Source      string         `json:"source"`
	DagID       string         `json:"dag_id,omitempty"`
	RunID       string         `json:"run_id,omitempty"`
	TaskID      string         `json:"task_id,omitempty"`
}

type IngestPayload struct {
	Dataset *DatasetPayload `json:"dataset,omitempty"`
	Lineage *LineagePayload `json:"lineage,omitempty"`
	Job     *JobPayload     `json:"job,omitempty"`
	Tags    []TagPayload    `json:"tags,omitempty"`
}

type DatasetPayload struct {
	Name        string           `json:"name" validate:"required"`
	URN         string           `json:"urn" validate:"required"`
	Platform    string           `json:"platform"`
	Database    string           `json:"database"`
	Schema      string           `json:"schema"`
	TableType   string           `json:"table_type"`
	Description string           `json:"description"`
	Owner       OwnerPayload     `json:"owner"`
	Columns     []ColumnDef      `json:"columns"`
}

type OwnerPayload struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type LineagePayload struct {
	SourceDataset string          `json:"source_dataset" validate:"required"`
	TargetDataset string          `json:"target_dataset" validate:"required"`
	Mappings      []ColumnMapping `json:"mappings"`
	TransformType string          `json:"transform_type"`
	JobID         string          `json:"job_id"`
}

type JobPayload struct {
	JobID       string    `json:"job_id" validate:"required"`
	DagID       string    `json:"dag_id"`
	TaskID      string    `json:"task_id"`
	RunID       string    `json:"run_id"`
	Status      string    `json:"status"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	TriggeredBy string    `json:"triggered_by"`
}

type TagPayload struct {
	Key      string `json:"key" validate:"required"`
	Value    string `json:"value"`
	Category string `json:"category"`
}