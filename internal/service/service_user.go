package service

import (
	"context"
	"fmt"

	"github.com/samandr77/bot-test/internal/models"
	"github.com/samandr77/bot-test/internal/repository"
)

const (
	ModelGPT        = "gpt"
	ModelSora       = "sora2"
	ModelNanoBanana = "nanobanana"
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
	balance, getErr := s.repo.GetBalance(ctx, userID)
	if getErr != nil {
		return "", fmt.Errorf("failed to get balance from repo: %w", getErr)
	}

	return fmt.Sprintf("Ваш баланс:\n- GPT: %d\n- Sora 2: %d\n- NanoBanana: %d",
		balance.GptCredits, balance.SoraCredits, balance.NanobananaCredits), nil
}

func (s *service) SaveUser(ctx context.Context, user *models.User) error {
	saveErr := s.repo.SaveUser(ctx, user)
	if saveErr != nil {
		return fmt.Errorf("failed to save user in repo: %w", saveErr)
	}
	return nil
}

func (s *service) UpdateUser(ctx context.Context, user *models.User) error {
	updateErr := s.repo.UpdateUser(ctx, user)
	if updateErr != nil {
		return fmt.Errorf("failed to update user in repo: %w", updateErr)
	}
	return nil
}

func (s *service) AddCredits(ctx context.Context, userID int64, modelType string, amount int) error {
	addErr := s.repo.AddCredits(ctx, userID, modelType, amount)
	if addErr != nil {
		return fmt.Errorf("failed to add credits in repo: %w", addErr)
	}
	return nil
}

func (s *service) HasCredits(ctx context.Context, userID int64, modelType string) (bool, error) {
	credits, getErr := s.repo.GetCredits(ctx, userID, modelType)
	if getErr != nil {
		return false, fmt.Errorf("failed to get credits from repo: %w", getErr)
	}
	return credits > 0, nil
}

func (s *service) UseCredits(ctx context.Context, userID int64, modelType string) error {
	balance, getErr := s.repo.GetBalance(ctx, userID)
	if getErr != nil {
		return fmt.Errorf("failed to get balance: %w", getErr)
	}

	switch modelType {
	case ModelGPT:
		if balance.GptCredits > 0 {
			balance.GptCredits--
		}
	case ModelSora:
		if balance.SoraCredits > 0 {
			balance.SoraCredits--
		}
	case ModelNanoBanana:
		if balance.NanobananaCredits > 0 {
			balance.NanobananaCredits--
		}
	default:
		return fmt.Errorf("unknown model type: %s", modelType)
	}

	if updateErr := s.repo.Update(ctx, balance); updateErr != nil {
		return fmt.Errorf("failed to update balance in repo: %w", updateErr)
	}

	return nil
}
