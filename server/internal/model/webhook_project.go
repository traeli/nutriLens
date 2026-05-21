package model

import "time"

// WebhookProject stores per-project configuration for Gitea webhook & deploy.
type WebhookProject struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Name             string    `gorm:"size:255;not null" json:"name"`                  // 项目名称
	RepoName         string    `gorm:"size:255;uniqueIndex;not null" json:"repo_name"` // Gitea 仓库全名，如 "SmartWearables/apiService"
	RepoPath         string    `gorm:"size:500;not null" json:"repo_path"`             // 本地 git 仓库路径
	FeishuWebhookURL string    `gorm:"size:500" json:"feishu_webhook_url"`             // 飞书机器人 webhook 地址
	FeishuKeyword    string    `gorm:"size:100" json:"feishu_keyword"`                 // 飞书机器人自定义关键词
	DeployScript     string    `gorm:"type:text" json:"deploy_script"`                 // git pull 后执行的部署脚本
	Enabled          bool      `gorm:"default:true" json:"enabled"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
