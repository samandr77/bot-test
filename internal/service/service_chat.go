package service

import (
	"context"
	"fmt"

	"github.com/samandr77/bot-test/internal/ai"
	"github.com/samandr77/bot-test/internal/models"
	"github.com/samandr77/bot-test/internal/repository"
)

type ChatService interface {
	AddMessage(ctx context.Context, userID int64, role, content string) error
	GetHistory(ctx context.Context, userID int64, limit int) ([]models.ChatMessage, error)
	GetAIChatContext(ctx context.Context, userID int64) ([]ai.Message, error)
	ClearHistory(ctx context.Context, userID int64) error
}

type chatService struct {
	repo *repository.Repo
}

func NewChatService(repo *repository.Repo) ChatService {
	return &chatService{
		repo: repo,
	}
}

func (s *chatService) AddMessage(ctx context.Context, userID int64, role, content string) error {
	msg := &models.ChatMessage{
		UserID:  userID,
		Role:    role,
		Content: content,
	}
	saveErr := s.repo.SaveChatMessage(ctx, msg)
	if saveErr != nil {
		return fmt.Errorf("failed to save chat message in repo: %w", saveErr)
	}
	return nil
}

func (s *chatService) GetHistory(ctx context.Context, userID int64, limit int) ([]models.ChatMessage, error) {
	history, getErr := s.repo.GetChatHistory(ctx, userID, limit)
	if getErr != nil {
		return nil, fmt.Errorf("failed to get history from repo: %w", getErr)
	}
	return history, nil
}

const chatContextLimit = 10

func (s *chatService) GetAIChatContext(ctx context.Context, userID int64) ([]ai.Message, error) {
	history, historyErr := s.repo.GetChatHistory(ctx, userID, chatContextLimit)
	if historyErr != nil {
		return nil, fmt.Errorf("failed to get history for AI context: %w", historyErr)
	}

	messages := make([]ai.Message, len(history))
	for i, m := range history {
		messages[i] = ai.Message{
			Role:    m.Role,
			Content: m.Content,
		}
	}
	return messages, nil
}

func (s *chatService) ClearHistory(ctx context.Context, userID int64) error {
	clearErr := s.repo.ClearChatHistory(ctx, userID)
	if clearErr != nil {
		return fmt.Errorf("failed to clear history in repo: %w", clearErr)
	}
	return nil
}
