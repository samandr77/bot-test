package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/samandr77/bot-test/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository interface {
	GetBalance(ctx context.Context, userID int64) (*models.Balance, error)
	SaveUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
	AddCredits(ctx context.Context, userID int64, modelType string, amount int) error
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
}

type Repo struct {
	DB *gorm.DB
}

func New(dsn string) (*Repo, error) {
	if err := RunMigrations(dsn); err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: false,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	slog.Info("Успешное подключение к базе данных")

	return &Repo{DB: db}, nil
}

func (r *Repo) GetBalance(ctx context.Context, userID int64) (*models.Balance, error) {
	var balance models.Balance
	err := r.DB.WithContext(ctx).Where("user_id = ?", userID).First(&balance).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &models.Balance{UserID: userID}, nil
		}
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return &balance, nil
}

func (r *Repo) SaveUser(ctx context.Context, user *models.User) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existingUser models.User
		err := tx.Where("id = ?", user.ID).First(&existingUser).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				if errCreate := tx.Create(user).Error; errCreate != nil {
					return fmt.Errorf("ошибка создания пользователя: %w", errCreate)
				}
				balance := &models.Balance{UserID: user.ID}
				if errBalance := tx.Create(balance).Error; errBalance != nil {
					return fmt.Errorf("ошибка создания баланса: %w", errBalance)
				}
				return nil
			}
			return fmt.Errorf("ошибка поиска пользователя: %w", err)
		}

		hasChanges := false
		if existingUser.Username != user.Username {
			existingUser.Username = user.Username
			hasChanges = true
		}
		if existingUser.FirstName != user.FirstName {
			existingUser.FirstName = user.FirstName
			hasChanges = true
		}
		if existingUser.LastName != user.LastName {
			existingUser.LastName = user.LastName
			hasChanges = true
		}

		if hasChanges {
			if err := tx.Save(&existingUser).Error; err != nil {
				return fmt.Errorf("ошибка обновления пользователя: %w", err)
			}
		}

		return nil
	})
}

func (r *Repo) UpdateUser(ctx context.Context, user *models.User) error {
	return r.DB.WithContext(ctx).Model(user).Updates(map[string]interface{}{
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
	}).Error
}

func (r *Repo) AddCredits(ctx context.Context, userID int64, modelType string, amount int) error {
	column := ""
	switch modelType {
	case "gpt":
		column = "gpt_credits"
	case "sora2":
		column = "sora_credits"
	case "nanobanano":
		column = "nanobanano_credits"
	default:
		return fmt.Errorf("неизвестный тип модели: %s", modelType)
	}

	return r.DB.WithContext(ctx).Model(&models.Balance{}).
		Where("user_id = ?", userID).
		UpdateColumn(column, gorm.Expr(column+" + ?", amount)).Error
}

func (r *Repo) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	if err := r.DB.WithContext(ctx).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &user, nil
}
