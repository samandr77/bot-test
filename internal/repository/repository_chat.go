package repository

import (
	"context"
	"fmt"

	"github.com/samandr77/bot-test/internal/models"
)

func (r *Repo) SaveChatMessage(ctx context.Context, msg *models.ChatMessage) error {
	createErr := r.DB.WithContext(ctx).Create(msg).Error
	if createErr != nil {
		return fmt.Errorf("failed to create chat message: %w", createErr)
	}
	return nil
}

func (r *Repo) GetChatHistory(ctx context.Context, userID int64, limit int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	findErr := r.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(limit).
		Find(&messages).Error

	if findErr != nil {
		return nil, fmt.Errorf("failed to find chat history: %w", findErr)
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *Repo) ClearChatHistory(ctx context.Context, userID int64) error {
	deleteErr := r.DB.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.ChatMessage{}).Error
	if deleteErr != nil {
		return fmt.Errorf("failed to clear chat history: %w", deleteErr)
	}
	return nil
}
