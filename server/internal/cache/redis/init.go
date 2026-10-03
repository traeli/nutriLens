// Package redis 负责 Redis 生命周期管理和登录会话持久化。
package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"shijibu/internal/config"
)

type Client struct {
	rdb       *goredis.Client
	keyPrefix string
}

const defaultKeyPrefix = "v1:"

var defaultClient *Client

// Init 根据配置创建 Redis 客户端并检查连接；是否启用 Redis 由应用入口决定。
func Init() error {
	if defaultClient != nil {
		return nil
	}
	cfg := config.Current().Redis
	client, err := newClient(cfg)
	if err != nil {
		return fmt.Errorf("initialize Redis client: %w", err)
	}
	if err := client.selfCheck(context.Background(), cfg.CheckTimeout); err != nil {
		_ = client.Close()
		return fmt.Errorf("Redis self-check: %w", err)
	}
	defaultClient = client
	return nil
}

// Default 返回由 Init 创建的 Redis 客户端。
func Default() *Client {
	if defaultClient == nil {
		panic("Redis is not initialized")
	}
	return defaultClient
}

// Close 释放包级管理的 Redis 连接池。
func Close() error {
	if defaultClient == nil {
		return nil
	}
	client := defaultClient
	defaultClient = nil
	return client.Close()
}

// New 创建不受包级生命周期管理的客户端，供集成测试和独立工具使用。
func New(dsn string) (*Client, error) {
	options, err := goredis.ParseURL(dsn)
	if err != nil {
		return nil, errors.New("invalid Redis connection URL")
	}
	return &Client{rdb: goredis.NewClient(options), keyPrefix: defaultKeyPrefix}, nil
}

func newClient(cfg config.RedisConfig) (*Client, error) {
	options, err := goredis.ParseURL(cfg.DSN)
	if err != nil {
		return nil, errors.New("invalid Redis connection URL")
	}
	options.PoolSize = cfg.PoolSize
	options.MinIdleConns = cfg.MinIdleConns
	options.MaxIdleConns = cfg.MaxIdleConns
	options.ConnMaxIdleTime = cfg.ConnMaxIdleTime
	options.ConnMaxLifetime = cfg.ConnMaxLifetime
	options.DialTimeout = cfg.DialTimeout
	options.ReadTimeout = cfg.ReadTimeout
	options.WriteTimeout = cfg.WriteTimeout
	return &Client{rdb: goredis.NewClient(options), keyPrefix: defaultKeyPrefix}, nil
}

// Close 释放当前客户端持有的 Redis 连接池。
func (c *Client) Close() error {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.Close()
}

// Init 对不受包级管理的客户端执行连接自检，主要供集成测试使用。
func (c *Client) Init(ctx context.Context) error {
	return c.selfCheck(ctx, 5*time.Second)
}

func (c *Client) selfCheck(ctx context.Context, timeout time.Duration) error {
	if c == nil || c.rdb == nil {
		return errors.New("Redis client is not initialized")
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return c.rdb.Ping(ctx).Err()
}
