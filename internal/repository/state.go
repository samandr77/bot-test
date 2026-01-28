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
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
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
	var data []byte
	var err error

	for attempt := 1; attempt <= redisMaxRetries; attempt++ {
		data, err = r.client.Get(ctx, r.key(userID)).Bytes()
		if err == nil {
			break
		}
		if err == redis.Nil {
			return models.NewUserState(), nil
		}
		if attempt < redisMaxRetries {
			slog.Warn("Redis Get failed, retrying",
				"attempt", attempt,
				"max_retries", redisMaxRetries,
				"user_id", userID,
				"error", err,
			)
			time.Sleep(redisRetryBaseDelay * time.Duration(attempt))
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get state from redis after %d retries: %w", redisMaxRetries, err)
	}

	var state models.UserState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user state: %w", err)
	}

	return &state, nil
}

func (r *RedisStateRepository) Save(ctx context.Context, userID int64, state *models.UserState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal user state: %w", err)
	}

	for attempt := 1; attempt <= redisMaxRetries; attempt++ {
		err = r.client.Set(ctx, r.key(userID), data, r.ttl).Err()
		if err == nil {
			return nil
		}
		if attempt < redisMaxRetries {
			slog.Warn("Redis Save failed, retrying",
				"attempt", attempt,
				"max_retries", redisMaxRetries,
				"user_id", userID,
				"error", err,
			)
			time.Sleep(redisRetryBaseDelay * time.Duration(attempt))
		}
	}

	return fmt.Errorf("failed to set state in redis after %d retries: %w", redisMaxRetries, err)
}

func (r *RedisStateRepository) Close() error {
	return r.client.Close()
}
