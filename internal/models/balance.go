package models

import "time"

type Balance struct {
	UserID            int64     `gorm:"primaryKey"`
	GptCredits        int       `gorm:"default:0"`
	SoraCredits       int       `gorm:"default:0"`
	NanobananaCredits int       `gorm:"default:0"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"`
}
