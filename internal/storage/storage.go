package storage

import (
	"fmt"
	"log/slog"

	"github.com/samandr77/bot-test/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Storage struct {
	DB *gorm.DB
}

func NewStorage(dsn string) (*Storage, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	slog.Info("Успешное подключение к базе данных")

	err = db.AutoMigrate(&models.User{}, &models.Balance{}, &models.GenerationHistory{})
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении автомиграции: %w", err)
	}

	slog.Info("Миграция базы данных завершена")

	return &Storage{DB: db}, nil
}
