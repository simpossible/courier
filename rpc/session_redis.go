package rpc

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/simpossible/courier/rpc/internal/sessionpb"
	"google.golang.org/protobuf/proto"
)

// RedisSessionStore implements SessionStore using Redis for multi-instance deployments.
// Sessions are serialized with protobuf for high performance.
type RedisSessionStore struct {
	client *redis.Client
	prefix string
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

func (r *RedisSessionStore) key(clientID string) string {
	return r.prefix + clientID
}

func (r *RedisSessionStore) Get(clientID string) (*Session, error) {
	data, err := r.client.Get(context.Background(), r.key(clientID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("courier/rpc: redis get session: %w", err)
	}

	pb := &sessionpb.SessionData{}
	if err := proto.Unmarshal(data, pb); err != nil {
		return nil, fmt.Errorf("courier/rpc: proto unmarshal session: %w", err)
	}

	return &Session{
		UserID:     pb.UserId,
		Data:       pb.Data,
		CreatedAt:  time.Unix(0, pb.CreatedAt),
		LastActive: time.Unix(0, pb.LastActive),
	}, nil
}

func (r *RedisSessionStore) Set(clientID string, sess *Session) error {
	sess.LastActive = time.Now()
	pb := &sessionpb.SessionData{
		UserId:     sess.UserID,
		Data:       sess.Data,
		CreatedAt:  sess.CreatedAt.UnixNano(),
		LastActive: sess.LastActive.UnixNano(),
	}
	data, err := proto.Marshal(pb)
	if err != nil {
		return fmt.Errorf("courier/rpc: proto marshal session: %w", err)
	}
	return r.client.Set(context.Background(), r.key(clientID), data, r.maxAge).Err()
}

func (r *RedisSessionStore) Delete(clientID string) error {
	return r.client.Del(context.Background(), r.key(clientID)).Err()
}

func (r *RedisSessionStore) CleanExpired(maxAge time.Duration) error {
	return nil
}
