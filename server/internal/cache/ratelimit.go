package cache

import (
	"context"
	"fmt"
	"time"
)

// RateLimiter provides fixed-window rate limiting backed by Redis.
//
// Key format: {prefix}:{identifier}:{date}
// TTL: auto-expires at end of the current day.
//
// Example usage:
//
//	limiter := cache.NewRateLimiter(redisClient, "ai_rate")
//	count, remaining, err := limiter.Allow(ctx, "user_123", 30)
type RateLimiter struct {
	client *Client
	prefix string
}

// NewRateLimiter creates a rate limiter with the given key prefix.
func NewRateLimiter(client *Client, prefix string) *RateLimiter {
	return &RateLimiter{client: client, prefix: prefix}
}

// Allow checks if the request is within limit and increments the counter.
// Returns (currentCount, remaining, error).
// If currentCount > limit, the request is denied (remaining will be 0).
func (r *RateLimiter) Allow(ctx context.Context, identifier string, limit int64) (count int64, remaining int64, err error) {
	key := r.buildKey(identifier)
	ttl := untilEndOfDay()

	count, err = r.client.IncrWithExpire(ctx, key, ttl)
	if err != nil {
		return 0, 0, fmt.Errorf("rate limit check failed: %w", err)
	}

	remaining = limit - count
	if remaining < 0 {
		remaining = 0
	}
	return count, remaining, nil
}

// GetUsage returns the current usage count without incrementing.
func (r *RateLimiter) GetUsage(ctx context.Context, identifier string) (int64, error) {
	key := r.buildKey(identifier)
	val, err := r.client.Get(ctx, key)
	if err != nil {
		// Key doesn't exist yet — zero usage
		return 0, nil
	}
	var count int64
	fmt.Sscanf(val, "%d", &count)
	return count, nil
}

// Reset clears the rate limit counter for an identifier.
func (r *RateLimiter) Reset(ctx context.Context, identifier string) error {
	key := r.buildKey(identifier)
	return r.client.Del(ctx, key)
}

// buildKey constructs the Redis key: prefix:identifier:YYYYMMDD
func (r *RateLimiter) buildKey(identifier string) string {
	return r.prefix + ":" + identifier + ":" + time.Now().Format("20060102")
}

// untilEndOfDay returns the duration from now until midnight (00:00 tomorrow).
func untilEndOfDay() time.Duration {
	now := time.Now()
	end := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return end.Sub(now)
}
