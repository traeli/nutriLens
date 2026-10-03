package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadFromYAMLAllowsEnvironmentOverrides(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "config.yaml")
	contents := []byte(`
server:
  addr: ":9090"
database:
  dsn: "file-dsn"
jwt:
  secret: "file-secret-with-at-least-32-characters"
  expire_hours: 72
wechat:
  app_id: "file-app-id"
  app_secret: "file-app-secret"
bailian:
  api_key: "file-bailian-key"
  base_url: "https://workspace.example.com/compatible-mode/v1"
  asr_model: "qwen3-asr-flash-test"
cors:
  allow_origins:
    - "http://localhost:5173"
`)
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("write test config: %v", err)
	}
	t.Setenv("SERVER_ADDR", ":8081")
	t.Setenv("DATABASE_DSN", "env-dsn")
	t.Setenv("JWT_SECRET", "env-secret-with-at-least-32-characters")
	t.Setenv("JWT_EXPIRE_HOURS", "168")
	t.Setenv("WECHAT_APP_ID", "env-app-id")
	t.Setenv("WECHAT_APP_SECRET", "env-app-secret")
	t.Setenv("CORS_ALLOW_ORIGINS", "http://127.0.0.1:5173, https://example.test")

	cfg, err := loadFile(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.ServerAddr != ":8081" || cfg.DatabaseDSN != "env-dsn" || cfg.JWTExpire != 168*time.Hour {
		t.Fatalf("file values were not loaded: %+v", cfg)
	}
	if cfg.WeChatAppID != "env-app-id" || cfg.WeChatAppSecret != "env-app-secret" {
		t.Fatalf("wechat file values were not loaded: %+v", cfg)
	}
	if cfg.BailianAPIKey != "file-bailian-key" || cfg.BailianBaseURL != "https://workspace.example.com/compatible-mode/v1" || cfg.BailianASRModel != "qwen3-asr-flash-test" {
		t.Fatalf("bailian file values were not loaded")
	}
	if len(cfg.CORSAllowOrigins) != 2 || cfg.CORSAllowOrigins[0] != "http://127.0.0.1:5173" || cfg.CORSAllowOrigins[1] != "https://example.test" {
		t.Fatalf("CORS_ALLOW_ORIGINS = %#v", cfg.CORSAllowOrigins)
	}
}

func TestProductionRequiresCompleteSecurityConfiguration(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "config.yaml")
	if err := os.WriteFile(path, []byte(`
database:
  dsn: "postgres-dsn"
jwt:
  secret: "file-secret-with-at-least-32-characters"
`), 0o600); err != nil {
		t.Fatalf("write test config: %v", err)
	}
	t.Setenv("APP_ENV", "production")

	if _, err := loadFile(path); err == nil {
		t.Fatal("loadFile() error = nil, want incomplete production configuration error")
	}
}

func TestLoadMissingFileReturnsError(t *testing.T) {
	_, err := loadFile(filepath.Join(t.TempDir(), "missing.yaml"))
	if err == nil {
		t.Fatal("loadFile() error = nil, want missing file error")
	}
}

func TestSMTPConfiguration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("database:\n  dsn: test\njwt:\n  secret: test-secret-with-at-least-32-characters\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("APP_ENV", "test")
	for _, key := range []string{"SMTP_HOST", "SMTP_PORT", "SMTP_USERNAME", "SMTP_PASSWORD", "SMTP_FROM", "SMTP_TLS_MODE"} {
		t.Setenv(key, "")
	}
	cfg, err := loadFile(path)
	if err != nil || cfg.SMTP.Host != "" {
		t.Fatalf("disabled SMTP: %v", err)
	}
	t.Setenv("SMTP_HOST", "smtp.example.com")
	if _, err := loadFile(path); err == nil {
		t.Fatal("partial SMTP config accepted")
	}
	t.Setenv("SMTP_USERNAME", "sender@example.com")
	t.Setenv("SMTP_PASSWORD", "test-only-password")
	t.Setenv("SMTP_FROM", "sender@example.com")
	t.Setenv("SMTP_TLS_MODE", "starttls")
	cfg, err = loadFile(path)
	if err != nil || cfg.SMTP.Port != 587 || cfg.SMTP.TLSMode != "starttls" {
		t.Fatalf("SMTP env config: %v", err)
	}
	t.Setenv("SMTP_PORT", "0")
	if _, err := loadFile(path); err == nil {
		t.Fatal("invalid port accepted")
	}
	t.Setenv("SMTP_PORT", "465")
	t.Setenv("SMTP_TLS_MODE", "none")
	if _, err := loadFile(path); err == nil {
		t.Fatal("plaintext SMTP accepted")
	}
}
