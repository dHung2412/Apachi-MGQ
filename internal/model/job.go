package model

import "time"

type Job struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	JobID       string    `json:"job_id" gorm:"uniqueIndex;not null"`
	DagID       string    `json:"dag_id" gorm:"index"`
	TaskID      string    `json:"task_id"`
	RunID       string    `json:"run_id"`
	Status      string    `json:"status"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	TriggeredBy string    `json:"triggered_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}