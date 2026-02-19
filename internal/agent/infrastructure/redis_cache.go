package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/agent/domain"
)

const (
	// messageCacheSize is the number of recent messages cached per conversation.
	messageCacheSize = 20
	// messageCacheTTL is the cache expiry; refreshed on each write.
	messageCacheTTL = 2 * time.Hour
)

// RedisMessageCache caches the last N messages per conversation to reduce DB reads during chat.
type RedisMessageCache struct {
	client *redis.Client
}

// NewRedisMessageCache connects to Redis using the provided URL.
func NewRedisMessageCache(redisURL string) (*RedisMessageCache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("redis_cache: parse url: %w", err)
	}
	return &RedisMessageCache{client: redis.NewClient(opts)}, nil
}

// cacheKey returns the Redis list key for a conversation's messages.
func cacheKey(conversationID uuid.UUID) string {
	return fmt.Sprintf("agent:conv:%s:messages", conversationID)
}

// GetMessages retrieves cached messages for a conversation (oldest-first).
// Returns nil, nil if cache miss.
func (c *RedisMessageCache) GetMessages(ctx context.Context, conversationID uuid.UUID) ([]*domain.Message, error) {
	key := cacheKey(conversationID)
	items, err := c.client.LRange(ctx, key, 0, -1).Result()
	if err == redis.Nil || len(items) == 0 {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis_cache: get messages: %w", err)
	}

	messages := make([]*domain.Message, 0, len(items))
	for _, item := range items {
		var m domain.Message
		if err := json.Unmarshal([]byte(item), &m); err != nil {
			return nil, fmt.Errorf("redis_cache: unmarshal message: %w", err)
		}
		messages = append(messages, &m)
	}
	return messages, nil
}

// AddMessage appends a message to the cache list, trimming to messageCacheSize.
// Refreshes the TTL on each write.
func (c *RedisMessageCache) AddMessage(ctx context.Context, conversationID uuid.UUID, m *domain.Message) error {
	key := cacheKey(conversationID)

	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("redis_cache: marshal message: %w", err)
	}

	pipe := c.client.Pipeline()
	pipe.RPush(ctx, key, string(data))
	pipe.LTrim(ctx, key, -messageCacheSize, -1)
	pipe.Expire(ctx, key, messageCacheTTL)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis_cache: add message: %w", err)
	}
	return nil
}

// InvalidateConversation removes a conversation's cached messages.
func (c *RedisMessageCache) InvalidateConversation(ctx context.Context, conversationID uuid.UUID) error {
	return c.client.Del(ctx, cacheKey(conversationID)).Err()
}

// Ping checks Redis connectivity.
func (c *RedisMessageCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
