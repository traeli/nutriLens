package config

import (
	"encoding/base64"
	"fmt"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	TLSMode  string `yaml:"tls_mode"`
}

type Config struct {
	SMTP              SMTPConfig
	ServerAddr        string
	DatabaseDSN       string
	RedisDSN          string
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
	SMTP   SMTPConfig `yaml:"smtp"`
	Server struct {
		Addr string `yaml:"addr"`
	} `yaml:"server"`
	Database struct {
		DSN string `yaml:"dsn"`
	} `yaml:"database"`
	Redis struct {
		DSN string `yaml:"dsn"`
	} `yaml:"redis"`
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
	path := strings.TrimSpace(os.Getenv("CONFIG_PATH"))
	if path == "" {
		path = "config/config.yaml"
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
		expireHours = 2
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
	smtpConfig := fileValues.SMTP
	smtpConfig.Host = envOr("SMTP_HOST", strings.TrimSpace(smtpConfig.Host))
	smtpConfig.Username = envOr("SMTP_USERNAME", strings.TrimSpace(smtpConfig.Username))
	smtpConfig.Password = envOr("SMTP_PASSWORD", smtpConfig.Password)
	smtpConfig.From = envOr("SMTP_FROM", strings.TrimSpace(smtpConfig.From))
	smtpConfig.TLSMode = strings.ToLower(envOr("SMTP_TLS_MODE", valueOr(smtpConfig.TLSMode, "tls")))
	if smtpConfig.Port == 0 {
		if smtpConfig.TLSMode == "starttls" {
			smtpConfig.Port = 587
		} else {
			smtpConfig.Port = 465
		}
	}
	if value := strings.TrimSpace(os.Getenv("SMTP_PORT")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, fmt.Errorf("SMTP_PORT must be an integer")
		}
		smtpConfig.Port = parsed
	}
	if smtpConfig.Host != "" || smtpConfig.Username != "" || smtpConfig.Password != "" || smtpConfig.From != "" {
		address, err := mail.ParseAddress(smtpConfig.From)
		if smtpConfig.Host == "" || smtpConfig.Username == "" || smtpConfig.Password == "" || err != nil || address.Address != smtpConfig.From || strings.ContainsAny(smtpConfig.From, "\r\n") {
			return Config{}, fmt.Errorf("SMTP requires host, username, password and a valid from address")
		}
	}
	if smtpConfig.Port < 1 || smtpConfig.Port > 65535 {
		return Config{}, fmt.Errorf("SMTP port must be between 1 and 65535")
	}
	if smtpConfig.TLSMode != "tls" && smtpConfig.TLSMode != "starttls" {
		return Config{}, fmt.Errorf("SMTP TLS mode must be tls or starttls")
	}
	cfg := Config{
		SMTP:              smtpConfig,
		ServerAddr:        envOr("SERVER_ADDR", valueOr(fileValues.Server.Addr, ":8080")),
		DatabaseDSN:       envOr("DATABASE_DSN", strings.TrimSpace(fileValues.Database.DSN)),
		RedisDSN:          envOr("REDIS_DSN", strings.TrimSpace(fileValues.Redis.DSN)),
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
