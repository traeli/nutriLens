package model

import "time"

type PublisherVerification struct {
	UserID              uint      `gorm:"primaryKey" json:"user_id"`
	PhoneEncrypted      string    `gorm:"type:text;not null" json:"-"`
	PhoneHash           string    `gorm:"size:64;not null;uniqueIndex" json:"-"`
	VerificationChannel string    `gorm:"size:32;not null" json:"verification_channel"`
	ProviderReference   string    `gorm:"size:128" json:"-"`
	Status              string    `gorm:"size:24;not null" json:"status"`
	VerifiedAt          time.Time `json:"verified_at"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (PublisherVerification) TableName() string { return "publisher_verifications" }

type ContentSafetyCheck struct {
	ID             uint         `gorm:"primaryKey" json:"id"`
	TargetType     string       `gorm:"size:32;not null;index" json:"target_type"`
	TargetID       uint         `gorm:"not null;index" json:"target_id"`
	Provider       string       `gorm:"size:32;not null" json:"provider"`
	CheckType      string       `gorm:"size:32;not null" json:"check_type"`
	Result         string       `gorm:"size:24;not null" json:"result"`
	RiskLabels     JSONDocument `gorm:"type:jsonb" json:"risk_labels"`
	RawResponseRef string       `gorm:"type:text" json:"-"`
	CheckedAt      time.Time    `json:"checked_at"`
}

func (ContentSafetyCheck) TableName() string { return "content_safety_checks" }

type ModerationTask struct {
	ID              uint         `gorm:"primaryKey" json:"id"`
	TaskType        string       `gorm:"size:32;not null" json:"task_type"`
	TargetType      string       `gorm:"size:32;not null" json:"target_type"`
	TargetID        uint         `gorm:"not null" json:"target_id"`
	RecordVersionID *uint        `json:"record_version_id,omitempty"`
	Priority        int          `json:"priority"`
	RiskLabels      JSONDocument `gorm:"type:jsonb" json:"risk_labels"`
	Status          string       `gorm:"size:24;not null" json:"status"`
	Decision        string       `gorm:"size:24" json:"decision"`
	DecisionReason  string       `gorm:"type:text" json:"decision_reason"`
	DueAt           *time.Time   `json:"due_at,omitempty"`
	DecidedAt       *time.Time   `json:"decided_at,omitempty"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

func (ModerationTask) TableName() string { return "moderation_tasks" }

type FeatureFlag struct {
	Key         string       `gorm:"primaryKey;size:64" json:"key"`
	Enabled     bool         `gorm:"not null" json:"enabled"`
	Value       JSONDocument `gorm:"type:jsonb" json:"value"`
	Description string       `gorm:"type:text" json:"description"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func (FeatureFlag) TableName() string { return "feature_flags" }
