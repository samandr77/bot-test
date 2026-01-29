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
	Update(ctx context.Context, balance *models.Balance) error
	AddCredits(ctx context.Context, userID int64, modelType string, amount int) error
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	GetCredits(ctx context.Context, userID int64, modelType string) (int, error)
}

type Repo struct {
	DB *gorm.DB
}

func New(dsn string) (*Repo, error) {
	if migrateErr := RunMigrations(dsn); migrateErr != nil {
		return nil, fmt.Errorf("migration error: %w", migrateErr)
	}

	db, dbErr := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: false,
	})
	if dbErr != nil {
		return nil, fmt.Errorf("database connection error: %w", dbErr)
	}

	slog.Info("Database connection established")

	return &Repo{DB: db}, nil
}

func (r *Repo) GetBalance(ctx context.Context, userID int64) (*models.Balance, error) {
	var balance models.Balance
	firstErr := r.DB.WithContext(ctx).Where("user_id = ?", userID).First(&balance).Error
	if firstErr != nil {
		if firstErr == gorm.ErrRecordNotFound {
			return &models.Balance{UserID: userID}, nil
		}
		return nil, fmt.Errorf("failed to get balance: %w", firstErr)
	}
	return &balance, nil
}

func (r *Repo) SaveUser(ctx context.Context, user *models.User) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existingUser models.User
		findErr := tx.Where("id = ?", user.ID).First(&existingUser).Error
		if findErr != nil {
			if findErr == gorm.ErrRecordNotFound {
				if createErr := tx.Create(user).Error; createErr != nil {
					return fmt.Errorf("failed to create user: %w", createErr)
				}
				balance := &models.Balance{UserID: user.ID}
				if balanceErr := tx.Create(balance).Error; balanceErr != nil {
					return fmt.Errorf("failed to create balance: %w", balanceErr)
				}
				return nil
			}
			return fmt.Errorf("failed to find user: %w", findErr)
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
			if saveErr := tx.Save(&existingUser).Error; saveErr != nil {
				return fmt.Errorf("failed to update user: %w", saveErr)
			}
		}

		return nil
	})
}

func (r *Repo) UpdateUser(ctx context.Context, user *models.User) error {
	updateErr := r.DB.WithContext(ctx).Model(user).Updates(map[string]interface{}{
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
	}).Error
	if updateErr != nil {
		return fmt.Errorf("failed to update user: %w", updateErr)
	}
	return nil
}

func (r *Repo) Update(ctx context.Context, balance *models.Balance) error {
	saveErr := r.DB.WithContext(ctx).Save(balance).Error
	if saveErr != nil {
		return fmt.Errorf("failed to save balance: %w", saveErr)
	}
	return nil
}

func (r *Repo) AddCredits(ctx context.Context, userID int64, modelType string, amount int) error {
	slog.Info("AddCredits called", "userID", userID, "modelType", modelType, "amount", amount)

	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var balance models.Balance
		if getErr := tx.Where("user_id = ?", userID).FirstOrCreate(&balance, models.Balance{UserID: userID}).Error; getErr != nil {
			slog.Error("Failed to get/create balance in transaction", "error", getErr, "userID", userID)
			return fmt.Errorf("failed to get/create balance: %w", getErr)
		}

		slog.Info("Balance before update", "userID", userID, "gpt", balance.GptCredits, "sora", balance.SoraCredits, "nano", balance.NanobananaCredits)

		switch modelType {
		case "gpt":
			balance.GptCredits += amount
		case "sora2":
			balance.SoraCredits += amount
		case "nanobanana", "nanobanano":
			balance.NanobananaCredits += amount
		default:
			return fmt.Errorf("unknown model type: %s", modelType)
		}

		slog.Info("Balance after update", "userID", userID, "gpt", balance.GptCredits, "sora", balance.SoraCredits, "nano", balance.NanobananaCredits)

		if saveErr := tx.Save(&balance).Error; saveErr != nil {
			slog.Error("Failed to save balance", "error", saveErr, "userID", userID)
			return fmt.Errorf("failed to save balance: %w", saveErr)
		}

		slog.Info("Balance saved successfully", "userID", userID)
		return nil
	})
}

func (r *Repo) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	if findErr := r.DB.WithContext(ctx).First(&user, id).Error; findErr != nil {
		if findErr == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found: %w", findErr)
		}
		return nil, fmt.Errorf("failed to get user by id: %w", findErr)
	}
	return &user, nil
}

func (r *Repo) GetCredits(ctx context.Context, userID int64, modelType string) (int, error) {
	balance, balanceErr := r.GetBalance(ctx, userID)
	if balanceErr != nil {
		return 0, fmt.Errorf("failed to get credits: %w", balanceErr)
	}

	switch modelType {
	case "gpt":
		return balance.GptCredits, nil
	case "sora2":
		return balance.SoraCredits, nil
	case "nanobanana", "nanobanano":
		return balance.NanobananaCredits, nil
	default:
		return 0, fmt.Errorf("unknown model type: %s", modelType)
	}
}
