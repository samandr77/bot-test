package service

import (
	"context"
	"fmt"

	"github.com/samandr77/bot-test/internal/models"
	"github.com/samandr77/bot-test/internal/repository"
)

type Service interface {
	GetBalance(ctx context.Context, userID int64) (string, error)
	SaveUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
	AddCredits(ctx context.Context, userID int64, modelType string, amount int) error
	HasCredits(ctx context.Context, userID int64, modelType string) (bool, error)
	UseCredits(ctx context.Context, userID int64, modelType string) error
}

type service struct {
	repo repository.Repository
}

func New(r repository.Repository) Service {
	return &service{
		repo: r,
	}
}

func (s *service) GetBalance(ctx context.Context, userID int64) (string, error) {
	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Ваш баланс:\n- GPT: %d\n- Sora 2: %d\n- NanoBanana: %d",
		balance.GptCredits, balance.SoraCredits, balance.NanobananaCredits), nil
}

func (s *service) SaveUser(ctx context.Context, user *models.User) error {
	return s.repo.SaveUser(ctx, user)
}

func (s *service) UpdateUser(ctx context.Context, user *models.User) error {
	return s.repo.UpdateUser(ctx, user)
}

func (s *service) AddCredits(ctx context.Context, userID int64, modelType string, amount int) error {
	return s.repo.AddCredits(ctx, userID, modelType, amount)
}

func (s *service) HasCredits(ctx context.Context, userID int64, modelType string) (bool, error) {
	credits, err := s.repo.GetCredits(ctx, userID, modelType)
	if err != nil {
		return false, err
	}
	return credits > 0, nil
}

func (s *service) UseCredits(ctx context.Context, userID int64, modelType string) error {
	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	switch modelType {
	case "gpt":
		if balance.GptCredits > 0 {
			balance.GptCredits--
		}
	case "sora2":
		if balance.SoraCredits > 0 {
			balance.SoraCredits--
		}
	case "nanobanana", "nanobanano":
		if balance.NanobananaCredits > 0 {
			balance.NanobananaCredits--
		}
	default:
		return fmt.Errorf("unknown model type: %s", modelType)
	}

	if err := s.repo.Update(ctx, balance); err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	return nil
}
