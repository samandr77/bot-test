package models

import (
	"time"

	"github.com/google/uuid"
)

type ChatMessage struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    int64     `gorm:"type:bigint;not null;index"`
	Role      string    `gorm:"type:text;not null"`
	Content   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
