package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

type Config struct {
	ServerAddr        string
	DatabaseDSN       string
	JWTSecret         string
	JWTExpire         time.Duration
	WeChatAppID       string
	WeChatAppSecret   string
	BailianAPIKey     string
	BailianBaseURL    string
	BailianASRModel   string
	CORSAllowOrigins  []string
	UploadDir         string
	PublicBaseURL     string
	DataEncryptionKey string
	AllowMockLogin    bool
}

type fileConfig struct {
	Server struct {
		Addr string `yaml:"addr"`
	} `yaml:"server"`
	Database struct {
		DSN string `yaml:"dsn"`
	} `yaml:"database"`
	JWT struct {
		Secret      string `yaml:"secret"`
		ExpireHours int    `yaml:"expire_hours"`
	} `yaml:"jwt"`
	WeChat struct {
		AppID     string `yaml:"app_id"`
		AppSecret string `yaml:"app_secret"`
	} `yaml:"wechat"`
	Bailian struct {
		APIKey   string `yaml:"api_key"`
		BaseURL  string `yaml:"base_url"`
		ASRModel string `yaml:"asr_model"`
	} `yaml:"bailian"`
	CORS struct {
		AllowOrigins []string `yaml:"allow_origins"`
	} `yaml:"cors"`
	Storage struct {
		UploadDir     string `yaml:"upload_dir"`
		PublicBaseURL string `yaml:"public_base_url"`
	} `yaml:"storage"`
	Security struct {
		DataEncryptionKey string `yaml:"data_encryption_key"`
		AllowMockLogin    bool   `yaml:"allow_mock_login"`
	} `yaml:"security"`
}

func Load() (Config, error) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		return Config{}, fmt.Errorf("resolve config file path")
	}
	path := strings.TrimSpace(os.Getenv("CONFIG_PATH"))
	if path == "" {
		path = filepath.Join(filepath.Dir(sourceFile), "config.yaml")
	}
	return loadFile(path)
}

func loadFile(path string) (Config, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) || strings.TrimSpace(os.Getenv("DATABASE_DSN")) == "" {
			return Config{}, fmt.Errorf("read config file %q: %w", path, err)
		}
		contents = []byte("{}")
	}
	var fileValues fileConfig
	if err := yaml.Unmarshal(contents, &fileValues); err != nil {
		return Config{}, fmt.Errorf("parse config file %q: %w", path, err)
	}
	expireHours := fileValues.JWT.ExpireHours
	if expireHours == 0 {
		expireHours = 168
	}
	if expireHours < 0 {
		return Config{}, fmt.Errorf("jwt.expire_hours must be a positive integer")
	}
	if value := strings.TrimSpace(os.Getenv("JWT_EXPIRE_HOURS")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			return Config{}, fmt.Errorf("JWT_EXPIRE_HOURS must be a positive integer")
		}
		expireHours = parsed
	}
	allowedOrigins := cleanValues(fileValues.CORS.AllowOrigins)
	if value := strings.TrimSpace(os.Getenv("CORS_ALLOW_ORIGINS")); value != "" {
		allowedOrigins = cleanValues(strings.Split(value, ","))
	}
	cfg := Config{
		ServerAddr:        envOr("SERVER_ADDR", valueOr(fileValues.Server.Addr, ":8080")),
		DatabaseDSN:       envOr("DATABASE_DSN", strings.TrimSpace(fileValues.Database.DSN)),
		JWTSecret:         envOr("JWT_SECRET", strings.TrimSpace(fileValues.JWT.Secret)),
		JWTExpire:         time.Duration(expireHours) * time.Hour,
		WeChatAppID:       envOr("WECHAT_APP_ID", strings.TrimSpace(fileValues.WeChat.AppID)),
		WeChatAppSecret:   envOr("WECHAT_APP_SECRET", strings.TrimSpace(fileValues.WeChat.AppSecret)),
		BailianAPIKey:     envOr("BAILIAN_API_KEY", strings.TrimSpace(fileValues.Bailian.APIKey)),
		BailianBaseURL:    envOr("BAILIAN_BASE_URL", valueOr(fileValues.Bailian.BaseURL, "https://dashscope.aliyuncs.com/compatible-mode/v1")),
		BailianASRModel:   envOr("BAILIAN_ASR_MODEL", valueOr(fileValues.Bailian.ASRModel, "qwen3-asr-flash")),
		CORSAllowOrigins:  allowedOrigins,
		UploadDir:         envOr("UPLOAD_DIR", valueOr(fileValues.Storage.UploadDir, "./uploads")),
		PublicBaseURL:     envOr("PUBLIC_BASE_URL", valueOr(fileValues.Storage.PublicBaseURL, "/uploads")),
		DataEncryptionKey: envOr("DATA_ENCRYPTION_KEY", strings.TrimSpace(fileValues.Security.DataEncryptionKey)),
		AllowMockLogin:    fileValues.Security.AllowMockLogin && strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV"))) != "production",
	}
	if cfg.DatabaseDSN == "" {
		return Config{}, fmt.Errorf("database.dsn is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("jwt.secret must contain at least 32 characters")
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production") {
		if cfg.WeChatAppID == "" || cfg.WeChatAppSecret == "" {
			return Config{}, fmt.Errorf("wechat credentials are required in production")
		}
		key, err := base64.StdEncoding.DecodeString(cfg.DataEncryptionKey)
		if err != nil || len(key) != 32 {
			return Config{}, fmt.Errorf("DATA_ENCRYPTION_KEY must be Base64 for exactly 32 bytes in production")
		}
		if !strings.HasPrefix(strings.ToLower(cfg.PublicBaseURL), "https://") {
			return Config{}, fmt.Errorf("PUBLIC_BASE_URL must use HTTPS in production")
		}
	}
	return cfg, nil
}

func envOr(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func valueOr(value, fallback string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return fallback
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
