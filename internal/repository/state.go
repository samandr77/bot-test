package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/samandr77/bot-test/internal/models"
)

type StateRepository interface {
	Get(ctx context.Context, userID int64) (*models.UserState, error)
	Save(ctx context.Context, userID int64, state *models.UserState) error
}

type RedisStateRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisStateRepository(redisURL string) (*RedisStateRepository, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &RedisStateRepository{
		client: client,
		ttl:    24 * time.Hour,
	}, nil
}

func (r *RedisStateRepository) key(userID int64) string {
	return fmt.Sprintf("user_state:%d", userID)
}

func (r *RedisStateRepository) Get(ctx context.Context, userID int64) (*models.UserState, error) {
	data, err := r.client.Get(ctx, r.key(userID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return models.NewUserState(), nil
		}
		return nil, err
	}

	var state models.UserState
	if err := json.Unmarshal(data, &state); err != nil {
		return models.NewUserState(), nil
	}

	return &state, nil
}

func (r *RedisStateRepository) Save(ctx context.Context, userID int64, state *models.UserState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(userID), data, r.ttl).Err()
}

func (r *RedisStateRepository) Close() error {
	return r.client.Close()
}
