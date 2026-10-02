package redis

import (
	"context"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Client struct {
	rdb       *goredis.Client
	keyPrefix string
}

const defaultKeyPrefix = "v1:"

func NewRedis(dsn string) (*Client, error) {
	// DSN 已包含地址、端口、密码和数据库编号，不应再手动拼接端口。
	options, err := goredis.ParseURL(dsn)
	if err != nil {
		// 连接地址可能含密码，不将原始 URL 或解析错误写入日志。
		return nil, errors.New("invalid Redis connection URL")
	}
	rdb := goredis.NewClient(options)
	// 前缀仅保存在封装中，后续读写方法需要显式使用它，不会自动添加到 Redis 键上。
	return &Client{rdb: rdb, keyPrefix: defaultKeyPrefix}, nil
}

// Close 在服务退出时释放底层连接池。
func (c *Client) Close() error {
	return c.rdb.Close()
}

func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.rdb == nil {
		return errors.New("redis client is not initialized")
	}
	return c.rdb.Ping(ctx).Err()
}

// Init checks connectivity before the application accepts requests.
func (c *Client) Init(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)

	defer cancel()
	return c.Ping(ctx)
}
