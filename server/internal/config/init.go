// Package config 使用 Viper 加载并校验应用配置。
package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	CheckTimeout    time.Duration
}

type RedisConfig struct {
	DSN             string
	PoolSize        int
	MinIdleConns    int
	MaxIdleConns    int
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	CheckTimeout    time.Duration
}

type Config struct {
	ServerAddr        string
	Database          DatabaseConfig
	Redis             RedisConfig
	JWTSecret         string
	JWTExpire         time.Duration
	WeChatAppID       string
	WeChatAppSecret   string
	BailianAPIKey     string
	BailianBaseURL    string
	BailianASRModel   string
	LLMAPIKey         string
	LLMBaseURL        string
	LLMModel          string
	LLMThinkingMode   string
	CORSAllowOrigins  []string
	UploadDir         string
	PublicBaseURL     string
	DataEncryptionKey string
	AllowMockLogin    bool
}

var current Config

// init 在其他包使用配置前完成加载；配置错误会直接终止启动，因为后续组件无法安全初始化。
func init() {
	cfg, err := loadFile(resolveConfigPath())
	if err != nil {
		panic(fmt.Errorf("initialize configuration: %w", err))
	}
	current = cfg
}

// Current 返回包初始化期间加载的只读配置副本。
func Current() Config {
	return current
}

func resolveConfigPath() string {
	if path := strings.TrimSpace(os.Getenv("CONFIG_PATH")); path != "" {
		return path
	}
	directory, err := os.Getwd()
	if err != nil {
		return "config/config.yaml"
	}
	for {
		for _, candidate := range []string{filepath.Join(directory, "config", "config.yaml"), filepath.Join(directory, "config.yaml")} {
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate
			}
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			break
		}
		directory = parent
	}
	return "config/config.yaml"
}

// loadFile 使用独立 Viper 实例，避免配置测试修改基础设施初始化所使用的包级配置。
func loadFile(path string) (Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		var missing viper.ConfigFileNotFoundError
		if (!errors.As(err, &missing) && !os.IsNotExist(err)) || strings.TrimSpace(v.GetString("database.dsn")) == "" {
			return Config{}, fmt.Errorf("read config file %q: %w", path, err)
		}
	}

	expireHours := v.GetInt("jwt.expire_hours")
	if raw := strings.TrimSpace(os.Getenv("JWT_EXPIRE_HOURS")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			return Config{}, fmt.Errorf("JWT_EXPIRE_HOURS must be a positive integer")
		}
		expireHours = parsed
	}
	if expireHours <= 0 {
		return Config{}, fmt.Errorf("jwt.expire_hours must be a positive integer")
	}

	allowedOrigins := cleanValues(v.GetStringSlice("cors.allow_origins"))
	if value := strings.TrimSpace(os.Getenv("CORS_ALLOW_ORIGINS")); value != "" {
		allowedOrigins = cleanValues(strings.Split(value, ","))
	}
	cfg := Config{
		ServerAddr: strings.TrimSpace(v.GetString("server.addr")),
		Database: DatabaseConfig{
			DSN:             strings.TrimSpace(v.GetString("database.dsn")),
			MaxOpenConns:    v.GetInt("database.max_open_conns"),
			MaxIdleConns:    v.GetInt("database.max_idle_conns"),
			ConnMaxLifetime: v.GetDuration("database.conn_max_lifetime"),
			ConnMaxIdleTime: v.GetDuration("database.conn_max_idle_time"),
			CheckTimeout:    v.GetDuration("database.check_timeout"),
		},
		Redis: RedisConfig{
			DSN:             strings.TrimSpace(v.GetString("redis.dsn")),
			PoolSize:        v.GetInt("redis.pool_size"),
			MinIdleConns:    v.GetInt("redis.min_idle_conns"),
			MaxIdleConns:    v.GetInt("redis.max_idle_conns"),
			ConnMaxIdleTime: v.GetDuration("redis.conn_max_idle_time"),
			ConnMaxLifetime: v.GetDuration("redis.conn_max_lifetime"),
			DialTimeout:     v.GetDuration("redis.dial_timeout"),
			ReadTimeout:     v.GetDuration("redis.read_timeout"),
			WriteTimeout:    v.GetDuration("redis.write_timeout"),
			CheckTimeout:    v.GetDuration("redis.check_timeout"),
		},
		JWTSecret:         strings.TrimSpace(v.GetString("jwt.secret")),
		JWTExpire:         time.Duration(expireHours) * time.Hour,
		WeChatAppID:       strings.TrimSpace(v.GetString("wechat.app_id")),
		WeChatAppSecret:   strings.TrimSpace(v.GetString("wechat.app_secret")),
		BailianAPIKey:     strings.TrimSpace(v.GetString("bailian.api_key")),
		BailianBaseURL:    strings.TrimSpace(v.GetString("bailian.base_url")),
		BailianASRModel:   strings.TrimSpace(v.GetString("bailian.asr_model")),
		LLMAPIKey:         strings.TrimSpace(v.GetString("llm.api_key")),
		LLMBaseURL:        strings.TrimSpace(v.GetString("llm.base_url")),
		LLMModel:          strings.TrimSpace(v.GetString("llm.model")),
		LLMThinkingMode:   strings.TrimSpace(v.GetString("llm.thinking_mode")),
		CORSAllowOrigins:  allowedOrigins,
		UploadDir:         strings.TrimSpace(v.GetString("storage.upload_dir")),
		PublicBaseURL:     strings.TrimSpace(v.GetString("storage.public_base_url")),
		DataEncryptionKey: strings.TrimSpace(v.GetString("security.data_encryption_key")),
		AllowMockLogin:    v.GetBool("security.allow_mock_login") && !isProduction(),
	}
	if cfg.LLMAPIKey == "" && cfg.LLMBaseURL == cfg.BailianBaseURL {
		cfg.LLMAPIKey = cfg.BailianAPIKey
	}
	if err := validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.addr", ":8080")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", "30m")
	v.SetDefault("database.conn_max_idle_time", "10m")
	v.SetDefault("database.check_timeout", "5s")
	v.SetDefault("redis.pool_size", 20)
	v.SetDefault("redis.min_idle_conns", 2)
	v.SetDefault("redis.max_idle_conns", 10)
	v.SetDefault("redis.conn_max_idle_time", "10m")
	v.SetDefault("redis.conn_max_lifetime", "30m")
	v.SetDefault("redis.dial_timeout", "5s")
	v.SetDefault("redis.read_timeout", "3s")
	v.SetDefault("redis.write_timeout", "3s")
	v.SetDefault("redis.check_timeout", "5s")
	v.SetDefault("jwt.expire_hours", 2)
	v.SetDefault("bailian.base_url", "https://dashscope.aliyuncs.com/compatible-mode/v1")
	v.SetDefault("bailian.asr_model", "qwen3-asr-flash")
	v.SetDefault("llm.base_url", "https://dashscope.aliyuncs.com/compatible-mode/v1")
	v.SetDefault("llm.model", "qwen-plus")
	v.SetDefault("llm.thinking_mode", "disabled")
	v.SetDefault("storage.upload_dir", "./uploads")
	v.SetDefault("storage.public_base_url", "/uploads")
}

func validate(cfg Config) error {
	if cfg.Database.DSN == "" {
		return fmt.Errorf("database.dsn is required")
	}
	if cfg.Database.MaxOpenConns <= 0 || cfg.Database.MaxIdleConns < 0 || cfg.Database.MaxIdleConns > cfg.Database.MaxOpenConns {
		return fmt.Errorf("database connection pool configuration is invalid")
	}
	if cfg.Database.ConnMaxLifetime <= 0 || cfg.Database.ConnMaxIdleTime <= 0 || cfg.Database.CheckTimeout <= 0 {
		return fmt.Errorf("database connection lifetime and timeout configuration is invalid")
	}
	if cfg.Redis.DSN == "" {
		return fmt.Errorf("redis.dsn is required")
	}
	if cfg.Redis.PoolSize <= 0 || cfg.Redis.MinIdleConns < 0 || cfg.Redis.MaxIdleConns <= 0 || cfg.Redis.MinIdleConns > cfg.Redis.MaxIdleConns || cfg.Redis.MaxIdleConns > cfg.Redis.PoolSize {
		return fmt.Errorf("redis connection pool configuration is invalid")
	}
	if cfg.Redis.ConnMaxIdleTime <= 0 || cfg.Redis.ConnMaxLifetime <= 0 || cfg.Redis.DialTimeout <= 0 || cfg.Redis.ReadTimeout <= 0 || cfg.Redis.WriteTimeout <= 0 || cfg.Redis.CheckTimeout <= 0 {
		return fmt.Errorf("redis connection lifetime and timeout configuration is invalid")
	}
	if len(cfg.JWTSecret) < 32 {
		return fmt.Errorf("jwt.secret must contain at least 32 characters")
	}
	if cfg.LLMThinkingMode != "disabled" && cfg.LLMThinkingMode != "enabled" && cfg.LLMThinkingMode != "provider_default" {
		return fmt.Errorf("llm.thinking_mode must be disabled, enabled, or provider_default")
	}
	if isProduction() {
		if cfg.WeChatAppID == "" || cfg.WeChatAppSecret == "" {
			return fmt.Errorf("wechat credentials are required in production")
		}
		key, err := base64.StdEncoding.DecodeString(cfg.DataEncryptionKey)
		if err != nil || len(key) != 32 {
			return fmt.Errorf("DATA_ENCRYPTION_KEY must be Base64 for exactly 32 bytes in production")
		}
		if !strings.HasPrefix(strings.ToLower(cfg.PublicBaseURL), "https://") {
			return fmt.Errorf("PUBLIC_BASE_URL must use HTTPS in production")
		}
	}
	return nil
}

func cleanValues(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func isProduction() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production")
}
