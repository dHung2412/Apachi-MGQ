package model

import (
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	JobID        string    `json:"job_id" gorm:"uniqueIndex;not null"`
	DagID        string    `json:"dag_id" gorm:"index"`
	TaskID       string    `json:"task_id"`
	RunID        string    `json:"run_id"`
	Status       string    `json:"status"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	TriggeredBy  uuid.UUID `json:"triggered_by" gorm:"type:uuid"`
	DurationSecs float64   `json:"duration_seconds"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}