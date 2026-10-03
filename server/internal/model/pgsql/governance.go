package pgsql

import "time"

type PublisherVerification struct {
	UserID              uint      `gorm:"primaryKey;comment:用户编号" json:"user_id"`
	PhoneEncrypted      string    `gorm:"type:text;not null;comment:加密手机号" json:"-"`
	PhoneHash           string    `gorm:"size:64;not null;unique;comment:手机号摘要" json:"-"`
	VerificationChannel string    `gorm:"size:32;not null;comment:核验渠道" json:"verification_channel"`
	ProviderReference   string    `gorm:"size:128;comment:服务方凭据编号" json:"-"`
	Status              string    `gorm:"size:24;not null;comment:核验状态" json:"status"`
	VerifiedAt          time.Time `gorm:"comment:核验时间" json:"verified_at"`
	CreatedAt           time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt           time.Time `gorm:"comment:更新时间" json:"updated_at"`
}

func (PublisherVerification) TableName() string { return "publisher_verifications" }

type ContentSafetyCheck struct {
	ID             uint         `gorm:"primaryKey;comment:安全检查编号" json:"id"`
	TargetType     string       `gorm:"size:32;not null;index;comment:检查对象类型" json:"target_type"`
	TargetID       uint         `gorm:"not null;index;comment:检查对象编号" json:"target_id"`
	Provider       string       `gorm:"size:32;not null;comment:检查服务提供方" json:"provider"`
	CheckType      string       `gorm:"size:32;not null;comment:检查类型" json:"check_type"`
	Result         string       `gorm:"size:24;not null;comment:检查结果" json:"result"`
	RiskLabels     JSONDocument `gorm:"type:jsonb;comment:风险标签" json:"risk_labels"`
	RawResponseRef string       `gorm:"type:text;comment:原始响应引用" json:"-"`
	CheckedAt      time.Time    `gorm:"comment:检查时间" json:"checked_at"`
}

func (ContentSafetyCheck) TableName() string { return "content_safety_checks" }

type ModerationTask struct {
	ID              uint         `gorm:"primaryKey;comment:审核任务编号" json:"id"`
	TaskType        string       `gorm:"size:32;not null;comment:任务类型" json:"task_type"`
	TargetType      string       `gorm:"size:32;not null;comment:审核对象类型" json:"target_type"`
	TargetID        uint         `gorm:"not null;index;comment:审核对象编号" json:"target_id"`
	RecordVersionID *uint        `gorm:"comment:足迹版本编号" json:"record_version_id,omitempty"`
	ReviewVersionID *uint        `gorm:"comment:评论版本编号" json:"review_version_id,omitempty"`
	Priority        int          `gorm:"comment:任务优先级" json:"priority"`
	RiskLabels      JSONDocument `gorm:"type:jsonb;comment:风险标签" json:"risk_labels"`
	Status          string       `gorm:"size:24;not null;comment:任务状态" json:"status"`
	Decision        string       `gorm:"size:24;comment:审核决定" json:"decision"`
	DecisionReason  string       `gorm:"type:text;comment:决定原因" json:"decision_reason"`
	DueAt           *time.Time   `gorm:"comment:处理截止时间" json:"due_at,omitempty"`
	DecidedAt       *time.Time   `gorm:"comment:决定时间" json:"decided_at,omitempty"`
	CreatedAt       time.Time    `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time    `gorm:"comment:更新时间" json:"updated_at"`
}

func (ModerationTask) TableName() string { return "moderation_tasks" }

type FeatureFlag struct {
	Key         string       `gorm:"primaryKey;size:64;comment:功能开关键" json:"key"`
	Enabled     bool         `gorm:"not null;comment:是否启用" json:"enabled"`
	Value       JSONDocument `gorm:"type:jsonb;comment:扩展配置" json:"value"`
	Description string       `gorm:"type:text;comment:功能说明" json:"description"`
	UpdatedAt   time.Time    `gorm:"comment:更新时间" json:"updated_at"`
}

func (FeatureFlag) TableName() string { return "feature_flags" }
