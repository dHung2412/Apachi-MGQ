package model

import (
	"time"

	"github.com/google/uuid"
)

type Tag struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Key      string   `json:"key" gorm:"uniqueIndex;not null"`
	Value    string   `json:"value"`
	Category string   `json:"category"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}