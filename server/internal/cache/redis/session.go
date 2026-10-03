package redis

import (
	"context"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var ErrSessionMissing = errors.New("session missing or expired")

// Session 只保存令牌摘要，原始 Bearer 凭据不会写入 Redis。
type Session struct {
	UserID      uint   `json:"user_id"`
	AccessHash  string `json:"access_hash"`
	RefreshHash string `json:"refresh_hash"`
}

func (c *Client) sessionKey(id string) string { return c.keyPrefix + "session:{" + id + "}" }

func (c *Client) CreateSession(ctx context.Context, id string, s Session, ttl time.Duration) error {
	_, err := c.rdb.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
		pipe.HSet(ctx, c.sessionKey(id), "user_id", s.UserID, "access_hash", s.AccessHash, "refresh_hash", s.RefreshHash)
		pipe.PExpire(ctx, c.sessionKey(id), ttl)
		return nil
	})
	return err
}

func (c *Client) MatchSession(ctx context.Context, id, field, hash string) (bool, error) {
	value, err := c.rdb.HGet(ctx, c.sessionKey(id), field).Result()
	if errors.Is(err, goredis.Nil) {
		return false, nil
	}
	return value == hash && value != "", err
}

// 在一条命令中完成比较与替换，避免并发刷新请求重复使用同一刷新凭据。
var rotateSession = goredis.NewScript(`
if redis.call('HGET', KEYS[1], 'refresh_hash') ~= ARGV[1] then return 0 end
redis.call('HSET', KEYS[1], 'access_hash', ARGV[2], 'refresh_hash', ARGV[3])
redis.call('PEXPIRE', KEYS[1], ARGV[4])
return 1
`)

func (c *Client) RotateSession(ctx context.Context, id, oldHash string, s Session, ttl time.Duration) error {
	result, err := rotateSession.Run(ctx, c.rdb, []string{c.sessionKey(id)}, oldHash, s.AccessHash, s.RefreshHash, ttl.Milliseconds()).Int()
	if err != nil {
		return err
	}
	if result != 1 {
		return ErrSessionMissing
	}
	return nil
}
