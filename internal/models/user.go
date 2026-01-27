package models

import "time"

type User struct {
	ID        int64     `gorm:"primaryKey"`
	Username  string    `gorm:"type:text"`
	FirstName string    `gorm:"type:text"`
	LastName  string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	Balance   Balance   `gorm:"foreignKey:UserID"`
}
