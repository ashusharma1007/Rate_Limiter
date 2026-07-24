package store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr string) *RedisStore {
	return &RedisStore{
		client: redis.NewClient(&redis.Options{Addr: addr}),
	}
}

// IncrTotal increments the total request count for the given IP
func (s *RedisStore) IncrTotal(ctx context.Context, ip string) error {
	return s.client.Incr(ctx, "client:"+ip+":total").Err()
}

// IncrBlocked increments the blocked request count for the given IP
func (s *RedisStore) IncrBlocked(ctx context.Context, ip string) error {
	return s.client.Incr(ctx, "client:"+ip+":blocked").Err()
}

// AddUserID adds a userID to the set of accounts seen from the given IP
func (s *RedisStore) AddUserID(ctx context.Context, ip string, userID string) error {
	return s.client.SAdd(ctx, "ip:"+ip+":userids", userID).Err()
}

// IncrEndpoint increments the hit count for a specific endpoint for the given IP
func (s *RedisStore) IncrEndpoint(ctx context.Context, ip string, endpoint string) error {
	return s.client.HIncrBy(ctx, "client:"+ip+":endpoints", endpoint, 1).Err()
}

// SetTTL resets the 24h expiry on all keys for the given IP
func (s *RedisStore) SetTTL(ctx context.Context, ip string) error {
	keys := []string{
		"client:" + ip + ":total",
		"client:" + ip + ":blocked",
		"ip:" + ip + ":userids",
		"client:" + ip + ":endpoints",
	}
	for _, key := range keys {
		if err := s.client.Expire(ctx, key, 24*time.Hour).Err(); err != nil {
			return err
		}
	}
	return nil
}
