package model

import "time"

// CodeReviewReport stores AI-generated code review reports for webhook pushes.
type CodeReviewReport struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ProjectID     uint      `gorm:"index;not null" json:"project_id"`          // 关联 WebhookProject
	ProjectName   string    `gorm:"size:255;not null" json:"project_name"`     // 项目名称（冗余，方便查询）
	RepoName      string    `gorm:"size:255;not null" json:"repo_name"`        // 仓库全名
	Branch        string    `gorm:"size:255" json:"branch"`                    // 分支
	CommitHash    string    `gorm:"size:40;index;not null" json:"commit_hash"` // 提交哈希
	CommitMsg     string    `gorm:"size:1000" json:"commit_msg"`               // 提交信息
	Author        string    `gorm:"size:255" json:"author"`                    // 提交者
	ChangedFiles  string    `gorm:"type:text" json:"changed_files"`            // 变更文件列表
	RiskLevel     string    `gorm:"size:20;index" json:"risk_level"`           // 总体风险等级: low/medium/high
	RiskSummary   string    `gorm:"size:1000" json:"risk_summary"`             // 风险摘要
	ReportContent string    `gorm:"type:text;not null" json:"report_content"`  // 完整评审报告（Markdown）
	FilePath      string    `gorm:"size:500" json:"file_path"`                 // 本地 Markdown 文件路径
	CreatedAt     time.Time `gorm:"index" json:"created_at"`
}

func (CodeReviewReport) TableName() string {
	return "code_review_reports"
}
