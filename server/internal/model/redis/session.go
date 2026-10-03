package redis

import (
	"context"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var ErrSessionMissing = errors.New("session missing or expired")

// Session stores only token digests; raw bearer credentials never enter Redis.
type Session struct {
	UserID      uint   `json:"user_id"`
	AccessHash  string `json:"access_hash"`
	RefreshHash string `json:"refresh_hash"`
}

// sessionKey 统一会话键格式；花括号中的会话 ID 作为 Redis Cluster 的 hash tag。
func (c *Client) sessionKey(id string) string { return c.keyPrefix + "session:{" + id + "}" }

// CreateSession 保存用户 ID 和两种 Token 的摘要，并设置整个会话键的过期时间。
// 使用 MULTI/EXEC 成组执行写入和 TTL 命令；错误交给调用方处理，不在此返回 Token。
func (c *Client) CreateSession(ctx context.Context, id string, s Session, ttl time.Duration) error {
	_, err := c.rdb.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
		pipe.HSet(ctx, c.sessionKey(id), "user_id", s.UserID, "access_hash", s.AccessHash, "refresh_hash", s.RefreshHash) // sessionKey 生成带统一前缀的会话键。
		pipe.PExpire(ctx, c.sessionKey(id), ttl)                                                                          // 为同一会话键设置有效期。
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

// Compare and replace in one command prevents two refresh requests from reusing a credential.
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
