package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/samandr77/bot-test/internal/models"
)

const (
	redisMaxRetries     = 3
	redisRetryBaseDelay = 100 * time.Millisecond
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
	opt, parseErr := redis.ParseURL(redisURL)
	if parseErr != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", parseErr)
	}

	opt.DialTimeout = 10 * time.Second
	opt.ReadTimeout = 5 * time.Second
	opt.WriteTimeout = 5 * time.Second
	opt.PoolTimeout = 10 * time.Second
	opt.MaxRetries = 3
	opt.MinRetryBackoff = 100 * time.Millisecond
	opt.MaxRetryBackoff = 500 * time.Millisecond

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if pingErr := client.Ping(ctx).Err(); pingErr != nil {
		return nil, fmt.Errorf("redis connection failed: %w", pingErr)
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
	var data []byte
	var lastErr error

	for attempt := 1; attempt <= redisMaxRetries; attempt++ {
		data, lastErr = r.client.Get(ctx, r.key(userID)).Bytes()
		if lastErr == nil {
			break
		}
		if lastErr == redis.Nil {
			return models.NewUserState(), nil
		}
		if attempt < redisMaxRetries {
			slog.Warn("Redis Get failed, retrying",
				"attempt", attempt,
				"max_retries", redisMaxRetries,
				"user_id", userID,
				"error", lastErr,
			)
			time.Sleep(redisRetryBaseDelay * time.Duration(attempt))
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to get state from redis after %d retries: %w", redisMaxRetries, lastErr)
	}

	var state models.UserState
	if unmarshalErr := json.Unmarshal(data, &state); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal user state: %w", unmarshalErr)
	}

	return &state, nil
}

func (r *RedisStateRepository) Save(ctx context.Context, userID int64, state *models.UserState) error {
	data, marshalErr := json.Marshal(state)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal user state: %w", marshalErr)
	}

	var lastErr error
	for attempt := 1; attempt <= redisMaxRetries; attempt++ {
		lastErr = r.client.Set(ctx, r.key(userID), data, r.ttl).Err()
		if lastErr == nil {
			return nil
		}
		if attempt < redisMaxRetries {
			slog.Warn("Redis Save failed, retrying",
				"attempt", attempt,
				"max_retries", redisMaxRetries,
				"user_id", userID,
				"error", lastErr,
			)
			time.Sleep(redisRetryBaseDelay * time.Duration(attempt))
		}
	}

	return fmt.Errorf("failed to set state in redis after %d retries: %w", redisMaxRetries, lastErr)
}

func (r *RedisStateRepository) Close() error {
	closeErr := r.client.Close()
	if closeErr != nil {
		return fmt.Errorf("failed to close redis client: %w", closeErr)
	}
	return nil
}
