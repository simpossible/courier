package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisSessionStore implements SessionStore using Redis for multi-instance deployments.
// All sessions have a TTL based on maxAge and are automatically expired by Redis.
type RedisSessionStore struct {
	client *redis.Client
	prefix string // key prefix, e.g. "courier:session:"
	maxAge time.Duration
}

// RedisSessionStoreOption configures a RedisSessionStore.
type RedisSessionStoreOption func(*RedisSessionStore)

// WithRedisPrefix sets the key prefix in Redis. Default: "courier:session:".
func WithRedisPrefix(prefix string) RedisSessionStoreOption {
	return func(r *RedisSessionStore) {
		r.prefix = prefix
	}
}

// WithRedisMaxAge sets the session TTL. Default: 30 minutes.
// Sessions are automatically expired by Redis TTL.
func WithRedisMaxAge(d time.Duration) RedisSessionStoreOption {
	return func(r *RedisSessionStore) {
		r.maxAge = d
	}
}

// NewRedisSessionStore creates a session store backed by Redis.
func NewRedisSessionStore(client *redis.Client, opts ...RedisSessionStoreOption) *RedisSessionStore {
	r := &RedisSessionStore{
		client: client,
		prefix: "courier:session:",
		maxAge: 30 * time.Minute,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *RedisSessionStore) key(deviceID string) string {
	return r.prefix + deviceID
}

func (r *RedisSessionStore) Get(deviceID string) (*Session, error) {
	data, err := r.client.Get(context.Background(), r.key(deviceID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("courier/rpc: redis get session: %w", err)
	}

	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("courier/rpc: redis unmarshal session: %w", err)
	}
	return &sess, nil
}

func (r *RedisSessionStore) Set(deviceID string, sess *Session) error {
	sess.LastActive = time.Now()
	data, err := json.Marshal(sess)
	if err != nil {
		return fmt.Errorf("courier/rpc: redis marshal session: %w", err)
	}
	return r.client.Set(context.Background(), r.key(deviceID), data, r.maxAge).Err()
}

func (r *RedisSessionStore) Delete(deviceID string) error {
	return r.client.Del(context.Background(), r.key(deviceID)).Err()
}

func (r *RedisSessionStore) CleanExpired(maxAge time.Duration) error {
	// Redis handles TTL-based expiration automatically.
	// This is a no-op for RedisSessionStore.
	return nil
}
