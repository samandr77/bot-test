package models

import (
	"time"

	"github.com/google/uuid"
)

type GenerationHistory struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    int64     `gorm:"type:bigint;not null"`
	ModelType string    `gorm:"type:text;not null"`
	Prompt    string    `gorm:"type:text;not null"`
	MediaURL  string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
