package config

import (
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Redis     RedisConfig     `yaml:"redis"`
	WeChat    WeChatConfig    `yaml:"wechat"`
	DeepSeek  DeepSeekConfig  `yaml:"deepseek"`
	JWT       JWTConfig       `yaml:"jwt"`
	Upload    UploadConfig    `yaml:"upload"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	COS       COSConfig       `yaml:"cos"`
	Notify    NotifyConfig    `yaml:"notify"`
	Webhook   WebhookConfig   `yaml:"webhook"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
	Mode string `yaml:"mode"` // debug / release
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

func (d DatabaseConfig) DSN() string {
	return "host=" + d.Host + " port=" + d.Port + " user=" + d.User +
		" password=" + d.Password + " dbname=" + d.DBName + " sslmode=" + d.SSLMode
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type WeChatConfig struct {
	AppID          string `yaml:"app_id"`
	AppSecret      string `yaml:"app_secret"`
	Token          string `yaml:"token"`
	EncodingAESKey string `yaml:"encoding_aes_key"`
}

type DeepSeekConfig struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
	Expire int    `yaml:"expire"` // hours
}

type UploadConfig struct {
	Dir string `yaml:"dir"` // local upload directory
}

type RateLimitConfig struct {
	Daily int64 `yaml:"daily"` // 每日AI调用上限
}

type COSConfig struct {
	SecretID  string `yaml:"secret_id"`
	SecretKey string `yaml:"secret_key"`
	Bucket    string `yaml:"bucket"`
	Region    string `yaml:"region"`
}

type NotifyConfig struct {
	TemplateID    string `yaml:"template_id"`
	BreakfastTime string `yaml:"breakfast_time"`
	LunchTime     string `yaml:"lunch_time"`
	DinnerTime    string `yaml:"dinner_time"`
	CheckInterval string `yaml:"check_interval"`
}

type WebhookConfig struct {
	FeishuWebhookURL string `yaml:"feishu_webhook_url"` // 飞书机器人 webhook
	RepoPath         string `yaml:"repo_path"`          // 代码仓库路径
	Secret           string `yaml:"secret"`             // webhook 签名密钥 (可选)
	Keyword          string `yaml:"keyword"`            // webhook 关键词校验 (可选)
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Replace ${ENV_VAR} with environment variable values
	expanded := expandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// expandEnv replaces ${VAR} or ${VAR:-default} patterns with environment variable values.
func expandEnv(s string) string {
	re := regexp.MustCompile(`\$\{([^}:]+)(?::-([^}]*))?\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		sub := re.FindStringSubmatch(match)
		name := sub[1]
		defaultVal := sub[2]
		if v := os.Getenv(name); v != "" {
			return v
		}
		return defaultVal
	})
}
