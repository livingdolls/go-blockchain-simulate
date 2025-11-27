package redis

import (
	"context"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/redis/go-redis/v9"
)

type MemoryAdapter interface {
	Get(ctx context.Context, key string) ([]byte, bool)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration)
	Del(ctx context.Context, key string)
	InvalidatePattern(ctx context.Context, pattern string) error
}

type cacheItem struct {
	value     []byte
	expiresAt time.Time
}

type memoryAdapter struct {
	redis *redis.Client
	local *lru.Cache[string, cacheItem]
}

func NewMemoryAdapter(redisClient *redis.Client, localSize int) (MemoryAdapter, error) {
	localCache, err := lru.New[string, cacheItem](localSize)

	if err != nil {
		return nil, err
	}

	return &memoryAdapter{
		redis: redisClient,
		local: localCache,
	}, nil
}

// Del implements port.MemoryAdapter.
func (m *memoryAdapter) Del(ctx context.Context, key string) {
	// delete from lru cache
	m.local.Remove(key)

	// delete from redis
	if m.redis != nil {
		_ = m.redis.Del(ctx, key).Err()
	}
}

// Get implements port.MemoryAdapter.
func (m *memoryAdapter) Get(ctx context.Context, key string) ([]byte, bool) {
	// try lru cache first
	if v, ok := m.local.Get(key); ok {
		if time.Now().Before(v.expiresAt) {
			return v.value, true
		}

		// expired, delete
		m.local.Remove(key)
	}

	// then try redis
	if m.redis != nil {
		// get from redis
		b, err := m.redis.Get(ctx, key).Bytes()

		if err == nil {
			ttl := 5 * time.Minute

			// set to lru cache
			m.local.Add(key, cacheItem{
				value:     b,
				expiresAt: time.Now().Add(ttl),
			})

			return b, true
		}
	}

	// not found
	return nil, false
}

// InvalidatePattern implements port.MemoryAdapter.
func (m *memoryAdapter) InvalidatePattern(ctx context.Context, pattern string) error {
	if m.redis == nil {
		return nil
	}

	// scan keys by pattern
	// 0 first cursor -> 1000 max keys per scan
	iter := m.redis.Scan(ctx, 0, pattern, 1000).Iterator()

	for iter.Next(ctx) {
		// delete from lru cache
		m.local.Remove(iter.Val())
		// delete from redis
		_ = m.redis.Del(ctx, iter.Val()).Err()
	}

	return iter.Err()
}

// Set implements port.MemoryAdapter.
func (m *memoryAdapter) Set(ctx context.Context, key string, value []byte, ttl time.Duration) {
	// set to lru cache
	m.local.Add(key, cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	})

	// set to redis
	if m.redis != nil {
		_ = m.redis.Set(ctx, key, value, ttl).Err()
	}
}
